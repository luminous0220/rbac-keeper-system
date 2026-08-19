package middleware

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"rbackeeper/global"
	"rbackeeper/response"
)

// 幂等中间件：基于 X-Request-Id + Redis SETNX。
// 同一 requestId 在 TTL 内只放行一次：第一次进入时 SETNX 成功、记入进行中标记；请求结束后 DEL 让重试可进入
// （避免「接口异常 + 重试」被永久拒绝）；若需「提交后才允许重试」的更强语义，可改为 EXPIRE 后才释放。
// 仅作用于写方法（POST/PUT/DELETE/PATCH）。
func Idempotent(ttl time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 仅写方法生效
		switch c.Request.Method {
		case "POST", "PUT", "DELETE", "PATCH":
		default:
			c.Next()
			return
		}

		reqID := c.GetHeader("X-Request-Id")
		if reqID == "" {
			reqID = uuid.NewString()
			c.Request.Header.Set("X-Request-Id", reqID)
		}
		c.Writer.Header().Set("X-Request-Id", reqID)
		c.Set("requestId", reqID)

		if global.Redis == nil {
			// 未连 Redis：放行（不阻塞业务）
			c.Next()
			return
		}
		key := "idem:" + reqID
		ctx, cancel := context.WithTimeout(c.Request.Context(), 200*time.Millisecond)
		defer cancel()
		ok, err := global.Redis.SetNX(ctx, key, "1", ttl).Result()
		if err != nil {
			// Redis 异常：放行 + 记录（不阻断业务）
			c.Next()
			return
		}
		if !ok {
			response.Fail(c, response.ErrIdempotent, "请勿重复提交")
			c.Abort()
			return
		}
		c.Next()

		// 请求结束释放幂等 key，使同一 requestId 的重试可再次进入
		// （避免「接口异常 + 重试」被 TTL 内永久拒绝）。用新 context：
		// 业务处理可能已超过初始化时 200ms 超时的 ctx。
		releaseCtx, releaseCancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer releaseCancel()
		if err := global.Redis.Del(releaseCtx, key).Err(); err != nil {
			log.Warn().Err(err).Str("key", key).Msg("幂等 key 释放失败")
		}
	}
}
