package router

import (
	"github.com/gin-gonic/gin"

	logmod "rbackeeper/internal/log"
	"rbackeeper/middleware"
)

// registerLogRoutes 日志模块（权限标识 monitor:log:*）
func registerLogRoutes(base *gin.RouterGroup) {
	lc := &logmod.Controller{}
	lg := base.Group("/log")
	lg.Use(middleware.JWTAuth(), middleware.RateLimit(50, 100))
	{
		// 登录日志分页列表
		lg.GET("/login", middleware.RBAC("monitor:log:login", "*"), lc.Login)
		// 操作日志分页列表
		lg.GET("/operation", middleware.RBAC("monitor:log:operation", "*"), lc.Operation)
		// 操作日志导出
		lg.GET("/export", middleware.RBAC("monitor:log:export", "*"), lc.Export)
	}
}
