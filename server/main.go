package main

import (
	"context"
	"rbackeeper/config"
	"rbackeeper/global"
	"rbackeeper/initialize"
	"rbackeeper/internal/log"
	"rbackeeper/router"
	"rbackeeper/utils/casbinx"

	zerolog "github.com/rs/zerolog/log"
)

// @title           RBAC Keeper System API
// @version         1.0
// @description     基于 RBAC 的权限管理系统后端
// @host            localhost:8080
// @BasePath        /api
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description 输入 Bearer {access_token}
func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Test()
	// 加载环境
	cfg := config.LoadConfig(".")
	global.CFG = &cfg

	// 注册 zerolog
	initialize.RegisterLogger(*global.CFG)

	// 注册基础设施：MySQL / Redis / RocketMQ / MinIO
	initialize.RegisterMySQL(*global.CFG)
	initialize.RegisterRedis(*global.CFG)
	initialize.RegisterRocketMQ(*global.CFG)

	// if err := initialize.RegisterMinIO(*global.CFG); err != nil {
	// 	log.Fatal().Err(err).Msg("注册 MinIO 失败：")
	// }

	// 数据库迁移（AutoMigrate）
	initialize.RegisterMigrate()
	// 内置数据初始化（失败仅告警不阻断启动）
	initialize.RegisterSeed()

	// 策略引擎初始化 + 首次全量加载（依赖 seed 后的 role_menus / user_roles）
	casbinx.Init()
	if err := casbinx.ReloadEnforcerPolicies(); err != nil {
		zerolog.Warn().Err(err).Msg("Casbin 策略首次加载失败")
	}

	// 先注册各业务的消费者
	log.InitOperationLog()

	// 启动前确保所有 topic 已存在（不存在则自动创建，避免 Consumer 启动报 topic not exist）
	initialize.EnsureConsumerTopics(ctx, initialize.Topics()...)
	// 启动所有消费者
	initialize.StartAllConsumer(ctx)

	// 启动健康检查自检
	if comps, ok := initialize.HealthCheck(); ok {
		zerolog.Info().Int("components", len(comps)).Msg("rbackeeper 启动完成，所有依赖健康")
	} else {
		zerolog.Warn().Int("components", len(comps)).Msg("rbackeeper 启动完成，但部分依赖异常")
	}

	// 启动 HTTP 服务（阻塞，支持优雅关闭）
	router.Run(*global.CFG, cancel)
}
