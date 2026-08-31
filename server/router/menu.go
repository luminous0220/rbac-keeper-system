package router

import (
	"time"

	"github.com/gin-gonic/gin"

	"rbackeeper/internal/menu"
	"rbackeeper/middleware"
)

// registerMenuRoutes 菜单模块（权限标识 system:menu:*）
func registerMenuRoutes(base *gin.RouterGroup) {
	mc := &menu.Controller{}
	mg := base.Group("/menu")
	mg.Use(middleware.JWTAuth(), middleware.RateLimit(50, 100), middleware.CaptureBody())
	{
		// 菜单树（目录/菜单/功能三层结构）
		mg.GET("/tree", middleware.RBAC("system:menu:list", "*"), mc.Tree)
		// 新增菜单
		mg.POST("/add", middleware.Idempotent(10*time.Second), middleware.RBAC("system:menu:add", "*"), mc.Add)
		// 编辑菜单（含 ParentID 循环引用校验）
		mg.POST("/edit", middleware.Idempotent(10*time.Second), middleware.RBAC("system:menu:edit", "*"), mc.Edit)
		// 删除菜单（级联清理 role_menus）
		mg.POST("/delete", middleware.Idempotent(10*time.Second), middleware.RBAC("system:menu:remove", "*"), mc.Delete)
	}
}
