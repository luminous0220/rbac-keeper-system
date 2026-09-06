package log

import (
	"time"

	"gorm.io/gorm"

	"rbackeeper/global"
	"rbackeeper/model"
)

// 导出行数上限，防 OOM
const logExportLimit = 100000

// InsertLoginLog 写入一条登录日志
func InsertLoginLog(l *model.LoginLog) error {
	return global.DB.Create(l).Error
}

// InsertOperationLog 写入一条操作日志（消费者落库 / 直接调用）
func InsertOperationLog(l *model.OperationLog) error {
	return global.DB.Create(l).Error
}

// applyLoginLogFilter 组装登录日志筛选条件（列表 / 导出共用）
func applyLoginLogFilter(tx *gorm.DB, username, ip string, status *int8, start, end time.Time) *gorm.DB {
	if username != "" {
		tx = tx.Where("username LIKE ?", "%"+username+"%")
	}
	if ip != "" {
		tx = tx.Where("ip = ?", ip)
	}
	if status != nil {
		tx = tx.Where("status = ?", *status)
	}
	if !start.IsZero() {
		tx = tx.Where("login_time >= ?", start)
	}
	if !end.IsZero() {
		tx = tx.Where("login_time <= ?", end)
	}
	return tx
}

// applyOperationLogFilter 组装操作日志筛选条件（列表 / 导出共用）
func applyOperationLogFilter(tx *gorm.DB, title, operName string, status *int8, start, end time.Time) *gorm.DB {
	if title != "" {
		tx = tx.Where("title LIKE ?", "%"+title+"%")
	}
	if operName != "" {
		tx = tx.Where("oper_name LIKE ?", "%"+operName+"%")
	}
	if status != nil {
		tx = tx.Where("status = ?", *status)
	}
	if !start.IsZero() {
		tx = tx.Where("oper_time >= ?", start)
	}
	if !end.IsZero() {
		tx = tx.Where("oper_time <= ?", end)
	}
	return tx
}

// GetLoginLogList 登录日志分页（按 username/ip/status/时间范围）
func GetLoginLogList(q LoginLogListQuery) ([]model.LoginLog, int64, error) {
	tx := applyLoginLogFilter(global.DB.Model(&model.LoginLog{}), q.Username, q.IP, q.Status, q.StartTime, q.EndTime)
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.LoginLog
	err := tx.Order("login_time DESC").
		Offset((q.PageNum - 1) * q.PageSize).Limit(q.PageSize).
		Find(&rows).Error
	return rows, total, err
}

// GetOperationLogList 操作日志分页
func GetOperationLogList(q OperationLogListQuery) ([]model.OperationLog, int64, error) {
	tx := applyOperationLogFilter(global.DB.Model(&model.OperationLog{}), q.Title, q.OperName, q.Status, q.StartTime, q.EndTime)
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.OperationLog
	err := tx.Order("oper_time DESC").
		Offset((q.PageNum - 1) * q.PageSize).Limit(q.PageSize).
		Find(&rows).Error
	return rows, total, err
}

// LoginLogExportRows 登录日志导出（筛选同列表，按 login_time 升序，上限防 OOM）
func LoginLogExportRows(q LoginLogExportQuery) ([]model.LoginLog, error) {
	var rows []model.LoginLog
	tx := applyLoginLogFilter(global.DB.Model(&model.LoginLog{}), q.Username, q.IP, q.Status, q.StartTime, q.EndTime)
	err := tx.Order("login_time ASC").Limit(logExportLimit).Find(&rows).Error
	return rows, err
}

// OperationLogExportRows 操作日志导出（筛选同列表，按 oper_time 升序）
func OperationLogExportRows(q OperationLogExportQuery) ([]model.OperationLog, error) {
	var rows []model.OperationLog
	tx := applyOperationLogFilter(global.DB.Model(&model.OperationLog{}), q.Title, q.OperName, q.Status, q.StartTime, q.EndTime)
	err := tx.Order("oper_time ASC").Limit(logExportLimit).Find(&rows).Error
	return rows, err
}

// ClearLoginLogs 物理清空登录日志（含历史软删行；模型内嵌软删除，故用 Unscoped 走 DELETE），返回删除行数
func ClearLoginLogs() (int64, error) {
	res := global.DB.Unscoped().Where("1 = 1").Delete(&model.LoginLog{})
	return res.RowsAffected, res.Error
}

// ClearOperationLogs 物理清空操作日志
func ClearOperationLogs() (int64, error) {
	res := global.DB.Unscoped().Where("1 = 1").Delete(&model.OperationLog{})
	return res.RowsAffected, res.Error
}
