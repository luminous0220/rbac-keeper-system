package initialize

import (
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"

	"rbackeeper/config"
	"rbackeeper/global"
)

// RegisterRedis 初始化 Redis 客户端并写入全局单例
// 启动时 ping 验证连通性，连接池大小按配置设定
func RegisterRedis(cfg config.CFG) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
		PoolSize: cfg.Redis.PoolSize,
	})

	ctx, cancel := contextWithTimeout(3 * time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		log.Fatal().Err(err).Msg("Redis ping 失败")
	}

	global.Redis = client
	log.Info().Str("module", "redis").Msg("注册 Redis 成功")
}
