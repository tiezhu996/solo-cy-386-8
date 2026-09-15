package util

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构：{ "code": 0, "message": "ok", "data": ... }。
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// OK 成功响应。
func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{Code: 0, Message: "ok", Data: data})
}

// OKMessage 成功响应（带自定义文案）。
func OKMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{Code: 0, Message: message, Data: data})
}

// Fail 失败响应。
func Fail(c *gin.Context, httpStatus, code int, message string) {
	c.AbortWithStatusJSON(httpStatus, Response{Code: code, Message: message})
}
