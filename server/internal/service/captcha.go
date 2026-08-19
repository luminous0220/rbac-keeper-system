package service

import (
	"context"
	"rbackeeper/global"
	"sync"
	"time"

	"github.com/mojocn/base64Captcha"
)

var (
	once          sync.Once
	captchaIns    *base64Captcha.Captcha
	captchaExpire = 5 * time.Minute
	CaptchaKey    = "captcha:"
)

type redisCaptchaStore struct {
	expire time.Duration
}

func (s *redisCaptchaStore) Set(id, value string) error {
	return global.Redis.Set(context.Background(), CaptchaKey+id, value, s.expire).Err()
}

func (s *redisCaptchaStore) Get(id string, clear bool) string {
	ctx := context.Background()
	key := CaptchaKey + id
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
	// 初始化
	var driver = base64Captcha.NewDriverDigit(
		80,  // 图片高度 ImgHeight
		240, // 图片宽度 ImgWidth
		4,   // 验证码字符长度 KeyLong
		0.7, // 干扰线条密度 0~1，越大越模糊
		80,  // 字体大小
	)
	captchaIns = base64Captcha.NewCaptcha(driver, store)
}

// 生成验证码，返回 id 与 base64 图片
func GenerateCaptcha() (id, b64s string, err error) {
	once.Do(initCaptcha)
	id, b64s, _, err = captchaIns.Generate()
	return
}

// 校验验证码（一次性，校验后即删）
func VerifyCaptcha(id, answer string) bool {
	once.Do(initCaptcha)
	return captchaIns.Verify(id, answer, true)
}
