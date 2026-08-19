package utils

import (
	"errors"
	"rbackeeper/global"
	"rbackeeper/internal/constants"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Claims 自定义声明
type Claims struct {
	UserID   uint                `json:"user_id"`
	Username string              `json:"username"`
	RoleKey  string              `json:"role_key"`
	Type     constants.TokenType `json:"type"`
	jwt.RegisteredClaims
}

// Generate 签发指定类型令牌，返回 token 字符串与其 jti。
// 配置直接读取 global.CFG.JWT，与项目全局单例模式保持一致，无需手动初始化。
func Generate(t constants.TokenType, userID uint, username, roleKey string) (token, jti string, err error) {
	cfg := global.CFG.JWT
	jti = NewUUID()
	now := time.Now()
	expire := cfg.AccessExpire
	if t == constants.RefreshToken {
		expire = cfg.RefreshExpire
	}
	claims := Claims{
		UserID:   userID,
		Username: username,
		RoleKey:  roleKey,
		Type:     t,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			Issuer:    cfg.Issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(expire)),
		},
	}
	token, err = jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(cfg.Secret))
	return
}

// Parse 解析并校验令牌签名与有效期
func Parse(token string) (*Claims, error) {
	claims := &Claims{}
	parsed, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("非法签名算法")
		}
		return []byte(global.CFG.JWT.Secret), nil
	})
	if err != nil {
		return nil, err
	}
	if !parsed.Valid {
		return nil, errors.New("令牌无效")
	}
	return claims, nil
}

// GetClaims 从上下文获取认证声明
func GetClaims(c *gin.Context) *Claims {
	v, ok := c.Get("claims")
	if !ok {
		return nil
	}
	claims, _ := v.(*Claims)
	return claims
}
