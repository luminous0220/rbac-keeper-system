package router

import (
	"time"

	"github.com/gin-gonic/gin"

	"rbackeeper/internal/controller"
	"rbackeeper/middleware"
)

// registerDeptRoutes 部门模块（权限标识 system:dept:*）
func registerDeptRoutes(base *gin.RouterGroup) {
	dept := &controller.DeptController{}
	dg := base.Group("/dept")
	dg.Use(middleware.JWTAuth(), middleware.RateLimit(50, 100), middleware.CaptureBody())
	{
		// 部门树（按数据权限过滤）
		dg.GET("/tree", middleware.RBAC("system:dept:list", "*"), middleware.DataScope(), dept.Tree)
		// 新增部门（父部门须在数据权限内）
		dg.POST("/add", middleware.Idempotent(10*time.Second), middleware.RBAC("system:dept:add", "*"), middleware.DataScope(), dept.Add)
		// 编辑部门（目标部门须在数据权限内）
		dg.POST("/edit", middleware.Idempotent(10*time.Second), middleware.RBAC("system:dept:edit", "*"), middleware.DataScope(), dept.Edit)
		// 删除部门（级联清理 role_depts；目标部门须在数据权限内）
		dg.POST("/delete", middleware.Idempotent(10*time.Second), middleware.RBAC("system:dept:remove", "*"), middleware.DataScope(), dept.Delete)
	}
}
