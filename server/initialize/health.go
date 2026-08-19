package initialize

import (
	"time"

	"rbackeeper/global"
)

// ComponentHealth 单个依赖组件的健康状态
type ComponentHealth struct {
	Name  string `json:"name"`
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

// HealthCheck 依次探活 MySQL/Redis/MinIO，返回各组件状态与总体是否健康
// 供后续 P2 阶段的 /healthz、/readyz 接口调用
func HealthCheck() ([]ComponentHealth, bool) {
	var comps []ComponentHealth
	healthy := true

	// MySQL
	if global.DB == nil {
		comps = append(comps, ComponentHealth{Name: "mysql", OK: false, Error: "not initialized"})
		healthy = false
	} else if sqlDB, err := global.DB.DB(); err != nil {
		comps = append(comps, ComponentHealth{Name: "mysql", OK: false, Error: err.Error()})
		healthy = false
	} else {
		ctx, cancel := contextWithTimeout(2 * time.Second)
		err := sqlDB.PingContext(ctx)
		cancel()
		if err != nil {
			comps = append(comps, ComponentHealth{Name: "mysql", OK: false, Error: err.Error()})
			healthy = false
		} else {
			comps = append(comps, ComponentHealth{Name: "mysql", OK: true})
		}
	}

	// Redis
	if global.Redis == nil {
		comps = append(comps, ComponentHealth{Name: "redis", OK: false, Error: "not initialized"})
		healthy = false
	} else {
		ctx, cancel := contextWithTimeout(2 * time.Second)
		err := global.Redis.Ping(ctx).Err()
		cancel()
		if err != nil {
			comps = append(comps, ComponentHealth{Name: "redis", OK: false, Error: err.Error()})
			healthy = false
		} else {
			comps = append(comps, ComponentHealth{Name: "redis", OK: true})
		}
	}

	// MinIO
	if global.MinIO == nil {
		comps = append(comps, ComponentHealth{Name: "minio", OK: false, Error: "not initialized"})
		healthy = false
	} else {
		ctx, cancel := contextWithTimeout(2 * time.Second)
		_, err := global.MinIO.ListBuckets(ctx)
		cancel()
		if err != nil {
			comps = append(comps, ComponentHealth{Name: "minio", OK: false, Error: err.Error()})
			healthy = false
		} else {
			comps = append(comps, ComponentHealth{Name: "minio", OK: true})
		}
	}

	return comps, healthy
}
