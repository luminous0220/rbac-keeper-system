package initialize

import (
	"errors"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"rbackeeper/global"
	"rbackeeper/model"
)

// 系统监控目录下的“登录日志/操作日志”M 菜单规格（按 path 稳定定位，不依赖位置自增 ID）
var logMenuMSpecs = []struct {
	path, name, component, icon string
	order                       int
}{
	{"loginlog", "登录日志", "Monitor/LoginLog/index.tsx", "ProfileOutlined", 1},
	{"oplog", "操作日志", "Monitor/OperationLog/index.tsx", "FormOutlined", 2},
}

// 日志族 F 功能行规格：parentPath 指向所属 M 菜单的 path，perms 为稳定查找键
var logMenuFSpecs = []struct {
	parentPath, name, perm string
	order                  int
}{
	{"loginlog", "登录日志查询", "monitor:log:login", 3},
	{"loginlog", "登录日志导出", "monitor:log:login:export", 5},
	{"loginlog", "登录日志清空", "monitor:log:login:clear", 6},
	{"oplog", "操作日志查询", "monitor:log:operation", 4},
	{"oplog", "操作日志导出", "monitor:log:operation:export", 7},
	{"oplog", "操作日志清空", "monitor:log:operation:clear", 8},
}

// ReconcileLogMenus 幂等校正“系统监控 → 登录/操作日志”菜单：
//  1. 修正两个日志 M 菜单的 component 为真实页面路径（老库 seed 曾写入错误大小写）；
//  2. 按 perms 补齐缺失的导出/清空 F 行；
//  3. 把日志族全部 F 菜单 id 补挂进 super 角色的 role_menus，保证前端 hasPerm 可见。
//
// 设计约束：
//   - 一律按 (menu_type, path)/(menu_type, perms) 稳定字段查找父级，绝不使用位置自增 ID，
//     避免 admin 在菜单管理里改过顺序/删除重建后校正出错；
//   - 每次启动都运行：全新库（seed 已插全量含新增行）为 no-op，老库负责修复；
//   - 失败仅告警，不阻断启动；super 为 casbin 通配，此处纯为前端权限集合补全。
func ReconcileLogMenus() {
	if global.DB == nil {
		return
	}
	if err := reconcileLogMenusTx(global.DB); err != nil {
		log.Warn().Err(err).Str("module", "seed").Msg("日志菜单校正失败（仅告警，不影响启动）")
	}
}

func reconcileLogMenusTx(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// 1. 确保目录“系统监控”（C, path=monitor）存在
		var monitor model.Menu
		err := tx.Where("menu_type = ? AND path = ?", "C", "monitor").First(&monitor).Error
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			monitor = model.Menu{
				MenuName: "系统监控", Path: "monitor", MenuType: "C",
				OrderNum: 2, Status: 1, Icon: "MonitorOutlined",
			}
			if err := tx.Create(&monitor).Error; err != nil {
				return err
			}
		case err != nil:
			return err
		}

		// 2. 确保两个日志 M 菜单存在并修正 component（与 admin 真实页面大小写一致）
		parentIDs := make(map[string]uint, len(logMenuMSpecs))
		for _, spec := range logMenuMSpecs {
			var m model.Menu
			err := tx.Where("menu_type = ? AND path = ?", "M", spec.path).First(&m).Error
			switch {
			case errors.Is(err, gorm.ErrRecordNotFound):
				m = model.Menu{
					MenuName: spec.name, Path: spec.path, Component: spec.component,
					MenuType: "M", OrderNum: spec.order, Status: 1, Icon: spec.icon,
					ParentID: monitor.ID,
				}
				if err := tx.Create(&m).Error; err != nil {
					return err
				}
			case err != nil:
				return err
			case m.Component != spec.component:
				// component 与真实文件不一致时校正（路由解析按大小写精确匹配）
				if err := tx.Model(&m).Update("component", spec.component).Error; err != nil {
					return err
				}
			}
			parentIDs[spec.path] = m.ID
		}

		// 3. 按 perms 补齐缺失的 F 功能行（含历史已有的查询行，幂等跳过）
		fIDs := make([]uint, 0, len(logMenuFSpecs))
		for _, spec := range logMenuFSpecs {
			var f model.Menu
			err := tx.Where("menu_type = ? AND perms = ?", "F", spec.perm).First(&f).Error
			switch {
			case errors.Is(err, gorm.ErrRecordNotFound):
				parentID, ok := parentIDs[spec.parentPath]
				if !ok {
					log.Warn().Str("perm", spec.perm).Msg("日志菜单校正：父菜单缺失，跳过该 F 行")
					continue
				}
				f = model.Menu{
					MenuName: spec.name, Perms: spec.perm, MenuType: "F",
					ParentID: parentID, OrderNum: spec.order, Status: 1, Icon: "#",
				}
				if err := tx.Create(&f).Error; err != nil {
					return err
				}
			case err != nil:
				return err
			}
			fIDs = append(fIDs, f.ID)
		}

		// 4. 把日志族 F 菜单 id 补挂进 super 角色（避免 Replace，保留 admin 后加的其它菜单）
		var super model.Role
		if err := tx.Where("role_key = ?", "super").First(&super).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil // super 尚不存在（seed 未跑/未成功），全量由 seed 负责
			}
			return err
		}
		var have []uint
		if err := tx.Table("role_menus").
			Where("role_id = ? AND menu_id IN ?", super.ID, fIDs).
			Pluck("menu_id", &have).Error; err != nil {
			return err
		}
		seen := make(map[uint]bool, len(have))
		for _, id := range have {
			seen[id] = true
		}
		var missing []*model.Menu
		for _, id := range fIDs {
			if !seen[id] {
				var mm model.Menu
				if err := tx.First(&mm, id).Error; err != nil {
					return err
				}
				missing = append(missing, &mm)
			}
		}
		if len(missing) > 0 {
			if err := tx.Model(&super).Association("Menus").Append(missing); err != nil {
				return err
			}
		}
		return nil
	})
}
