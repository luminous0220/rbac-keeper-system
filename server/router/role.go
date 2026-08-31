package router

import (
	"time"

	"github.com/gin-gonic/gin"

	"rbackeeper/internal/role"
	"rbackeeper/middleware"
)

// registerRoleRoutes 角色模块（权限标识 system:role:*）
func registerRoleRoutes(base *gin.RouterGroup) {
	rc := &role.Controller{}
	rg := base.Group("/role")
	rg.Use(middleware.JWTAuth(), middleware.RateLimit(50, 100), middleware.CaptureBody())
	{
		// 角色分页列表
		rg.GET("/list", middleware.RBAC("system:role:list", "*"), rc.List)
		// 角色详情（按 roleKey；仅需登录，引导期前端据此取数据权限/菜单/部门）
		rg.GET("/info", rc.Info)
		// 新增角色（含菜单/部门授权）
		rg.POST("/add", middleware.Idempotent(10*time.Second), middleware.RBAC("system:role:add", "*"), rc.Add)
		// 编辑角色（含菜单/部门授权）
		rg.POST("/edit", middleware.Idempotent(10*time.Second), middleware.RBAC("system:role:edit", "*"), rc.Edit)
		// 删除角色（按 roleKey 定位）
		rg.POST("/delete", middleware.Idempotent(10*time.Second), middleware.RBAC("system:role:remove", "*"), rc.Delete)
	}
}
