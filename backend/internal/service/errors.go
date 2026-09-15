package service

import "github.com/marketpal/marketpal/internal/util"

// utilAppError 便捷构造业务错误（service 层手动拼接 message，透传至 handler）。
func utilAppError(code int, message string, err error) error {
	return util.NewAppError(code, message, err)
}
