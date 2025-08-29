package model

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"

	"github.com/OpenListTeam/OpenList/v4/pkg/utils"
	uuid "github.com/google/uuid"
	"golang.org/x/crypto/argon2"
	"gorm.io/gorm"
)

type User struct {

	// 属性名首字母大写，表示该属性对外导出，类似public。小写则是private
	Id           uuid.UUID `gorm:"type:uuid;primarykey" json:"id"`
	Username     string    `gorm:"unique;not null;size:50" json:"username"`
	Email        string    `gorm:"unique;not null;size:100" json:"email"`
	PasswordHash string    `gorm:"not null" json:"-"`                      // 密码哈希值
	Salt         string    `gorm:"unique;not null" json:"-"`               // 每个用户的Salt唯一，防止彩虹表攻击
	Password     string    `gorm:"-" json:"password"`                      // 明文密码
	BasePath     string    `json:"base_path"`                              // 用户的基础路径
	Identity     int       `gorm:"not null" json:"identity"`               // 区分管理员和用户，0是Admin，1是Guest
	Disabled     bool      `gorm:"not null;default:false" json:"disabled"` // 用户是否被禁用
	PasswordTS   int64     `json:"password_ts"`                            // 密码时间戳，用于验证密码是否更改
}

// 用于指定模型对应的数据库表名，模型的属性也会自动转化为列。默认为蛇形复数形式。
func (User) TableName() string {
	return "users"
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.Id == uuid.Nil {
		newUUID, err := uuid.NewV7()
		if err != nil {
			return err
		}
		u.Id = newUUID
	}
	return nil
}

// 用户身份相关

const (
	ADMIN = iota //语法糖，ADMIN=0, GUEST=1, GENERAL=2
	GENERAL
	GUEST
)

func (u *User) IsAdmin() bool {
	return u.Identity == ADMIN
}

func (u *User) IsGuest() bool {
	return u.Identity == GUEST
}

// func (u *User) CanWrite() bool {
// 	return !u.Disabled && (u.Identity == ADMIN || u.Identity == GENERAL)
// }

// 密码加密相关

// argon2密码加密所需参数
const (
	argon2Iterations  = 1         // 迭代次数
	argon2SaltLength  = 16        // Salt长度
	argon2Parallelism = 4         // 使用的线程数
	argon2Memory      = 64 * 1024 // 内存消耗
	argon2KeyLength   = 32        // 哈希值的长度
)

// 根据密码明文计算哈希值
func (u *User) SetPassword(password string) error {
	salt := make([]byte, argon2SaltLength)

	if _, err := rand.Read(salt); err != nil {
		return err
	}

	// 计算哈希值
	hash := argon2.IDKey([]byte(password), salt, argon2Iterations, argon2Memory, argon2Parallelism, argon2KeyLength)

	// 存储用户的Salt和哈希值
	u.Salt = base64.RawStdEncoding.EncodeToString(salt)
	u.PasswordHash = base64.RawStdEncoding.EncodeToString(hash)

	return nil
}

func (u *User) CheckPassword(password string) (bool, error) {
	salt, err := base64.RawStdEncoding.DecodeString(u.Salt)
	if err != nil {
		return false, fmt.Errorf("decode Salt error: %w", err)
	}

	hash, err := base64.RawStdEncoding.DecodeString(u.PasswordHash)
	if err != nil {
		return false, fmt.Errorf("decode PasswordHash error: %w", err)
	}

	// 哈希值检验
	comparisonHash := argon2.IDKey([]byte(password), salt, argon2Iterations, argon2Memory, argon2Parallelism, argon2KeyLength)

	// 匹配执行时间基本相同，防止攻击者靠响应时间来猜测密码正确性
	if subtle.ConstantTimeCompare(hash, comparisonHash) == 1 {
		return true, nil
	}

	return false, nil
}

// 路径合并功能，把请求path加到BasePath的后缀
func (u *User) JoinPath(reqPath string) (string, error) {
	return utils.JoinBasePath(u.BasePath, reqPath)
}
