package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"rbackeeper/internal/constants"
	"rbackeeper/internal/service"
	"rbackeeper/response"
	"rbackeeper/utils"
)

// JWTAuth 认证中间件：解析 Access Token -> 黑名单校验 -> 密码修改时间校验 -> 注入身份
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			response.Fail(c, response.ErrUnauthorized, "缺少认证令牌")
			c.Abort()
			return
		}
		claims, err := utils.Parse(token)
		if err != nil {
			response.Fail(c, response.ErrTokenInvalid, "认证令牌无效或已过期")
			c.Abort()
			return
		}
		if claims.Type != constants.AccessToken {
			response.Fail(c, response.ErrTokenInvalid, "令牌类型错误")
			c.Abort()
			return
		}
		if revoked, msg := service.IsTokenRevoked(c.Request.Context(), claims); revoked {
			response.Fail(c, response.ErrTokenInvalid, msg)
			c.Abort()
			return
		}
		c.Set("claims", claims)
		c.Next()
	}
}

// extractToken 从 Authorization 头提取 Bearer 令牌
func extractToken(c *gin.Context) string {
	auth := c.GetHeader("Authorization")
	if len(auth) > 7 && strings.HasPrefix(auth, "Bearer ") {
		return auth[7:]
	}
	return ""
}
