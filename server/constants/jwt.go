package constants

// TokenType 令牌类型
type TokenType int

const (
	AccessToken TokenType = iota
	RefreshToken
)
