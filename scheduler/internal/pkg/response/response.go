package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Body[T any] struct {
	Code    int    `json:"code"`
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    T      `json:"data,omitempty"`
}

type PageData[T any] struct {
	List  []T `json:"list"`
	Total int `json:"total"`
	Page  int `json:"page"`
	Size  int `json:"size"`
}

func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Body[any]{Code: 200, Success: true, Message: "success", Data: data})
}

func Page[T any](c *gin.Context, list []T, total, page, size int) {
	if list == nil {
		list = []T{}
	}
	OK(c, PageData[T]{List: list, Total: total, Page: page, Size: size})
}

func Fail(c *gin.Context, code int, message string) {
	c.JSON(http.StatusOK, Body[any]{Code: code, Success: false, Message: message})
}

func Unauthorized(c *gin.Context, message string) {
	if message == "" {
		message = "未授权"
	}
	c.JSON(http.StatusUnauthorized, Body[any]{Code: 401, Success: false, Message: message})
}
