package model

// Role 角色模型
type Role struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	RoleName  string `gorm:"size:30;not null;comment:角色名称" json:"role_name"`
	RoleKey   string `gorm:"size:100;uniqueIndex;not null;comment:角色标识" json:"role_key"`
	RoleSort  int    `gorm:"default:0;comment:显示顺序" json:"role_sort"`
	DataScope int8   `gorm:"default:1;comment:数据权限 1全部 2自定义 3本部门 4本部门及以下 5仅本人" json:"data_scope"`
	Status    int8   `gorm:"default:1;comment:状态 0禁用 1正常" json:"status"`

	// 关联关系：GORM 自动管理中间表
	Menus []*Menu `gorm:"many2many:role_menus;" json:"menus"` // 无约束中间表
	// 这里的 Depts 代表部门数据权限范围 仅当 DataScope==2 时才会被读取生效，其余范围忽略。
	Depts  []*Dept `gorm:"many2many:role_depts;" json:"depts"`
	Remark string  `gorm:"size:500;comment:备注" json:"remark"`
	BaseModel
}
