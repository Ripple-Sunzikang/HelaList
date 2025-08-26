package common

import "github.com/gin-gonic/gin"

// common负责在handle部分返回通用的response信息

// Response类型
type Response[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

// 这里Response里带string只是因为我在测试。
func ErrorResponse(c *gin.Context, err error, code int) {
	c.JSON(200, Response[string]{
		Code:    code,
		Message: "error",
		Data:    "error data",
	})
	c.Abort()
}

func SuccessResponse(c *gin.Context) {
	c.JSON(200, Response[string]{
		Code:    200,
		Message: "success",
		Data:    "success data",
	})
}
