package auth

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mojocn/base64Captcha"
	"gorm.io/gorm"

	"rbackeeper/constants"
	"rbackeeper/global"
	"rbackeeper/internal/log"
	"rbackeeper/internal/user"
	"rbackeeper/model"
	"rbackeeper/response"
	"rbackeeper/utils"
)

const (
	failKeyPrefix      = "fail:"
	lockKeyPrefix      = "lock:"
	blacklistKeyPrefix = "blacklist:"
	pwchangedKeyPrefix = "pwchanged:" // 密码修改时间戳，用于强制登出

	failThreshold = 5
	failWindow    = 15 * time.Minute
	lockWindow    = 15 * time.Minute
)

// ==================== 验证码 ====================

var (
	captchaOnce   sync.Once
	captchaIns    *base64Captcha.Captcha
	captchaExpire = 5 * time.Minute
	captchaKey    = "captcha:"
)

type redisCaptchaStore struct {
	expire time.Duration
}

func (s *redisCaptchaStore) Set(id, value string) error {
	return global.Redis.Set(context.Background(), captchaKey+id, value, s.expire).Err()
}

func (s *redisCaptchaStore) Get(id string, clear bool) string {
	ctx := context.Background()
	key := captchaKey + id
	val, err := global.Redis.Get(ctx, key).Result()
	if err != nil {
		return ""
	}
	if clear {
		global.Redis.Del(ctx, key)
	}
	return val
}

func (s *redisCaptchaStore) Verify(id, answer string, clear bool) bool {
	return s.Get(id, clear) == answer
}

// 初始化验证码实例（依赖 global.Redis，须在 Redis 就绪后调用）
func initCaptcha() {
	store := &redisCaptchaStore{
		expire: captchaExpire,
	}
	driver := base64Captcha.NewDriverDigit(
		80,  // 图片高度 ImgHeight
		240, // 图片宽度 ImgWidth
		4,   // 验证码字符长度 KeyLong
		0.7, // 干扰线条密度 0~1，越大越模糊
		80,  // 字体大小
	)
	captchaIns = base64Captcha.NewCaptcha(driver, store)
}

// GenerateCaptcha 生成验证码，返回 id 与 base64 图片
func GenerateCaptcha() (id, b64s string, err error) {
	captchaOnce.Do(initCaptcha)
	id, b64s, _, err = captchaIns.Generate()
	return
}

// verifyCaptcha 校验验证码（一次性，校验后即删）
func verifyCaptcha(id, answer string) bool {
	captchaOnce.Do(initCaptcha)
	return captchaIns.Verify(id, answer, true)
}

// ==================== 登录/令牌 ====================

// Login 登录
func Login(ctx context.Context, body LoginBody, ip, ua string) (*LoginResp, error) {
	// 校验验证码（一次性）
	if !verifyCaptcha(body.CaptchaID, body.CaptchaValue) {
		return nil, response.NewBizError(response.ErrInvalidParam, "验证码错误或已失效")
	}

	// 账号锁定检查
	if locked, _ := global.Redis.Exists(ctx, lockKeyPrefix+body.Username).Result(); locked > 0 {
		return nil, response.NewBizError(response.ErrForbidden, "账号已锁定，请 15 分钟后再试")
	}

	// 查用户并校验密码（用户不存在与密码错误返回同一提示，防枚举）
	u, err := user.GetUserAndRolesByUsername(body.Username)
	if err != nil || !utils.Verify(u.Password, body.Password) {
		recordLoginFailure(ctx, body.Username, ip, ua)
		return nil, response.NewBizError(response.ErrUnauthorized, "用户名或密码错误")
	}

	// 账号状态
	if u.Status != 1 {
		return nil, response.NewBizError(response.ErrForbidden, "账号已禁用")
	}

	// 登录成功：清失败计数
	global.Redis.Del(ctx, failKeyPrefix+body.Username)

	// 签发双令牌
	roleKey := primaryRoleKey(u.Roles)
	access, _, err := utils.Generate(constants.AccessToken, u.ID, u.Username, roleKey)
	if err != nil {
		return nil, err
	}
	refresh, _, err := utils.Generate(constants.RefreshToken, u.ID, u.Username, roleKey)
	if err != nil {
		return nil, err
	}

	// 回写登录信息 + 登录日志
	_ = user.UpdateLoginInfo(u.Username, ip)
	writeLoginLog(u.Username, ip, ua, true, "登录成功")

	return &LoginResp{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresIn:    int64(global.CFG.JWT.AccessExpire.Seconds()),
		Username:     u.Username,
		Nickname:     u.Nickname,
	}, nil
}

// Refresh 校验 Refresh Token 并签发新 Access Token
func Refresh(ctx context.Context, body RefreshBody) (*RefreshResp, error) {
	claims, err := utils.Parse(body.RefreshToken)
	if err != nil {
		return nil, response.NewBizError(response.ErrTokenInvalid, "刷新令牌无效或已过期")
	}
	if claims.Type != constants.RefreshToken {
		return nil, response.NewBizError(response.ErrTokenInvalid, "令牌类型错误")
	}
	if revoked, msg := IsTokenRevoked(ctx, claims); revoked {
		return nil, response.NewBizError(response.ErrTokenInvalid, msg)
	}
	access, _, err := utils.Generate(constants.AccessToken, claims.UserID, claims.Username, claims.RoleKey)
	if err != nil {
		return nil, err
	}
	return &RefreshResp{
		AccessToken: access,
		ExpiresIn:   int64(global.CFG.JWT.AccessExpire.Seconds()),
	}, nil
}

// Logout 登出：将 Access Token 的 jti 加入 Redis 黑名单，TTL 为剩余有效期
func Logout(ctx context.Context, claims *utils.Claims) error {
	if claims.ID == "" || claims.ExpiresAt == nil {
		return nil
	}
	ttl := time.Until(claims.ExpiresAt.Time)
	if ttl <= 0 {
		return nil // 已过期，无需入黑名单
	}
	return global.Redis.Set(ctx, blacklistKeyPrefix+claims.ID, "1", ttl).Err()
}

// ResetPassword 密码重置：校验旧密码 -> 更新 -> 记录密码修改时间（强制登出所有会话）
func ResetPassword(ctx context.Context, claims *utils.Claims, body ResetPasswordBody) error {
	if !validPassword(body.NewPassword) {
		return response.NewBizError(response.ErrInvalidParam, "密码至少 8 位，需同时包含字母和数字")
	}
	u, err := user.GetUserAndRolesByUsername(claims.Username)
	if err != nil {
		return response.NewBizError(response.ErrNotFound, "用户不存在")
	}
	if !utils.Verify(u.Password, body.OldPassword) {
		return response.NewBizError(response.ErrInvalidParam, "旧密码错误")
	}
	if body.OldPassword == body.NewPassword {
		return response.NewBizError(response.ErrInvalidParam, "新密码不能与旧密码相同")
	}
	hashed, err := utils.Hash(body.NewPassword)
	if err != nil {
		return err
	}
	if err := user.UpdatePassword(u.Username, hashed); err != nil {
		return err
	}
	// 强制登出：记录密码修改时间，所有早于该时间的令牌（含 refresh）即刻失效
	global.Redis.Set(ctx, pwchangedKeyPrefix+strconv.FormatUint(uint64(claims.UserID), 10), time.Now().Unix(), 0)
	return nil
}

// UpdateProfile 修改个人信息（仅当前登录用户本人，自作用域无需数据权限）。
// 指针字段 nil 不修改；返回更新后的本人信息（含角色并集菜单），前端据此同步 store。
func UpdateProfile(ctx context.Context, claims *utils.Claims, body ProfileBody) (*user.UserItem, error) {
	if claims == nil || claims.Username == "" {
		return nil, response.NewBizError(response.ErrUnauthorized, "登录状态失效，请重新登录")
	}
	updates := map[string]any{}
	if body.Nickname != nil {
		updates["nickname"] = *body.Nickname
	}
	if body.Phone != nil {
		updates["phone"] = *body.Phone
	}
	if body.Email != nil {
		updates["email"] = *body.Email
	}
	if body.Sex != nil {
		updates["sex"] = *body.Sex
	}
	if body.Avatar != nil {
		updates["avatar"] = *body.Avatar
	}
	if len(updates) == 0 {
		return nil, response.NewBizError(response.ErrInvalidParam, "没有需要修改的字段")
	}
	if err := user.UpdateUserByUsername(global.DB, claims.Username, updates); err != nil {
		return nil, err
	}
	return user.Info(ctx, claims.Username)
}

// Permissions 返回当前用户的菜单权限列表与数据权限部门列表（引导期权限唯一来源）。
// 菜单：启用角色并集，去重后剪枝（停用菜单及子孙一并剔除）。
// 部门：按 dataScope 判断（依赖 DataScope 中间件注入的 scope），并排除停用部门。
func Permissions(ctx *gin.Context) (*PermissionsResp, error) {
	claims := utils.GetClaims(ctx)
	if claims == nil || claims.Username == "" {
		return nil, response.NewBizError(response.ErrUnauthorized, "登录状态失效，请重新登录")
	}

	u, err := user.GetUserDetailByUsernameScoped(nil, global.DB, claims.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, response.NewBizError(response.ErrUnauthorized, "用户不存在")
		}
		return nil, err
	}

	menus := make([]*model.Menu, 0)
	for _, r := range u.Roles {
		if r.Status != constants.StatusEnabled {
			continue
		}
		menus = append(menus, r.Menus...)
	}
	menus = utils.UniqueBy(menus, func(m *model.Menu) uint { return m.ID })
	menus = utils.PruneSubtrees(menus,
		func(m *model.Menu) uint { return m.ID },
		func(m *model.Menu) uint { return m.ParentID },
		func(m *model.Menu) bool { return m.Status == constants.StatusEnabled },
	)

	depts := make([]model.Dept, 0)
	if err := global.DB.Model(&model.Dept{}).
		Scopes(utils.GetDataScope(ctx)).
		Where("status = ?", constants.StatusEnabled).
		Order("order_num ASC, id ASC").
		Find(&depts).Error; err != nil {
		return nil, err
	}

	return &PermissionsResp{Menus: menus, Depts: depts}, nil
}

// IsTokenRevoked 检查令牌是否已吊销：黑名单（登出）或签发早于密码修改时间（密码重置强制登出）
func IsTokenRevoked(ctx context.Context, claims *utils.Claims) (bool, string) {
	if isBlacklisted(ctx, claims.ID) {
		return true, "令牌已失效，请重新登录"
	}
	if isPasswordChangedAfter(ctx, claims) {
		return true, "登录状态已失效，请重新登录"
	}
	return false, ""
}

// 记录登录失败：原子计数 + 达阈值锁定 + 写失败日志
func recordLoginFailure(ctx context.Context, username, ip, ua string) {
	key := failKeyPrefix + username
	pipe := global.Redis.Pipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, failWindow)
	_, _ = pipe.Exec(ctx)
	if incr.Err() == nil && incr.Val() >= failThreshold {
		global.Redis.Set(ctx, lockKeyPrefix+username, "1", lockWindow)
	}
	writeLoginLog(username, ip, ua, false, "用户名或密码错误")
}

// 写入登录日志（同步；操作日志异步写入在日志阶段实现）
func writeLoginLog(username, ip, ua string, ok bool, msg string) {
	browser, os := utils.ParseUA(ua)
	status := int8(0) // 0 成功
	if !ok {
		status = 1 // 1 失败
	}
	_ = log.InsertLoginLog(&model.LoginLog{
		Username:  username,
		IP:        ip,
		IPAddr:    utils.ParseIPLocation(ip),
		Browser:   browser,
		OS:        os,
		Status:    status,
		Msg:       msg,
		LoginTime: time.Now(),
	})
}

// 是否在黑名单中
func isBlacklisted(ctx context.Context, jti string) bool {
	if jti == "" {
		return false
	}
	n, err := global.Redis.Exists(ctx, blacklistKeyPrefix+jti).Result()
	return err == nil && n > 0
}

// 令牌签发时间是否早于密码修改时间
func isPasswordChangedAfter(ctx context.Context, claims *utils.Claims) bool {
	if claims.IssuedAt == nil {
		return false
	}
	val, err := global.Redis.Get(ctx, pwchangedKeyPrefix+strconv.FormatUint(uint64(claims.UserID), 10)).Int64()
	if err != nil {
		return false // 未修改过密码
	}
	return claims.IssuedAt.Time.Before(time.Unix(val, 0))
}

// 取用户首个角色标识（多角色场景后续由权限中间件处理）
func primaryRoleKey(roles []*model.Role) string {
	if len(roles) == 0 {
		return ""
	}
	return roles[0].RoleKey
}

// 密码复杂度：>=8 位且含字母与数字
func validPassword(p string) bool {
	if len(p) < 8 {
		return false
	}
	var hasLetter, hasDigit bool
	for _, r := range p {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z':
			hasLetter = true
		case r >= '0' && r <= '9':
			hasDigit = true
		}
	}
	return hasLetter && hasDigit
}
