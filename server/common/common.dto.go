package common

// 分页参数
type Page struct {
	// 页码
	PageNum int `form:"pageNum" binding:"required,min=1"`
	// 条数
	PageSize int `form:"pageSize" binding:"required,min=1,max=100"`
}

// 分页响应
type PageResp[T any] struct {
	PageNum  int   `json:"pageNum"`
	PageSize int   `json:"pageSize"`
	Total    int64 `json:"total"`
	List     []T   `json:"list"`
}

// 按用户名查询
type UsernameQuery struct {
	// 用户名
	Username string `form:"username" binding:"required,min=3,max=30"`
}

// 按主键 id 查询
type IdQuery struct {
	// ID
	ID uint `form:"id" binding:"required"`
}

// 按角色标识查询
type RoleKeyQuery struct {
	// 角色标识
	RoleKey string `form:"roleKey" binding:"required"`
}
