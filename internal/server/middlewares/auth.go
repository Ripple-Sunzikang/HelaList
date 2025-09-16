package middlewares

import (
	"HelaList/configs"
	"HelaList/internal/model"
	"HelaList/internal/op"
	"HelaList/internal/server/common"
	"errors"

	"github.com/gin-gonic/gin"
)

func Auth(required bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取token
		token := c.GetHeader("Authorization")
		if token != "" && len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		} else {
			// 尝试从查询参数获取token
			token = c.Query("token")
		}

		if token == "" {
			if required {
				common.ErrorResponse(c, errors.New("未提供认证令牌"), 401)
				c.Abort()
				return
			}
			// 如果不需要认证且没有token，使用guest用户
			guestUser, err := op.GetGuest()
			if err != nil {
				common.ErrorResponse(c, errors.New("系统错误"), 500)
				c.Abort()
				return
			}
			c.Request = c.Request.WithContext(common.ContentWithValue(c.Request.Context(), configs.UserKey, guestUser))
			c.Next()
			return
		}

		// 解析token
		claims, err := common.ParseToken(token)
		if err != nil {
			if required {
				// 【关键修改】将来自ParseToken的原始错误err直接传递出去
				common.ErrorResponse(c, err, 401)
				c.Abort()
				return
			}
			// 如果不需要认证且token无效，使用guest用户
			guestUser, err := op.GetGuest()
			if err != nil {
				common.ErrorResponse(c, errors.New("系统错误"), 500)
				c.Abort()
				return
			}
			c.Request = c.Request.WithContext(common.ContentWithValue(c.Request.Context(), configs.UserKey, guestUser))
			c.Next()
			return
		}

		// 获取用户信息
		user, err := op.GetUserByName(claims.Username)
		if err != nil {
			if required {
				common.ErrorResponse(c, errors.New("用户不存在"), 401)
				c.Abort()
				return
			}
			// 如果不需要认证且用户不存在，使用guest用户
			guestUser, err := op.GetGuest()
			if err != nil {
				common.ErrorResponse(c, errors.New("系统错误"), 500)
				c.Abort()
				return
			}
			c.Request = c.Request.WithContext(common.ContentWithValue(c.Request.Context(), configs.UserKey, guestUser))
			c.Next()
			return
		}

		// 检查密码是否已更改
		if user.PasswordTS != claims.PwdTS {
			if required {
				common.ErrorResponse(c, errors.New("密码已更改，请重新登录"), 401)
				c.Abort()
				return
			}
			// 如果不需要认证且密码已更改，使用guest用户
			guestUser, err := op.GetGuest()
			if err != nil {
				common.ErrorResponse(c, errors.New("系统错误"), 500)
				c.Abort()
				return
			}
			c.Request = c.Request.WithContext(common.ContentWithValue(c.Request.Context(), configs.UserKey, guestUser))
			c.Next()
			return
		}

		// 检查用户是否被禁用
		if user.Disabled {
			if required {
				common.ErrorResponse(c, errors.New("用户已被禁用"), 403)
				c.Abort()
				return
			}
			// 如果不需要认证且用户被禁用，使用guest用户
			guestUser, err := op.GetGuest()
			if err != nil {
				common.ErrorResponse(c, errors.New("系统错误"), 500)
				c.Abort()
				return
			}
			c.Request = c.Request.WithContext(common.ContentWithValue(c.Request.Context(), configs.UserKey, guestUser))
			c.Next()
			return
		}

		// 将用户信息放入上下文
		c.Request = c.Request.WithContext(common.ContentWithValue(c.Request.Context(), configs.UserKey, user))
		c.Next()
	}
}

func AuthNotGuest(c *gin.Context) {
	user := c.Request.Context().Value(configs.UserKey).(*model.User)
	if user.IsGuest() {
		common.ErrorResponse(c, errors.New("访客用户无权限访问"), 403)
		c.Abort()
	} else {
		c.Next()
	}
}

func AuthAdmin(c *gin.Context) {
	user := c.Request.Context().Value(configs.UserKey).(*model.User)
	if !user.IsAdmin() {
		common.ErrorResponse(c, errors.New("需要管理员权限"), 403)
		c.Abort()
	} else {
		c.Next()
	}
}
