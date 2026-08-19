package utils

// validPhone 手机号校验：11 位纯数字
func IsValidPhone(p string) bool {
	if len(p) != 11 {
		return false
	}
	for _, r := range p {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// 校验规则：只能包含英文字母和数字，且不能是纯数字
func IsAlphaNumeric(username string) bool {
	hasLetter := false

	// 2. 遍历字节
	for i := 0; i < len(username); i++ {
		c := username[i]

		// 如果是英文字母（大小写）
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
			hasLetter = true
			continue
		}

		// 如果是数字
		if c >= '0' && c <= '9' {
			continue
		}

		// 遇到了非法字符，直接返回 false
		return false
	}

	// 3. 必须包含至少一个英文字母（不能纯数字）
	return hasLetter
}
