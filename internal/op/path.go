package op

import (
	"HelaList/internal/driver"
	stdpath "path"
	"strings"

	"github.com/OpenListTeam/OpenList/v4/pkg/generic_sync"
	"github.com/OpenListTeam/OpenList/v4/pkg/utils"
	"github.com/pkg/errors"
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
	// log.Debugln("use storage: ", storage.GetStorage().MountPath)
	mountPath := utils.GetActualMountPath(storage.GetStorage().MountPath)
	actualPath = utils.FixAndCleanPath(strings.TrimPrefix(rawPath, mountPath))
	return
}

// urlTreeSplitLineFormPath 分割path中分割真实路径和UrlTree定义字符串
func urlTreeSplitLineFormPath(path string) (pp string, file string) {
	// url.PathUnescape 会移除 // ，手动加回去
	path = strings.Replace(path, "https:/", "https://", 1)
	path = strings.Replace(path, "http:/", "http://", 1)
	if strings.Contains(path, ":https:/") || strings.Contains(path, ":http:/") {
		// URL-Tree模式 /url_tree_drivr/file_name[:size[:time]]:https://example.com/file
		fPath := strings.SplitN(path, ":", 2)[0]
		pp, _ = stdpath.Split(fPath)
		file = path[len(pp):]
	} else if strings.Contains(path, "/https:/") || strings.Contains(path, "/http:/") {
		// URL-Tree模式 /url_tree_drivr/https://example.com/file
		index := strings.Index(path, "/http://")
		if index == -1 {
			index = strings.Index(path, "/https://")
		}
		pp = path[:index]
		file = path[index+1:]
	} else {
		pp, file = stdpath.Split(path)
	}
	if pp == "" {
		pp = "/"
	}
	return
}

var balanceMap generic_sync.MapOf[string, int]

// GetBalancedStorage get storage by path
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
		i, _ := balanceMap.LoadOrStore(virtualPath, 0)
		i = (i + 1) % storageNum
		balanceMap.Store(virtualPath, i)
		return storages[i]
	}
}
