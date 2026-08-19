package initialize

import (
	"fmt"
	"rbackeeper/config"

	"github.com/rs/zerolog/log"
)

// gormWriter 实现 gorm logger.Writer，按配置输出到控制台与 zap 日志
type gormWriter struct {
	config config.MySQLConfig
}

func newGormWriter(c config.MySQLConfig) *gormWriter {
	return &gormWriter{config: c}
}

// Printf 格式化打印 SQL 日志
func (g *gormWriter) Printf(msg string, data ...any) {
	logMsg := fmt.Sprintf(msg, data...)

	// 使用 zerolog 的 Log() 方法，它支持链式调用且保证顺序
	switch g.config.LogLevel {
	case 0:
		log.Debug().Msg("============== SQL START ==============")
		log.Debug().Msg(logMsg)
		log.Debug().Msg("============== SQL END ===============")
	case 1:
		log.Error().Msg("============== SQL START ==============")
		log.Error().Msg(logMsg)
		log.Error().Msg("============== SQL END ===============")
	case 2:
		log.Warn().Msg("============== SQL START ==============")
		log.Warn().Msg(logMsg)
		log.Warn().Msg("============== SQL END ===============")
	default:
		log.Info().Msg("============== SQL START ==============")
		log.Info().Msg(logMsg)
		log.Info().Msg("============== SQL END ===============")
	}
}
