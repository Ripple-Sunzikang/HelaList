package database

import (
	"HelaList/internal/bootstrap"
	"HelaList/internal/model"
	"errors"
	"fmt"
)

// 与数据库交互得到用户数据

func CreateUser(user *model.User) error {
	if user.Password == "" {
		return errors.New("Password cannot be empty.")
	}

	if err := user.SetPassword(user.Password); err != nil {
		return fmt.Errorf("failed to set password: %w", err)
	}

	result := bootstrap.Db.Create(user)

	if result.Error != nil {
		return fmt.Errorf("failed to create user: %w", result.Error)
	}

	return nil
}

