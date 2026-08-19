package service

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"rbackeeper/global"
	"rbackeeper/internal/dao"
	"rbackeeper/internal/dto"
	"rbackeeper/internal/model"
	"rbackeeper/response"
)

// ==================== DictType 字典类型 ====================

// 字典类型列表
func DictTypeList(ctx context.Context, query dto.DictTypeListQuery) (*dto.PageResp[model.DictType], error) {
	list, total, err := dao.GetDictTypeList(query)
	if err != nil {
		return nil, err
	}
	return &dto.PageResp[model.DictType]{
		PageNum: query.PageNum, PageSize: query.PageSize, Total: total, List: list,
	}, nil
}

// 新增字典类型（dict_type 编码唯一性校验）。新类型暂无数据，无需动缓存。
func DictTypeAdd(ctx context.Context, body dto.DictTypeAddBody) error {
	if exist, err := dao.ExistsDictTypeByCode(global.DB, body.DictType, 0); err != nil {
		return err
	} else if exist {
		return response.NewBizError(response.ErrInvalidParam, "字典类型已存在")
	}

	t := &model.DictType{
		DictName: body.DictName,
		DictType: body.DictType,
		Status:   body.Status,
	}
	return global.DB.Transaction(func(tx *gorm.DB) error {
		return dao.CreateDictType(tx, t)
	})
}

// 修改字典类型
// dict_type 编码变更时：唯一性校验 + 联动更新字典数据 dict_type 字段 + 新旧编码缓存延时双删。
// 仅名称/状态变更不影响数据缓存（缓存键以 dict_type 编码为准）。
func DictTypeEdit(ctx context.Context, id uint, body dto.DictTypeEditBody) error {
	t, err := dao.GetDictTypeByID(global.DB, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.NewBizError(response.ErrNotFound, "字典类型不存在")
		}
		return err
	}

	updates := map[string]any{}
	if body.DictName != nil {
		updates["dict_name"] = *body.DictName
	}
	if body.Status != nil {
		updates["status"] = *body.Status
	}

	// dict_type 编码变更处理
	codeChanged := body.DictType != nil && *body.DictType != t.DictType
	if codeChanged {
		if exist, err := dao.ExistsDictTypeByCode(global.DB, *body.DictType, id); err != nil {
			return err
		} else if exist {
			return response.NewBizError(response.ErrInvalidParam, "字典类型已存在")
		}
		updates["dict_type"] = *body.DictType
	}

	cacheTypes := []string{}
	if codeChanged {
		// 新旧编码缓存都需失效：旧编码下的数据将迁移到新编码
		cacheTypes = []string{t.DictType, *body.DictType}
		// 延时双删：先删缓存
		deleteDictCaches(cacheTypes)
	}

	if err := global.DB.Transaction(func(tx *gorm.DB) error {
		if err := dao.UpdateDictTypeByID(tx, id, updates); err != nil {
			return err
		}
		// 编码变更：联动更新字典数据的 dict_type 字段
		if codeChanged {
			if err := dao.UpdateDictDataCodeByType(tx, t.DictType, *body.DictType); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}

	if codeChanged {
		// 延时双删：延迟 500ms 再删一次
		delayedDeleteDictCaches(cacheTypes)
	}
	return nil
}

// 删除字典类型：联动软删该类型下全部字典数据，并清理缓存。
func DictTypeDelete(ctx context.Context, id uint) error {
	t, err := dao.GetDictTypeByID(global.DB, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.NewBizError(response.ErrNotFound, "字典类型不存在")
		}
		return err
	}

	cacheTypes := []string{t.DictType}
	// 延时双删：先删缓存
	deleteDictCaches(cacheTypes)

	if err := global.DB.Transaction(func(tx *gorm.DB) error {
		return dao.DeleteDictTypeByID(tx, id)
	}); err != nil {
		return err
	}

	// 延时双删：延迟 500ms 再删一次
	delayedDeleteDictCaches(cacheTypes)
	return nil
}

// ==================== DictData 字典数据 ====================

// 字典数据列表（分页，不走缓存：含停用项，供管理端查看）
func DictDataList(ctx context.Context, query dto.DictDataListQuery) (*dto.PageResp[model.DictData], error) {
	list, total, err := dao.GetDictDataList(query)
	if err != nil {
		return nil, err
	}
	return &dto.PageResp[model.DictData]{
		PageNum: query.PageNum, PageSize: query.PageSize, Total: total, List: list,
	}, nil
}

// 新增字典数据：类型存在性 + 同类型下 dict_value 唯一性
func DictDataAdd(ctx context.Context, body dto.DictDataAddBody) error {
	// 字典类型存在性校验
	if exist, err := dao.ExistsDictTypeByCode(global.DB, body.DictType, 0); err != nil {
		return err
	} else if !exist {
		return response.NewBizError(response.ErrInvalidParam, "字典类型不存在")
	}
	// 同类型下 dict_value 唯一性
	if dup, err := dao.ExistsDictDataByValue(global.DB, body.DictType, body.DictValue, 0); err != nil {
		return err
	} else if dup {
		return response.NewBizError(response.ErrInvalidParam, "字典键值已存在")
	}

	d := &model.DictData{
		DictSort:  body.DictSort,
		DictLabel: body.DictLabel,
		DictValue: body.DictValue,
		DictType:  body.DictType,
		CssClass:  body.CssClass,
		ListClass: body.ListClass,
		IsDefault: body.IsDefault,
		Status:    body.Status,
	}

	if err := global.DB.Transaction(func(tx *gorm.DB) error {
		return dao.CreateDictData(tx, d)
	}); err != nil {
		return err
	}

	return nil
}

// 修改字典数据（按 id 定位；dict_type 不可变）。
// dict_value 变更走同类型唯一性校验，缓存按原 dict_type 延时双删。
func DictDataEdit(ctx context.Context, id uint, body dto.DictDataEditBody) error {
	d, err := dao.GetDictDataByID(global.DB, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.NewBizError(response.ErrNotFound, "字典数据不存在")
		}
		return err
	}

	updates := map[string]any{}
	if body.DictSort != nil {
		updates["dict_sort"] = *body.DictSort
	}
	if body.DictLabel != nil {
		updates["dict_label"] = *body.DictLabel
	}
	if body.CssClass != nil {
		updates["css_class"] = *body.CssClass
	}
	if body.ListClass != nil {
		updates["list_class"] = *body.ListClass
	}
	if body.IsDefault != nil {
		updates["is_default"] = *body.IsDefault
	}
	if body.Status != nil {
		updates["status"] = *body.Status
	}
	// dict_value 变更：同类型下唯一性校验（dict_type 不可变，按原类型校验）
	if body.DictValue != nil && *body.DictValue != d.DictValue {
		if dup, err := dao.ExistsDictDataByValue(global.DB, d.DictType, *body.DictValue, id); err != nil {
			return err
		} else if dup {
			return response.NewBizError(response.ErrInvalidParam, "字典键值已存在")
		}
		updates["dict_value"] = *body.DictValue
	}

	cacheTypes := []string{d.DictType}
	// 延时双删：先删缓存
	deleteDictCaches(cacheTypes)

	if err := global.DB.Transaction(func(tx *gorm.DB) error {
		return dao.UpdateDictDataByID(tx, id, updates)
	}); err != nil {
		return err
	}

	// 延时双删：延迟 500ms 再删一次
	delayedDeleteDictCaches(cacheTypes)
	return nil
}

// 删除字典数据（按 id 定位）：清理所属类型缓存。
func DictDataDelete(ctx context.Context, id uint) error {
	d, err := dao.GetDictDataByID(global.DB, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.NewBizError(response.ErrNotFound, "字典数据不存在")
		}
		return err
	}

	cacheTypes := []string{d.DictType}
	// 延时双删：先删缓存
	deleteDictCaches(cacheTypes)

	if err := global.DB.Transaction(func(tx *gorm.DB) error {
		return dao.DeleteDictDataByID(tx, id)
	}); err != nil {
		return err
	}

	// 延时双删：延迟 500ms 再删一次
	delayedDeleteDictCaches(cacheTypes)
	return nil
}
