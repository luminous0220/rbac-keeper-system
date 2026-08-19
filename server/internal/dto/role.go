package dto

// 角色列表查询
type RoleListQuery struct {
	// 角色名称(模糊)
	RoleName string `form:"roleName"`
	// 角色标识(模糊)
	RoleKey string `form:"roleKey"`
	// 状态 0禁用 1正常 不传则不筛选
	Status *int8 `form:"status" binding:"omitempty,oneof=0 1"`
}

type RoleItem struct {
	// 角色名
	RoleName string `json:"roleName"`
	// 角色key
	RoleKey string `json:"roleKey"`
	// 顺序
	RoleSort int `json:"roleSort"`
	// 角色数据权限 （1全部 2自定义 3本部门 4本部门及以下 5仅本人）
	DataScope int8 `json:"dataScope"`
	// 授权菜单 ID 列表（传空数组清空）
	MenuIDs []uint `json:"menuIds"`
	// 自定义数据权限部门（传空数组清空，dataScope 改为非 2 时自动清空）
	DeptIDs []uint `json:"deptIds"`
	// 备注
	Remark string `json:"remark"`
}

// 新增角色（role_key 唯一且创建后不可变）
type RoleAddBody struct {
	// 角色名
	RoleName string `json:"roleName" binding:"required,min=2,max=30"`
	// 角色key
	RoleKey string `json:"roleKey" binding:"required,min=2,max=32"`
	// 顺序
	RoleSort int `json:"roleSort"`
	// 角色数据权限 （1全部 2自定义 3本部门 4本部门及以下 5仅本人）
	DataScope int8 `json:"dataScope" binding:"required,oneof=1 2 3 4 5"`
	// 授权菜单 ID 列表
	MenuIDs []uint `json:"menuIds"`
	// 自定义数据权限部门
	DeptIDs []uint `json:"deptIds" binding:"required_if=DataScope 2"`
	// 备注
	Remark string `json:"remark" binding:"omitempty,max=500"`
}

// 修改角色
type RoleEditBody struct {
	// 角色名
	RoleName *string `json:"roleName" binding:"omitempty,min=2,max=30"`
	// 顺序
	RoleSort *int `json:"roleSort"`
	// 角色数据权限 （1全部 2自定义 3本部门 4本部门及以下 5仅本人）
	DataScope *int8 `json:"dataScope" binding:"omitempty,oneof=1 2 3 4 5"`
	// 状态
	Status *int8 `json:"status" binding:"omitempty,oneof=0 1"`
	// 授权菜单 ID 列表（传空数组清空）
	MenuIDs []uint `json:"menuIds"`
	// 自定义数据权限部门（传空数组清空，dataScope 改为非 2 时自动清空）
	DeptIDs []uint `json:"deptIds" binding:"required_if=DataScope 2"`
	// 备注
	Remark *string `json:"remark" binding:"omitempty,max=500"`
}
