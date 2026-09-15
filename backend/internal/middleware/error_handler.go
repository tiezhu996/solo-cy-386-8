package middleware

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/marketpal/marketpal/internal/constants"
	"github.com/marketpal/marketpal/internal/util"
)

// ErrorHandler 全局错误处理中间件：将 panic 与 c.Error 统一转为标准 JSON 响应。
func ErrorHandler(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("request panic recovered", "path", c.Request.URL.Path, "err", r)
				util.Fail(c, http.StatusInternalServerError, constants.CodeInternalError, "服务器内部错误，请稍后重试")
			}
		}()
		c.Next()
		if len(c.Errors) == 0 {
			return
		}
		err := c.Errors.Last().Err
		var appErr *util.AppError
		if errors.As(err, &appErr) {
			status := http.StatusBadRequest
			switch appErr.Code {
			case constants.CodeUnauthorized:
				status = http.StatusUnauthorized
			case constants.CodeForbidden:
				status = http.StatusForbidden
			case constants.CodeNotFound, constants.CodeUserNotFound, constants.CodeProductNotFound,
				constants.CodeOrderNotFound, constants.CodeAddressNotFound, constants.CodeCartItemNotFound,
				constants.CodeMessageNotFound:
				status = http.StatusNotFound
			case constants.CodeConflict, constants.CodeOrderStateInvalid, constants.CodeProductSold,
				constants.CodeReviewExists:
				status = http.StatusConflict
			case constants.CodeInternalError:
				status = http.StatusInternalServerError
			}
			util.Fail(c, status, appErr.Code, appErr.Message)
			return
		}
		logger.Error("unhandled error", "path", c.Request.URL.Path, "err", err)
		util.Fail(c, http.StatusInternalServerError, constants.CodeInternalError, "服务器内部错误，请稍后重试")
	}
}
