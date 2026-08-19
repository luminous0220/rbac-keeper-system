package utils

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"rbackeeper/internal/constants"
)

// 核心通用方法
// 它返回一个 gorm.Scope 函数，可以直接链式调用
func GetDataScope(ctx *gin.Context) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		// 1. 从 Context 获取权限信息
		scopeVal, exists := ctx.Get(constants.CtxDataScope)
		if !exists {
			return db // 如果没有权限信息，默认不过滤（或返回空，视安全策略而定）
		}

		dataScope := scopeVal.(int8)

		// 2. 超级管理员或全部权限，直接返回（不过滤）
		if dataScope == constants.DsAll {
			return db
		}

		// 3. 根据权限类型构建查询
		switch dataScope {
		case constants.DsCustom, constants.DsSubDept:
			// 自定义：获取授权部门列表
			depts, _ := ctx.Get(constants.CtxDataScopeDepts)
			deptIDs := depts.([]uint)
			if len(deptIDs) == 0 {
				return db.Where("1 = 0") // 空集合，查不到数据
			}
			return db.Where("id IN ?", deptIDs)

		case constants.DsDept:
			// 本部门
			deptID, _ := ctx.Get(constants.CtxDataScopeDept)
			return db.Where("id = ?", deptID)

		case constants.DsSelf:
			// 仅本人：表示无权限查看部门数据
			return db.Where("1 = 0")

		default:
			return db.Where("1 = 0")
		}
	}
}

// 用户表数据权限过滤（用户列表用，基于 dept_id / id）。
// 规则：DsAll(1) → 不过滤；DsCustom(2)/DsSubDept(4) → dept_id IN 授权部门集合。
// DsDept(3) → dept_id = 本部门；DsSelf(5) → id = 本人。未注入权限信息时不过滤（与 GetDataScope 一致）。
func GetUserDataScope(ctx *gin.Context) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		scopeVal, exists := ctx.Get(constants.CtxDataScope)
		if !exists {
			return db
		}
		dataScope, ok := scopeVal.(int8)
		if !ok || dataScope == constants.DsAll {
			return db
		}

		switch dataScope {
		case constants.DsCustom, constants.DsSubDept:
			deptsVal, _ := ctx.Get(constants.CtxDataScopeDepts)
			deptIDs, ok := deptsVal.([]uint)
			if !ok || len(deptIDs) == 0 {
				return db.Where("1 = 0")
			}
			return db.Where("dept_id IN ?", deptIDs)

		case constants.DsDept:
			deptVal, _ := ctx.Get(constants.CtxDataScopeDept)
			deptID, ok := deptVal.(uint)
			if !ok {
				return db.Where("1 = 0")
			}
			return db.Where("dept_id = ?", deptID)

		case constants.DsSelf:
			uidVal, _ := ctx.Get(constants.CtxDataScopeUserID)
			uid, ok := uidVal.(uint)
			if !ok {
				return db.Where("1 = 0")
			}
			return db.Where("id = ?", uid)

		default:
			return db.Where("1 = 0")
		}
	}
}
