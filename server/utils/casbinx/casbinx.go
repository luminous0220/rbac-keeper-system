package casbinx

import (
	"github.com/rs/zerolog/log"

	"rbackeeper/internal/constants"
	"rbackeeper/internal/dao"
)

// 全量重载 Casbin 策略（p + g）。
// 流程：
//  1. 收集所有 role-key 列表（给 super 加 *,* 通配策略）
//  2. 收集 role → menu perms 映射，构建 p 规则
//  3. 收集 user → role 映射，构建 g 规则
//  4. casbinx.Reload(policies, groupings) 原子切换
//
// 数据量按当前规模（10 用户 / 4 角色 / 12 菜单）< 50ms；后续可改造增量 AddPolicy/RemovePolicy。
func ReloadEnforcerPolicies() error {
	// 1. role 列表（用于 super 通配）
	roleKeys, err := dao.ListAllRoleKeys()
	if err != nil {
		log.Warn().Err(err).Msg("Casbin 加载：取 role 列表失败")
		return err
	}

	// 2. role → perms 映射（p 规则）
	rolePerms, err := dao.ListRoleMenuPerms()
	if err != nil {
		log.Warn().Err(err).Msg("Casbin 加载：取 role-perms 失败")
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
	userRoles, err := dao.ListUserRoles()
	if err != nil {
		log.Warn().Err(err).Msg("Casbin 加载：取 user-roles 失败")
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
