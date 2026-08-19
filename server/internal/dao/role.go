package dao

import (
	"gorm.io/gorm"

	"rbackeeper/global"
	"rbackeeper/internal/dto"
	"rbackeeper/internal/model"
)

// 分页查询角色列表，返回列表与总数。
func GetRoleList(q dto.RoleListQuery) ([]model.Role, error) {
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

// 按 role_key 查询角色（含菜单、部门预加载），未找到返回 gorm.ErrRecordNotFound
func GetRoleByKey(roleKey string) (*model.Role, error) {
	var r model.Role
	err := global.DB.Preload("Menus").Preload("Depts").
		Where("role_key = ?", roleKey).First(&r).Error
	if err != nil {
		return nil, err
	}
	return &r, nil
}

// 创建角色（事务内调用，传入 tx；成功后 role.ID 回填）
func CreateRole(db *gorm.DB, role *model.Role) error {
	return db.Create(role).Error
}

// 按主键更新指定字段（map 写入避免零值被忽略）
func UpdateRoleByID(db *gorm.DB, roleID uint, updates map[string]any) error {
	return db.Model(&model.Role{}).Where("id = ?", roleID).Updates(updates).Error
}

// 替换角色菜单关联
func ReplaceRoleMenus(db *gorm.DB, roleID uint, menus []model.Menu) error {
	return db.Model(&model.Role{ID: roleID}).
		Association("Menus").Replace(menus)
}

// 替换角色部门关联（自定义数据权限，中间表 role_depts）。
// 传入空切片表示清空；仅当角色 data_scope=2 时由数据权限中间件读取。
func ReplaceRoleDepts(db *gorm.DB, roleID uint, depts []model.Dept) error {
	return db.Model(&model.Role{ID: roleID}).
		Association("Depts").Replace(depts)
}

// 逻辑删除角色并清理全部关联（事务内调用）。
func DeleteRoleByID(db *gorm.DB, roleID uint) error {
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

// 校验 role_key 是否已存在
func ExistsRoleKey(db *gorm.DB, roleKey string) (bool, error) {
	var cnt int64
	q := db.Model(&model.Role{}).Where("role_key = ?", roleKey)
	if err := q.Count(&cnt).Error; err != nil {
		return false, err
	}
	return cnt > 0, nil
}

// 统计关联该角色的用户数（删除前校验，防权限真空）
func CountUsersByRoleID(db *gorm.DB, roleID uint) (int64, error) {
	var cnt int64
	err := db.Table("user_roles").Where("role_id = ?", roleID).Count(&cnt).Error
	return cnt, err
}

// 按 ID 批量查询菜单（授权校验用，自动过滤逻辑删除）
func GetMenusByIDs(db *gorm.DB, ids []uint) ([]model.Menu, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var menus []model.Menu
	err := db.Where("id IN ?", ids).Find(&menus).Error
	return menus, err
}

// 按 ID 批量查询部门（自定义数据权限校验用，自动过滤逻辑删除）
func GetDeptsByIDs(db *gorm.DB, ids []uint) ([]model.Dept, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var depts []model.Dept
	err := db.Where("id IN ?", ids).Find(&depts).Error
	return depts, err
}

// 返回全部未删除角色的 role_key
func ListAllRoleKeys() ([]string, error) {
	var keys []string
	err := global.DB.Model(&model.Role{}).
		Where("deleted_at IS NULL").
		Pluck("role_key", &keys).Error
	return keys, err
}
