package router

import (
	"time"

	"github.com/gin-gonic/gin"

	"rbackeeper/internal/controller"
	"rbackeeper/middleware"
)

// registerUserRoutes 用户模块（权限标识 system:user:*）
func registerUserRoutes(base *gin.RouterGroup) {
	user := &controller.UserController{}
	ug := base.Group("/user")
	ug.Use(middleware.JWTAuth(), middleware.RateLimit(50, 100), middleware.CaptureBody())
	{
		// 用户分页列表（按数据权限过滤）
		ug.GET("/list", middleware.RBAC("system:user:list", "*"), middleware.DataScope(), user.List)
		// 用户详情（本人信息 + 角色并集菜单；仅需登录，属自作用域，不挂 RBAC 权限）
		ug.GET("/info", user.Info)
		// 新增用户（含角色分配；分配部门须在数据权限内）
		ug.POST("/add", middleware.Idempotent(10*time.Second), middleware.RBAC("system:user:add", "*"), middleware.DataScope(), user.Add)
		// 编辑用户（含角色分配；nil 字段不修改；分配部门须在数据权限内）
		ug.POST("/edit", middleware.Idempotent(10*time.Second), middleware.RBAC("system:user:edit", "*"), middleware.DataScope(), user.Edit)
	}
}
