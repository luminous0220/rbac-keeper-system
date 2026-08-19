package controller

import (
	"github.com/gin-gonic/gin"

	"rbackeeper/internal/dto"
	"rbackeeper/internal/service"
	"rbackeeper/response"
)

// MenuController 菜单控制器
type MenuController struct{}

// @Summary 菜单树
// @Tags 菜单
// @Security BearerAuth
// @Param query query dto.MenuTreeQuery true "-"
// @Success 200 {object} response.Result[[]model.Menu]
// @Router /menu/tree [get]
func (m *MenuController) Tree(c *gin.Context) {
	var query dto.MenuTreeQuery
	if !response.BindQuery(c, &query) {
		return
	}
	resp, err := service.MenuTree(c.Request.Context(), query)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	response.OK(c, resp)
}

// @Summary 新增菜单
// @Tags 菜单
// @Security BearerAuth
// @Param body body dto.MenuAddBody true "-"
// @Success 200 {object} response.Result[any]
// @Router /menu/add [post]
func (m *MenuController) Add(c *gin.Context) {
	var body dto.MenuAddBody
	if !response.BindJSON(c, &body) {
		return
	}
	if err := service.MenuAdd(c.Request.Context(), body); err != nil {
		response.FailWithError(c, err)
		return
	}
	service.CreateLogWithMq(c, "菜单管理")
	response.OK(c, nil)
}

// @Summary 修改菜单
// @Tags 菜单
// @Security BearerAuth
// @Param id query int true "-"
// @Param body body dto.MenuEditBody true "-"
// @Success 200 {object} response.Result[any]
// @Router /menu/edit [post]
func (m *MenuController) Edit(c *gin.Context) {
	var query dto.IdQuery
	if !response.BindQuery(c, &query) {
		return
	}

	var body dto.MenuEditBody
	if !response.BindJSON(c, &body) {
		return
	}
	if err := service.MenuEdit(c.Request.Context(), query.ID, body); err != nil {
		response.FailWithError(c, err)
		return
	}
	service.CreateLogWithMq(c, "菜单管理")
	response.OK(c, nil)
}

// @Summary 删除菜单
// @Tags 菜单
// @Security BearerAuth
// @Param id query int true "-"
// @Success 200 {object} response.Result[any]
// @Router /menu/delete [post]
func (m *MenuController) Delete(c *gin.Context) {
	var query dto.IdQuery
	if !response.BindQuery(c, &query) {
		return
	}
	if err := service.MenuDelete(c.Request.Context(), query.ID); err != nil {
		response.FailWithError(c, err)
		return
	}
	service.CreateLogWithMq(c, "菜单管理")
	response.OK(c, nil)
}
