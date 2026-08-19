package utils

import "github.com/google/uuid"

// NewUUID 生成 UUID，用于 JWT jti / 防重放 requestId
func NewUUID() string {
	return uuid.NewString()
}
