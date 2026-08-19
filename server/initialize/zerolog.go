package initialize

import (
	"fmt"
	"os"
	"path/filepath"
	"rbackeeper/config"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// initializeLogger 初始化全局 zerolog 日志器
// 保证每条日志稳定携带四要素：具体时间(本地)、错误级别、文件&行号、错误信息
// 控制台整行按级别着色输出，文件按日期 + 级别分文件写入 logs/<年>/<月-日>/<级别>.log
// config：应用配置，根据 Environment 区分开发/生产日志形态
func RegisterLogger(config config.CFG) {
	// 1. 全局时间序列化格式：毫秒级精度，便于精确排序与问题定位
	zerolog.TimeFieldFormat = "2006-01-02 15:04:05.000"
	// 显式指定本地时间函数，避免容器/云环境系统时区为 UTC 导致日志时间偏移
	zerolog.TimestampFunc = time.Now
	// 2. caller 格式化：仅保留 文件名:行号，去除冗长绝对路径，便于快速定位
	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		return fmt.Sprintf("%s:%d", filepath.Base(file), line)
	}

	// 3. 文件切割器：按日期 + 级别分文件，写入 logs/<年>/<月-日>/<级别>.log，跨天自动切换
	fileCutter := newFileCutter("logs", config.LogMaxAge)

	if config.Mode == "debug" {
		// 开发环境：彩色控制台(整行按级别着色) + 按日期级别分文件 JSON
		consoleWriter := newConsoleWriter(os.Stderr)
		// MultiLevelWriter 多路分发同一份 JSON：
		// coloredConsoleWriter 实现 WriteLevel -> 解析 JSON 整行按级别着色输出控制台
		// FileCutter 实现 WriteLevel -> 按级别路由到对应日期文件
		multi := zerolog.MultiLevelWriter(consoleWriter, fileCutter)
		log.Logger = zerolog.New(multi).With().Timestamp().Caller().Logger()
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		// 生产环境：JSON 输出到 stderr + 按日期级别分文件，便于 ELK/Loki 采集
		multi := zerolog.MultiLevelWriter(os.Stderr, fileCutter)
		log.Logger = zerolog.New(multi).With().Timestamp().Caller().Logger()
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	log.Info().Msg("注册 zerolog 成功!!")
}
