package user

import (
	"context"
	"errors"
	"unicode"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"rbackeeper/constants"
	"rbackeeper/global"
	"rbackeeper/internal/dept"
	"rbackeeper/internal/role"
	"rbackeeper/model"
	"rbackeeper/response"
	"rbackeeper/utils"
)

// List 用户列表（应用 dataScope：super 全部；DEPT/DEPT_AND_CHILDREN/CUSTOM 走部门过滤；SELF 仅本人）
func List(ctx *gin.Context, query UserListQuery) (*UserListResp, error) {
	q := UserListQuery{
		Page:     query.Page,
		Username: query.Username,
		Phone:    query.Phone,
		Status:   query.Status,
		DeptID:   query.DeptID,
	}
	// 部门筛选：选中部门连同其全部子部门一起参与过滤（部门树左侧筛选的语义）。
	// 需全量部门列表（不应用数据权限）展开子树，否则子部门被权限裁剪后会漏过滤；
	// 用户行的可见性仍由 dao 中的 GetUserDataScope 单独约束。
	if query.DeptID != nil {
		all, err := dept.GetAllDeptsScoped(nil, nil)
		if err != nil {
			return nil, err
		}
		q.DeptIDs = utils.CollectSubtreeIDs(all,
			func(d model.Dept) uint { return d.ID },
			func(d model.Dept) uint { return d.ParentID },
			*query.DeptID,
		)
	}
	users, total, err := GetUserList(ctx, q)

	if err != nil {
		return nil, err
	}
	list := make([]UserItem, 0, len(users))
	for i := range users {
		list = append(list, toUserItem(&users[i]))
	}
	return &UserListResp{List: list, Total: total,
		PageNum:  query.PageNum,
		PageSize: query.PageSize}, nil
}

// Info 用户详情
func Info(ctx context.Context, username string) (*UserInfoResp, error) {
	user, err := GetUserInfoByUsernameScoped(nil, global.DB, username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, response.NewBizError(response.ErrNotFound, "用户不存在")
		}
		return nil, err
	}

	// 一个用户有多个角色，每个角色都菜单权限列表可能会有重复的，取并集
	userItem := toUserItem(user)
	menus := make([]*model.Menu, 0)
	for _, v := range user.Roles {
		menus = append(menus, v.Menus...)
	}
	uniqueMenus := utils.UniqueBy(menus, func(m *model.Menu) uint {
		return m.ID
	})

	return &UserInfoResp{
		UserItem: userItem,
		Menus:    uniqueMenus,
	}, nil
}

// Add 新增用户（分配的部门须在当前用户数据权限内）
func Add(ctx *gin.Context, body UserAddBody) error {
	if !isAlphaNumeric(body.Username) {
		return response.NewBizError(response.ErrInvalidParam, "Username 只能包含英文字母和数字，且不能是纯数字")
	}

	// 唯一性校验
	if exists, err := ExistsUsername(global.DB, body.Username); err != nil {
		return err
	} else if exists {
		return response.NewBizError(response.ErrInvalidParam, "用户名已存在")
	}

	// 部门存在性（须在当前用户数据权限内）
	if ok, err := dept.ExistsByIDScoped(ctx, global.DB, body.DeptID); err != nil {
		return err
	} else if !ok {
		return response.NewBizError(response.ErrInvalidParam, "所选部门不存在")
	}

	// 密码复杂度 + 哈希
	if !validPassword(body.Password) {
		return response.NewBizError(response.ErrInvalidParam, "密码至少 8 位，需同时包含字母和数字")
	}
	hashed, err := utils.Hash(body.Password)
	if err != nil {
		return err
	}

	// 获取角色
	roles, err := role.GetByKeys(global.DB, body.RoleKeys)
	if err != nil {
		return err
	}

	if len(roles) <= 0 {
		return response.NewBizError(response.ErrInvalidParam, "RoleKeys 中包含不存在的角色key")
	}

	// 组装用户
	user := &model.User{
		Username: body.Username,
		Password: hashed,
		Nickname: body.Nickname,
		Phone:    body.Phone,
		Email:    body.Email,
		Sex:      body.Sex,
		DeptID:   *body.DeptID,
		Remark:   body.Remark,
		Status:   1, // 新增用户默认正常
	}

	// 事务：建用户 + 关联角色
	return global.DB.Transaction(func(tx *gorm.DB) error {
		if err := CreateUser(tx, user); err != nil {
			return err
		}
		if len(roles) > 0 {
			if err := ReplaceUserRoles(tx, user.ID, roles); err != nil {
				return err
			}
		}
		return nil
	})
}

// Edit 修改用户（分配的部门须在当前用户数据权限内）
func Edit(ctx *gin.Context, username string, body UserEditBody) error {
	if username == constants.SuperUsername {
		return response.NewBizError(response.ErrNotFound, "禁止修改")
	}

	user, err := GetUserInfoByUsernameScoped(ctx, global.DB, username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.NewBizError(response.ErrNotFound, "用户不存在")
		}
		return err
	}

	// 部门存在性（须在当前用户数据权限内）
	if ok, err := dept.ExistsByIDScoped(ctx, global.DB, body.DeptID); err != nil {
		return err
	} else if !ok {
		return response.NewBizError(response.ErrInvalidParam, "所选部门不存在")
	}

	// 组装更新字段（map 写入，避免零值忽略；nil 字段不参与更新）
	updates := map[string]any{}
	if body.DeptID != nil {
		updates["dept_id"] = *body.DeptID
	}
	if body.Status != nil {
		updates["status"] = *body.Status
	}
	// 可选密码重置
	if body.Password != nil {
		if !validPassword(*body.Password) {
			return response.NewBizError(response.ErrInvalidParam, "密码至少 8 位，需同时包含字母和数字")
		}
		hashed, err := utils.Hash(*body.Password)
		if err != nil {
			return err
		}
		updates["password"] = hashed
	}

	// 角色替换（nil=不改；空切片=清空；非空=校验后替换）。
	var roles []model.Role
	manageRoles := body.RoleKeys != nil
	if manageRoles && len(body.RoleKeys) > 0 {
		if roles, err = role.GetByKeys(global.DB, body.RoleKeys); err != nil {
			return err
		}
		if len(roles) != len(body.RoleKeys) {
			return response.NewBizError(response.ErrInvalidParam, "RoleKeys 中包含不存在的角色key")
		}
	}

	return global.DB.Transaction(func(tx *gorm.DB) error {
		if err := UpdateUserByID(tx, user.ID, updates); err != nil {
			return err
		}
		if manageRoles {
			if err := ReplaceUserRoles(tx, user.ID, roles); err != nil {
				return err
			}
		}
		return nil
	})
}

// 列表项（脱敏：不含 id/密码）
func toUserItem(u *model.User) UserItem {
	item := UserItem{
		Username:  u.Username,
		Nickname:  u.Nickname,
		Sex:       u.Sex,
		Avatar:    u.Avatar,
		Status:    u.Status,
		LoginIP:   u.LoginIP,
		CreatedAt: u.CreatedAt,
		Phone:     u.Phone,
		Email:     u.Email,
		Roles:     getBriefRoles(u.Roles),
		Dept: &BriefDept{
			ID:       u.Dept.ID,
			DeptName: u.Dept.DeptName,
		},
	}
	return item
}

// 获取简洁的 roles，只包含RoleKey和RoleName
func getBriefRoles(roles []*model.Role) []*BriefRole {
	list := make([]*BriefRole, len(roles))
	for i, r := range roles {
		list[i] = &BriefRole{
			RoleKey:  r.RoleKey,
			RoleName: r.RoleName,
		}
	}
	return list
}

// 密码复杂度：>=8 位且含字母与数字
func validPassword(p string) bool {
	if len(p) < 8 {
		return false
	}
	var hasLetter, hasDigit bool
	for _, r := range p {
		switch {
		case unicode.IsLetter(r):
			hasLetter = true
		case unicode.IsDigit(r):
			hasDigit = true
		}
	}
	return hasLetter && hasDigit
}

// isAlphaNumeric 校验用户名仅含字母数字，且不能是纯数字
func isAlphaNumeric(s string) bool {
	if s == "" {
		return false
	}
	hasLetter := false
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z':
			hasLetter = true
		case r >= '0' && r <= '9':
		default:
			return false
		}
	}
	return hasLetter
}
