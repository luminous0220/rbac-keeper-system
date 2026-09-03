package model

// Menu 菜单模型
type Menu struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	MenuName  string `gorm:"size:50;not null;comment:菜单名称" json:"menu_name"`
	ParentID  uint   `gorm:"default:0;comment:父菜单ID" json:"parent_id"`
	OrderNum  int    `gorm:"default:0;comment:显示顺序" json:"order_num"`
	Path      string `gorm:"size:200;comment:路由地址" json:"path"`
	Component string `gorm:"size:255;comment:组件路径" json:"component"`
	Query     string `gorm:"size:255;comment:路由参数" json:"query"`
	MenuType  string `gorm:"size:1;not null;comment:菜单类型 M菜单 C目录 F功能" json:"menu_type"`
	Status    int8   `gorm:"default:0;comment:菜单状态 1正常 0停用" json:"status"`
	Perms     string `gorm:"size:100;comment:权限标识(仅功能F使用)" json:"perms"`
	Icon      string `gorm:"size:100;comment:菜单图标" json:"icon"`
	Visible   int8   `gorm:"default:1;comment:显示状态 1显示 0隐藏" json:"visible"`
	Remark    string `gorm:"size:500;comment:备注" json:"remark"`
	// 关联关系：自关联子菜单（指针切片，便于 O(n) 递归挂载与多级构建）
	Children []*Menu `gorm:"foreignKey:ParentID" json:"children"`
	BaseModel
}
