package controller

import (
	"github.com/gin-gonic/gin"

	"rbackeeper/internal/dto"
	"rbackeeper/internal/service"
	"rbackeeper/response"
	"rbackeeper/utils"
)

// UserController 用户控制器
type UserController struct{}

// @Summary 用户列表
// @Tags 用户
// @Security BearerAuth
// @Param query query dto.UserListQuery true "-"
// @Success 200 {object} response.Result[dto.UserListResp]
// @Router /user/list [get]
func (u *UserController) List(c *gin.Context) {
	var query dto.UserListQuery
	if !response.BindQuery(c, &query) {
		return
	}
	resp, err := service.UserList(c, query)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	response.OK(c, resp)
}

// @Summary 新增用户
// @Tags 用户
// @Security BearerAuth
// @Param body body dto.UserAddBody true "-"
// @Success 200 {object} response.Result[any]
// @Router /user/add [post]
func (u *UserController) Add(c *gin.Context) {
	var body dto.UserAddBody
	if !response.BindJSON(c, &body) {
		return
	}
	if err := service.UserAdd(c, body); err != nil {
		response.FailWithError(c, err)
		return
	}
	service.CreateLogWithMq(c, "用户管理")
	response.OK(c, nil)
}

// @Summary 管理员修改用户信息（密码、部门、角色、状态等）
// @Tags 用户
// @Security BearerAuth
// @Param username query string true "-"
// @Param body body dto.UserEditBody true "-"
// @Success 200 {object} response.Result[any]
// @Router /user/edit [post]
func (u *UserController) Edit(c *gin.Context) {
	var query dto.UsernameQuery
	if !response.BindQuery(c, &query) {
		return
	}

	var body dto.UserEditBody
	if !response.BindJSON(c, &body) {
		return
	}
	if err := service.UserEdit(c, query.Username, body); err != nil {
		response.FailWithError(c, err)
		return
	}
	service.CreateLogWithMq(c, "用户管理")
	response.OK(c, nil)
}

// @Summary 用户信息
// @Tags 用户
// @Security BearerAuth
// @Success 200 {object} response.Result[dto.UserItem]
// @Router /user/info [get]
func (u *UserController) Info(c *gin.Context) {
	resp, err := service.UserInfo(c.Request.Context(), utils.GetClaims(c).Username)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	response.OK(c, resp)
}
