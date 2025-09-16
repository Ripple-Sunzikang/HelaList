package op

import (
	"HelaList/internal/driver"
	"HelaList/internal/model"
)

// 给Obj用的Hook
type ObjsUpdateHook = func(parent string, objs []model.Obj)

var (
	objsUpdateHooks = make([]ObjsUpdateHook, 0)
)

func RegisterObjsUpdateHook(hook ObjsUpdateHook) {
	objsUpdateHooks = append(objsUpdateHooks, hook)
}

func HandleObjsUpdateHook(parent string, objs []model.Obj) {
	for _, hook := range objsUpdateHooks {
		hook(parent, objs)
	}
}

// Storage用的Hook
type StorageHook func(typ string, storage driver.Driver)

var storageHooks = make([]StorageHook, 0)

func callStorageHooks(typ string, storage driver.Driver) {
	for _, hook := range storageHooks {
		hook(typ, storage)
	}
}

func RegisterStorageHook(hook StorageHook) {
	storageHooks = append(storageHooks, hook)
}
