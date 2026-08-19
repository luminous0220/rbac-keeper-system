package dto

// 部门模块 DTO：树形配置数据，节点需暴露 id 供前端定位父子与编辑/删除。

// DeptTreeQuery 部门树查询
type DeptTreeQuery struct {
	// 模糊匹配（返回匹配节点及其祖先链，保证树完整）
	DeptName string `form:"deptName"`
	// 状态
	Status *int8 `form:"status" binding:"omitempty,oneof=0 1"`
}

// parentId=0 为根
type DeptAddBody struct {
	// 父部门ID，默认0表示根节点，
	ParentID uint `json:"parentId"`
	// 部门名称
	DeptName string `json:"deptName" binding:"required,min=2,max=30"`
	// 显示排序，默认0
	OrderNum int `json:"orderNum"`
	// 负责人
	Leader string `json:"leader" binding:"omitempty,max=20"`
	// 联系电话
	Phone string `json:"phone" binding:"omitempty,max=11"`
	// 邮箱
	Email string `json:"email" binding:"omitempty,max=50"`
	// 备注
	Remark string `json:"remark" binding:"omitempty,max=500"`
}

// 修改部门
type DeptEditBody struct {
	// 父部门ID
	ParentID *uint `json:"parentId"`
	// 部门名称
	DeptName *string `json:"deptName" binding:"omitempty,min=2,max=30"`
	// 显示排序
	OrderNum *int `json:"orderNum"`
	// 负责人
	Leader *string `json:"leader" binding:"omitempty,max=20"`
	// 联系电话
	Phone *string `json:"phone" binding:"omitempty,max=11"`
	// 邮箱
	Email *string `json:"email" binding:"omitempty,max=50"`
	// 状态：0-正常，1-停用
	Status *int8 `json:"status" binding:"omitempty,oneof=0 1"`
	// 备注
	Remark *string `json:"remark" binding:"omitempty,max=500"`
}
