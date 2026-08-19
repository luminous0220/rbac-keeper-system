package dto

import (
	"rbackeeper/internal/model"
	"time"
)

// 用户列表查询
type UserListQuery struct {
	Page
	// 用户名(模糊)
	Username string `form:"username"`
	// 手机号(模糊)
	Phone string `form:"phone"`
	// 状态 0禁用 1正常 不传则查询全部
	Status *int8 `form:"status" binding:"omitempty,oneof=0 1"`
	// 部门ID
	DeptID *uint `form:"deptId"`
}

type BriefRole struct {
	// 角色名称
	RoleName string `json:"role_name"`
	// 角色标识
	RoleKey string `json:"role_key"`
}
type BriefDept struct {
	ID uint `json:"id"`
	// 部门名称
	DeptName string `json:"dept_name"`
}

// 用户列表项（脱敏：不含 id 与密码）
type UserItem struct {
	// 用户名
	Username string `json:"username"`
	// 用户昵称
	Nickname string `json:"nickname"`
	// 手机号码
	Phone string `json:"phone"`
	// 电子邮箱
	Email string `json:"email"`
	// 性别（1-男，2-女）
	Sex int8 `json:"sex"`
	// 头像地址
	Avatar string `json:"avatar"`
	// 账号状态（0-正常，1-停用）
	Status int8 `json:"status"`
	// 最后登录 IP
	LoginIP string `json:"loginIp"`
	// 所属部门
	Dept *BriefDept `json:"dept"`
	// 拥有的角色列表
	Roles []*BriefRole `json:"roles"`
	// 创建时间
	CreatedAt time.Time `json:"createdAt"`
}

type UserInfoResp struct {
	UserItem
	Menus []*model.Menu `json:"menus"`
}

// 分页用户列表响应
type UserListResp = PageResp[UserItem]

// 新增用户请求
type UserAddBody struct {
	// 用户名
	Username string `json:"username" binding:"required,min=6,max=30,alphanum"`
	// 密码
	Password string `json:"password" binding:"required,min=8,max=32"`
	// 用户昵称
	Nickname string `json:"nickname" binding:"omitempty,max=30"`
	// 手机号码
	Phone string `json:"phone" binding:"omitempty,max=11"`
	// 电子邮箱
	Email string `json:"email" binding:"omitempty,max=50"`
	// 性别（0-未知，1-男，2-女），与 model/UserItem 语义一致。
	// 用 omitempty 而非 required：required 会把 0 视为空导致「女」无法提交。
	Sex int8 `json:"sex" binding:"omitempty,oneof=0 1 2"`
	// 所属部门 ID
	DeptID *uint `json:"deptId"`
	// 角色标识列表
	RoleKeys []string `json:"roleKeys" binding:"required,min=1"`
	// 备注信息
	Remark string `json:"remark" binding:"max=500"`
}

// 修改用户请求
type UserEditBody struct {
	// 新密码
	Password *string `json:"password" binding:"omitempty,min=8,max=32"`
	// 所属部门 ID
	DeptID *uint `json:"deptId"`
	// 角色标识列表
	RoleKeys []string `json:"roleKeys" `
	// 账号状态（0-停用，1-正常）
	Status *int8 `json:"status" binding:"omitempty,oneof=0 1"`
}
