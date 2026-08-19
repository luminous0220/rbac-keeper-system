package router

import (
	"github.com/gin-gonic/gin"

	"rbackeeper/internal/controller"
	"rbackeeper/middleware"
)

// registerAuthRoutes 认证模块
func registerAuthRoutes(base *gin.RouterGroup) {
	auth := &controller.AuthController{}
	ag := base.Group("/auth")
	{
		// 获取图形验证码（captcha:<id>，一次性）
		ag.GET("/captcha", auth.Captcha)
		// 登录：校验验证码+用户名密码，签发 Access/Refresh 双令牌
		ag.POST("/login", auth.Login)
		// 用 Refresh Token 重签 Access Token
		ag.POST("/refresh", auth.Refresh)
		// 需认证
		authed := ag.Group("")
		authed.Use(middleware.JWTAuth())
		{
			// 登出：令牌进黑名单吊销
			authed.POST("/logout", auth.Logout)
			// 修改密码：强制该用户所有令牌失效
			authed.POST("/reset", auth.Reset)
			// 修改个人信息（昵称/手机/邮箱/性别/头像，仅本人）
			authed.POST("/profile", auth.Profile)
		}
	}
}
