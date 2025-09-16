package service

// repository层负责实现最基本的、近乎原语的对数据库的操作，而service层是在repository层的基础上，进一步封装了数据合法性的检测等功能。
import (
	"HelaList/internal/model"
	"HelaList/internal/repository"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

// CreateUser 调用 repository 创建用户，进行简单校验并包装错误
func CreateUser(u *model.User) error {
	if u == nil {
		return errors.New("user is nil")
	}
	if err := repository.CreateUser(u); err != nil {
		return errors.Wrapf(err, "failed create user")
	}
	return nil
}

// UpdateUser 调用 repository 更新用户并包装错误
func UpdateUser(u *model.User) error {
	if u == nil {
		return errors.New("user is nil")
	}
	if err := repository.UpdateUser(u); err != nil {
		return errors.Wrapf(err, "failed update user")
	}
	return nil
}

// GetUserByIdentity 通过身份获取用户并包装错误
func GetUserByIdentity(identity int) (*model.User, error) {
	u, err := repository.GetUserByIdentity(identity)
	if err != nil {
		return nil, errors.Wrapf(err, "failed get user by identity")
	}
	return u, nil
}

// GetUserByName 通过用户名获取用户并包装错误
func GetUserByName(username string) (*model.User, error) {
	u, err := repository.GetUserByName(username)
	if err != nil {
		return nil, errors.Wrapf(err, "failed find user")
	}
	return u, nil
}

// GetUserById 通过 ID 获取用户并包装错误
func GetUserById(id uuid.UUID) (*model.User, error) {
	u, err := repository.GetUserById(id)
	if err != nil {
		return nil, errors.Wrapf(err, "failed get old user")
	}
	return u, nil
}

// DeleteUserById 通过 ID 删除用户并包装错误
func DeleteUserById(id uuid.UUID) error {
	if err := repository.DeleteUserById(id); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// DeleteUserByUsername 通过用户名删除用户并包装错误
func DeleteUserByUsername(username string) error {
	if err := repository.DeleteUserByUsername(username); err != nil {
		return errors.WithStack(err)
	}
	return nil
}
