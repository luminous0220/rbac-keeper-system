package role

import (
	"gorm.io/gorm"

	"rbackeeper/global"
	"rbackeeper/model"
)

// GetList 分页查询角色列表，返回列表与总数。
func GetList(q RoleListQuery) ([]model.Role, error) {
	tx := global.DB.Model(&model.Role{})
	if q.RoleName != "" {
		tx = tx.Where("role_name LIKE ?", "%"+q.RoleName+"%")
	}
	if q.RoleKey != "" {
		tx = tx.Where("role_key LIKE ?", "%"+q.RoleKey+"%")
	}
	if q.Status != nil {
		tx = tx.Where("status = ?", *q.Status)
	}

	var roles []model.Role
	err := tx.Preload("Menus").Preload("Depts").Order("role_sort ASC, created_at DESC").
		Find(&roles).Error
	if err != nil {
		return nil, err
	}
	return roles, nil
}

// GetByKey 按 role_key 查询角色（含菜单、部门预加载），未找到返回 gorm.ErrRecordNotFound
func GetByKey(roleKey string) (*model.Role, error) {
	var r model.Role
	err := global.DB.Preload("Menus").Preload("Depts").
		Where("role_key = ?", roleKey).First(&r).Error
	if err != nil {
		return nil, err
	}
	return &r, nil
}

// GetByKeys 按 role_key 批量查询角色（新增/修改用户时关联校验用）
func GetByKeys(db *gorm.DB, roleKeys []string) ([]model.Role, error) {
	if len(roleKeys) == 0 {
		return nil, nil
	}
	var roles []model.Role
	err := db.Where("role_key IN ?", roleKeys).Find(&roles).Error
	return roles, err
}

// Create 创建角色（事务内调用，传入 tx；成功后 role.ID 回填）
func Create(db *gorm.DB, role *model.Role) error {
	return db.Create(role).Error
}

// UpdateByID 按主键更新指定字段（map 写入避免零值被忽略）
func UpdateByID(db *gorm.DB, roleID uint, updates map[string]any) error {
	return db.Model(&model.Role{}).Where("id = ?", roleID).Updates(updates).Error
}

// ReplaceMenus 替换角色菜单关联
func ReplaceMenus(db *gorm.DB, roleID uint, menus []model.Menu) error {
	return db.Model(&model.Role{ID: roleID}).
		Association("Menus").Replace(menus)
}

// ReplaceDepts 替换角色部门关联（自定义数据权限，中间表 role_depts）。
// 传入空切片表示清空；仅当角色 data_scope=2 时由数据权限中间件读取。
func ReplaceDepts(db *gorm.DB, roleID uint, depts []model.Dept) error {
	return db.Model(&model.Role{ID: roleID}).
		Association("Depts").Replace(depts)
}

// DeleteByID 逻辑删除角色并清理全部关联（事务内调用）。
func DeleteByID(db *gorm.DB, roleID uint) error {
	if err := db.Exec("DELETE FROM role_menus WHERE role_id = ?", roleID).Error; err != nil {
		return err
	}
	if err := db.Exec("DELETE FROM role_depts WHERE role_id = ?", roleID).Error; err != nil {
		return err
	}
	if err := db.Exec("DELETE FROM user_roles WHERE role_id = ?", roleID).Error; err != nil {
		return err
	}
	return db.Delete(&model.Role{}, roleID).Error
}

// ExistsKey 校验 role_key 是否已存在
func ExistsKey(db *gorm.DB, roleKey string) (bool, error) {
	var cnt int64
	q := db.Model(&model.Role{}).Where("role_key = ?", roleKey)
	if err := q.Count(&cnt).Error; err != nil {
		return false, err
	}
	return cnt > 0, nil
}

// CountUsersByID 统计关联该角色的用户数（删除前校验，防权限真空）
func CountUsersByID(db *gorm.DB, roleID uint) (int64, error) {
	var cnt int64
	err := db.Table("user_roles").Where("role_id = ?", roleID).Count(&cnt).Error
	return cnt, err
}
