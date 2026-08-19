package utils

import (
	"bytes"
	"encoding/json"
	"strings"
)

// 敏感字段判定规则：键名包含以下任意子串（不区分大小写）即整值打码。
// 用于操作日志 OperParam 落库前脱敏，避免密码/令牌/密钥明文写入日志。
var sensitiveKeyParts = []string{"password", "token", "secret"}

// 对 JSON 请求体做脱敏后返回字符串。
// 空 body 返回空串；非 JSON 或序列化失败时原样返回。递归处理嵌套对象与数组。
func MaskSensitiveJSON(body []byte) string {
	if len(bytes.TrimSpace(body)) == 0 {
		return ""
	}
	var v any
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.UseNumber() // 保留数字原样格式，避免 float64 重排
	if err := dec.Decode(&v); err != nil {
		return string(body)
	}
	maskSensitive(v)
	out, err := json.Marshal(v)
	if err != nil {
		return string(body)
	}
	return string(out)
}

func maskSensitive(v any) {
	switch t := v.(type) {
	case map[string]any:
		for k, val := range t {
			if isSensitiveKey(k) {
				t[k] = "******"
			} else {
				maskSensitive(val)
			}
		}
	case []any:
		for _, item := range t {
			maskSensitive(item)
		}
	}
}

func isSensitiveKey(k string) bool {
	lower := strings.ToLower(k)
	for _, s := range sensitiveKeyParts {
		if strings.Contains(lower, s) {
			return true
		}
	}
	return false
}
