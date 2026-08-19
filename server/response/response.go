package response

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// Result 统一响应结构
type Result[T any] struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data T      `json:"data,omitempty"`
}

// 错误码
const (
	Success            = 0
	ErrInvalidParam    = 40001
	ErrUnauthorized    = 40101
	ErrTokenExpired    = 40102
	ErrTokenInvalid    = 40103
	ErrForbidden       = 40301
	ErrTooManyRequests = 40302 // 限流：请求过于频繁
	ErrNotFound        = 40401
	ErrIdempotent      = 40002 // 幂等：重复提交
	ErrServer          = 50001
)

// BizError 业务错误，携带错误码与提示信息，供中间件/服务层统一处理
type BizError struct {
	Code int
	Msg  string
}

func (e BizError) Error() string { return e.Msg }

// NewBizError 构造业务错误
func NewBizError(code int, msg string) BizError {
	return BizError{Code: code, Msg: msg}
}

// OK 成功响应
func OK(ctx *gin.Context, data any) {
	ctx.JSON(http.StatusOK, Result[any]{Code: Success, Msg: "success", Data: data})
}

// OKMsg 成功响应（自定义提示）
func OKMsg(ctx *gin.Context, msg string, data any) {
	ctx.JSON(http.StatusOK, Result[any]{Code: Success, Msg: msg, Data: data})
}

// Fail 失败响应
func Fail(ctx *gin.Context, code int, msg string) {
	ctx.JSON(http.StatusOK, Result[any]{Code: code, Msg: msg})
}

// FailWithError 失败响应，自动识别 BizError
func FailWithError(ctx *gin.Context, err error) {
	var be BizError
	if errors.As(err, &be) {
		Fail(ctx, be.Code, be.Msg)
		return
	}
	Fail(ctx, ErrServer, err.Error())
}

// 绑定并校验 JSON 入参，失败时自动写入错误响应并返回 false
func BindJSON(ctx *gin.Context, obj any) bool {
	if err := ctx.ShouldBindJSON(obj); err != nil {
		var valErrs validator.ValidationErrors
		if errors.As(err, &valErrs) {
			Fail(ctx, ErrInvalidParam, translate(valErrs))
		} else {
			Fail(ctx, ErrInvalidParam, err.Error())
		}
		return false
	}
	return true
}

// 绑定并校验 Query 入参（GET 列表查询），失败时自动写入错误响应并返回 false
func BindQuery(ctx *gin.Context, obj any) bool {
	if err := ctx.ShouldBindQuery(obj); err != nil {
		var valErrs validator.ValidationErrors
		if errors.As(err, &valErrs) {
			Fail(ctx, ErrInvalidParam, translate(valErrs))
		} else {
			Fail(ctx, ErrInvalidParam, err.Error())
		}
		return false
	}
	return true
}
