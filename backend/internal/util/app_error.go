package util

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

// AppError 业务错误：code 对应 constants 错误码，message 为透传提示文案。
type AppError struct {
	Code    int
	Message string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("code=%d message=%s cause=%v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("code=%d message=%s", e.Code, e.Message)
}

// Unwrap 支持 errors.Is/As 错误链。
func (e *AppError) Unwrap() error { return e.Err }

// NewAppError 构造业务错误，message 中必须包含实体名/字段名/角色名等上下文。
func NewAppError(code int, message string, err error) *AppError {
	return &AppError{Code: code, Message: message, Err: err}
}

// AbortWithError 由 handler 包装 service 返回的错误并输出统一响应（屎山约束：handler 必须再次包装）。
func AbortWithError(c *gin.Context, err error) {
	_ = c.Error(err)
}
