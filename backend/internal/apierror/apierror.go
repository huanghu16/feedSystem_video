package apierror

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// APIError 是统一的错误响应结构
// 所有 API 错误都通过这个结构返回，保证前端收到的格式一致
type APIError struct {
	Code    int    `json:"code"`    // 业务错误码
	Message string `json:"message"` // 错误描述
}

// Response 是统一的 API 响应结构
type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"` // omitempty: 数据为 nil 时不输出
}

// OK 返回成功响应
func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

// Fail 返回业务错误响应
func Fail(c *gin.Context, httpCode int, bizCode int, msg string) {
	c.JSON(httpCode, Response{
		Code:    bizCode,
		Message: msg,
	})
}

// FailServer 返回服务器内部错误
func FailServer(c *gin.Context, msg string) {
	Fail(c, http.StatusInternalServerError, 500, msg)
}

// FailParam 返回参数错误
func FailParam(c *gin.Context, msg string) {
	Fail(c, http.StatusBadRequest, 400, msg)
}

// FailAuth 返回认证错误
func FailAuth(c *gin.Context, msg string) {
	Fail(c, http.StatusUnauthorized, 401, msg)
}
