package handler

import (
	"HelaList/internal/model"
	"HelaList/internal/op"
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func CreateUser(c *gin.Context) {
	var request model.User
	if err := c.ShouldBind(&request); err != nil {
		ErrorResponse(c, err, 400)
		return
	}
	if request.IsAdmin() || request.IsGuest() {
		ErrorResponse(c, errors.New("error"), 400)
		return
	}
	request.SetPassword(request.Password)
	request.Password = ""
	if err := op.CreateUser(&request); err != nil {
		ErrorResponse(c, err, 500, true)
	} else {
		SuccessResponse(c)
	}
}

func UpdateUser(c *gin.Context) {
	var request model.User
	if err := c.ShouldBind(&request); err != nil {
		ErrorResponse(c, err, 400)
		return
	}
	user, err := op.GetUserById(request.Id)
	if err != nil {
		ErrorResponse(c, err, 500)
		return
	}
	if user.Identity != request.Identity {
		ErrorResponse(c, errors.New("role can not be changed"), 400)
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
		ErrorResponse(c, err, 500)
	} else {
		SuccessResponse(c)
	}
}

func DeleteUser(c *gin.Context) {
	idStr := c.Query("id")
	if idStr == "" {
		ErrorResponse(c, errors.New("missing id"), 400)
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		ErrorResponse(c, err, 400)
		return
	}

	if err := op.DeleteUserById(id); err != nil {
		ErrorResponse(c, err, 500)
		return
	}
	SuccessResponse(c)
}

func GetUser(c *gin.Context) {
	idStr := c.Query("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		ErrorResponse(c, err, 400)
		return
	}
	user, err := op.GetUserById(uuid.UUID(id))
	if err != nil {
		ErrorResponse(c, err, 500, true)
		return
	}
	SuccessResponse(c, user)
}
