package op

import (
	"HelaList/internal/driver"
	"strings"
	"sync/atomic"

	"github.com/OpenListTeam/OpenList/v4/pkg/generic_sync"
	"github.com/OpenListTeam/OpenList/v4/pkg/utils"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func GetStorageAndActualPath(rawPath string) (storage driver.Driver, actualPath string, err error) {
	rawPath = utils.FixAndCleanPath(rawPath)
	storage = GetBalancedStorage(rawPath)
	if storage == nil {
		if rawPath == "/" {
			err = errors.New("please add a storage first")
			return
		}
		err = errors.Errorf("storage not found for rawPath: %s", rawPath)
		return
	}
	logrus.Debugln("use storage: ", storage.GetStorage().MountPath)
	mountPath := utils.GetActualMountPath(storage.GetStorage().MountPath)
	actualPath = utils.FixAndCleanPath(strings.TrimPrefix(rawPath, mountPath))
	return
}

// 当多个虚拟网盘挂载到同一个文件下时，如果要查找网盘，需要轮询查找
// 这里我有往上提交pr,建议改成*int64进行原子操作，避免高并行出现混乱
var balanceMap generic_sync.MapOf[string, *int64]

func GetBalancedStorage(path string) driver.Driver {
	path = utils.FixAndCleanPath(path)
	storages := getStoragesByPath(path)
	storageNum := len(storages)
	switch storageNum {
	case 0:
		return nil
	case 1:
		return storages[0]
	default:
		virtualPath := utils.GetActualMountPath(storages[0].GetStorage().MountPath)
		// 如果不存在则存入一个新的 *int64(0)，LoadOrStore 会返回已存在的值或新值
		p, _ := balanceMap.LoadOrStore(virtualPath, new(int64))
		// 原子自增并转换为 0 基索引
		idx := atomic.AddInt64(p, 1) - 1
		i := int(idx % int64(storageNum))
		return storages[i]
	}
}
