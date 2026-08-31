package dict

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	zlog "github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"rbackeeper/common"
	"rbackeeper/global"
	"rbackeeper/model"
	"rbackeeper/response"
)

// ==================== 字典缓存 ====================

const (
	// cachePrefix 字典数据缓存键前缀，完整键为 sys_dict:{dictType}
	cachePrefix = "sys_dict:"
	// cacheExpire 缓存有效期，作为延时双删失败时的兜底失效手段
	cacheExpire = 30 * time.Minute
	// doubleDeleteDelay 延时双删二次删除延迟，覆盖「DB 提交前回源、提交后回填」的并发旧值窗口
	doubleDeleteDelay = 500 * time.Millisecond
)

// cacheKey 拼接字典缓存键
func cacheKey(dictType string) string {
	return cachePrefix + dictType
}

// 按字典类型查询可用数据（status=0），缓存优先：
//  1. 命中缓存直接返回；
//  2. 未命中回源 DB 并回填缓存（回填失败仅告警，不影响主流程）。
//
// 供前端下拉/回显使用，写操作通过延时双删保证最终一致。
func DataByType(ctx context.Context, dictType string) ([]model.DictData, error) {
	key := cacheKey(dictType)

	// 1. 查缓存
	if raw, err := global.Redis.Get(ctx, key).Result(); err == nil && raw != "" {
		var cached []model.DictData
		if err := json.Unmarshal([]byte(raw), &cached); err == nil {
			return cached, nil
		}
		// 反序列化失败：删除脏缓存，继续回源
		global.Redis.Del(ctx, key)
	}

	// 2. 回源 DB（仅取状态正常数据，按 dict_sort 排序）
	list, err := GetDataByType(dictType)
	if err != nil {
		return nil, err
	}

	// 3. 回填缓存
	if data, err := json.Marshal(list); err == nil {
		if err := global.Redis.Set(ctx, key, data, cacheExpire).Err(); err != nil {
			zlog.Warn().Err(err).Str("dict_type", dictType).Msg("回填字典缓存失败")
		}
	}
	return list, nil
}

// 立即删除指定类型集合的字典缓存（best-effort，失败仅告警）。
// 用于延时双删的第一次删除（写操作前）。使用独立 context，避免请求结束后 ctx 取消导致删除丢失。
func deleteCaches(dictTypes []string) {
	keys := collectCacheKeys(dictTypes)
	if len(keys) == 0 {
		return
	}
	if err := global.Redis.Del(context.Background(), keys...).Err(); err != nil {
		zlog.Warn().Err(err).Msg("删除字典缓存失败")
	}
}

// 延时双删的第二次删除：异步等待 500ms 后再删一次。
// 覆盖并发场景下「读请求在 DB 提交前回源旧值、提交后回填缓存」导致的脏缓存。
func delayedDeleteCaches(dictTypes []string) {
	keys := collectCacheKeys(dictTypes)
	if len(keys) == 0 {
		return
	}
	go func(keys []string) {
		time.Sleep(doubleDeleteDelay)
		if err := global.Redis.Del(context.Background(), keys...).Err(); err != nil {
			zlog.Warn().Err(err).Msg("延时删除字典缓存失败")
		}
	}(keys)
}

// collectCacheKeys 过滤空编码并拼接缓存键
func collectCacheKeys(dictTypes []string) []string {
	keys := make([]string, 0, len(dictTypes))
	for _, t := range dictTypes {
		if t != "" {
			keys = append(keys, cacheKey(t))
		}
	}
	return keys
}

// ==================== DictType 字典类型 ====================

// TypeList 字典类型列表
func TypeList(ctx context.Context, query DictTypeListQuery) (*common.PageResp[model.DictType], error) {
	list, total, err := GetTypeList(query)
	if err != nil {
		return nil, err
	}
	return &common.PageResp[model.DictType]{
		PageNum: query.PageNum, PageSize: query.PageSize, Total: total, List: list,
	}, nil
}

// TypeAdd 新增字典类型（dict_type 编码唯一性校验）。新类型暂无数据，无需动缓存。
func TypeAdd(ctx context.Context, body DictTypeAddBody) error {
	if exist, err := ExistsTypeByCode(global.DB, body.DictType, 0); err != nil {
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
		return CreateType(tx, t)
	})
}

// TypeEdit 修改字典类型
// dict_type 编码变更时：唯一性校验 + 联动更新字典数据 dict_type 字段 + 新旧编码缓存延时双删。
// 仅名称/状态变更不影响数据缓存（缓存键以 dict_type 编码为准）。
func TypeEdit(ctx context.Context, id uint, body DictTypeEditBody) error {
	t, err := GetTypeByID(global.DB, id)
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
		if exist, err := ExistsTypeByCode(global.DB, *body.DictType, id); err != nil {
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
		deleteCaches(cacheTypes)
	}

	if err := global.DB.Transaction(func(tx *gorm.DB) error {
		if err := UpdateTypeByID(tx, id, updates); err != nil {
			return err
		}
		// 编码变更：联动更新字典数据的 dict_type 字段
		if codeChanged {
			if err := UpdateDataCodeByType(tx, t.DictType, *body.DictType); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}

	if codeChanged {
		// 延时双删：延迟 500ms 再删一次
		delayedDeleteCaches(cacheTypes)
	}
	return nil
}

// TypeDelete 删除字典类型：联动软删该类型下全部字典数据，并清理缓存。
func TypeDelete(ctx context.Context, id uint) error {
	t, err := GetTypeByID(global.DB, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.NewBizError(response.ErrNotFound, "字典类型不存在")
		}
		return err
	}

	cacheTypes := []string{t.DictType}
	// 延时双删：先删缓存
	deleteCaches(cacheTypes)

	if err := global.DB.Transaction(func(tx *gorm.DB) error {
		return DeleteTypeByID(tx, id)
	}); err != nil {
		return err
	}

	// 延时双删：延迟 500ms 再删一次
	delayedDeleteCaches(cacheTypes)
	return nil
}

// ==================== DictData 字典数据 ====================

// DataList 字典数据列表（分页，不走缓存：含停用项，供管理端查看）
func DataList(ctx context.Context, query DictDataListQuery) (*common.PageResp[model.DictData], error) {
	list, total, err := GetDataList(query)
	if err != nil {
		return nil, err
	}
	return &common.PageResp[model.DictData]{
		PageNum: query.PageNum, PageSize: query.PageSize, Total: total, List: list,
	}, nil
}

// DataAdd 新增字典数据：类型存在性 + 同类型下 dict_value 唯一性
func DataAdd(ctx context.Context, body DictDataAddBody) error {
	// 字典类型存在性校验
	if exist, err := ExistsTypeByCode(global.DB, body.DictType, 0); err != nil {
		return err
	} else if !exist {
		return response.NewBizError(response.ErrInvalidParam, "字典类型不存在")
	}
	// 同类型下 dict_value 唯一性
	if dup, err := ExistsDataByValue(global.DB, body.DictType, body.DictValue, 0); err != nil {
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
		return CreateData(tx, d)
	}); err != nil {
		return err
	}

	return nil
}

// DataEdit 修改字典数据（按 id 定位；dict_type 不可变）。
// dict_value 变更走同类型唯一性校验，缓存按原 dict_type 延时双删。
func DataEdit(ctx context.Context, id uint, body DictDataEditBody) error {
	d, err := GetDataByID(global.DB, id)
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
		if dup, err := ExistsDataByValue(global.DB, d.DictType, *body.DictValue, id); err != nil {
			return err
		} else if dup {
			return response.NewBizError(response.ErrInvalidParam, "字典键值已存在")
		}
		updates["dict_value"] = *body.DictValue
	}

	cacheTypes := []string{d.DictType}
	// 延时双删：先删缓存
	deleteCaches(cacheTypes)

	if err := global.DB.Transaction(func(tx *gorm.DB) error {
		return UpdateDataByID(tx, id, updates)
	}); err != nil {
		return err
	}

	// 延时双删：延迟 500ms 再删一次
	delayedDeleteCaches(cacheTypes)
	return nil
}

// DataDelete 删除字典数据（按 id 定位）：清理所属类型缓存。
func DataDelete(ctx context.Context, id uint) error {
	d, err := GetDataByID(global.DB, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.NewBizError(response.ErrNotFound, "字典数据不存在")
		}
		return err
	}

	cacheTypes := []string{d.DictType}
	// 延时双删：先删缓存
	deleteCaches(cacheTypes)

	if err := global.DB.Transaction(func(tx *gorm.DB) error {
		return DeleteDataByID(tx, id)
	}); err != nil {
		return err
	}

	// 延时双删：延迟 500ms 再删一次
	delayedDeleteCaches(cacheTypes)
	return nil
}
