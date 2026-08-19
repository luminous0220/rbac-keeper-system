package utils

import "golang.org/x/crypto/bcrypt"

const cost = bcrypt.DefaultCost

// Hash 对明文密码加盐哈希
func Hash(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Verify 校验明文密码与哈希是否匹配
func Verify(hashed, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password)) == nil
}
