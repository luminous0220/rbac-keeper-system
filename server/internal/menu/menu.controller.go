package menu

import (
	"github.com/gin-gonic/gin"

	"rbackeeper/common"
	"rbackeeper/internal/log"
	"rbackeeper/response"
)

// Controller 菜单控制器
type Controller struct{}

// @Summary 菜单树
// @Tags 菜单
// @Security BearerAuth
// @Param query query MenuTreeQuery true "-"
// @Success 200 {object} response.Result[[]model.Menu]
// @Router /menu/tree [get]
func (m *Controller) Tree(c *gin.Context) {
	var query MenuTreeQuery
	if !response.BindQuery(c, &query) {
		return
	}
	resp, err := Tree(c.Request.Context(), query)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	response.OK(c, resp)
}

// @Summary 新增菜单
// @Tags 菜单
// @Security BearerAuth
// @Param body body MenuAddBody true "-"
// @Success 200 {object} response.Result[any]
// @Router /menu/add [post]
func (m *Controller) Add(c *gin.Context) {
	var body MenuAddBody
	if !response.BindJSON(c, &body) {
		return
	}
	if err := Add(c.Request.Context(), body); err != nil {
		response.FailWithError(c, err)
		return
	}
	log.CreateWithMq(c, "菜单管理")
	response.OK(c, nil)
}

// @Summary 修改菜单
// @Tags 菜单
// @Security BearerAuth
// @Param id query int true "-"
// @Param body body MenuEditBody true "-"
// @Success 200 {object} response.Result[any]
// @Router /menu/edit [post]
func (m *Controller) Edit(c *gin.Context) {
	var query common.IdQuery
	if !response.BindQuery(c, &query) {
		return
	}

	var body MenuEditBody
	if !response.BindJSON(c, &body) {
		return
	}
	if err := Edit(c.Request.Context(), query.ID, body); err != nil {
		response.FailWithError(c, err)
		return
	}
	log.CreateWithMq(c, "菜单管理")
	response.OK(c, nil)
}

// @Summary 删除菜单
// @Tags 菜单
// @Security BearerAuth
// @Param id query int true "-"
// @Success 200 {object} response.Result[any]
// @Router /menu/delete [post]
func (m *Controller) Delete(c *gin.Context) {
	var query common.IdQuery
	if !response.BindQuery(c, &query) {
		return
	}
	if err := Delete(c.Request.Context(), query.ID); err != nil {
		response.FailWithError(c, err)
		return
	}
	log.CreateWithMq(c, "菜单管理")
	response.OK(c, nil)
}
