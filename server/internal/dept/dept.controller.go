package dept

import (
	"github.com/gin-gonic/gin"

	"rbackeeper/common"
	"rbackeeper/internal/log"
	"rbackeeper/response"
)

// Controller 部门控制器
type Controller struct{}

// @Summary 部门树
// @Tags 部门
// @Security BearerAuth
// @Param query query DeptTreeQuery true "-"
// @Success 200 {object} response.Result[[]model.Dept]
// @Router /dept/tree [get]
func (d *Controller) Tree(c *gin.Context) {
	var query DeptTreeQuery
	if !response.BindQuery(c, &query) {
		return
	}
	resp, err := Tree(c, query)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	response.OK(c, resp)
}

// @Summary 新增部门
// @Tags 部门
// @Security BearerAuth
// @Param body body DeptAddBody true "-"
// @Success 200 {object} response.Result[any]
// @Router /dept/add [post]
func (d *Controller) Add(c *gin.Context) {
	var body DeptAddBody
	if !response.BindJSON(c, &body) {
		return
	}
	if err := Add(c, body); err != nil {
		response.FailWithError(c, err)
		return
	}
	log.CreateWithMq(c, "部门管理")
	response.OK(c, nil)
}

// @Summary 修改部门
// @Tags 部门
// @Security BearerAuth
// @Param id query int true "-"
// @Param body body DeptEditBody true "-"
// @Success 200 {object} response.Result[any]
// @Router /dept/edit [post]
func (d *Controller) Edit(ctx *gin.Context) {
	var query common.IdQuery
	if !response.BindQuery(ctx, &query) {
		return
	}

	var body DeptEditBody
	if !response.BindJSON(ctx, &body) {
		return
	}
	if err := Edit(ctx, query.ID, body); err != nil {
		response.FailWithError(ctx, err)
		return
	}
	log.CreateWithMq(ctx, "部门管理")
	response.OK(ctx, nil)
}

// @Summary 删除部门
// @Tags 部门
// @Security BearerAuth
// @Param id query int true "-"
// @Success 200 {object} response.Result[any]
// @Router /dept/delete [post]
func (d *Controller) Delete(c *gin.Context) {
	var query common.IdQuery
	if !response.BindQuery(c, &query) {
		return
	}
	if err := Delete(c, query.ID); err != nil {
		response.FailWithError(c, err)
		return
	}
	log.CreateWithMq(c, "部门管理")
	response.OK(c, nil)
}
