package dict

import (
	"rbackeeper/common"
	"rbackeeper/model"
)

// ==================== DictType 字典类型 ====================

// DictTypeListQuery 字典类型列表查询
type DictTypeListQuery struct {
	common.Page
	// 字典名称(模糊)
	DictName string `form:"dictName"`
	// 字典类型(模糊)
	DictType string `form:"dictType"`
	// 状态 0停用 1正常  不传则不筛选
	Status *int8 `form:"status" binding:"omitempty,oneof=0 1"`
}

// DictTypeListResp 字典类型列表响应
type DictTypeListResp struct {
	common.PageResp[model.DictType]
}

// DictTypeAddBody 新增字典类型
type DictTypeAddBody struct {
	// 字典名称
	DictName string `json:"dictName" binding:"required,min=2,max=100"`
	// 字典类型（唯一编码）
	DictType string `json:"dictType" binding:"required,min=2,max=100"`
	// 状态 0停用 1正常
	Status int8 `json:"status" binding:"omitempty,oneof=0 1"`
}

// DictTypeEditBody 修改字典类型
type DictTypeEditBody struct {
	// 字典名称
	DictName *string `json:"dictName" binding:"omitempty,min=2,max=100"`
	// 字典类型（唯一编码，变更时联动数据与缓存）
	DictType *string `json:"dictType" binding:"omitempty,min=2,max=100"`
	// 状态 0停用 1正常
	Status *int8 `json:"status" binding:"omitempty,oneof=0 1"`
}

// ==================== DictData 字典数据 ====================

// DictDataListQuery 字典数据列表查询（按 dictType 分组过滤）
type DictDataListQuery struct {
	common.Page
	// 字典类型（精确匹配，按类型分组展示）
	DictType string `form:"dictType"`
	// 字典标签(模糊)
	DictLabel string `form:"dictLabel"`
	// 状态 0停用 1正常  不传则不筛选
	Status *int8 `form:"status" binding:"omitempty,oneof=0 1"`
}

// DictDataListResp 字典数据列表响应
type DictDataListResp struct {
	common.PageResp[model.DictData]
}

// DictDataQuery 按字典类型查询可用数据（缓存读取，供前端下拉/回显）
type DictDataQuery struct {
	// 字典类型
	DictType string `form:"dictType" binding:"required,min=1,max=100"`
}

// DictDataAddBody 新增字典数据
type DictDataAddBody struct {
	// 字典排序
	DictSort int `json:"dictSort"`
	// 字典标签
	DictLabel string `json:"dictLabel" binding:"required,min=1,max=100"`
	// 字典键值
	DictValue string `json:"dictValue" binding:"required,min=1,max=100"`
	// 字典类型
	DictType string `json:"dictType" binding:"required,min=1,max=100"`
	// 样式属性
	CssClass string `json:"cssClass" binding:"omitempty,max=100"`
	// 表格回显样式
	ListClass string `json:"listClass" binding:"omitempty,max=100"`
	// 是否默认 0否 1是
	IsDefault int8 `json:"isDefault" binding:"omitempty,oneof=0 1"`
	// 状态 0停用 1正常
	Status int8 `json:"status" binding:"omitempty,oneof=0 1"`
}

// DictDataEditBody 修改字典数据（dict_type 不可变，如需迁移请删除后新增）
type DictDataEditBody struct {
	// 字典排序
	DictSort *int `json:"dictSort"`
	// 字典标签
	DictLabel *string `json:"dictLabel" binding:"omitempty,min=1,max=100"`
	// 字典键值
	DictValue *string `json:"dictValue" binding:"omitempty,min=1,max=100"`
	// 样式属性
	CssClass *string `json:"cssClass" binding:"omitempty,max=100"`
	// 表格回显样式
	ListClass *string `json:"listClass" binding:"omitempty,max=100"`
	// 是否默认 0否 1是
	IsDefault *int8 `json:"isDefault" binding:"omitempty,oneof=0 1"`
	// 状态 0停用 1正常
	Status *int8 `json:"status" binding:"omitempty,oneof=0 1"`
}

// ==================== 批量删除 ====================

// DictTypeBatchDeleteBody 批量删除字典类型（级联清理其下字典数据）
type DictTypeBatchDeleteBody struct {
	IDs []uint `json:"ids" binding:"required,min=1"`
}

// DictDataBatchDeleteBody 批量删除字典数据
type DictDataBatchDeleteBody struct {
	IDs []uint `json:"ids" binding:"required,min=1"`
}
