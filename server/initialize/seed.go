package initialize

import (
	"fmt"
	"slices"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"rbackeeper/global"
	"rbackeeper/model"
	"rbackeeper/utils"
)

const (
	seedPassword = "123456aaa" // 内置用户统一初始密码（8 位含字母+数字，符合复杂度规则）
	seedSuperKey = "super"     // 超级管理员角色标识，同时作为幂等检测锚点
)

// 内置用户与其角色的配对，便于事务内分步写入主表与中间表。
type seedUser struct {
	user *model.User
	role model.Role
}

// 返回 10 条模拟部门（1 总公司 + 5 一级部门 + 4 二级部门）。
func seedDepts() []model.Dept {
	return []model.Dept{
		{DeptName: "总公司", ParentID: 0, OrderNum: 1, Leader: "张伟", Phone: "13800000001", Email: "hq@rbackeeper.com", Status: 1},
		{DeptName: "技术部", ParentID: 1, OrderNum: 1, Leader: "李强", Phone: "13800000002", Email: "tech@rbackeeper.com", Status: 1},
		{DeptName: "产品部", ParentID: 1, OrderNum: 2, Leader: "王芳", Phone: "13800000003", Email: "product@rbackeeper.com", Status: 1},
		{DeptName: "市场部", ParentID: 1, OrderNum: 3, Leader: "刘洋", Phone: "13800000004", Email: "market@rbackeeper.com", Status: 1},
		{DeptName: "财务部", ParentID: 1, OrderNum: 4, Leader: "陈静", Phone: "13800000005", Email: "finance@rbackeeper.com", Status: 1},
		{DeptName: "人力资源部", ParentID: 1, OrderNum: 5, Leader: "赵敏", Phone: "13800000006", Email: "hr@rbackeeper.com", Status: 1},
		{DeptName: "研发组", ParentID: 2, OrderNum: 1, Leader: "孙磊", Phone: "13800000007", Email: "rd@rbackeeper.com", Status: 1},
		{DeptName: "测试组", ParentID: 2, OrderNum: 2, Leader: "周婷", Phone: "13800000008", Email: "qa@rbackeeper.com", Status: 1},
		{DeptName: "设计组", ParentID: 3, OrderNum: 1, Leader: "吴鹏", Phone: "13800000009", Email: "design@rbackeeper.com", Status: 1},
		{DeptName: "销售组", ParentID: 4, OrderNum: 1, Leader: "郑琳", Phone: "13800000010", Email: "sales@rbackeeper.com", Status: 1},
	}
}

// 返回 41 条内置菜单（M 菜单 / C 目录 / F 功能的三层结构）。
// 注意：ParentID 依赖 Create 批量插入时自增 ID 从 1 起按切片顺序分配（根节点为 0），与 seedDepts 同款约定。
func seedMenus() []model.Menu {
	m := func(name, menuType, perms, path, component, icon string, parent, order int) model.Menu {
		return model.Menu{
			MenuName:  name,
			ParentID:  uint(parent),
			OrderNum:  order,
			Path:      path,
			Component: component,
			MenuType:  menuType,
			Status:    1, // 1 正常
			Perms:     perms,
			Icon:      icon,
		}
	}
	return []model.Menu{
		m("数据看板", "M", "", "/dash-board", "Dashboard/index.tsx", "DashboardOutlined", 0, 10),
		// 系统管理（目录 C）
		m("系统管理", "C", "", "system", "", "", 0, 1),
		// 用户管理（M 菜单 + F 功能）
		m("用户管理", "M", "", "user", "System/User/index.tsx", "UserOutlined", 2, 1),
		m("用户新增", "F", "system:user:add", "", "", "#", 3, 1),
		m("用户修改", "F", "system:user:edit", "", "", "#", 3, 2),
		m("用户删除", "F", "system:user:remove", "", "", "#", 3, 3),
		m("用户查询", "F", "system:user:list", "", "", "#", 3, 4),
		m("用户详情", "F", "system:user:query", "", "", "#", 3, 5),
		m("用户导出", "F", "system:user:export", "", "", "#", 3, 6),
		m("重置密码", "F", "system:user:reset", "", "", "#", 3, 7),
		// 角色管理（M 菜单 + F 功能）
		m("角色管理", "M", "", "role", "System/Role/index.tsx", "TeamOutlined", 2, 2),
		m("角色新增", "F", "system:role:add", "", "", "#", 11, 1),
		m("角色查询", "F", "system:role:list", "", "", "#", 11, 2),
		m("角色导出", "F", "system:role:export", "", "", "#", 11, 3),
		m("角色删除", "F", "system:role:remove", "", "", "#", 11, 4),
		// 菜单管理（M 菜单 + F 功能）
		m("菜单管理", "M", "", "MenuOutlined", "System/Menu/index.tsx", "MenuOutlined", 2, 3),
		m("菜单查询", "F", "system:menu:list", "", "", "#", 15, 1),
		m("菜单新增", "F", "system:menu:add", "", "", "#", 15, 2),
		m("菜单修改", "F", "system:menu:edit", "", "", "#", 15, 3),
		m("菜单删除", "F", "system:menu:remove", "", "", "#", 15, 4),
		// 部门管理（M 菜单 + F 功能）
		m("部门管理", "M", "", "dept", "System/Dept/index.tsx", "ApartmentOutlined", 2, 4),
		m("部门查询", "F", "system:dept:list", "", "", "#", 21, 1),
		m("部门新增", "F", "system:dept:add", "", "", "#", 21, 2),
		m("部门修改", "F", "system:dept:edit", "", "", "#", 21, 3),
		m("部门删除", "F", "system:dept:remove", "", "", "#", 21, 4),
		// 字典管理（M 菜单 + F 功能）
		m("字典管理", "M", "", "ApartmentOutlined", "System/Dict/index.tsx", "ContainerOutlined", 2, 5),
		m("字典类型查询", "F", "system:dict:type:list", "", "", "#", 26, 1),
		m("字典类型新增", "F", "system:dict:type:add", "", "", "#", 26, 2),
		m("字典类型修改", "F", "system:dict:type:edit", "", "", "#", 26, 3),
		m("字典类型删除", "F", "system:dict:type:delete", "", "", "#", 26, 4),
		m("字典数据查询", "F", "system:dict:data:list", "", "", "#", 26, 5),
		m("字典数据按类型查询", "F", "system:dict:data:type", "", "", "#", 26, 6),
		m("字典数据新增", "F", "system:dict:data:add", "", "", "#", 26, 7),
		m("字典数据修改", "F", "system:dict:data:edit", "", "", "#", 26, 8),
		m("字典数据删除", "F", "system:dict:data:delete", "", "", "#", 26, 9),
		// 系统监控（目录 C + 日志 M 菜单 / F 功能）
		m("系统监控", "C", "", "monitor", "", "MonitorOutlined", 0, 2),
		m("登录日志", "M", "", "loginlog", "monitor/Loginlog/index.tsx", "ProfileOutlined", 36, 1),
		m("登录日志查询", "F", "monitor:log:login", "", "", "#", 37, 3),
		m("操作日志", "M", "", "oplog", "monitor/Oplog/index.tsx", "FormOutlined", 36, 2),
		m("操作日志查询", "F", "monitor:log:operation", "", "", "#", 39, 4),
	}
}

// 返回 4 个内置角色。
func seedRoles() []model.Role {
	return []model.Role{
		{RoleName: "超级管理员", RoleKey: "super", RoleSort: 1, DataScope: 1, Status: 1, Remark: "拥有全部权限，受保护不可删除"},
		{RoleName: "普通管理员", RoleKey: "admin", RoleSort: 2, DataScope: 1, Status: 1, Remark: "日常运营管理"},
		{RoleName: "普通用户", RoleKey: "user", RoleSort: 3, DataScope: 5, Status: 1, Remark: "仅可见本人数据"},
		{RoleName: "员工", RoleKey: "staff", RoleSort: 4, DataScope: 3, Status: 1, Remark: "可见本部门数据"},
	}
}

// 返回 10 个内置用户（1 super + 2 admin + 3 user + 4 staff）与各自角色的配对。
func seedUsers(hashed string, depts []model.Dept, roles []model.Role) []seedUser {
	type spec struct {
		username, nickname string
		sex                int8
		phone              string
		deptIdx            int
		roleKey            string
		loginDate          time.Time
	}
	specs := []spec{
		{"super", "超级管理员", 1, "13900000001", 0, "super", time.Now()},
		{"admin01", "管理员甲", 1, "13900000002", 1, "admin", time.Now()},
		{"admin02", "管理员乙", 2, "13900000003", 2, "admin", time.Now()},
		{"user01", "用户甲", 1, "13900000004", 1, "user", time.Now()},
		{"user02", "用户乙", 2, "13900000005", 6, "user", time.Now()},
		{"user03", "用户丙", 1, "13900000006", 7, "user", time.Now()},
		{"staff01", "员工甲", 1, "13900000007", 3, "staff", time.Now()},
		{"staff02", "员工乙", 2, "13900000008", 9, "staff", time.Now()},
		{"staff03", "员工丙", 1, "13900000009", 4, "staff", time.Now()},
		{"staff04", "员工丁", 2, "13900000010", 5, "staff", time.Now()},
	}
	out := make([]seedUser, 0, len(specs))
	for _, s := range specs {

		idx := slices.IndexFunc(roles, func(r model.Role) bool {
			return r.RoleKey == s.roleKey
		})

		out = append(out, seedUser{
			user: &model.User{
				Username:  s.username,
				Password:  hashed,
				Nickname:  s.nickname,
				Phone:     s.phone,
				Email:     s.username + "@rbackeeper.com",
				Sex:       s.sex,
				DeptID:    depts[s.deptIdx].ID,
				LoginDate: &s.loginDate,
				Status:    1, // 1 正常
			},
			role: roles[idx],
		})
	}
	return out
}

// 返回 12 条内置字典类型。
// 键值语义见 seedDictData，全部与对应模型字段注释保持一致。
func seedDictTypes() []model.DictType {
	t := func(name, dictType string) model.DictType {
		return model.DictType{
			DictName: name,
			DictType: dictType,
			Status:   1, // 1 正常
		}
	}
	return []model.DictType{
		t("用户性别", "sys_user_sex"),
		t("系统开关", "sys_normal_disable"),
		t("显示状态", "sys_show_hide"),
		t("系统是否", "sys_yes_no"),
		t("数据权限", "sys_data_scope"),
		t("菜单类型", "sys_menu_type"),
		t("操作类别", "sys_oper_type"),
		t("操作状态", "sys_oper_status"),
		t("登录状态", "sys_common_status"),
		t("通知类型", "sys_notice_type"),
		t("通知状态", "sys_notice_status"),
		t("任务状态", "sys_job_status"),
	}
}

// 返回 12 条内置字典类型对应的字典数据。
// 键值严格遵循各模型字段注释：性别 0未知1男2女、通用状态 0停用1正常、可见 0隐藏1显示、
// 数据权限 1-5、菜单类型 M菜单 C目录 F功能、操作类别 0其它1后台2手机端、登录状态 0成功1失败、操作状态 0正常1异常。
func seedDictData() []model.DictData {
	d := func(sort int, label, value, dictType, listClass string) model.DictData {
		return model.DictData{
			DictSort:  sort,
			DictLabel: label,
			DictValue: value,
			DictType:  dictType,
			ListClass: listClass,
			Status:    1, // 1 正常
		}
	}
	return []model.DictData{
		// 用户性别（model.User.Sex：0 未知 1 男 2 女）
		d(1, "未知", "0", "sys_user_sex", ""),
		d(2, "男", "1", "sys_user_sex", ""),
		d(3, "女", "2", "sys_user_sex", ""),
		// 系统开关（通用 Status 字段：1 正常 0 停用）
		d(1, "正常", "1", "sys_normal_disable", "success"),
		d(2, "停用", "0", "sys_normal_disable", "danger"),
		// 显示状态（model.Menu.Visible：1 显示 0 隐藏）
		d(1, "显示", "1", "sys_show_hide", "success"),
		d(2, "隐藏", "0", "sys_show_hide", "warning"),
		// 系统是否
		d(1, "是", "Y", "sys_yes_no", "success"),
		d(2, "否", "N", "sys_yes_no", "danger"),
		// 数据权限（constants/data_scope.go：1 全部 2 自定义 3 本部门 4 本部门及以下 5 仅本人）
		d(1, "全部数据权限", "1", "sys_data_scope", ""),
		d(2, "自定义数据权限", "2", "sys_data_scope", ""),
		d(3, "本部门数据权限", "3", "sys_data_scope", ""),
		d(4, "本部门及以下数据权限", "4", "sys_data_scope", ""),
		d(5, "仅本人数据权限", "5", "sys_data_scope", ""),
		// 菜单类型（model.Menu.MenuType：M 菜单 C 目录 F 功能）
		d(1, "目录", "C", "sys_menu_type", ""),
		d(2, "菜单", "M", "sys_menu_type", ""),
		d(3, "功能", "F", "sys_menu_type", ""),
		// 操作类别（model.OperationLog.OperatorType：0 其它 1 后台用户 2 手机端用户）
		d(1, "其它", "0", "sys_oper_type", "info"),
		d(2, "后台用户", "1", "sys_oper_type", "primary"),
		d(3, "手机端用户", "2", "sys_oper_type", "warning"),
		// 操作状态（model.OperationLog.Status：0 正常 1 异常）
		d(1, "正常", "0", "sys_oper_status", "success"),
		d(2, "异常", "1", "sys_oper_status", "danger"),
		// 登录状态（model.LoginLog.Status：0 成功 1 失败）
		d(1, "成功", "0", "sys_common_status", "success"),
		d(2, "失败", "1", "sys_common_status", "danger"),
		// 通知类型
		d(1, "通知", "1", "sys_notice_type", "warning"),
		d(2, "公告", "2", "sys_notice_type", "success"),
		// 通知状态
		d(1, "正常", "0", "sys_notice_status", "success"),
		d(2, "关闭", "1", "sys_notice_status", "danger"),
		// 任务状态
		d(1, "正常", "0", "sys_job_status", "success"),
		d(2, "暂停", "1", "sys_job_status", "danger"),
	}
}

// 初始化内置数据（部门 / 菜单 / 角色 / 用户 / 字典）。
func RegisterSeed() {
	if global.DB == nil {
		log.Fatal().Msg("数据库未初始化")
		return
	}

	// super 角色存在视为已播种
	var seeded int64
	if err := global.DB.Model(&model.Role{}).Where("role_key = ?", seedSuperKey).Count(&seeded).Error; err != nil {
		log.Fatal().Err(err).Msg("seed 幂等检测失败")
	}
	if seeded > 0 {
		log.Info().Str("module", "seed").Msg("内置数据已存在，跳过初始化")
		return
	}

	// 统一初始密码哈希（所有内置用户共用，仅哈希一次）
	hashed, err := utils.Hash(seedPassword)
	if err != nil {
		log.Fatal().Err(err).Msg("seed 密码哈希失败")
	}

	global.DB.Transaction(func(tx *gorm.DB) error {
		// 1. 部门（10 条，扁平存储 + ParentID 表达层级，读取时由 utils.Build 建树）
		depts := seedDepts()
		if err := tx.Create(&depts).Error; err != nil {
			log.Fatal().Err(err).Msg("seed 部门失败")
		}

		// 2. 菜单（39 条，M 菜单 / C 目录 / F 功能的三层结构）
		menus := seedMenus()
		if err := tx.Create(&menus).Error; err != nil {
			log.Fatal().Err(err).Msg("seed 菜单失败")
		}

		// 3. 角色（4 条：super / admin / user / staff）
		roles := seedRoles()
		if err := tx.Create(&roles).Error; err != nil {
			log.Fatal().Err(err).Msg("seed 角色失败")
		}

		// 4. super 角色关联全部菜单（RBAC g 规则命中，super 拥有全量权限）
		var superRole model.Role
		if err := tx.Where("role_key = ?", seedSuperKey).First(&superRole).Error; err != nil {
			log.Fatal().Err(err).Msg("seed 查询 super 角色失败")
		}
		if err := tx.Model(&superRole).Association("Menus").Replace(menus); err != nil {
			log.Fatal().Err(err).Msg("seed super 角色菜单关联失败")
			return nil
		}

		// 5. 用户（10 条：1 super + 2 admin + 3 user + 4 staff）
		//    先建用户主记录拿到 ID，再用 Replace 写中间表 user_roles，
		//    避免批量插入时 many2many 关联行为不可控。

		var persistedRoles []model.Role
		tx.Find(&persistedRoles) // 查询所有已插入的角色
		for _, su := range seedUsers(hashed, depts, persistedRoles) {
			if err := tx.Create(su.user).Error; err != nil {
				log.Fatal().Err(err).Msg(fmt.Sprintf("seed 用户 %s 失败", su.user.Username))
			}
			if err := tx.Model(&su.user).
				Association("Roles").Replace(su.role); err != nil {
				log.Fatal().Msg(fmt.Sprintf("seed 用户 %s 角色关联失败: %v", su.user.Username, err))
				log.Fatal().Err(err).Msg(fmt.Sprintf("seed 用户 %s 角色关联失败", su.user.Username))
				return nil
			}
		}

		// 6. 字典类型（12 条）与字典数据（对应键值），供前端下拉 / 标签回显
		dictTypes := seedDictTypes()
		if err := tx.Create(&dictTypes).Error; err != nil {
			log.Fatal().Err(err).Msg("seed 字典类型失败")
		}
		dictData := seedDictData()
		if err := tx.Create(&dictData).Error; err != nil {
			log.Fatal().Err(err).Msg("seed 字典数据失败")
		}

		log.Info().
			Str("module", "seed").
			Int("depts", len(depts)).
			Int("menus", len(menus)).
			Int("roles", len(roles)).
			Int("users", 10).
			Int("dictTypes", len(dictTypes)).
			Int("dictData", len(dictData)).
			Str("password", seedPassword).
			Msg("内置数据初始化完成")
		return nil
	})
}
