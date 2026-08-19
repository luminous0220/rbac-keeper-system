package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"rbackeeper/response"
)

// Recovery 捕获 panic，记录堆栈并返回 500，避免单次请求崩溃拖垮进程
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Error().Interface("error", r).Bytes("stack", debug.Stack()).Msg("panic recovered")
				c.AbortWithStatusJSON(http.StatusInternalServerError, response.Result[any]{
					Code: response.ErrServer,
					Msg:  "服务器内部错误",
				})
			}
		}()
		c.Next()
	}
}
