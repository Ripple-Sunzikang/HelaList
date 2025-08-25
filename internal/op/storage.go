package op

import (
	"HelaList/internal/model"
	"HelaList/internal/pkg"
	"context"
	"sort"
	"strings"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
)

// 将Storage保存到数据库，并返回数据库的uuid
// 然后匹配对应Drive,并保存到内存中
func CreateStorage(ctx context.Context, storage model.Storage) (uuid.UUID, error) {
	storage.ModifiedTime = time.Now()
	// MountPath挂载点不用动，因为你还没写Path的规范化函数。
	driverName := storage.Driver
	driverNew, err := GetDriver(driverName)

}

var storagesMap = pkg.NewSyncStorageMap()

// 根据给定路径，返回该路径下的所有虚拟文件夹
func GetVirtualObjsByPath(path string) []model.Obj {
	objs := make([]model.Obj, 0)
	storages := storagesMap.Values()

	// 排序，按Order、MountPath顺序排列
	sort.Slice(storages, func(i int, j int) bool {
		if storages[i].GetStorage().Order == storages[j].GetStorage().Order {
			return storages[i].GetStorage().MountPath < storages[j].GetStorage().MountPath
		}
		return storages[i].GetStorage().Order < storages[j].GetStorage().Order
	})

	// 用于存放路径下的文件名
	mapSet := mapset.NewSet[string]()

	for _, v := range storages {
		mountPath := v.GetStorage().MountPath
		/*
			// 需要过滤掉不需要的存储挂载路径，只保留path的直接子目录的挂载点，所以用一个if判别
			// 但是IsSubPath你只在lock.go里实现了一个内部函数。
			if len(path) >= len(mountPath) || IsSubPath(path, mountPath) {
				continue
			}
		*/
		// 从挂载路径中提取相对于path的第一级目录名，也就是路径下的文件名
		name := strings.SplitN(strings.TrimPrefix(mountPath[len(path):], "/"), "/", 2)[0]
		if mapSet.Add(name) {
			objs = append(objs, &model.Object{
				Name:         name,
				Size:         0,
				ModifiedTime: v.GetStorage().ModifiedTime,
				IsFolder:     true,
			})
		}
	}
	return objs
}
