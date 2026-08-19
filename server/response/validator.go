package response

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

// translate 将 validator 校验错误翻译为可读中文提示（取首条错误）
func translate(errs validator.ValidationErrors) string {
	if len(errs) == 0 {
		return "参数校验失败"
	}
	e := errs[0]
	switch e.Tag() {
	case "required":
		return fmt.Sprintf("字段 %s 不能为空", e.Field())
	case "min":
		return fmt.Sprintf("字段 %s 长度/值不能小于 %s", e.Field(), e.Param())
	case "max":
		return fmt.Sprintf("字段 %s 长度/值不能大于 %s", e.Field(), e.Param())
	case "len":
		return fmt.Sprintf("字段 %s 长度必须为 %s", e.Field(), e.Param())
	case "oneof":
		return fmt.Sprintf("字段 %s 只能为 %s 之一", e.Field(), e.Param())
	case "optional_enum":
		return fmt.Sprintf("字段 %s 只能为 %s 之一", e.Field(), e.Param())
	case "email":
		return fmt.Sprintf("字段 %s 邮箱格式不正确", e.Field())
	case "alpha":
		return fmt.Sprintf("字段 %s 仅允许纯英文", e.Field())
	case "alphanum":
		return fmt.Sprintf("字段 %s 仅允许纯英文和数字", e.Field())
	case "required_if":
		return fmt.Sprintf("字段 %s 必须传入", e.Field())
	default:
		return fmt.Sprintf("字段 %s 校验失败: %s=%s", e.Field(), e.Tag(), e.Param())
	}
}
