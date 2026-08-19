package controller

import (
	"github.com/gin-gonic/gin"

	"rbackeeper/internal/dto"
	"rbackeeper/internal/service"
	"rbackeeper/response"
)

// RoleController 角色控制器
type RoleController struct{}

// @Summary 角色列表
// @Tags 角色
// @Security BearerAuth
// @Param query query dto.RoleListQuery true "-"
// @Success 200 {object} response.Result[[]dto.RoleItem]
// @Router /role/list [get]
func (r *RoleController) List(c *gin.Context) {
	var query dto.RoleListQuery
	if !response.BindQuery(c, &query) {
		return
	}
	resp, err := service.RoleList(c.Request.Context(), query)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	response.OK(c, resp)
}

// @Summary 角色详情（按 roleKey，返回含菜单/部门授权）
// @Tags 角色
// @Security BearerAuth
// @Param roleKey query string true "-"
// @Success 200 {object} response.Result[any]
// @Router /role/info [get]
func (r *RoleController) Info(c *gin.Context) {
	var query dto.RoleKeyQuery
	if !response.BindQuery(c, &query) {
		return
	}
	role, err := service.RoleDetail(c.Request.Context(), query.RoleKey)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	response.OK(c, role)
}

// @Summary 新增角色
// @Tags 角色
// @Security BearerAuth
// @Param body body dto.RoleAddBody true "-"
// @Success 200 {object} response.Result[any]
// @Router /role/add [post]
func (r *RoleController) Add(c *gin.Context) {
	var body dto.RoleAddBody
	if !response.BindJSON(c, &body) {
		return
	}
	if err := service.RoleAdd(c.Request.Context(), body); err != nil {
		response.FailWithError(c, err)
		return
	}
	service.CreateLogWithMq(c, "角色管理")
	response.OK(c, nil)
}

// @Summary 修改角色
// @Tags 角色
// @Security BearerAuth
// @Param roleKey query string true "-"
// @Param body body dto.RoleEditBody true "-"
// @Success 200 {object} response.Result[any]
// @Router /role/edit [post]
func (r *RoleController) Edit(c *gin.Context) {
	var query dto.RoleKeyQuery
	if !response.BindQuery(c, &query) {
		return
	}

	var body dto.RoleEditBody
	if !response.BindJSON(c, &body) {
		return
	}
	if err := service.RoleEdit(c.Request.Context(), query.RoleKey, body); err != nil {
		response.FailWithError(c, err)
		return
	}
	service.CreateLogWithMq(c, "角色管理")
	response.OK(c, nil)
}

// @Summary 删除角色
// @Tags 角色
// @Security BearerAuth
// @Param roleKey query string true "-"
// @Success 200 {object} response.Result[any]
// @Router /role/delete [post]
func (r *RoleController) Delete(c *gin.Context) {
	var query dto.RoleKeyQuery
	if !response.BindQuery(c, &query) {
		return
	}
	if err := service.RoleDelete(c.Request.Context(), query.RoleKey); err != nil {
		response.FailWithError(c, err)
		return
	}
	service.CreateLogWithMq(c, "角色管理")
	response.OK(c, nil)
}
