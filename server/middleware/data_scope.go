package middleware

import (
	"github.com/gin-gonic/gin"

	"rbackeeper/constants"
	"rbackeeper/global"
	"rbackeeper/internal/dept"
	"rbackeeper/internal/user"
	"rbackeeper/model"
	"rbackeeper/utils"
)

// 计算并注入到 ctx（四个 key 恒写入，空集合写空/0）：
// 多角色取「最宽」权限（min(data_scope)）。
// super 强制 DsAll（不写过滤）。
func DataScope() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := utils.GetClaims(c)
		if claims == nil {
			c.Next()
			return
		}
		// 四个 key 恒写入（空值写空/0），service 端无需再区分「未设置」与「空集合」
		setScope := func(scope int8, depts []uint, deptID, uid uint) {
			c.Set(constants.CtxDataScope, scope)
			c.Set(constants.CtxDataScopeDepts, depts)
			c.Set(constants.CtxDataScopeDept, deptID)
			c.Set(constants.CtxDataScopeUserID, uid)
		}
		// super 直接放行
		// if claims.RoleKey == constants.SuperUsername {
		// 	setScope(constants.DsAll, nil, 0, 0)
		// 	c.Next()
		// 	return
		// }
		// 取用户全部角色
		// 这里必须全量查询，第一个参数传 nil
		u, err := user.GetUserInfoByUsernameScoped(nil, global.DB, claims.Username)
		if err != nil || u == nil {
			setScope(constants.DsSelf, nil, 0, claims.UserID)
			c.Next()
			return
		}
		scope, depts, deptID, uid := computeScope(u)
		setScope(scope, depts, deptID, uid)
		c.Next()
	}
}

// 取最宽权限（min(data_scope)）。
// 返回：(scope, customDepts, deptID, userID)
func computeScope(u *model.User) (int8, []uint, uint, uint) {
	if len(u.Roles) == 0 {
		return constants.DsSelf, nil, 0, u.ID
	}
	scope := constants.DsSelf
	for _, r := range u.Roles {
		if r.DataScope < scope {
			scope = r.DataScope
		}
	}
	switch scope {
	// 查看所有
	case constants.DsAll:
		return constants.DsAll, nil, 0, 0
	// 仅自己
	case constants.DsSelf:
		return constants.DsSelf, nil, u.DeptID, u.ID
	// 仅本部门
	case constants.DsDept:
		return constants.DsDept, nil, u.DeptID, 0
	// 本部门及子部门、自定义
	case constants.DsSubDept, constants.DsCustom:
		ids := map[uint]struct{}{}
		if scope == constants.DsSubDept && u.DeptID > 0 {
			// 计算"本部门及以下"需要全量部门树，不应用数据权限
			all, err := dept.GetAllDeptsScoped(nil, nil)
			if err == nil {
				for _, id := range utils.CollectSubtreeIDs(all,
					func(d model.Dept) uint { return d.ID },
					func(d model.Dept) uint { return d.ParentID },
					u.DeptID,
				) {
					ids[id] = struct{}{}
				}
			}
		} else {
			// DsCustom：聚合所有角色的 role_depts
			for _, r := range u.Roles {
				for _, d := range r.Depts {
					ids[d.ID] = struct{}{}
				}
			}
		}
		// 保证用户对自己的部门/自己可见且可管理（DsSubDept 已以本部门为根，此处为幂等）
		if u.DeptID > 0 {
			ids[u.DeptID] = struct{}{}
		}
		out := make([]uint, 0, len(ids))
		for k := range ids {
			out = append(out, k)
		}
		return scope, out, 0, 0
	// 默认
	default:
		return constants.DsSelf, nil, 0, u.ID
	}
}
