package model

import "time"

// User 用户模型
type User struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	Username  string     `gorm:"size:30;uniqueIndex;not null;comment:用户名" json:"username"`
	Password  string     `gorm:"size:128;not null;comment:密码哈希" json:"-"` // json:"-" 永不外泄密码
	Nickname  string     `gorm:"size:30;comment:昵称" json:"nickname"`
	Phone     string     `gorm:"size:11;comment:手机号" json:"phone"`
	Email     string     `gorm:"size:50;comment:邮箱" json:"email"`
	Sex       int8       `gorm:"default:0;comment:性别 0未知 1男 2女" json:"sex"`
	Avatar    string     `gorm:"size:255;comment:头像" json:"avatar"`
	Status    int8       `gorm:"default:1;comment:状态 0禁用 1正常" json:"status"`
	LoginIP   string     `gorm:"size:128;comment:最后登录IP" json:"login_ip"`
	LoginDate *time.Time `gorm:"comment:最后登录时间" json:"login_date"` // 指针：未登录时为 NULL，避免零值时间被 MySQL NO_ZERO_DATE 拒绝

	// 关联关系：GORM 自动管理中间表
	DeptID uint `gorm:"comment:部门ID" json:"dept_id"`
	// 行政归属
	Dept  *Dept   `gorm:"foreignKey:DeptID" json:"dept,omitempty"`
	Roles []*Role `gorm:"many2many:user_roles;" json:"roles,omitempty"` // 无约束中间表

	Remark string `gorm:"size:500;comment:备注" json:"remark"`

	BaseModel
}
