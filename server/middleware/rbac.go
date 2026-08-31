package middleware

import (
	"github.com/gin-gonic/gin"

	logmod "rbackeeper/internal/log"
	"rbackeeper/response"
	"rbackeeper/utils"
	"rbackeeper/utils/casbinx"
)

// 权限校验中间件。
// 用法：router.GET("/list", middleware.RBAC("system:user:list", "*"), handler)
// 行为：未登录 → 401；Enforce 失败 → 403（异步记录越权日志）；enforcer 未初始化 → 拒绝访问。
func RBAC(obj, act string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := utils.GetClaims(c)
		if claims == nil {
			response.Fail(c, response.ErrUnauthorized, "未登录")
			c.Abort()
			return
		}
		e := casbinx.GetCasbin()
		if e == nil {
			response.Fail(c, response.ErrServer, "权限系统未初始化")
			c.Abort()
			return
		}
		ok, err := e.Enforce(claims.Username, obj, act)
		if err != nil {
			response.Fail(c, response.ErrServer, "权限校验失败")
			c.Abort()
			return
		}
		if !ok {
			// 异步记录越权日志（不阻塞业务响应）
			logmod.CreateWithMq(c, "权限拒绝")
			response.Fail(c, response.ErrForbidden, "无权限访问")
			c.Abort()
			return
		}
		c.Next()
	}
}
