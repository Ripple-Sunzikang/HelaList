package handler

import (
	"HelaList/internal/model"
	"HelaList/internal/op"
	"HelaList/internal/server/common"
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string      `json:"token"`
	User  *model.User `json:"user"`
}

func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ErrorResponse(c, err, 400)
		return
	}

	// 获取用户
	user, err := op.GetUserByName(req.Username)
	if err != nil {
		common.ErrorResponse(c, errors.New("用户名或密码错误"), 401)
		return
	}

	// 验证密码
	if ok, err := user.CheckPassword(req.Password); !ok || err != nil {
		common.ErrorResponse(c, errors.New("用户名或密码错误"), 401)
		return
	}

	// 检查用户是否被禁用
	if user.Disabled {
		common.ErrorResponse(c, errors.New("用户已被禁用"), 403)
		return
	}

	// 生成JWT token
	token, err := common.GenerateToken(user)
	if err != nil {
		common.ErrorResponse(c, err, 500)
		return
	}

	response := LoginResponse{
		Token: token,
		User:  user,
	}
	common.SuccessResponse(c, response)
}

func Logout(c *gin.Context) {
	// 从请求头获取token
	token := c.GetHeader("Authorization")
	if token != "" && len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
		common.InvalidateToken(token)
	}
	common.SuccessResponse(c, "登出成功")
}

func CreateUser(c *gin.Context) {
	var request model.User
	if err := c.ShouldBind(&request); err != nil {
		common.ErrorResponse(c, err, 400)
		return
	}
	request.SetPassword(request.Password)
	request.Password = ""
	if err := op.CreateUser(&request); err != nil {
		common.ErrorResponse(c, err, 500, true)
	} else {
		common.SuccessResponse(c)
	}
}

func UpdateUser(c *gin.Context) {
	var request model.User
	if err := c.ShouldBind(&request); err != nil {
		common.ErrorResponse(c, err, 400)
		return
	}
	user, err := op.GetUserById(request.Id)
	if err != nil {
		common.ErrorResponse(c, err, 500)
		return
	}
	if user.Identity != request.Identity {
		common.ErrorResponse(c, errors.New("role can not be changed"), 400)
		return
	}
	if request.Password == "" {
		request.PasswordHash = user.PasswordHash
		request.Salt = user.Salt
	} else {
		request.SetPassword(request.Password)
		request.Password = ""
	}
	if err := op.UpdateUser(&request); err != nil {
		common.ErrorResponse(c, err, 500)
	} else {
		common.SuccessResponse(c)
	}
}

func DeleteUser(c *gin.Context) {
	idStr := c.Query("id")
	if idStr == "" {
		common.ErrorResponse(c, errors.New("missing id"), 400)
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		common.ErrorResponse(c, err, 400)
		return
	}

	if err := op.DeleteUserById(id); err != nil {
		common.ErrorResponse(c, err, 500)
		return
	}
	common.SuccessResponse(c)
}

func GetUser(c *gin.Context) {
	idStr := c.Query("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		common.ErrorResponse(c, err, 400)
		return
	}
	user, err := op.GetUserById(uuid.UUID(id))
	if err != nil {
		common.ErrorResponse(c, err, 500, true)
		return
	}
	common.SuccessResponse(c, user)
}
