package middleware

import (
	"bytes"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// bodyWriter 自定义 ResponseWriter，用于捕获响应体
type bodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// 记录每个请求的方法、路径、状态码、耗时与客户端 IP
func RequestLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// 1. 替换 ResponseWriter，用于捕获响应体
		blw := &bodyWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBuffer([]byte{}),
		}
		c.Writer = blw

		c.Next()

		// 2. 获取响应状态码和响应体
		status := c.Writer.Status()
		respBody := blw.body.String()

		// 3. 根据状态码区分日志级别和内容
		event := log.Info()
		if status >= 400 {
			event = log.Error()
		}

		event.
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Int("status", status).
			Float64("latency_ms", float64(time.Since(start).Milliseconds())).
			Str("ip", c.ClientIP())

		// 4. 仅在错误响应时记录响应体（避免 info 日志膨胀），上传文件接口不记录body
		if status >= 400 && !strings.Contains(c.GetHeader("Content-Type"), "multipart/form-data") {
			event.Str("response_body", respBody)
		}

		event.Msg("request")
	}
}
