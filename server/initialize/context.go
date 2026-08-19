package initialize

import (
	"context"
	"time"
)

// contextWithTimeout 统一构造带超时的 context，避免各初始化器重复样板代码
func contextWithTimeout(d time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), d)
}
