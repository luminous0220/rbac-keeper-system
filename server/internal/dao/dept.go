package dao

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"rbackeeper/global"
	"rbackeeper/internal/model"
	"rbackeeper/utils"
)

// 校验部门是否存在，且在当前用户数据权限内（父部门选择/用户分配部门等写路径校验用）。
// deptID 为 nil（前端不传时） 视为"不分配部门"，合法直接放行。
// 不传ctx，表示不做权限隔离
func ExistsDeptByIDScoped(ctx *gin.Context, db *gorm.DB, deptID *uint) (bool, error) {
	if deptID == nil {
		return true, nil
	}

	var cnt int64

	if ctx != nil {
		db = db.Scopes(utils.GetDataScope(ctx))
	}

	err := db.
		Model(&model.Dept{}).
		Where("id = ?", deptID).
		Count(&cnt).Error
	return cnt > 0, err
}

// 校验部门deptName唯一性
func ExistsDeptByDeptName(db *gorm.DB, deptName string, excludeID uint) (bool, error) {
	if deptName == "" {
		return true, nil // 空名称视为无效，根据业务需求也可返回 false
	}
	var cnt int64
	q := db.Model(&model.Dept{}).Where("dept_name = ?", deptName)
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	err := q.Count(&cnt).Error
	return cnt > 0, err
}

// 查询全部部门（树构建用），可选 status 过滤，按 order_num、id 排序。
// 应用数据权限：目标行须在当前用户可见范围内。
// 不传ctx，表示不做权限隔离
func GetAllDeptsScoped(ctx *gin.Context, status *int8) ([]model.Dept, error) {
	db := global.DB
	if ctx != nil {
		db = db.Scopes(utils.GetDataScope(ctx))
	}

	if status != nil {
		db = db.Where("status = ?", *status)
	}
	var depts []model.Dept
	err := db.Order("order_num ASC, id ASC").Find(&depts).Error

	return depts, err
}

// 按主键查询部门，目标部门须在当前用户数据权限内。
// 未找到或不在权限内均返回 gorm.ErrRecordNotFound（fail-closed，不泄露行存在性）。
// 不传ctx，表示不做权限隔离。
func GetDeptByIDScoped(ctx *gin.Context, db *gorm.DB, id uint) (*model.Dept, error) {
	var d model.Dept

	if ctx != nil {
		db = db.Scopes(utils.GetDataScope(ctx))
	}

	err := db.Where("id = ?", id).First(&d).Error
	if err != nil {
		return nil, err
	}
	return &d, nil
}

// 统计子部门数（删除前校验）
func CountDeptsByParentID(db *gorm.DB, parentID uint) (int64, error) {
	var cnt int64
	err := db.Model(&model.Dept{}).Where("parent_id = ?", parentID).Count(&cnt).Error
	return cnt, err
}

// 统计部门下用户数（删除前校验，防用户悬挂）
func CountUsersByDeptID(db *gorm.DB, deptID uint) (int64, error) {
	var cnt int64
	err := db.Model(&model.User{}).Where("dept_id = ?", deptID).Count(&cnt).Error
	return cnt, err
}

// 创建部门（事务内调用，传入 tx；成功后 dept.ID 回填）
func CreateDept(db *gorm.DB, dept *model.Dept) error {
	return db.Create(dept).Error
}

// 按主键更新指定字段（map 写入避免零值被忽略）
func UpdateDeptByID(db *gorm.DB, deptID uint, updates map[string]any) error {
	return db.Model(&model.Dept{}).Where("id = ?", deptID).Updates(updates).Error
}

// 逻辑删除部门并清理角色-部门数据权限关联（事务内调用）。
// 中间表 role_depts 无 GORM 模型，直接原生 SQL 清理；用户 dept_id 引用由删除前校验拦截。
func DeleteDeptByID(db *gorm.DB, deptID uint) error {
	if err := db.Exec("DELETE FROM role_depts WHERE dept_id = ?", deptID).Error; err != nil {
		return err
	}
	return db.Delete(&model.Dept{}, deptID).Error
}
