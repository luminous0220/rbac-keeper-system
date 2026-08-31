package auth

import (
	"github.com/gin-gonic/gin"

	"rbackeeper/internal/log"
	"rbackeeper/response"
	"rbackeeper/utils"
)

// Controller 认证控制器
type Controller struct{}

// @Summary 获取验证码
// @Tags 认证
// @Success 200 {object} response.Result[CaptchaResp]
// @Router /auth/captcha [get]
func (a *Controller) Captcha(ctx *gin.Context) {
	id, b64s, err := GenerateCaptcha()
	if err != nil {
		response.FailWithError(ctx, err)
		return
	}
	response.OK(ctx, CaptchaResp{CaptchaID: id, Img: b64s})
}

// @Summary 用户登录
// @Tags 认证
// @Param body body LoginBody true "-"
// @Success 200 {object} response.Result[LoginResp]
// @Router /auth/login [post]
func (a *Controller) Login(ctx *gin.Context) {
	var body LoginBody
	if !response.BindJSON(ctx, &body) {
		return
	}
	resp, err := Login(ctx.Request.Context(), body, ctx.ClientIP(), ctx.GetHeader("User-Agent"))
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
func (a *Controller) Logout(ctx *gin.Context) {
	claims := utils.GetClaims(ctx)
	if err := Logout(ctx.Request.Context(), claims); err != nil {
		response.FailWithError(ctx, err)
		return
	}
	// 登出动作并入操作日志（本期无独立登出日志表）
	log.CreateWithMq(ctx, "退出登录")
	response.OK(ctx, nil)
}

// @Summary 刷新令牌
// @Tags 认证
// @Param body body RefreshBody true "-"
// @Success 200 {object} response.Result[RefreshResp]
// @Router /auth/refresh [post]
func (a *Controller) Refresh(ctx *gin.Context) {
	var body RefreshBody
	if !response.BindJSON(ctx, &body) {
		return
	}
	resp, err := Refresh(ctx.Request.Context(), body)
	if err != nil {
		response.FailWithError(ctx, err)
		return
	}
	response.OK(ctx, resp)
}

// @Summary 密码重置
// @Tags 认证
// @Security BearerAuth
// @Param body body ResetPasswordBody true "-"
// @Success 200 {object} response.Result[any]
// @Router /auth/reset [post]
func (a *Controller) Reset(ctx *gin.Context) {
	var body ResetPasswordBody
	if !response.BindJSON(ctx, &body) {
		return
	}
	if err := ResetPassword(ctx.Request.Context(), utils.GetClaims(ctx), body); err != nil {
		response.FailWithError(ctx, err)
		return
	}
	response.OK(ctx, nil)
}

// @Summary 修改个人信息（昵称/手机/邮箱/性别/头像，仅本人）
// @Tags 认证
// @Security BearerAuth
// @Param body body ProfileBody true "-"
// @Success 200 {object} response.Result[any]
// @Router /auth/profile [post]
func (a *Controller) Profile(ctx *gin.Context) {
	var body ProfileBody
	if !response.BindJSON(ctx, &body) {
		return
	}
	resp, err := UpdateProfile(ctx.Request.Context(), utils.GetClaims(ctx), body)
	if err != nil {
		response.FailWithError(ctx, err)
		return
	}
	response.OK(ctx, resp)
}

// @Summary 当前用户权限（菜单权限列表 + 数据权限部门列表）
// @Tags 认证
// @Security BearerAuth
// @Success 200 {object} response.Result[PermissionsResp]
// @Router /auth/permissions [get]
func (a *Controller) Permissions(ctx *gin.Context) {
	resp, err := Permissions(ctx)
	if err != nil {
		response.FailWithError(ctx, err)
		return
	}
	response.OK(ctx, resp)
}
