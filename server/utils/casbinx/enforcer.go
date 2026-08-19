package casbinx

import (
	"errors"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"rbackeeper/global"

	"github.com/casbin/casbin/v3"
	"github.com/casbin/casbin/v3/model"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/rs/zerolog/log"
)

// modelConfPath Casbin RBAC 模型文件（相对运行目录，与 main.go 同目录）。
var modelConfPath = "utils/casbinx/rbac_model.conf"

// Enforcer 全局 Casbin enforcer。nil 表示未初始化或初始化失败。
// 并发模型：Reload 走 reloadMu 串行；Enforce 自身有内部读锁保护（SyncedEnforcer）。
var (
	Enforcer *casbin.SyncedCachedEnforcer
	reloadMu sync.Mutex
	once     sync.Once
	initErr  error
)

// Init 创建全局 enforcer（幂等，仅成功时写入 Enforcer）。启动阶段显式调用一次。
// 前置条件：global.DB 已就绪（main 中 MySQL 初始化之后）。策略经 gorm-adapter 持久化到 casbin_rule 表。
func Init() {
	once.Do(func() {
		if global.DB == nil {
			log.Panic().Msg("数据库未初始化，无法挂载 casbin adapter")
			return
		}
		modelPath, _ := filepath.Abs(modelConfPath)

		m, err := model.NewModelFromFile(modelPath)
		if err != nil {
			log.Panic().Msg(fmt.Sprintf("casbin model 解析失败: %v", err))
			return
		}
		a, err := gormadapter.NewAdapterByDB(global.DB)
		if err != nil {
			log.Panic().Msg(fmt.Sprintf("casbin gorm adapter 创建失败: %v", err))
			return
		}
		e, err := casbin.NewSyncedCachedEnforcer(m, a)
		if err != nil {
			log.Panic().Msg(fmt.Sprintf("casbin enforcer 创建失败: %v", err))
			return
		}
		// 决策缓存 1 小时（v3 的 SetExpireTime 参数为 time.Duration；旧代码 60*60 被当作纳秒，缓存形同虚设）
		e.SetExpireTime(1 * time.Hour)
		Enforcer = e
	})
}

// GetCasbin 懒初始化并返回全局 enforcer（初始化失败时返回 nil，调用方需判空）。
func GetCasbin() *casbin.SyncedCachedEnforcer {
	Init()
	return Enforcer
}

// 全量替换策略：清空 casbin_rule（DB）与内存后，写入新的 p/g 规则。
// 保证已撤销的角色/权限即时生效（旧实现只 AddPolicies 不清空，撤销永不到达且表数据残留）。
// 流程：adapter 清空 p/g → LoadPolicy 重建空内存（同时清决策缓存）→ AddPolicies → AddGroupingPolicies。
func Reload(policies, groupings [][]string) error {
	e := GetCasbin()
	if e == nil {
		return errors.New("casbin enforcer 未初始化")
	}

	reloadMu.Lock()
	defer reloadMu.Unlock()

	// 1. 清空 DB 中的全部 p/g 规则（casbin_rule 表按 ptype 删除，避免陈旧策略残留）
	adapter := e.GetAdapter()

	if err := adapter.RemoveFilteredPolicy("p", "p", -1); err != nil {
		return fmt.Errorf("清空 casbin p 规则失败: %w", err)
	}
	if err := adapter.RemoveFilteredPolicy("g", "g", -1); err != nil {
		return fmt.Errorf("清空 casbin g 规则失败: %w", err)
	}

	// 2. 从空表重载：内存清空 + 决策缓存清空（SyncedCachedEnforcer.LoadPolicy 两者都做）
	if err := e.LoadPolicy(); err != nil {
		return fmt.Errorf("casbin 策略重载失败: %w", err)
	}

	// 3. 全量写入（落 DB + 内存；g 关系图由 AddGroupingPolicies 增量构建）
	if len(policies) > 0 {
		if _, err := e.AddPolicies(policies); err != nil {
			return fmt.Errorf("casbin AddPolicies 失败: %w", err)
		}
	}
	if len(groupings) > 0 {
		if _, err := e.AddGroupingPolicies(groupings); err != nil {
			return fmt.Errorf("casbin AddGroupingPolicies 失败: %w", err)
		}
	}
	log.Info().Int("p", len(policies)).Int("g", len(groupings)).Msg("Casbin 策略已重载")
	return nil
}
