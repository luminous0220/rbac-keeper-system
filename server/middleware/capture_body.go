package middleware

import (
	"bytes"
	"io"

	"github.com/gin-gonic/gin"
)

// 缓存请求体到 gin context（键 requestBody），供操作日志落库前取原始参数做脱敏。
// 读取后把 body 流原样写回，不影响后续 BindJSON。GET 等无 body 请求为空，跳过。
func CaptureBody() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Body == nil {
			c.Next()
			return
		}
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.Next()
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
		c.Set("requestBody", body)
		c.Next()
	}
}
