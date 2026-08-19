package controller

import (
	"github.com/gin-gonic/gin"

	"rbackeeper/internal/dto"
	"rbackeeper/internal/service"
	"rbackeeper/response"
)

// DeptController 部门控制器
type DeptController struct{}

// @Summary 部门树
// @Tags 部门
// @Security BearerAuth
// @Param query query dto.DeptTreeQuery true "-"
// @Success 200 {object} response.Result[[]model.Dept]
// @Router /dept/tree [get]
func (d *DeptController) Tree(c *gin.Context) {
	var query dto.DeptTreeQuery
	if !response.BindQuery(c, &query) {
		return
	}
	resp, err := service.DeptTree(c, query)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	response.OK(c, resp)
}

// @Summary 新增部门
// @Tags 部门
// @Security BearerAuth
// @Param body body dto.DeptAddBody true "-"
// @Success 200 {object} response.Result[any]
// @Router /dept/add [post]
func (d *DeptController) Add(c *gin.Context) {
	var body dto.DeptAddBody
	if !response.BindJSON(c, &body) {
		return
	}
	if err := service.DeptAdd(c, body); err != nil {
		response.FailWithError(c, err)
		return
	}
	service.CreateLogWithMq(c, "部门管理")
	response.OK(c, nil)
}

// @Summary 修改部门
// @Tags 部门
// @Security BearerAuth
// @Param id query int true "-"
// @Param body body dto.DeptEditBody true "-"
// @Success 200 {object} response.Result[any]
// @Router /dept/edit [post]
func (d *DeptController) Edit(ctx *gin.Context) {
	var query dto.IdQuery
	if !response.BindQuery(ctx, &query) {
		return
	}

	var body dto.DeptEditBody
	if !response.BindJSON(ctx, &body) {
		return
	}
	if err := service.DeptEdit(ctx, query.ID, body); err != nil {
		response.FailWithError(ctx, err)
		return
	}
	service.CreateLogWithMq(ctx, "部门管理")
	response.OK(ctx, nil)
}

// @Summary 删除部门
// @Tags 部门
// @Security BearerAuth
// @Param id query int true "-"
// @Success 200 {object} response.Result[any]
// @Router /dept/delete [post]
func (d *DeptController) Delete(c *gin.Context) {
	var query dto.IdQuery
	if !response.BindQuery(c, &query) {
		return
	}
	if err := service.DeptDelete(c, query.ID); err != nil {
		response.FailWithError(c, err)
		return
	}
	service.CreateLogWithMq(c, "部门管理")
	response.OK(c, nil)
}
