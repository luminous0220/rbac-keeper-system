package controller

import (
	"github.com/gin-gonic/gin"

	"rbackeeper/internal/dto"
	"rbackeeper/internal/service"
	"rbackeeper/response"
)

// DictController 字典控制器（字典类型 + 字典数据）
type DictController struct{}

// ==================== 字典类型 ====================

// @Summary 字典类型列表
// @Tags 字典类型
// @Security BearerAuth
// @Param query query dto.DictTypeListQuery true "-"
// @Success 200 {object} response.Result[dto.DictTypeListResp]
// @Router /dict/type/list [get]
func (dc *DictController) TypeList(c *gin.Context) {
	var query dto.DictTypeListQuery
	if !response.BindQuery(c, &query) {
		return
	}
	resp, err := service.DictTypeList(c.Request.Context(), query)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	response.OK(c, resp)
}

// @Summary 新增字典类型
// @Tags 字典类型
// @Security BearerAuth
// @Param body body dto.DictTypeAddBody true "-"
// @Success 200 {object} response.Result[any]
// @Router /dict/type/add [post]
func (dc *DictController) TypeAdd(c *gin.Context) {
	var body dto.DictTypeAddBody
	if !response.BindJSON(c, &body) {
		return
	}
	if err := service.DictTypeAdd(c.Request.Context(), body); err != nil {
		response.FailWithError(c, err)
		return
	}
	service.CreateLogWithMq(c, "字典类型")
	response.OK(c, nil)
}

// @Summary 修改字典类型
// @Tags 字典类型
// @Security BearerAuth
// @Param id query int true "-"
// @Param body body dto.DictTypeEditBody true "-"
// @Success 200 {object} response.Result[any]
// @Router /dict/type/edit [post]
func (dc *DictController) TypeEdit(c *gin.Context) {
	var query dto.IdQuery
	if !response.BindQuery(c, &query) {
		return
	}
	var body dto.DictTypeEditBody
	if !response.BindJSON(c, &body) {
		return
	}
	if err := service.DictTypeEdit(c.Request.Context(), query.ID, body); err != nil {
		response.FailWithError(c, err)
		return
	}
	service.CreateLogWithMq(c, "字典类型")
	response.OK(c, nil)
}

// @Summary 删除字典类型
// @Tags 字典类型
// @Security BearerAuth
// @Param id query int true "-"
// @Success 200 {object} response.Result[any]
// @Router /dict/type/delete [post]
func (dc *DictController) TypeDelete(c *gin.Context) {
	var query dto.IdQuery
	if !response.BindQuery(c, &query) {
		return
	}
	if err := service.DictTypeDelete(c.Request.Context(), query.ID); err != nil {
		response.FailWithError(c, err)
		return
	}
	service.CreateLogWithMq(c, "字典类型")
	response.OK(c, nil)
}

// ==================== 字典数据 ====================

// @Summary 字典数据列表
// @Tags 字典数据
// @Security BearerAuth
// @Param query query dto.DictDataListQuery true "-"
// @Success 200 {object} response.Result[dto.DictDataListResp]
// @Router /dict/data/list [get]
func (dc *DictController) DataList(c *gin.Context) {
	var query dto.DictDataListQuery
	if !response.BindQuery(c, &query) {
		return
	}
	resp, err := service.DictDataList(c.Request.Context(), query)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	response.OK(c, resp)
}

// @Summary 按字典类型查询可用数据（缓存）
// @Tags 字典数据
// @Security BearerAuth
// @Param query query dto.DictDataQuery true "-"
// @Success 200 {object} response.Result[[]model.DictData]
// @Router /dict/data/type [get]
func (dc *DictController) DataByType(c *gin.Context) {
	var query dto.DictDataQuery
	if !response.BindQuery(c, &query) {
		return
	}
	resp, err := service.DictDataByType(c.Request.Context(), query.DictType)
	if err != nil {
		response.FailWithError(c, err)
		return
	}
	response.OK(c, resp)
}

// @Summary 新增字典数据
// @Tags 字典数据
// @Security BearerAuth
// @Param body body dto.DictDataAddBody true "-"
// @Success 200 {object} response.Result[any]
// @Router /dict/data/add [post]
func (dc *DictController) DataAdd(c *gin.Context) {
	var body dto.DictDataAddBody
	if !response.BindJSON(c, &body) {
		return
	}
	if err := service.DictDataAdd(c.Request.Context(), body); err != nil {
		response.FailWithError(c, err)
		return
	}
	service.CreateLogWithMq(c, "字典数据")
	response.OK(c, nil)
}

// @Summary 修改字典数据
// @Tags 字典数据
// @Security BearerAuth
// @Param id query int true "-"
// @Param body body dto.DictDataEditBody true "-"
// @Success 200 {object} response.Result[any]
// @Router /dict/data/edit [post]
func (dc *DictController) DataEdit(c *gin.Context) {
	var query dto.IdQuery
	if !response.BindQuery(c, &query) {
		return
	}
	var body dto.DictDataEditBody
	if !response.BindJSON(c, &body) {
		return
	}
	if err := service.DictDataEdit(c.Request.Context(), query.ID, body); err != nil {
		response.FailWithError(c, err)
		return
	}
	service.CreateLogWithMq(c, "字典数据")
	response.OK(c, nil)
}

// @Summary 删除字典数据
// @Tags 字典数据
// @Security BearerAuth
// @Param id query int true "-"
// @Success 200 {object} response.Result[any]
// @Router /dict/data/delete [post]
func (dc *DictController) DataDelete(c *gin.Context) {
	var query dto.IdQuery
	if !response.BindQuery(c, &query) {
		return
	}
	if err := service.DictDataDelete(c.Request.Context(), query.ID); err != nil {
		response.FailWithError(c, err)
		return
	}
	service.CreateLogWithMq(c, "字典数据")
	response.OK(c, nil)
}
