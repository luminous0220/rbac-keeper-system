package dto

type Page struct {
	// 页码
	PageNum int `form:"pageNum" binding:"required,min=1"`
	// 条数
	PageSize int `form:"pageSize" binding:"required,min=1,max=100"`
}

type PageResp[T any] struct {
	PageNum  int   `json:"pageNum"`
	PageSize int   `json:"pageSize"`
	Total    int64 `json:"total"`
	List     []T   `json:"list"`
}

type UsernameQuery struct {
	// 用户名
	Username string `form:"username" binding:"required,min=3,max=30"`
}

type IdQuery struct {
	// ID
	ID uint `form:"id" binding:"required"`
}

type RoleKeyQuery struct {
	// 角色标识
	RoleKey string `form:"roleKey" binding:"required"`
}
