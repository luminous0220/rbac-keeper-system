package dict

import (
	"github.com/gin-gonic/gin"

	"rbackeeper/common"
	"rbackeeper/internal/log"
	"rbackeeper/response"
)

// Controller 字典控制器（字典类型 + 字典数据）
type Controller struct{}

// ==================== 字典类型 ====================

// @Summary 字典类型列表
// @Tags 字典类型
// @Security BearerAuth
// @Param query query DictTypeListQuery true "-"
// @Success 200 {object} response.Result[DictTypeListResp]
// @Router /dict/type/list [get]
func (dc *Controller) TypeList(c *gin.Context) {
	var query DictTypeListQuery
	if !response.BindQuery(c, &query) {
		return
	}
	resp, err := TypeList(c.Request.Context(), query)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	response.OK(c, resp)
}

// @Summary 新增字典类型
// @Tags 字典类型
// @Security BearerAuth
// @Param body body DictTypeAddBody true "-"
// @Success 200 {object} response.Result[any]
// @Router /dict/type/add [post]
func (dc *Controller) TypeAdd(c *gin.Context) {
	var body DictTypeAddBody
	if !response.BindJSON(c, &body) {
		return
	}
	if err := TypeAdd(c.Request.Context(), body); err != nil {
		response.FailWithError(c, err)
		return
	}
	log.CreateWithMq(c, "字典类型")
	response.OK(c, nil)
}

// @Summary 修改字典类型
// @Tags 字典类型
// @Security BearerAuth
// @Param id query int true "-"
// @Param body body DictTypeEditBody true "-"
// @Success 200 {object} response.Result[any]
// @Router /dict/type/edit [post]
func (dc *Controller) TypeEdit(c *gin.Context) {
	var query common.IdQuery
	if !response.BindQuery(c, &query) {
		return
	}
	var body DictTypeEditBody
	if !response.BindJSON(c, &body) {
		return
	}
	if err := TypeEdit(c.Request.Context(), query.ID, body); err != nil {
		response.FailWithError(c, err)
		return
	}
	log.CreateWithMq(c, "字典类型")
	response.OK(c, nil)
}

// @Summary 删除字典类型
// @Tags 字典类型
// @Security BearerAuth
// @Param id query int true "-"
// @Success 200 {object} response.Result[any]
// @Router /dict/type/delete [post]
func (dc *Controller) TypeDelete(c *gin.Context) {
	var query common.IdQuery
	if !response.BindQuery(c, &query) {
		return
	}
	if err := TypeDelete(c.Request.Context(), query.ID); err != nil {
		response.FailWithError(c, err)
		return
	}
	log.CreateWithMq(c, "字典类型")
	response.OK(c, nil)
}

// ==================== 字典数据 ====================

// @Summary 字典数据列表
// @Tags 字典数据
// @Security BearerAuth
// @Param query query DictDataListQuery true "-"
// @Success 200 {object} response.Result[DictDataListResp]
// @Router /dict/data/list [get]
func (dc *Controller) DataList(c *gin.Context) {
	var query DictDataListQuery
	if !response.BindQuery(c, &query) {
		return
	}
	resp, err := DataList(c.Request.Context(), query)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	response.OK(c, resp)
}

// @Summary 按字典类型查询可用数据（缓存）
// @Tags 字典数据
// @Security BearerAuth
// @Param query query DictDataQuery true "-"
// @Success 200 {object} response.Result[[]model.DictData]
// @Router /dict/data/type [get]
func (dc *Controller) DataByType(c *gin.Context) {
	var query DictDataQuery
	if !response.BindQuery(c, &query) {
		return
	}
	resp, err := DataByType(c.Request.Context(), query.DictType)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	response.OK(c, resp)
}

// @Summary 新增字典数据
// @Tags 字典数据
// @Security BearerAuth
// @Param body body DictDataAddBody true "-"
// @Success 200 {object} response.Result[any]
// @Router /dict/data/add [post]
func (dc *Controller) DataAdd(c *gin.Context) {
	var body DictDataAddBody
	if !response.BindJSON(c, &body) {
		return
	}
	if err := DataAdd(c.Request.Context(), body); err != nil {
		response.FailWithError(c, err)
		return
	}
	log.CreateWithMq(c, "字典数据")
	response.OK(c, nil)
}

// @Summary 修改字典数据
// @Tags 字典数据
// @Security BearerAuth
// @Param id query int true "-"
// @Param body body DictDataEditBody true "-"
// @Success 200 {object} response.Result[any]
// @Router /dict/data/edit [post]
func (dc *Controller) DataEdit(c *gin.Context) {
	var query common.IdQuery
	if !response.BindQuery(c, &query) {
		return
	}
	var body DictDataEditBody
	if !response.BindJSON(c, &body) {
		return
	}
	if err := DataEdit(c.Request.Context(), query.ID, body); err != nil {
		response.FailWithError(c, err)
		return
	}
	log.CreateWithMq(c, "字典数据")
	response.OK(c, nil)
}

// @Summary 删除字典数据
// @Tags 字典数据
// @Security BearerAuth
// @Param id query int true "-"
// @Success 200 {object} response.Result[any]
// @Router /dict/data/delete [post]
func (dc *Controller) DataDelete(c *gin.Context) {
	var query common.IdQuery
	if !response.BindQuery(c, &query) {
		return
	}
	if err := DataDelete(c.Request.Context(), query.ID); err != nil {
		response.FailWithError(c, err)
		return
	}
	log.CreateWithMq(c, "字典数据")
	response.OK(c, nil)
}
