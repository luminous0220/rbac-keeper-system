package dao

import (
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"rbackeeper/global"
	"rbackeeper/internal/constants"
	"rbackeeper/internal/dto"
	"rbackeeper/internal/model"
	"rbackeeper/utils"
)

// 按用户名查询用户（含角色），未找到返回 gorm.ErrRecordNotFound
func GetUserAndRolesByUsername(username string) (*model.User, error) {
	var u model.User
	err := global.DB.Preload("Roles").Where("username = ?", username).First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// 回写最后登录 IP 与时间
func UpdateLoginInfo(username, ip string) error {
	return global.DB.Model(&model.User{}).
		Where("username = ?", username).
		Updates(map[string]any{
			"login_ip":   ip,
			"login_date": time.Now(),
		}).Error
}

// 更新密码哈希
func UpdatePassword(username, hashed string) error {
	return global.DB.Model(&model.User{}).
		Where("username = ?", username).
		Update("password", hashed).Error
}

// 分页查询用户列表（按数据权限过滤 dept_id/id），预加载角色与部门，返回列表与总数。
func GetUserList(ctx *gin.Context, q dto.UserListQuery) ([]model.User, int64, error) {
	tx := global.DB.Model(&model.User{}).Scopes(utils.GetUserDataScope(ctx))
	if q.Username != "" {
		tx = tx.Where("username LIKE ?", "%"+q.Username+"%")
	}
	if q.Phone != "" {
		tx = tx.Where("phone LIKE ?", "%"+q.Phone+"%")
	}
	if q.Status != nil {
		tx = tx.Where("status = ?", *q.Status)
	}
	if len(q.DeptIDs) > 0 {
		tx = tx.Where("dept_id IN ?", q.DeptIDs)
	}

	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var users []model.User
	err := tx.Preload("Roles").Preload("Dept").
		Offset((q.Page.PageNum-1)*q.PageSize).Limit(q.PageSize).
		Order("created_at DESC").Where("id <> ?", constants.SuperId).
		Find(&users).Error
	if err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

// 按 username 查询用户（含角色/菜单/数据权限部门/行政部门）。
// 目标用户须在当前用户数据权限内；未找到或不在权限内均返回 gorm.ErrRecordNotFound。
// 不传入ctx，表示不需数据隔离。
func GetUserInfoByUsernameScoped(ctx *gin.Context, db *gorm.DB, username string) (*model.User, error) {
	var u model.User

	if ctx != nil {
		db = db.Scopes(utils.GetUserDataScope(ctx))
	}

	err := db.Preload("Roles.Menus").Preload("Roles.Depts").Preload("Dept").
		Where("username = ?", username).First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// 创建用户
func CreateUser(db *gorm.DB, user *model.User) error {
	return db.Create(user).Error
}

// CreateUsers 批量创建用户（事务内调用，传入 tx）。
// 使用 CreateInBatches 分批插入，减少往返；成功后各 user.ID 回填。
func CreateUsers(db *gorm.DB, users []*model.User, batchSize int) error {
	if len(users) == 0 {
		return nil
	}
	if batchSize <= 0 {
		batchSize = 100
	}
	return db.CreateInBatches(users, batchSize).Error
}

// 按主键更新指定字段（map 写入避免零值被忽略）
func UpdateUserByID(db *gorm.DB, userID uint, updates map[string]any) error {
	return db.Model(&model.User{}).Where("id = ?", userID).Updates(updates).Error
}

// 按用户名更新指定字段（map 写入避免零值被忽略；个人资料修改场景）
func UpdateUserByUsername(db *gorm.DB, username string, updates map[string]any) error {
	return db.Model(&model.User{}).Where("username = ?", username).Updates(updates).Error
}

// 替换用户角色关联：清空原关联后重建（many2many 中间表 user_roles）。
func ReplaceUserRoles(db *gorm.DB, userID uint, roles []model.Role) error {
	return db.Model(&model.User{ID: userID}).
		Association("Roles").Replace(roles)
}

// 按 role_key 批量查询角色（新增/修改用户时关联校验用）
func GetRolesByKeys(db *gorm.DB, roleKeys []string) ([]model.Role, error) {
	if len(roleKeys) == 0 {
		return nil, nil
	}
	var roles []model.Role
	err := db.Where("role_key IN ?", roleKeys).Find(&roles).Error
	return roles, err
}

// 用户名是否已存在，excludeID>0 时排除自身（修改场景）
func ExistsUsername(db *gorm.DB, username string) (bool, error) {

	var cnt int64
	q := db.Model(&model.User{}).Where("username = ?", username)
	// 必须先 Count 再判断，否则 cnt 在 Count 执行前为 0，恒返回 false
	if err := q.Count(&cnt).Error; err != nil {
		return false, err
	}
	return cnt > 0, nil
}

// 验证手机号是否已存在
func ExistsPhone(db *gorm.DB, phone string) (bool, error) {
	var cnt int64
	q := db.Model(&model.User{}).Where("phone = ?", phone)

	if err := q.Count(&cnt).Error; err != nil {
		return false, err
	}
	return cnt > 0, nil
}

// 返回所有未逻辑删除用户的 username（Casbin g 规则：user -> role）
func ListAllUsernames() ([]string, error) {
	var names []string
	err := global.DB.Model(&model.User{}).
		Where("deleted_at IS NULL").
		Pluck("username", &names).Error
	return names, err
}

// 返回 (username, roleKey) 列表（g 规则）
func ListUserRoles() (map[string][]string, error) {
	type row struct {
		Username string
		RoleKey  string
	}
	var rows []row
	err := global.DB.Table("user_roles ur").
		Select("u.username, r.role_key").
		Joins("JOIN users u ON u.id = ur.user_id").
		Joins("JOIN roles r ON r.id = ur.role_id").
		Where("u.deleted_at IS NULL AND r.deleted_at IS NULL").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make(map[string][]string, len(rows))
	for _, r := range rows {
		out[r.Username] = append(out[r.Username], r.RoleKey)
	}
	return out, nil
}
