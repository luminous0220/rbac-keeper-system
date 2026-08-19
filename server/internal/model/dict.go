package model

// DictType 字典类型模型
type DictType struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	DictName string `gorm:"size:100;not null;comment:字典名称" json:"dict_name"`
	DictType string `gorm:"size:100;uniqueIndex;not null;comment:字典类型" json:"dict_type"`
	Status   int8   `gorm:"default:0;comment:状态 0正常 1停用" json:"status"`

	// 关联关系：字典数据（一对多，按 dict_type 字符串关联）
	Data []DictData `gorm:"foreignKey:DictType;references:DictType" json:"data,omitempty"`
	BaseModel
}

// DictData 字典数据模型
type DictData struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	DictSort  int    `gorm:"default:0;comment:字典排序" json:"dict_sort"`
	DictLabel string `gorm:"size:100;not null;comment:字典标签" json:"dict_label"`
	DictValue string `gorm:"size:100;not null;comment:字典键值" json:"dict_value"`
	DictType  string `gorm:"size:100;index;not null;comment:字典类型" json:"dict_type"`
	CssClass  string `gorm:"size:100;comment:样式属性" json:"css_class"`
	ListClass string `gorm:"size:100;comment:表格回显样式" json:"list_class"`
	IsDefault int8   `gorm:"default:0;comment:是否默认 0否 1是" json:"is_default"`
	Status    int8   `gorm:"default:0;comment:状态 0正常 1停用" json:"status"`
	BaseModel
}
