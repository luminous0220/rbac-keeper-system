package router

import (
	"github.com/gin-gonic/gin"

	"rbackeeper/internal/auth"
	"rbackeeper/middleware"
)

// registerAuthRoutes 认证模块
func registerAuthRoutes(base *gin.RouterGroup) {
	ac := &auth.Controller{}
	ag := base.Group("/auth")
	{
		// 获取图形验证码（captcha:<id>，一次性）
		ag.GET("/captcha", ac.Captcha)
		// 登录：校验验证码+用户名密码，签发 Access/Refresh 双令牌
		ag.POST("/login", ac.Login)
		// 用 Refresh Token 重签 Access Token
		ag.POST("/refresh", ac.Refresh)
		// 需认证
		authed := ag.Group("")
		authed.Use(middleware.JWTAuth())
		{
			// 登出：令牌进黑名单吊销
			authed.POST("/logout", ac.Logout)
			// 修改密码：强制该用户所有令牌失效
			authed.POST("/reset", ac.Reset)
			// 修改个人信息（昵称/手机/邮箱/性别/头像，仅本人）
			authed.POST("/profile", ac.Profile)
			// 当前用户权限（菜单权限列表 + 数据权限部门列表，按 dataScope 过滤部门）
			authed.GET("/permissions", middleware.DataScope(), ac.Permissions)
		}
	}
}
