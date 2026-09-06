package router

import (
	"time"

	"github.com/gin-gonic/gin"

	logmod "rbackeeper/internal/log"
	"rbackeeper/middleware"
)

// registerLogRoutes 日志模块（权限标识 monitor:log:*）
// 导出为 GET（幂等、携带列表同款筛选），清空为变更操作走 POST + 幂等键。
func registerLogRoutes(base *gin.RouterGroup) {
	lc := &logmod.Controller{}
	lg := base.Group("/log")
	lg.Use(middleware.JWTAuth(), middleware.RateLimit(50, 100))
	{
		// 登录日志分页列表
		lg.GET("/login", middleware.RBAC("monitor:log:login", "*"), lc.Login)
		// 登录日志导出
		lg.GET("/login/export", middleware.RBAC("monitor:log:login:export", "*"), lc.LoginExport)
		// 清空登录日志
		lg.POST("/login/clear", middleware.Idempotent(10*time.Second), middleware.RBAC("monitor:log:login:clear", "*"), lc.LoginClear)

		// 操作日志分页列表
		lg.GET("/operation", middleware.RBAC("monitor:log:operation", "*"), lc.Operation)
		// 操作日志导出
		lg.GET("/operation/export", middleware.RBAC("monitor:log:operation:export", "*"), lc.OperationExport)
		// 清空操作日志
		lg.POST("/operation/clear", middleware.Idempotent(10*time.Second), middleware.RBAC("monitor:log:operation:clear", "*"), lc.OperationClear)
	}
}
