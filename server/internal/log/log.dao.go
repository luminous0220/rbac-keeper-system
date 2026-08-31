package log

import (
	"time"

	"rbackeeper/global"
	"rbackeeper/model"
)

// InsertLoginLog 写入一条登录日志
func InsertLoginLog(l *model.LoginLog) error {
	return global.DB.Create(l).Error
}

// InsertOperationLog 写入一条操作日志（消费者落库 / 直接调用）
func InsertOperationLog(l *model.OperationLog) error {
	return global.DB.Create(l).Error
}

// GetLoginLogList 登录日志分页（按 username/ip/status/时间范围）
func GetLoginLogList(q LoginLogListQuery) ([]model.LoginLog, int64, error) {
	tx := global.DB.Model(&model.LoginLog{})
	if q.Username != "" {
		tx = tx.Where("username LIKE ?", "%"+q.Username+"%")
	}
	if q.IP != "" {
		tx = tx.Where("ip = ?", q.IP)
	}
	if q.Status != nil {
		tx = tx.Where("status = ?", *q.Status)
	}
	if !q.StartTime.IsZero() {
		tx = tx.Where("login_time >= ?", q.StartTime)
	}
	if !q.EndTime.IsZero() {
		tx = tx.Where("login_time <= ?", q.EndTime)
	}
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
	tx := global.DB.Model(&model.OperationLog{})
	if q.Title != "" {
		tx = tx.Where("title LIKE ?", "%"+q.Title+"%")
	}
	if q.OperName != "" {
		tx = tx.Where("oper_name LIKE ?", "%"+q.OperName+"%")
	}
	if q.Status != nil {
		tx = tx.Where("status = ?", *q.Status)
	}
	if !q.StartTime.IsZero() {
		tx = tx.Where("oper_time >= ?", q.StartTime)
	}
	if !q.EndTime.IsZero() {
		tx = tx.Where("oper_time <= ?", q.EndTime)
	}
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

// StreamLoginLogs 流式拉取登录日志（导出用，按时间升序）
func StreamLoginLogs(start, end time.Time) ([]model.LoginLog, error) {
	var rows []model.LoginLog
	tx := global.DB.Model(&model.LoginLog{})
	if !start.IsZero() {
		tx = tx.Where("login_time >= ?", start)
	}
	if !end.IsZero() {
		tx = tx.Where("login_time <= ?", end)
	}
	err := tx.Order("login_time ASC").Find(&rows).Error
	return rows, err
}

// StreamOperationLogs 流式拉取操作日志（导出用）
func StreamOperationLogs(start, end time.Time) ([]model.OperationLog, error) {
	var rows []model.OperationLog
	tx := global.DB.Model(&model.OperationLog{})
	if !start.IsZero() {
		tx = tx.Where("oper_time >= ?", start)
	}
	if !end.IsZero() {
		tx = tx.Where("oper_time <= ?", end)
	}
	err := tx.Order("oper_time ASC").Find(&rows).Error
	return rows, err
}
