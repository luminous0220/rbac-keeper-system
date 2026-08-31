package menu

import "time"

// MenuTreeNode 菜单树节点
type MenuTreeNode struct {
	ID        uint           `json:"id"`
	ParentID  uint           `json:"parentId"`
	MenuName  string         `json:"menuName"`
	MenuType  string         `json:"menuType"` // M菜单 C目录 F功能
	OrderNum  int            `json:"orderNum"`
	Path      string         `json:"path"`
	Component string         `json:"component"`
	Query     string         `json:"query"`
	Visible   int8           `json:"visible"` // 0隐藏 1显示
	Status    int8           `json:"status"`  // 0停用 1正常
	Perms     string         `json:"perms"`
	Icon      string         `json:"icon"`
	Children  []MenuTreeNode `json:"children,omitempty"`
}

// MenuTreeQuery 菜单树查询
type MenuTreeQuery struct {
	// 模糊匹配（返回匹配节点及其祖先链，保证树完整）
	MenuName string `form:"menuName"`
	Status   *int8  `form:"status" binding:"omitempty,oneof=0 1"`
}

// MenuTreeResp 菜单树响应（直接返回节点数组）
type MenuTreeResp = []MenuTreeNode

// MenuDetailResp 菜单详情
type MenuDetailResp struct {
	ID        uint      `json:"id"`
	ParentID  uint      `json:"parentId"`
	MenuName  string    `json:"menuName"`
	MenuType  string    `json:"menuType"`
	OrderNum  int       `json:"orderNum"`
	Path      string    `json:"path"`
	Component string    `json:"component"`
	Query     string    `json:"query"`
	Visible   int8      `json:"visible"`
	Status    int8      `json:"status"`
	Perms     string    `json:"perms"`
	Icon      string    `json:"icon"`
	Remark    string    `json:"remark"`
	CreatedAt time.Time `json:"createdAt"`
}

// MenuAddBody 新增菜单（parentId=0 为根）
type MenuAddBody struct {
	// 父级菜单ID
	ParentID uint `json:"parentId"`
	// 菜单名称
	MenuName string `json:"menuName" binding:"required,min=2,max=50"`
	// 菜单类型（M=菜单, C=目录, F=功能）
	MenuType string `json:"menuType" binding:"required,oneof=M C F"`
	// 显示顺序
	OrderNum int `json:"orderNum"`
	// 路由地址
	Path string `json:"path" binding:"omitempty,max=200"`
	// 组件路径
	Component string `json:"component" binding:"omitempty,max=255"`
	// 路由参数
	Query string `json:"query" binding:"omitempty,max=255"`
	// 菜单状态（1=正常, 0=停用），与 model.Menu 语义一致。
	// omitempty 而非 required：required 会把 0 视为空导致「停用」无法提交。
	Status int8 `json:"status" binding:"omitempty,oneof=0 1"`
	// 权限标识
	Perms string `json:"perms" binding:"omitempty,max=100"`
	// 菜单图标
	Icon string `json:"icon" binding:"omitempty,max=100"`
	// 显示状态 0隐藏 1显示
	Visible int8 `json:"visible" binding:"omitempty,oneof=0 1"`
	// 备注
	Remark string `json:"remark" binding:"omitempty,max=500"`
}

// MenuEditBody 修改菜单
type MenuEditBody struct {
	// 父级菜单ID
	ParentID *uint `json:"parentId"`
	// 菜单名称
	MenuName *string `json:"menuName" binding:"omitempty,min=2,max=50"`
	// 菜单类型（M=菜单, C=目录, F=功能）
	MenuType *string `json:"menuType" binding:"omitempty,oneof=M C F"`
	// 显示顺序
	OrderNum *int `json:"orderNum"`
	// 路由地址
	Path *string `json:"path" binding:"omitempty,max=200"`
	// 组件路径
	Component *string `json:"component" binding:"omitempty,max=255"`
	// 路由参数
	Query *string `json:"query" binding:"omitempty,max=255"`
	// 菜单状态（1=正常, 0=停用），与 model.Menu 语义一致
	Status *int8 `json:"status" binding:"omitempty,oneof=0 1"`
	// 权限标识
	Perms *string `json:"perms" binding:"omitempty,max=100"`
	// 菜单图标
	Icon *string `json:"icon" binding:"omitempty,max=100"`
	// 显示状态 0隐藏 1显示
	Visible *int8 `json:"visible" binding:"omitempty,oneof=0 1"`
	// 备注
	Remark *string `json:"remark" binding:"omitempty,max=500"`
}
