package repository

import (
	"HelaList/internal/bootstrap"
	"HelaList/internal/model"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

// 与数据库交互得到用户数据
/*
其实这带出来一个问题，就是数据库内容的检验应该放在哪里。
*/
func CreateUser(u *model.User) error {
	return errors.WithStack(bootstrap.Db.Create(u).Error)
	return nil
}

func UpdateUser(u *model.User) error {
	return errors.WithStack(bootstrap.Db.Save(u).Error)
}

func GetUserByIdentity(identity int) (*model.User, error) {
	user := model.User{Identity: identity}
	if err := bootstrap.Db.Where(user).Take(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func GetUserByName(username string) (*model.User, error) {
	user := model.User{Username: username}
	if err := bootstrap.Db.Where(user).First(&user).Error; err != nil {
		return nil, errors.Wrapf(err, "failed find user")
	}
	return &user, nil
}

func GetUserById(id uuid.UUID) (*model.User, error) {
	var u model.User
	if err := bootstrap.Db.First(&u, id).Error; err != nil {
		return nil, errors.Wrapf(err, "failed get old user")
	}
	return &u, nil
}

func DeleteUserById(id uuid.UUID) error {
	return errors.WithStack(bootstrap.Db.Delete(&model.User{}, id).Error)
}

func DeleteUserByUsername(username string) error {
	return errors.WithStack(bootstrap.Db.Delete(&model.User{}, username).Error)
}
