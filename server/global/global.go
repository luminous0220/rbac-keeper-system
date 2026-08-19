package global

import (
	"context"
	"sync"

	"rbackeeper/config"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/minio/minio-go/v7"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var (
	CFG   *config.CFG
	DB    *gorm.DB
	Redis *redis.Client
	MinIO *minio.Client

	// RocketMQ 生产者全局单例，业务侧（如操作日志入队）直接 SendSync 发送。
	// 注意：使用前需先 initialize.RegisterRocketMQ 完成 Start()。
	RocketMQ rocketmq.Producer
)

// 顶层 context：用于在退出时统一取消后台 goroutine（MQ 消费者等）。
var (
	AppCtx    context.Context
	AppCancel context.CancelFunc
	once      sync.Once
)

// InitAppContext 幂等初始化顶层 context（main 启动时调用一次）
func InitAppContext() (context.Context, context.CancelFunc) {
	var cancel context.CancelFunc
	once.Do(func() {
		AppCtx, AppCancel = context.WithCancel(context.Background())
		cancel = AppCancel
	})
	if AppCancel == nil {
		AppCtx, AppCancel = context.WithCancel(context.Background())
		cancel = AppCancel
	}
	return AppCtx, cancel
}
