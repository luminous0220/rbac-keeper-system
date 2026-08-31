package menu

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"rbackeeper/global"
	"rbackeeper/model"
	"rbackeeper/response"
	"rbackeeper/utils"
	"rbackeeper/utils/casbinx"
)

// permsRe 权限标识三段式格式：模块:实体:动作，各段以小写字母开头，含小写字母/数字/下划线
var permsRe = regexp.MustCompile(`^[a-z][a-z0-9_]*:[a-z][a-z0-9_]*:[a-z][a-z0-9_]*$`)

// Tree 菜单树
func Tree(ctx context.Context, query MenuTreeQuery) ([]model.Menu, error) {
	menus, err := GetAllMenus(query.Status)
	if err != nil {
		return nil, err
	}
	if query.MenuName != "" {
		menus = utils.FilterTreeWithAncestors(menus,
			func(m model.Menu) uint { return m.ID },
			func(m model.Menu) uint { return m.ParentID },
			func(m model.Menu) bool { return strings.Contains(m.MenuName, query.MenuName) },
		)
	}
	tree := utils.Build(menus,
		func(m model.Menu) uint { return m.ID },
		func(m model.Menu) uint { return m.ParentID },
		func(p *model.Menu, kids []*model.Menu) { p.Children = kids },
	)
	return tree, nil
}

// Add 新增菜单
func Add(ctx context.Context, body MenuAddBody) error {
	// 父菜单存在性（0=根合法）
	if body.ParentID != 0 {
		if ok, err := ExistsMenuID(global.DB, body.ParentID); err != nil {
			return err
		} else if !ok {
			return response.NewBizError(response.ErrInvalidParam, "父菜单不存在")
		}
	}

	// perms 校验：按钮(F)必填，目录(M)/菜单(C)可选；非空时校验格式 + 唯一性
	if err := validateMenuPerms(body.MenuType, body.Perms, 0); err != nil {
		return err
	}

	menu := &model.Menu{
		ParentID:  body.ParentID,
		MenuName:  body.MenuName,
		MenuType:  body.MenuType,
		OrderNum:  body.OrderNum,
		Path:      body.Path,
		Component: body.Component,
		Query:     body.Query,
		Status:    body.Status,
		Perms:     body.Perms,
		Icon:      body.Icon,
		Visible:   body.Visible,
		Remark:    body.Remark,
	}
	if err := global.DB.Transaction(func(tx *gorm.DB) error {
		return CreateMenu(tx, menu)
	}); err != nil {
		return err
	}

	// 写后全量重载 Casbin 策略（菜单 perms 是 p 规则的来源之一）
	if err := casbinx.ReloadEnforcerPolicies(); err != nil {
		log.Warn().Err(err).Msg("Casbin 策略重载失败")
	}
	return nil
}

// Edit 修改菜单
func Edit(ctx context.Context, id uint, body MenuEditBody) error {
	menu, err := GetMenuByID(global.DB, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.NewBizError(response.ErrNotFound, "菜单不存在")
		}
		return err
	}

	updates := map[string]any{}
	if body.MenuName != nil {
		updates["menu_name"] = *body.MenuName
	}
	if body.MenuType != nil {
		updates["menu_type"] = *body.MenuType
	}
	if body.OrderNum != nil {
		updates["order_num"] = *body.OrderNum
	}
	if body.Path != nil {
		updates["path"] = *body.Path
	}
	if body.Component != nil {
		updates["component"] = *body.Component
	}
	if body.Query != nil {
		updates["query"] = *body.Query
	}
	if body.Status != nil {
		updates["status"] = *body.Status
	}
	if body.Icon != nil {
		updates["icon"] = *body.Icon
	}
	if body.Remark != nil {
		updates["remark"] = *body.Remark
	}
	if body.Visible != nil {
		updates["visible"] = *body.Visible
	}

	// 父菜单变更：存在性 + 环检测（自身不能成为自身后代的子节点）
	if body.ParentID != nil {
		newParent := *body.ParentID
		if newParent != 0 {
			if ok, err := ExistsMenuID(global.DB, newParent); err != nil {
				return err
			} else if !ok {
				return response.NewBizError(response.ErrInvalidParam, "父菜单不存在")
			}
			// 环检测需全量菜单 id->parentID 映射
			all, err := GetAllMenus(nil)
			if err != nil {
				return err
			}
			parentMap := make(map[uint]uint, len(all))
			for i := range all {
				parentMap[all[i].ID] = all[i].ParentID
			}
			if utils.WouldCreateCycle(parentMap, menu.ID, newParent) {
				return response.NewBizError(response.ErrInvalidParam, "不能将菜单挂载到自身或其子菜单下")
			}
		}
		updates["parent_id"] = newParent
	}

	// perms 变更：按新 menuType（若有）校验
	if body.Perms != nil {
		menuType := menu.MenuType
		if body.MenuType != nil {
			menuType = *body.MenuType
		}
		if err := validateMenuPerms(menuType, *body.Perms, menu.ID); err != nil {
			return err
		}
		updates["perms"] = *body.Perms
	}

	if err := global.DB.Transaction(func(tx *gorm.DB) error {
		return UpdateMenuByID(tx, menu.ID, updates)
	}); err != nil {
		return err
	}

	// perms/visible 等可能影响 p 规则，写后重载
	if err := casbinx.ReloadEnforcerPolicies(); err != nil {
		log.Warn().Err(err).Msg("Casbin 策略重载失败")
	}
	return nil
}

// Delete 删除菜单
func Delete(ctx context.Context, id uint) error {
	cnt, err := CountMenusByParentID(global.DB, id)
	if err != nil {
		return err
	}
	if cnt > 0 {
		return response.NewBizError(response.ErrForbidden, "存在子菜单或按钮，无法删除")
	}
	if err := global.DB.Transaction(func(tx *gorm.DB) error {
		return DeleteMenuByID(tx, id)
	}); err != nil {
		return err
	}

	if err := casbinx.ReloadEnforcerPolicies(); err != nil {
		log.Warn().Err(err).Msg("Casbin 策略重载失败")
	}
	return nil
}

// 校验权限标识：按钮(F)必填；非空时校验三段式格式与唯一性
func validateMenuPerms(menuType, perms string, excludeID uint) error {
	if perms == "" {
		if menuType == "F" {
			return response.NewBizError(response.ErrInvalidParam, "按钮类型必须填写权限标识")
		}
		return nil
	}
	if !permsRe.MatchString(perms) {
		return response.NewBizError(response.ErrInvalidParam, "权限标识需为 模块:实体:动作 三段式（小写字母/数字/下划线）")
	}
	if exists, err := ExistsMenuPerms(global.DB, perms, excludeID); err != nil {
		return err
	} else if exists {
		return response.NewBizError(response.ErrInvalidParam, "权限标识已存在")
	}
	return nil
}
