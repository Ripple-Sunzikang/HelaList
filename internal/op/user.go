package op

import (
	"HelaList/internal/model"
	"HelaList/internal/service"
	"errors"
	"time"

	"github.com/OpenListTeam/OpenList/v4/pkg/singleflight"
	"github.com/OpenListTeam/OpenList/v4/pkg/utils"
	"github.com/OpenListTeam/go-cache"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// 需要修改
var userCache = cache.NewMemCache(cache.WithShards[*model.User](2))
var userG singleflight.Group[*model.User]
var guestUser *model.User
var adminUser *model.User

func GetAdmin() (*model.User, error) {
	if adminUser == nil {
		user, err := service.GetUserByIdentity(model.ADMIN)
		if err != nil {
			return nil, err
		}
		adminUser = user
	}
	return adminUser, nil
}

func GetGuest() (*model.User, error) {
	if guestUser == nil {
		user, err := service.GetUserByIdentity(model.GUEST)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// Create guest user
				guest := &model.User{
					Username: "guest",
					Email:    "guest@helalist.com",
					Identity: model.GUEST,
					Disabled: false,
					BasePath: "/",
				}
				guest.SetPassword("guest") // Or some random password
				if err := service.CreateUser(guest); err != nil {
					return nil, err
				}
				guestUser = guest
				return guestUser, nil
			}
			return nil, err
		}
		guestUser = user
	}
	return guestUser, nil
}

func GetUserByIdentity(identity int) (*model.User, error) {
	return service.GetUserByIdentity(identity)
}

func GetUserByName(username string) (*model.User, error) {
	if username == "" {
		return nil, errors.New("用户名为空")
	}
	if user, ok := userCache.Get(username); ok {
		return user, nil
	}
	user, err, _ := userG.Do(username, func() (*model.User, error) {
		_user, err := service.GetUserByName(username)
		if err != nil {
			return nil, err
		}
		userCache.Set(username, _user, cache.WithEx[*model.User](time.Hour))
		return _user, nil
	})
	return user, err
}

func GetUserById(id uuid.UUID) (*model.User, error) {
	return service.GetUserById(id)
}

func CreateUser(u *model.User) error {
	u.BasePath = utils.FixAndCleanPath(u.BasePath)
	return service.CreateUser(u)
}

func DeleteUserById(id uuid.UUID) error {
	old, err := service.GetUserById(id)
	if err != nil {
		return err
	}
	if old.IsAdmin() || old.IsGuest() {
		return errors.New("旧用户原来有身份的")
	}
	userCache.Del(old.Username)
	return service.DeleteUserById(id)
}

func UpdateUser(u *model.User) error {
	old, err := service.GetUserById(u.Id)
	if err != nil {
		return err
	}
	if u.IsAdmin() {
		adminUser = nil
	}
	if u.IsGuest() {
		guestUser = nil
	}
	userCache.Del(old.Username)
	u.BasePath = utils.FixAndCleanPath(u.BasePath)
	return service.UpdateUser(u)
}

func DelUserCache(username string) error {
	user, err := GetUserByName(username)
	if err != nil {
		return err
	}
	if user.IsAdmin() {
		adminUser = nil
	}
	if user.IsGuest() {
		guestUser = nil
	}
	userCache.Del(username)
	return nil
}
