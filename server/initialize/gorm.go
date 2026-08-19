package initialize

import (
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"rbackeeper/config"
	"rbackeeper/global"
)

// RegisterMySQL 初始化 MySQL 连接池并写入全局单例
// 配置连接池上限/空闲/最大存活时间，启动时 ping 验证连通性
func RegisterMySQL(cfg config.CFG) {
	db, err := gorm.Open(mysql.Open(cfg.MySQL.DSN), &gorm.Config{
		// 禁用 AutoMigrate 自动创建物理外键。
		DisableForeignKeyConstraintWhenMigrating: true,
		Logger: logger.New(newGormWriter(cfg.MySQL), logger.Config{
			SlowThreshold: 200 * time.Millisecond,
			LogLevel:      logger.LogLevel(cfg.MySQL.LogLevel),
		}),
	})
	if err != nil {
		log.Fatal().Err(err).Msg("mysql gorm 初始化失败")
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal().Err(err).Msg("获取底层 sql.DB 失败")
	}
	if cfg.MySQL.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(cfg.MySQL.MaxOpenConns)
	}
	if cfg.MySQL.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(cfg.MySQL.MaxIdleConns)
	}
	if cfg.MySQL.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(cfg.MySQL.ConnMaxLifetime)
	}

	ctx, cancel := contextWithTimeout(3 * time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		log.Fatal().Err(err).Msg("MySQL ping 失败")
	}

	global.DB = db

	log.Info().Str("module", "mysql").Msg("注册 MySQL 成功")
}
