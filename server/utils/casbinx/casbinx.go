package casbinx

import (
	zlog "github.com/rs/zerolog/log"

	"rbackeeper/constants"
	"rbackeeper/global"
	"rbackeeper/model"
)

// 全量重载 Casbin 策略（p + g）。
// 流程：
//  1. 收集所有 role-key 列表（给 super 加 *,* 通配策略）
//  2. 收集 role → menu perms 映射，构建 p 规则
//  3. 收集 user → role 映射，构建 g 规则
//  4. casbinx.Reload(policies, groupings) 原子切换
//
// 数据量按当前规模（10 用户 / 4 角色 / 12 菜单）< 50ms；后续可改造增量 AddPolicy/RemovePolicy。
// 为避免与业务模块（menu/role/user）形成导入环，Casbin 加载所需的 DB 查询就地实现。
func ReloadEnforcerPolicies() error {
	// 1. role 列表（用于 super 通配）
	roleKeys, err := listAllRoleKeys()
	if err != nil {
		zlog.Warn().Err(err).Msg("Casbin 加载：取 role 列表失败")
		return err
	}

	// 2. role → perms 映射（p 规则）
	rolePerms, err := listRoleMenuPerms()
	if err != nil {
		zlog.Warn().Err(err).Msg("Casbin 加载：取 role-perms 失败")
		return err
	}
	var policies [][]string
	for _, rk := range roleKeys {
		if rk == constants.SuperRoleKey {
			// super 全通
			policies = append(policies, []string{rk, "*", "*"})
			continue
		}
		for _, perm := range rolePerms[rk] {
			policies = append(policies, []string{rk, perm, "*"})
		}
	}

	// 3. user → roles 映射（g 规则）
	userRoles, err := listUserRoles()
	if err != nil {
		zlog.Warn().Err(err).Msg("Casbin 加载：取 user-roles 失败")
		return err
	}
	var groupings [][]string
	for u, rs := range userRoles {
		for _, rk := range rs {
			groupings = append(groupings, []string{u, rk})
		}
	}

	return Reload(policies, groupings)
}

// listAllRoleKeys 返回全部未删除角色的 role_key
func listAllRoleKeys() ([]string, error) {
	var keys []string
	err := global.DB.Model(&model.Role{}).
		Where("deleted_at IS NULL AND status = ?", constants.StatusEnabled).
		Pluck("role_key", &keys).Error
	return keys, err
}

// listRoleMenuPerms 返回 roleKey -> []perm，用于构建 Casbin p 规则
func listRoleMenuPerms() (map[string][]string, error) {
	type row struct {
		RoleKey string
		Perms   string
	}
	var rows []row
	err := global.DB.Table("role_menus rm").
		Select("r.role_key, m.perms").
		Joins("JOIN roles r ON r.id = rm.role_id").
		Joins("JOIN menus m ON m.id = rm.menu_id").
		Where("m.perms <> '' AND m.menu_type = 'F' AND r.deleted_at IS NULL AND m.deleted_at IS NULL AND r.status = ? AND m.status = ?",
			constants.StatusEnabled, constants.StatusEnabled).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make(map[string][]string, len(rows))
	for _, r := range rows {
		out[r.RoleKey] = append(out[r.RoleKey], r.Perms)
	}
	return out, nil
}

// listUserRoles 返回 username -> []roleKey，用于构建 Casbin g 规则
func listUserRoles() (map[string][]string, error) {
	type row struct {
		Username string
		RoleKey  string
	}
	var rows []row
	err := global.DB.Table("user_roles ur").
		Select("u.username, r.role_key").
		Joins("JOIN users u ON u.id = ur.user_id").
		Joins("JOIN roles r ON r.id = ur.role_id").
		Where("u.deleted_at IS NULL AND r.deleted_at IS NULL AND r.status = ?",
			constants.StatusEnabled).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make(map[string][]string, len(rows))
	for _, r := range rows {
		out[r.Username] = append(out[r.Username], r.RoleKey)
	}
	return out, nil
}
