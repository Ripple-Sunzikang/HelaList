package op

import (
	"HelaList/internal/model"
	"HelaList/internal/driver"
)

// 将Storage保存到数据库，并返回数据库的uuid
// 然后匹配对应Drive,并保存到内存中
func CreateStorage(ctx context.Context, storage model.Storage) (uuid.UUID, error) {
	storage.ModifiedTime = time.Now()
	// MountPath挂载点不用动，因为你还没写Path的规范化函数。
	driverName := storage.Driver
	driverNew, err := GetDriver(driverName)
}

// 起名恐惧症了，因为
func GetVirtualObjsByPath(path string) []model.Obj {
	objs := make([]model.Obj, 0)
	storages :=
}
