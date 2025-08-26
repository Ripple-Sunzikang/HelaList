package op

import "HelaList/internal/model"

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
