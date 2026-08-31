package dict

import (
	"gorm.io/gorm"

	"rbackeeper/global"
	"rbackeeper/model"
)

// ==================== DictType 字典类型 ====================

// GetTypeList 分页查询字典类型列表，返回列表与总数。
func GetTypeList(q DictTypeListQuery) ([]model.DictType, int64, error) {
	tx := global.DB.Model(&model.DictType{})
	if q.DictName != "" {
		tx = tx.Where("dict_name LIKE ?", "%"+q.DictName+"%")
	}
	if q.DictType != "" {
		tx = tx.Where("dict_type LIKE ?", "%"+q.DictType+"%")
	}
	if q.Status != nil {
		tx = tx.Where("status = ?", *q.Status)
	}

	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []model.DictType
	err := tx.Order("id DESC").
		Offset((q.Page.PageNum - 1) * q.PageSize).Limit(q.PageSize).
		Find(&list).Error
	if err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// GetTypeByID 按主键查询字典类型，未找到返回 gorm.ErrRecordNotFound
func GetTypeByID(db *gorm.DB, id uint) (*model.DictType, error) {
	var t model.DictType
	err := db.Where("id = ?", id).First(&t).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// ExistsTypeByCode 校验 dict_type 编码唯一性（excludeID 排除自身，编辑时传 id；新增/存在性校验传 0）。
func ExistsTypeByCode(db *gorm.DB, dictType string, excludeID uint) (bool, error) {
	var cnt int64
	q := db.Model(&model.DictType{}).Where("dict_type = ?", dictType)
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	if err := q.Count(&cnt).Error; err != nil {
		return false, err
	}
	return cnt > 0, nil
}

// CreateType 创建字典类型（事务内调用，传入 tx；成功后 ID 回填）
func CreateType(db *gorm.DB, t *model.DictType) error {
	return db.Create(t).Error
}

// UpdateTypeByID 按主键更新指定字段（map 写入避免零值被忽略）
func UpdateTypeByID(db *gorm.DB, id uint, updates map[string]any) error {
	return db.Model(&model.DictType{}).Where("id = ?", id).Updates(updates).Error
}

// UpdateDataCodeByType 按 dict_type 编码批量更新字典数据的 dict_type 字段（类型编码变更时联动，事务内调用）。
func UpdateDataCodeByType(db *gorm.DB, oldCode, newCode string) error {
	return db.Model(&model.DictData{}).Where("dict_type = ?", oldCode).Update("dict_type", newCode).Error
}

// DeleteTypeByID 逻辑删除字典类型并联动软删该类型下全部字典数据（事务内调用）。
// dict_data 通过字符串 dict_type 关联（非主键），无中间表，直接按编码软删。
func DeleteTypeByID(db *gorm.DB, id uint) error {
	var t model.DictType
	if err := db.Where("id = ?", id).First(&t).Error; err != nil {
		return err
	}
	// 联动软删字典数据
	if err := db.Where("dict_type = ?", t.DictType).Delete(&model.DictData{}).Error; err != nil {
		return err
	}
	return db.Delete(&model.DictType{}, id).Error
}

// ==================== DictData 字典数据 ====================

// GetDataList 分页查询字典数据列表，返回列表与总数。
func GetDataList(q DictDataListQuery) ([]model.DictData, int64, error) {
	tx := global.DB.Model(&model.DictData{})
	if q.DictType != "" {
		tx = tx.Where("dict_type = ?", q.DictType)
	}
	if q.DictLabel != "" {
		tx = tx.Where("dict_label LIKE ?", "%"+q.DictLabel+"%")
	}
	if q.Status != nil {
		tx = tx.Where("status = ?", *q.Status)
	}

	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []model.DictData
	err := tx.Order("dict_sort ASC, id ASC").
		Offset((q.Page.PageNum - 1) * q.PageSize).Limit(q.PageSize).
		Find(&list).Error
	if err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// GetDataByType 查询某字典类型下状态正常(status=0)的数据（缓存回源用，按 dict_sort 排序）。
func GetDataByType(dictType string) ([]model.DictData, error) {
	var list []model.DictData
	err := global.DB.Where("dict_type = ? AND status = ?", dictType, 0).
		Order("dict_sort ASC, id ASC").Find(&list).Error
	return list, err
}

// GetDataByID 按主键查询字典数据，未找到返回 gorm.ErrRecordNotFound
func GetDataByID(db *gorm.DB, id uint) (*model.DictData, error) {
	var d model.DictData
	err := db.Where("id = ?", id).First(&d).Error
	if err != nil {
		return nil, err
	}
	return &d, nil
}

// ExistsDataByValue 校验同一字典类型下 dict_value 唯一性（excludeID 排除自身）。
func ExistsDataByValue(db *gorm.DB, dictType, dictValue string, excludeID uint) (bool, error) {
	var cnt int64
	q := db.Model(&model.DictData{}).Where("dict_type = ? AND dict_value = ?", dictType, dictValue)
	if excludeID > 0 {
		q = q.Where("id <> ?", excludeID)
	}
	if err := q.Count(&cnt).Error; err != nil {
		return false, err
	}
	return cnt > 0, nil
}

// CreateData 创建字典数据（事务内调用，传入 tx；成功后 ID 回填）
func CreateData(db *gorm.DB, d *model.DictData) error {
	return db.Create(d).Error
}

// UpdateDataByID 按主键更新指定字段（map 写入避免零值被忽略）
func UpdateDataByID(db *gorm.DB, id uint, updates map[string]any) error {
	return db.Model(&model.DictData{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteDataByID 逻辑删除字典数据（事务内调用）
func DeleteDataByID(db *gorm.DB, id uint) error {
	return db.Delete(&model.DictData{}, id).Error
}
