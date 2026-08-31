package user

import (
	"github.com/gin-gonic/gin"

	"rbackeeper/common"
	"rbackeeper/internal/log"
	"rbackeeper/response"
	"rbackeeper/utils"
)

// Controller 用户控制器
type Controller struct{}

// @Summary 用户列表
// @Tags 用户
// @Security BearerAuth
// @Param query query UserListQuery true "-"
// @Success 200 {object} response.Result[UserListResp]
// @Router /user/list [get]
func (u *Controller) List(c *gin.Context) {
	var query UserListQuery
	if !response.BindQuery(c, &query) {
		return
	}
	resp, err := List(c, query)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	response.OK(c, resp)
}

// @Summary 新增用户
// @Tags 用户
// @Security BearerAuth
// @Param body body UserAddBody true "-"
// @Success 200 {object} response.Result[any]
// @Router /user/add [post]
func (u *Controller) Add(c *gin.Context) {
	var body UserAddBody
	if !response.BindJSON(c, &body) {
		return
	}
	if err := Add(c, body); err != nil {
		response.FailWithError(c, err)
		return
	}
	log.CreateWithMq(c, "用户管理")
	response.OK(c, nil)
}

// @Summary 管理员修改用户信息（密码、部门、角色、状态等）
// @Tags 用户
// @Security BearerAuth
// @Param username query string true "-"
// @Param body body UserEditBody true "-"
// @Success 200 {object} response.Result[any]
// @Router /user/edit [post]
func (u *Controller) Edit(c *gin.Context) {
	var query common.UsernameQuery
	if !response.BindQuery(c, &query) {
		return
	}

	var body UserEditBody
	if !response.BindJSON(c, &body) {
		return
	}
	if err := Edit(c, query.Username, body); err != nil {
		response.FailWithError(c, err)
		return
	}
	log.CreateWithMq(c, "用户管理")
	response.OK(c, nil)
}

// @Summary 用户信息
// @Tags 用户
// @Security BearerAuth
// @Success 200 {object} response.Result[UserItem]
// @Router /user/info [get]
func (u *Controller) Info(c *gin.Context) {
	resp, err := Info(c.Request.Context(), utils.GetClaims(c).Username)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	response.OK(c, resp)
}
