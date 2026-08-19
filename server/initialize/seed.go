package initialize

import (
	"fmt"
	"slices"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"rbackeeper/global"
	"rbackeeper/internal/model"
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
		{DeptName: "总公司", ParentID: 0, OrderNum: 1, Leader: "张伟", Phone: "13800000001", Email: "hq@rbackeeper.com", Status: 0},
		{DeptName: "技术部", ParentID: 1, OrderNum: 1, Leader: "李强", Phone: "13800000002", Email: "tech@rbackeeper.com", Status: 0},
		{DeptName: "产品部", ParentID: 1, OrderNum: 2, Leader: "王芳", Phone: "13800000003", Email: "product@rbackeeper.com", Status: 0},
		{DeptName: "市场部", ParentID: 1, OrderNum: 3, Leader: "刘洋", Phone: "13800000004", Email: "market@rbackeeper.com", Status: 0},
		{DeptName: "财务部", ParentID: 1, OrderNum: 4, Leader: "陈静", Phone: "13800000005", Email: "finance@rbackeeper.com", Status: 0},
		{DeptName: "人力资源部", ParentID: 1, OrderNum: 5, Leader: "赵敏", Phone: "13800000006", Email: "hr@rbackeeper.com", Status: 0},
		{DeptName: "研发组", ParentID: 2, OrderNum: 1, Leader: "孙磊", Phone: "13800000007", Email: "rd@rbackeeper.com", Status: 0},
		{DeptName: "测试组", ParentID: 2, OrderNum: 2, Leader: "周婷", Phone: "13800000008", Email: "qa@rbackeeper.com", Status: 0},
		{DeptName: "设计组", ParentID: 3, OrderNum: 1, Leader: "吴鹏", Phone: "13800000009", Email: "design@rbackeeper.com", Status: 0},
		{DeptName: "销售组", ParentID: 4, OrderNum: 1, Leader: "郑琳", Phone: "13800000010", Email: "sales@rbackeeper.com", Status: 0},
	}
}

// 返回 39 条内置菜单（M 目录 / C 菜单 / F 按钮的三层结构）。
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
		// 系统管理（目录 M）
		m("系统管理", "M", "system", "system", "", "system", 0, 1),
		// 用户管理（C 菜单 + F 按钮）
		m("用户管理", "C", "system:user", "user", "system/user/index", "user", 1, 1),
		m("用户新增", "F", "system:user:add", "", "", "#", 2, 1),
		m("用户修改", "F", "system:user:edit", "", "", "#", 2, 2),
		m("用户删除", "F", "system:user:remove", "", "", "#", 2, 3),
		m("用户查询", "F", "system:user:list", "", "", "#", 2, 4),
		m("用户详情", "F", "system:user:query", "", "", "#", 2, 5),
		m("用户导出", "F", "system:user:export", "", "", "#", 2, 6),
		m("重置密码", "F", "system:user:reset", "", "", "#", 2, 7),
		// 角色管理（C 菜单 + F 按钮）
		m("角色管理", "C", "system:role", "role", "system/role/index", "peoples", 1, 2),
		m("角色新增", "F", "system:role:add", "", "", "#", 10, 1),
		m("角色查询", "F", "system:role:list", "", "", "#", 10, 2),
		m("角色导出", "F", "system:role:export", "", "", "#", 10, 3),
		m("角色删除", "F", "system:role:remove", "", "", "#", 10, 4),
		// 菜单管理（C 菜单 + F 按钮）
		m("菜单管理", "C", "system:menu", "menu", "system/menu/index", "menu", 1, 3),
		m("菜单查询", "F", "system:menu:list", "", "", "#", 15, 1),
		m("菜单新增", "F", "system:menu:add", "", "", "#", 15, 2),
		m("菜单修改", "F", "system:menu:edit", "", "", "#", 15, 3),
		m("菜单删除", "F", "system:menu:remove", "", "", "#", 15, 4),
		// 部门管理（C 菜单 + F 按钮）
		m("部门管理", "C", "system:dept", "dept", "system/dept/index", "tree", 1, 4),
		m("部门查询", "F", "system:dept:list", "", "", "#", 20, 1),
		m("部门新增", "F", "system:dept:add", "", "", "#", 20, 2),
		m("部门修改", "F", "system:dept:edit", "", "", "#", 20, 3),
		m("部门删除", "F", "system:dept:remove", "", "", "#", 20, 4),
		// 字典（阶段七待补业务，先挂 F 按钮权限）
		m("字典类型查询", "F", "system:dict:type:list", "", "", "#", 1, 5),
		m("字典类型新增", "F", "system:dict:type:add", "", "", "#", 1, 6),
		m("字典类型修改", "F", "system:dict:type:edit", "", "", "#", 1, 7),
		m("字典类型删除", "F", "system:dict:type:delete", "", "", "#", 1, 8),
		m("字典数据查询", "F", "system:dict:data:list", "", "", "#", 1, 9),
		m("字典数据按类型", "F", "system:dict:data:type", "", "", "#", 1, 10),
		m("字典数据新增", "F", "system:dict:data:add", "", "", "#", 1, 11),
		m("字典数据修改", "F", "system:dict:data:edit", "", "", "#", 1, 12),
		m("字典数据删除", "F", "system:dict:data:delete", "", "", "#", 1, 13),
		// 系统监控（目录 M + 日志 C 菜单 / F 按钮）
		m("系统监控", "M", "monitor", "monitor", "", "monitor", 0, 2),
		m("登录日志", "C", "monitor:loginlog", "loginlog", "monitor/loginlog/index", "logininfor", 34, 1),
		m("操作日志", "C", "monitor:oplog", "oplog", "monitor/oplog/index", "form", 34, 2),
		m("登录日志查询", "F", "monitor:log:login", "", "", "#", 34, 3),
		m("操作日志查询", "F", "monitor:log:operation", "", "", "#", 34, 4),
		m("日志导出", "F", "monitor:log:export", "", "", "#", 34, 5),
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

// 初始化内置数据（部门 / 菜单 / 角色 / 用户）。
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

		// 2. 菜单（39 条，M 目录 / C 菜单 / F 按钮的三层结构）
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

		log.Info().
			Str("module", "seed").
			Int("depts", len(depts)).
			Int("menus", len(menus)).
			Int("roles", len(roles)).
			Int("users", 10).
			Str("password", seedPassword).
			Msg("内置数据初始化完成")
		return nil
	})
}
