package controller

import (
	"github.com/gin-gonic/gin"

	"rbackeeper/internal/dto"
	"rbackeeper/internal/service"
	"rbackeeper/response"
	"rbackeeper/utils"
)

// AuthController 认证控制器
type AuthController struct{}

// @Summary 获取验证码
// @Tags 认证
// @Success 200 {object} response.Result[dto.CaptchaResp]
// @Router /auth/captcha [get]
func (a *AuthController) Captcha(ctx *gin.Context) {

	id, b64s, err := service.GenerateCaptcha()
	if err != nil {
		response.FailWithError(ctx, err)
		return
	}
	response.OK(ctx, dto.CaptchaResp{CaptchaID: id, Img: b64s})
}

// @Summary 用户登录
// @Tags 认证
// @Param body body dto.LoginBody true "-"
// @Success 200 {object} response.Result[dto.LoginResp]
// @Router /auth/login [post]
func (a *AuthController) Login(ctx *gin.Context) {
	var body dto.LoginBody
	if !response.BindJSON(ctx, &body) {
		return
	}
	resp, err := service.Login(ctx.Request.Context(), body, ctx.ClientIP(), ctx.GetHeader("User-Agent"))
	if err != nil {
		response.FailWithError(ctx, err)
		return
	}
	response.OK(ctx, resp)
}

// @Summary 用户登出
// @Tags 认证
// @Security BearerAuth
// @Success 200 {object} response.Result[any]
// @Router /auth/logout [post]
func (a *AuthController) Logout(ctx *gin.Context) {
	claims := utils.GetClaims(ctx)
	if err := service.Logout(ctx.Request.Context(), claims); err != nil {
		response.FailWithError(ctx, err)
		return
	}
	// 登出动作并入操作日志（本期无独立登出日志表）
	service.CreateLogWithMq(ctx, "退出登录")
	response.OK(ctx, nil)
}

// @Summary 刷新令牌
// @Tags 认证
// @Param body body dto.RefreshBody true "-"
// @Success 200 {object} response.Result[dto.RefreshResp]
// @Router /auth/refresh [post]
func (a *AuthController) Refresh(ctx *gin.Context) {
	var body dto.RefreshBody
	if !response.BindJSON(ctx, &body) {
		return
	}
	resp, err := service.Refresh(ctx.Request.Context(), body)
	if err != nil {
		response.FailWithError(ctx, err)
		return
	}
	response.OK(ctx, resp)
}

// @Summary 密码重置
// @Tags 认证
// @Security BearerAuth
// @Param body body dto.ResetPasswordBody true "-"
// @Success 200 {object} response.Result[any]
// @Router /auth/reset [post]
func (a *AuthController) Reset(ctx *gin.Context) {
	var body dto.ResetPasswordBody
	if !response.BindJSON(ctx, &body) {
		return
	}
	if err := service.ResetPassword(ctx.Request.Context(), utils.GetClaims(ctx), body); err != nil {
		response.FailWithError(ctx, err)
		return
	}
	response.OK(ctx, nil)
}

// @Summary 修改个人信息（昵称/手机/邮箱/性别/头像，仅本人）
// @Tags 认证
// @Security BearerAuth
// @Param body body dto.ProfileBody true "-"
// @Success 200 {object} response.Result[any]
// @Router /auth/profile [post]
func (a *AuthController) Profile(ctx *gin.Context) {
	var body dto.ProfileBody
	if !response.BindJSON(ctx, &body) {
		return
	}
	resp, err := service.UpdateProfile(ctx.Request.Context(), utils.GetClaims(ctx), body)
	if err != nil {
		response.FailWithError(ctx, err)
		return
	}
	response.OK(ctx, resp)
}
