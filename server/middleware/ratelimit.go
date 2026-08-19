package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"rbackeeper/response"
)

// 令牌桶限流中间件（单机内存），防止恶意用户或突发流量将服务器压垮
// key 默认按「IP+路径」分桶；burst / rate 由调用方注入。生产多副本部署建议改用 Redis 分布式计数，
// 当前实现以「单机粒度 + 短时间窗」作为快速防线。
type rateBucket struct {
	lim  *rate.Limiter // Go 标准库提供的令牌桶限流器
	seen time.Time     // 记录该桶最后一次被访问的时间
}

var (
	bucketMu sync.Mutex                     // 互斥锁，保护并发安全
	buckets  = make(map[string]*rateBucket) // 全局的桶映射表
	// 使用 sync.Once 保证清理协程只启动一次
	cleanupOnce sync.Once
)

// 定期清理冷桶，防止长期运行内存膨胀。
func startCleanup() {
	cleanupOnce.Do(func() {
		t := time.NewTicker(5 * time.Minute)
		go func() {
			for range t.C {
				bucketMu.Lock()
				cutoff := time.Now().Add(-30 * time.Minute)
				for k, b := range buckets {
					if b.seen.Before(cutoff) {
						delete(buckets, k)
					}
				}
				bucketMu.Unlock()
			}
		}()
	})
}

// 构造令牌桶限流
// rps：每秒往桶里放多少个令牌
// burst：桶的最大容量。允许瞬间的并发峰值
func RateLimit(rps float64, burst int) gin.HandlerFunc {

	// 首次调用时，自动触发清理协程的初始化
	startCleanup()

	if rps <= 0 {
		rps = 50
	}
	if burst <= 0 {
		burst = 100
	}
	return func(c *gin.Context) {

		// 生成限流 Key，使用 IP + 请求路径 作为 Key：同一个 IP 访问 /api/login 和 /api/user 会各自独立计算限流，互不影响
		key := c.ClientIP() + "|" + c.FullPath()

		// 获取或创建限流桶
		bucketMu.Lock()
		b, ok := buckets[key]
		if !ok {
			b = &rateBucket{lim: rate.NewLimiter(rate.Limit(rps), burst)}
			buckets[key] = b
		}
		b.seen = time.Now()
		bucketMu.Unlock()

		if !b.lim.Allow() {
			c.Header("Retry-After", "1")
			response.Fail(c, response.ErrTooManyRequests, "请求过于频繁，请稍后再试")
			c.AbortWithStatus(http.StatusTooManyRequests)
			return
		}
		c.Next()
	}
}
