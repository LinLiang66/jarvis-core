package response

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"jarvis/backend/internal/pkg/serialize"
)

type Body[T any] struct {
	Code    int    `json:"code"`
	Success bool   `json:"success"` // 是否成功
	Message string `json:"message"`
	Data    T      `json:"data,omitempty"`
}

type PageData[T any] struct {
	List  []T `json:"list"`
	Total int `json:"total"`
	Page  int `json:"page"`
	Size  int `json:"size"`
}

func write(c *gin.Context, httpStatus int, body any) {
	b, err := serialize.MarshalJSON(body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "success": false, "message": "响应序列化失败"})
		return
	}
	c.Data(httpStatus, "application/json; charset=utf-8", b)
}

func OK(c *gin.Context, data any) {
	write(c, http.StatusOK, Body[any]{Code: 200, Success: true, Message: "success", Data: data})
}

func Page[T any](c *gin.Context, list []T, total, page, size int) {
	if list == nil {
		list = []T{}
	}
	OK(c, PageData[T]{List: list, Total: total, Page: page, Size: size})
}

func Fail(c *gin.Context, code int, message string) {
	write(c, http.StatusOK, Body[any]{Code: code, Success: false, Message: message})
}

func Unauthorized(c *gin.Context, message string) {
	if message == "" {
		message = "未授权"
	}
	write(c, http.StatusUnauthorized, Body[any]{Code: 401, Success: false, Message: message})
}
