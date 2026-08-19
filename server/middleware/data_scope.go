package middleware

import (
	"github.com/gin-gonic/gin"

	"rbackeeper/global"
	"rbackeeper/internal/constants"
	"rbackeeper/internal/dao"
	"rbackeeper/internal/model"
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
		user, err := dao.GetUserInfoByUsernameScoped(nil, global.DB, claims.Username)
		if err != nil || user == nil {
			setScope(constants.DsSelf, nil, 0, claims.UserID)
			c.Next()
			return
		}
		scope, depts, deptID, uid := computeScope(user)
		setScope(scope, depts, deptID, uid)
		c.Next()
	}
}

// 取最宽权限（min(data_scope)）。
// 返回：(scope, customDepts, deptID, userID)
func computeScope(user *model.User) (int8, []uint, uint, uint) {
	if len(user.Roles) == 0 {
		return constants.DsSelf, nil, 0, user.ID
	}
	scope := constants.DsSelf
	for _, r := range user.Roles {
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
		return constants.DsSelf, nil, user.DeptID, user.ID
	// 仅本部门
	case constants.DsDept:
		return constants.DsDept, nil, user.DeptID, 0
	// 本部门及子部门、自定义
	case constants.DsSubDept, constants.DsCustom:
		ids := map[uint]struct{}{}
		if scope == constants.DsSubDept && user.DeptID > 0 {
			// 计算"本部门及以下"需要全量部门树，不应用数据权限
			all, err := dao.GetAllDeptsScoped(nil, nil)
			if err == nil {
				// 找本部门及所有子部门
				childrenMap := map[uint][]uint{}
				for _, d := range all {
					childrenMap[d.ParentID] = append(childrenMap[d.ParentID], d.ID)
				}
				queue := []uint{user.DeptID}
				for len(queue) > 0 {
					cur := queue[0]
					queue = queue[1:]
					ids[cur] = struct{}{}
					queue = append(queue, childrenMap[cur]...)
				}
			}
		} else {
			// DsCustom：聚合所有角色的 role_depts
			for _, r := range user.Roles {
				for _, d := range r.Depts {
					ids[d.ID] = struct{}{}
				}
			}
		}
		// 保证用户对自己的部门/自己可见且可管理（DsSubDept 已以本部门为根，此处为幂等）
		if user.DeptID > 0 {
			ids[user.DeptID] = struct{}{}
		}
		out := make([]uint, 0, len(ids))
		for k := range ids {
			out = append(out, k)
		}
		return scope, out, 0, 0
	// 默认
	default:
		return constants.DsSelf, nil, 0, user.ID
	}
}
