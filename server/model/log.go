package model

import "time"

// LoginLog 登录日志模型
type LoginLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"size:50;index;comment:用户名" json:"username"`
	IP        string    `gorm:"size:128;comment:登录IP" json:"ip"`
	IPAddr    string    `gorm:"size:255;comment:登录地点" json:"ip_addr"`
	Browser   string    `gorm:"size:50;comment:浏览器" json:"browser"`
	OS        string    `gorm:"size:50;comment:操作系统" json:"os"`
	Status    int8      `gorm:"comment:登录状态 0成功 1失败" json:"status"`
	Msg       string    `gorm:"size:255;comment:提示消息" json:"msg"`
	LoginTime time.Time `gorm:"index;comment:登录时间" json:"login_time"`
	BaseModel
}

// OperationLog 操作日志模型
type OperationLog struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	Title         string    `gorm:"size:50;comment:模块标题" json:"title"`
	RequestMethod string    `gorm:"size:10;comment:请求方式" json:"request_method"`
	OperatorType  int8      `gorm:"comment:操作类别 0其它 1后台用户 2手机端用户" json:"operator_type"`
	OperName      string    `gorm:"size:50;comment:操作人员" json:"oper_name"`
	OperURL       string    `gorm:"size:255;comment:请求URL" json:"oper_url"`
	OperIP        string    `gorm:"size:128;comment:主机地址" json:"oper_ip"`
	OperParam     string    `gorm:"type:text;comment:请求参数" json:"oper_param"`
	Status        int8      `gorm:"default:0;comment:操作状态 0正常 1异常" json:"status"`
	ErrorMsg      string    `gorm:"size:2000;comment:错误消息" json:"error_msg"`
	OperTime      time.Time `gorm:"index;comment:操作时间" json:"oper_time"`
	BaseModel
}
