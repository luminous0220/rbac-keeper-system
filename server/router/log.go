package router

import (
	"github.com/gin-gonic/gin"

	"rbackeeper/internal/controller"
	"rbackeeper/middleware"
)

// registerLogRoutes 日志模块（权限标识 monitor:log:*）
func registerLogRoutes(base *gin.RouterGroup) {
	logc := &controller.LogController{}
	lg := base.Group("/log")
	lg.Use(middleware.JWTAuth(), middleware.RateLimit(50, 100))
	{
		// 登录日志分页列表
		lg.GET("/login", middleware.RBAC("monitor:log:login", "*"), logc.Login)
		// 操作日志分页列表
		lg.GET("/operation", middleware.RBAC("monitor:log:operation", "*"), logc.Operation)
		// 操作日志导出
		lg.GET("/export", middleware.RBAC("monitor:log:export", "*"), logc.Export)
	}
}
