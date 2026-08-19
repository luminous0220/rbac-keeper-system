package dao

import (
	"gorm.io/gorm"

	"rbackeeper/global"
	"rbackeeper/internal/model"
)

// 查询全部菜单（树构建用），可选 status 过滤，按 order_num、id 排序。
func GetAllMenus(status *int8) ([]model.Menu, error) {
	q := global.DB.Model(&model.Menu{})
	if status != nil {
		q = q.Where("status = ?", *status)
	}
	var menus []model.Menu
	err := q.Order("order_num ASC, id ASC").Find(&menus).Error
	return menus, err
}

// 按主键查询菜单，未找到返回 gorm.ErrRecordNotFound
func GetMenuByID(db *gorm.DB, id uint) (*model.Menu, error) {
	var m model.Menu
	err := db.Where("id = ?", id).First(&m).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// 菜单是否存在（父菜单存在性校验用，0=根视为合法）
func ExistsMenuID(db *gorm.DB, id uint) (bool, error) {
	if id == 0 {
		return true, nil
	}
	var cnt int64
	err := db.Model(&model.Menu{}).Where("id = ?", id).Count(&cnt).Error
	return cnt > 0, err
}

// 权限标识是否已存在，excludeID>0 时排除自身（修改场景）
func ExistsMenuPerms(db *gorm.DB, perms string, excludeID uint) (bool, error) {
	if perms == "" {
		return false, nil
	}
	var cnt int64
	q := db.Model(&model.Menu{}).Where("perms = ?", perms)
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	err := q.Count(&cnt).Error
	return cnt > 0, err
}

// 统计子菜单/按钮数（删除前校验，防孤立节点）
func CountMenusByParentID(db *gorm.DB, parentID uint) (int64, error) {
	var cnt int64
	err := db.Model(&model.Menu{}).Where("parent_id = ?", parentID).Count(&cnt).Error
	return cnt, err
}

// 创建菜单（事务内调用，传入 tx；成功后 menu.ID 回填）
func CreateMenu(db *gorm.DB, menu *model.Menu) error {
	return db.Create(menu).Error
}

// 按主键更新指定字段（map 写入避免零值被忽略）
func UpdateMenuByID(db *gorm.DB, menuID uint, updates map[string]any) error {
	return db.Model(&model.Menu{}).Where("id = ?", menuID).Updates(updates).Error
}

// 逻辑删除菜单并清理角色-菜单关联（事务内调用）。
// 中间表 role_menus 无 GORM 模型，直接原生 SQL 清理，避免悬挂引用。
func DeleteMenuByID(db *gorm.DB, menuID uint) error {
	if err := db.Exec("DELETE FROM role_menus WHERE menu_id = ?", menuID).Error; err != nil {
		return err
	}
	return db.Delete(&model.Menu{}, menuID).Error
}

// ListAllWithPerms 返回所有未逻辑删除菜单（仅 id/perms/roleKey 三列供 Casbin loader 用）
type MenuPermRow struct {
	ID       uint
	Perms    string
	MenuType string
}

func ListAllWithPerms() ([]MenuPermRow, error) {
	var rows []MenuPermRow
	err := global.DB.Model(&model.Menu{}).
		Select("id, perms, menu_type").
		Where("perms <> '' AND menu_type = 'F'").
		Find(&rows).Error
	return rows, err
}

// 返回 (roleKey, perm) 列表（用于构建 Casbin p 规则）
func ListRoleMenuPerms() (map[string][]string, error) {
	type row struct {
		RoleKey string
		Perms   string
	}
	var rows []row
	err := global.DB.Table("role_menus rm").
		Select("r.role_key, m.perms").
		Joins("JOIN roles r ON r.id = rm.role_id").
		Joins("JOIN menus m ON m.id = rm.menu_id").
		Where("m.perms <> '' AND m.menu_type = 'F' AND r.deleted_at IS NULL AND m.deleted_at IS NULL").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make(map[string][]string, len(rows))
	for _, r := range rows {
		out[r.RoleKey] = append(out[r.RoleKey], r.Perms)
	}
	return out, nil
}
