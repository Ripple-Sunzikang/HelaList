package handler

import (
	"github.com/gin-gonic/gin"
)

type Response[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

func ErrorResponse(c *gin.Context, err error, code int, l ...bool) {
	c.JSON(200, Response[interface{}]{
		Code:    code,
		Message: "error",
		Data:    nil,
	})
	c.Abort()
}

func SuccessResponse(c *gin.Context, data ...interface{}) {
	var respData interface{}
	if len(data) > 0 {
		respData = data[0]
	}

	c.JSON(200, Response[interface{}]{
		Code:    200,
		Message: "success",
		Data:    respData,
	})
}
