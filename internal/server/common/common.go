package common

import (
	"HelaList/configs"
	"context"
	"log"

	"github.com/gin-gonic/gin"
)

type Response[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

func ErrorResponse(c *gin.Context, err error, code int, l ...bool) {
	// 打印详细错误到服务器控制台
	log.Printf("Error response: [code %d] %v", code, err)

	c.JSON(code, Response[interface{}]{ // 注意：将HTTP状态码也改为传入的code
		Code:    code,
		Message: err.Error(), // 将具体的错误信息返回
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

func GinWithValue(c *gin.Context, keyAndValue ...any) {
	c.Request = c.Request.WithContext(
		ContentWithValue(c.Request.Context(), keyAndValue...),
	)
}

func ContentWithValue(ctx context.Context, keyAndValue ...any) context.Context {
	if len(keyAndValue) < 1 || len(keyAndValue)%2 != 0 {
		panic("keyAndValue must be an even number of arguments (key, value, ...)")
	}
	for len(keyAndValue) > 0 {
		ctx = context.WithValue(ctx, keyAndValue[0], keyAndValue[1])
		keyAndValue = keyAndValue[2:]
	}
	return ctx
}

func GetApiUrl(ctx context.Context) string {
	api, _ := ctx.Value(configs.ApiUrlKey).(string)
	return api
}
