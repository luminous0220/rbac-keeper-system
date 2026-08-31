package role

import (
	"context"
	"errors"

	zlog "github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"rbackeeper/constants"
	"rbackeeper/global"
	"rbackeeper/internal/dept"
	"rbackeeper/internal/menu"
	"rbackeeper/model"
	"rbackeeper/response"
	"rbackeeper/utils"
	"rbackeeper/utils/casbinx"
)

// List 角色列表
func List(ctx context.Context, query RoleListQuery) ([]model.Role, error) {
	roles, err := GetList(RoleListQuery{
		RoleName: query.RoleName,
		RoleKey:  query.RoleKey,
		Status:   query.Status,
	})
	if err != nil {
		return nil, err
	}
	return roles, nil
}

// Detail 角色详情
func Detail(ctx context.Context, roleKey string) (*model.Role, error) {
	r, err := GetByKey(roleKey)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, response.NewBizError(response.ErrNotFound, "角色不存在")
		}
		return nil, err
	}
	return r, nil
}

// Add 新增角色
func Add(ctx context.Context, body RoleAddBody) error {
	if !utils.IsAlphaNumeric(body.RoleKey) {
		return response.NewBizError(response.ErrInvalidParam, "RoleKey 只能包含英文字母和数字，且不能是纯数字")
	}

	exists, err := ExistsKey(global.DB, body.RoleKey)
	// 校验 role_key 是否已存在
	if err != nil {
		return err
	}
	if exists {
		return response.NewBizError(response.ErrInvalidParam, "RoleKey 已存在")
	}

	var menus []model.Menu
	if len(body.MenuIDs) > 0 {
		// 菜单授权校验（去重后比对命中数，防悬挂引用）
		menuIDs := uniqueUint(body.MenuIDs)
		menus, err = menu.GetByIDs(global.DB, menuIDs)
		if err != nil {
			return err
		}
		if len(menus) != len(menuIDs) {
			return response.NewBizError(response.ErrInvalidParam, "包含不存在的菜单ID")
		}
	}

	// 5. 自定义数据权限部门校验（仅 data_scope=2 时存储，其余范围忽略 DeptIDs）
	// required_if 对空数组不生效（validator 对 slice 只查 nil），这里显式强制：dataScope=2 必须选部门
	if body.DataScope == constants.DsCustom && len(body.DeptIDs) == 0 {
		return response.NewBizError(response.ErrInvalidParam, "自定义数据权限必须选择部门")
	}
	var depts []model.Dept
	if body.DataScope == constants.DsCustom && len(body.DeptIDs) > 0 {
		deptIDs := uniqueUint(body.DeptIDs)
		depts, err = dept.GetByIDs(global.DB, deptIDs)
		if err != nil {
			return err
		}
		if len(depts) != len(deptIDs) {
			return response.NewBizError(response.ErrInvalidParam, "包含不存在的部门ID")
		}
	}

	// 6. 组装角色
	r := &model.Role{
		RoleName:  body.RoleName,
		RoleKey:   body.RoleKey,
		RoleSort:  body.RoleSort,
		DataScope: body.DataScope,
		Status:    1,
		Remark:    body.Remark,
	}

	// 7. 事务：建角色 + 关联菜单 + 关联部门
	if err := global.DB.Transaction(func(tx *gorm.DB) error {
		if err := Create(tx, r); err != nil {
			return err
		}
		if err := ReplaceMenus(tx, r.ID, menus); err != nil {
			return err
		}
		if err := ReplaceDepts(tx, r.ID, depts); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	// 写后全量重载 Casbin 策略（角色-菜单变更影响 p 规则）
	if err := casbinx.ReloadEnforcerPolicies(); err != nil {
		zlog.Warn().Err(err).Str("role_key", body.RoleKey).Msg("Casbin 策略重载失败")
	}
	return nil
}

// Edit 修改角色
// role_key 不可变更；data_scope 改为非自定义时（2）自动清空部门关联。
func Edit(ctx context.Context, roleKey string, body RoleEditBody) error {
	r, err := GetByKey(roleKey)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.NewBizError(response.ErrNotFound, "角色不存在")
		}
		return err
	}

	// 组装更新字段
	updates := map[string]any{}

	if body.DataScope != nil {
		updates["data_scope"] = *body.DataScope
	}
	if body.RoleName != nil {
		updates["role_name"] = *body.RoleName
	}

	if body.RoleSort != nil {
		updates["role_sort"] = *body.RoleSort
	}

	if body.Status != nil {
		updates["status"] = *body.Status
	}
	if body.Remark != nil {
		updates["remark"] = *body.Remark
	}

	// 菜单授权替换
	var menus []model.Menu
	manageMenus := body.MenuIDs != nil
	if manageMenus {
		menuIDs := uniqueUint(body.MenuIDs)
		menus, err = menu.GetByIDs(global.DB, menuIDs)
		if err != nil {
			return err
		}
		if len(menus) != len(menuIDs) {
			return response.NewBizError(response.ErrInvalidParam, "包含不存在的菜单ID")
		}
	}

	// 部门授权替换：仅 data_scope=2 时存储
	// - data_scope 不为 2 清空关联
	// - data_scope 为 2 且未传 DeptIDs：保持原有关联
	// - data_scope 为 2 且传空 DeptIDs：清空关联
	// - data_scope 为 2 且 len(DeptIDs) >0：替换关联
	var depts []model.Dept
	manageDepts := false

	if body.DataScope != nil {
		// 非自定义范围：清空部门关联
		if *body.DataScope != constants.DsCustom {
			manageDepts = true
		} else {
			// 自定义范围：必须选择部门（防止 required_if 空数组逃逸），再去重校验后替换
			if len(body.DeptIDs) == 0 {
				return response.NewBizError(response.ErrInvalidParam, "自定义数据权限必须选择部门")
			}
			deptIDs := uniqueUint(body.DeptIDs)
			depts, err = dept.GetByIDs(global.DB, deptIDs)
			if err != nil {
				return err
			}
			if len(depts) != len(deptIDs) {
				return response.NewBizError(response.ErrInvalidParam, "包含不存在的部门ID")
			}
			manageDepts = true
		}
	}

	if err := global.DB.Transaction(func(tx *gorm.DB) error {
		// 始终记录修改人（与用户模块一致，关联变更也计入审计）
		if err := UpdateByID(tx, r.ID, updates); err != nil {
			return err
		}
		if manageMenus {
			if err := ReplaceMenus(tx, r.ID, menus); err != nil {
				return err
			}
		}
		if manageDepts {
			if err := ReplaceDepts(tx, r.ID, depts); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}

	// 写后全量重载 Casbin 策略（角色-菜单 / 角色-部门变更）
	if err := casbinx.ReloadEnforcerPolicies(); err != nil {
		zlog.Warn().Err(err).Str("role_key", roleKey).Msg("Casbin 策略重载失败")
	}
	return nil
}

// Delete 删除角色
// 禁止删除超级管理员角色，防止权限体系崩塌。
func Delete(ctx context.Context, roleKey string) error {
	if roleKey == constants.SuperRoleKey {
		return response.NewBizError(response.ErrForbidden, "超级管理员角色不允许删除")
	}
	r, err := GetByKey(roleKey)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.NewBizError(response.ErrNotFound, "角色不存在")
		}
		return err
	}
	// 关联用户校验：有用户持有该角色则拦截，防权限真空
	cnt, err := CountUsersByID(global.DB, r.ID)
	if err != nil {
		return err
	}
	if cnt > 0 {
		return response.NewBizError(response.ErrForbidden, "该角色已分配给用户，无法删除")
	}
	if err := global.DB.Transaction(func(tx *gorm.DB) error {
		return DeleteByID(tx, r.ID)
	}); err != nil {
		return err
	}

	// 写后全量重载 Casbin 策略
	if err := casbinx.ReloadEnforcerPolicies(); err != nil {
		zlog.Warn().Err(err).Str("role_key", roleKey).Msg("Casbin 策略重载失败")
	}
	return nil
}

// 去重 uint 切片，保持首次出现顺序（授权 ID 去重，避免重复计数导致存在性校验误判）
func uniqueUint(ids []uint) []uint {
	if len(ids) == 0 {
		return ids
	}
	seen := make(map[uint]struct{}, len(ids))
	out := make([]uint, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}
