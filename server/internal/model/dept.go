package model

// Dept 部门模型
type Dept struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	DeptName string `gorm:"size:30;not null;unique;comment:部门名称" json:"dept_name"`
	// 根节点传 0
	ParentID uint   `gorm:"default:0;comment:父部门ID" json:"parent_id"`
	OrderNum int    `gorm:"default:0;comment:显示顺序" json:"order_num"`
	Leader   string `gorm:"size:20;comment:负责人" json:"leader"`
	Phone    string `gorm:"size:11;comment:联系电话" json:"phone"`
	Email    string `gorm:"size:50;comment:邮箱" json:"email"`
	Status   int8   `gorm:"default:0;comment:状态 0正常 1停用" json:"status"`
	Remark   string `gorm:"size:500;comment:备注" json:"remark"`
	// 关联关系：自关联子部门
	Children []*Dept `gorm:"foreignKey:ParentID" json:"children"`
	BaseModel
}
