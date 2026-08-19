package initialize

import (
	"fmt"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/rs/zerolog/log"

	"rbackeeper/config"
	"rbackeeper/global"
)

// RegisterMinIO 初始化 MinIO 客户端并写入全局单例
// 启动时探测 bucket 是否存在，不存在则自动创建
func RegisterMinIO(cfg config.CFG) error {
	client, err := minio.New(cfg.MinIO.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinIO.AccessKey, cfg.MinIO.SecretKey, ""),
		Secure: cfg.MinIO.UseSSL,
	})
	if err != nil {
		return fmt.Errorf("创建 MinIO 客户端失败: %w", err)
	}

	ctx, cancel := contextWithTimeout(3 * time.Second)
	defer cancel()
	exists, err := client.BucketExists(ctx, cfg.MinIO.Bucket)
	if err != nil {
		return fmt.Errorf("MinIO 探测失败: %w", err)
	}
	if !exists {
		if err := client.MakeBucket(ctx, cfg.MinIO.Bucket, minio.MakeBucketOptions{}); err != nil {
			return fmt.Errorf("创建 MinIO bucket 失败: %w", err)
		}
		log.Info().Str("module", "minio").Str("bucket", cfg.MinIO.Bucket).Msg("bucket 不存在，已自动创建")
	}

	global.MinIO = client
	log.Info().Str("module", "minio").Msg("注册 MinIO 成功")
	return nil
}
