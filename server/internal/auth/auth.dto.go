package auth

// LoginBody 登录请求
type LoginBody struct {
	Username     string `json:"username" binding:"required"`
	Password     string `json:"password" binding:"required"`
	CaptchaID    string `json:"captchaId" binding:"required"`
	CaptchaValue string `json:"captchaValue" binding:"required"`
}

// LoginResp 登录响应
type LoginResp struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	// Access Token 有效期（时间戳）
	ExpiresIn int64  `json:"expiresIn"`
	Username  string `json:"username"`
	Nickname  string `json:"nickname"`
}

// RefreshBody 刷新令牌请求
type RefreshBody struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

// RefreshResp 刷新令牌响应
type RefreshResp struct {
	AccessToken string `json:"accessToken"`
	ExpiresIn   int64  `json:"expiresIn"`
}

// ResetPasswordBody 密码重置请求（当前用户从 Token 取，不入参 user id）
type ResetPasswordBody struct {
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required,min=8,max=32,alphanum"`
}

// CaptchaResp 验证码响应
type CaptchaResp struct {
	CaptchaID string `json:"captchaId"`
	Img       string `json:"img"` // base64 图片
}

// ProfileBody 修改个人信息请求（当前用户从 Token 取，指针字段，nil=不改，zero=清空）
type ProfileBody struct {
	// 昵称
	Nickname *string `json:"nickname" binding:"omitempty,max=30"`
	// 手机号
	Phone *string `json:"phone" binding:"omitempty,max=11"`
	// 邮箱
	Email *string `json:"email" binding:"omitempty,max=50"`
	// 性别（0未知 1男 2女）
	Sex *int8 `json:"sex" binding:"omitempty,oneof=0 1 2"`
	// 头像地址
	Avatar *string `json:"avatar" binding:"omitempty,max=255"`
}
