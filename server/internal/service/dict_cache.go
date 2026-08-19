package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/rs/zerolog/log"

	"rbackeeper/global"
	"rbackeeper/internal/dao"
	"rbackeeper/internal/model"
)

const (
	// dictCachePrefix 字典数据缓存键前缀，完整键为 sys_dict:{dictType}
	dictCachePrefix = "sys_dict:"
	// dictCacheExpire 缓存有效期，作为延时双删失败时的兜底失效手段
	dictCacheExpire = 30 * time.Minute
	// dictDoubleDeleteDelay 延时双删二次删除延迟，覆盖「DB 提交前回源、提交后回填」的并发旧值窗口
	dictDoubleDeleteDelay = 500 * time.Millisecond
)

// dictCacheKey 拼接字典缓存键
func dictCacheKey(dictType string) string {
	return dictCachePrefix + dictType
}

// 按字典类型查询可用数据（status=0），缓存优先：
//  1. 命中缓存直接返回；
//  2. 未命中回源 DB 并回填缓存（回填失败仅告警，不影响主流程）。
//
// 供前端下拉/回显使用，写操作通过延时双删保证最终一致。
func DictDataByType(ctx context.Context, dictType string) ([]model.DictData, error) {
	key := dictCacheKey(dictType)

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
	list, err := dao.GetDictDataByType(dictType)
	if err != nil {
		return nil, err
	}

	// 3. 回填缓存
	if data, err := json.Marshal(list); err == nil {
		if err := global.Redis.Set(ctx, key, data, dictCacheExpire).Err(); err != nil {
			log.Warn().Err(err).Str("dict_type", dictType).Msg("回填字典缓存失败")
		}
	}
	return list, nil
}

// 立即删除指定类型集合的字典缓存（best-effort，失败仅告警）。
// 用于延时双删的第一次删除（写操作前）。使用独立 context，避免请求结束后 ctx 取消导致删除丢失。
func deleteDictCaches(dictTypes []string) {
	keys := collectCacheKeys(dictTypes)
	if len(keys) == 0 {
		return
	}
	if err := global.Redis.Del(context.Background(), keys...).Err(); err != nil {
		log.Warn().Err(err).Msg("删除字典缓存失败")
	}
}

// 延时双删的第二次删除：异步等待 500ms 后再删一次。
// 覆盖并发场景下「读请求在 DB 提交前回源旧值、提交后回填缓存」导致的脏缓存。
func delayedDeleteDictCaches(dictTypes []string) {
	keys := collectCacheKeys(dictTypes)
	if len(keys) == 0 {
		return
	}
	go func(keys []string) {
		time.Sleep(dictDoubleDeleteDelay)
		if err := global.Redis.Del(context.Background(), keys...).Err(); err != nil {
			log.Warn().Err(err).Msg("延时删除字典缓存失败")
		}
	}(keys)
}

// collectCacheKeys 过滤空编码并拼接缓存键
func collectCacheKeys(dictTypes []string) []string {
	keys := make([]string, 0, len(dictTypes))
	for _, t := range dictTypes {
		if t != "" {
			keys = append(keys, dictCacheKey(t))
		}
	}
	return keys
}
