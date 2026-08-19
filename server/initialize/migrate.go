package initialize

import (
	"github.com/rs/zerolog/log"

	"rbackeeper/global"
	"rbackeeper/internal/model"
)

// RegisterMigrate 自动建表/同步表结构。开发环境使用，生产建议改用迁移脚本
func RegisterMigrate() {
	if global.DB == nil {
		log.Fatal().Msg("数据库未初始化")
		return

	}
	// AutoMigrate 直接返回 error
	if err := global.DB.AutoMigrate(
		&model.User{}, &model.Role{}, &model.Menu{}, &model.Dept{},
		&model.DictType{}, &model.DictData{},
		&model.LoginLog{}, &model.OperationLog{},
	); err != nil {
		log.Fatal().Err(err).Msg("AutoMigrate 失败")
		return
	}
	log.Info().Str("module", "migrate").Msg("数据库迁移完成")
}
