package repository

import (
	"HelaList/internal/bootstrap"
	"HelaList/internal/model"
	"fmt"

	"github.com/google/uuid"
)

// 与数据库交互得到用户数据
/*
其实这带出来一个问题，就是数据库内容的检验应该放在哪里。
*/
func CreateUser(user *model.User) error {
	result := bootstrap.Db.Create(user)
	if result.Error != nil {
		return fmt.Errorf("failed to create user: %w", result.Error)
	}
	return nil
}

func UpdateUser(user *model.User) error {
	result := bootstrap.Db.Save(user)
	if result.Error != nil {
		return fmt.Errorf("failed to update user: %w", result.Error)
	}
	return nil
}

func DeleteUserById(id uuid.UUID) error {
	if err := bootstrap.Db.Delete(&model.User{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete mount: %w", err)
	}
	return nil
}

func DeleteUserByUsername(username string) error {
	if err := bootstrap.Db.Delete(&model.User{}, username).Error; err != nil {
		return fmt.Errorf("failed to delete mount: %w", err)
	}
	return nil
}
