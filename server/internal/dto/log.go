package dto

import "time"

// LoginLogListQuery 登录日志查询
type LoginLogListQuery struct {
	Page
	// 用户名，模糊匹配
	Username string `form:"username"`
	// 登录 IP，精确匹配
	IP string `form:"ip"`
	// 登录状态（0 成功 / 1 失败）
	Status *int8 `form:"status" binding:"omitempty,oneof=0 1"`
	// 查询起始时间
	StartTime time.Time `form:"startTime" time_format:"2006-01-02 15:04:05"`
	// 查询结束时间
	EndTime time.Time `form:"endTime" time_format:"2006-01-02 15:04:05"`
}

// LoginLogListResp 登录日志分页响应
type LoginLogListResp = PageResp[LoginLogItem]

// LoginLogItem 登录日志行
type LoginLogItem struct {
	// 日志 ID
	ID uint `json:"id"`
	// 登录用户名
	Username string `json:"username"`
	// 登录 IP
	IP string `json:"ip"`
	// IP 归属地
	IPAddr string `json:"ipAddr"`
	// 浏览器标识
	Browser string `json:"browser"`
	// 操作系统
	OS string `json:"os"`
	// 登录状态（0 成功 / 1 失败）
	Status int8 `json:"status"`
	// 提示消息
	Msg string `json:"msg"`
	// 登录时间
	LoginTime time.Time `json:"loginTime"`
}

// OperationLogListQuery 操作日志查询
type OperationLogListQuery struct {
	Page
	// 模块标题，模糊匹配
	Title string `form:"title"`
	// 操作人员，模糊匹配
	OperName string `form:"operName"`
	// 操作状态（0 正常 / 1 异常）
	Status *int8 `form:"status" binding:"omitempty,oneof=0 1"`
	// 查询起始时间
	StartTime time.Time `form:"startTime" time_format:"2006-01-02 15:04:05"`
	// 查询结束时间
	EndTime time.Time `form:"endTime" time_format:"2006-01-02 15:04:05"`
}

// OperationLogListResp 操作日志分页响应
type OperationLogListResp = PageResp[OperationLogItem]

// OperationLogItem 操作日志行
type OperationLogItem struct {
	// 日志 ID
	ID uint `json:"id"`
	// 模块标题
	Title string `json:"title"`
	// 请求方式（GET / POST 等）
	RequestMethod string `json:"requestMethod"`
	// 操作人员
	OperName string `json:"operName"`
	// 请求 URL
	OperURL string `json:"operUrl"`
	// 请求方 IP
	OperIP string `json:"operIp"`
	// 操作状态（0 正常 / 1 异常）
	Status int8 `json:"status"`
	// 错误消息（操作异常时记录）
	ErrorMsg string `json:"errorMsg"`
	// 操作时间
	OperTime time.Time `json:"operTime"`
}

// LogExportQuery 导出查询
type LogExportQuery struct {
	// 导出类型（login 登录日志 / operation 操作日志）
	Type string `form:"type" binding:"required,oneof=login operation"`
}
