package role

import (
	"github.com/gin-gonic/gin"

	"rbackeeper/common"
	"rbackeeper/internal/log"
	"rbackeeper/response"
)

// Controller 角色控制器
type Controller struct{}

// @Summary 角色列表
// @Tags 角色
// @Security BearerAuth
// @Param query query RoleListQuery true "-"
// @Success 200 {object} response.Result[[]RoleItem]
// @Router /role/list [get]
func (r *Controller) List(c *gin.Context) {
	var query RoleListQuery
	if !response.BindQuery(c, &query) {
		return
	}
	resp, err := List(c.Request.Context(), query)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	response.OK(c, resp)
}

// @Summary 新增角色
// @Tags 角色
// @Security BearerAuth
// @Param body body RoleAddBody true "-"
// @Success 200 {object} response.Result[any]
// @Router /role/add [post]
func (r *Controller) Add(c *gin.Context) {
	var body RoleAddBody
	if !response.BindJSON(c, &body) {
		return
	}
	if err := Add(c.Request.Context(), body); err != nil {
		response.FailWithError(c, err)
		return
	}
	log.CreateWithMq(c, "角色管理")
	response.OK(c, nil)
}

// @Summary 修改角色
// @Tags 角色
// @Security BearerAuth
// @Param roleKey query string true "-"
// @Param body body RoleEditBody true "-"
// @Success 200 {object} response.Result[any]
// @Router /role/edit [post]
func (r *Controller) Edit(c *gin.Context) {
	var query common.RoleKeyQuery
	if !response.BindQuery(c, &query) {
		return
	}

	var body RoleEditBody
	if !response.BindJSON(c, &body) {
		return
	}
	if err := Edit(c.Request.Context(), query.RoleKey, body); err != nil {
		response.FailWithError(c, err)
		return
	}
	log.CreateWithMq(c, "角色管理")
	response.OK(c, nil)
}

// @Summary 删除角色
// @Tags 角色
// @Security BearerAuth
// @Param roleKey query string true "-"
// @Success 200 {object} response.Result[any]
// @Router /role/delete [post]
func (r *Controller) Delete(c *gin.Context) {
	var query common.RoleKeyQuery
	if !response.BindQuery(c, &query) {
		return
	}
	if err := Delete(c.Request.Context(), query.RoleKey); err != nil {
		response.FailWithError(c, err)
		return
	}
	log.CreateWithMq(c, "角色管理")
	response.OK(c, nil)
}
