package op

import (
	"HelaList/configs"
	"HelaList/internal/driver"
	"HelaList/internal/model"
	"HelaList/internal/service"
	"context"
	"fmt"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/OpenListTeam/OpenList/v4/pkg/generic_sync"
	"github.com/OpenListTeam/OpenList/v4/pkg/utils"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var storagesMap generic_sync.MapOf[string, driver.Driver]

func GetAllStorages() []driver.Driver {
	return storagesMap.Values()
}

func HasStorage(mountPath string) bool {
	return storagesMap.Has(utils.FixAndCleanPath(mountPath))
}

func GetStorageByMountPath(mountPath string) (driver.Driver, error) {
	mountPath = utils.FixAndCleanPath(mountPath)
	storageDriver, ok := storagesMap.Load(mountPath)
	if !ok {
		return nil, errors.Errorf("no mount path for an storage is: %s", mountPath)
	}
	return storageDriver, nil
}

func CreateStorage(ctx context.Context, storage model.Storage) (uuid.UUID, error) {
	storage.ModifiedTime = time.Now()
	storage.MountPath = utils.FixAndCleanPath(storage.MountPath)
	var err error
	driverName := storage.Driver
	driverNew, err := GetDriver(driverName)
	if err != nil {
		return uuid.Nil, errors.WithMessage(err, "failed get driver new")
	}
	storageDriver := driverNew()

	// 把storage插入数据库
	err = service.CreateStorage(&storage)
	if err != nil {
		return storage.Id, errors.WithMessage(err, "failed create storage in database")
	}

	err = initStorage(ctx, storage, storageDriver)
	go callStorageHooks("add", storageDriver)
	if err != nil {
		return storage.Id, errors.Wrap(err, "failed init storage but storage is already created")
	}
	logrus.Debugf("storage %+v is created", storageDriver)
	return storage.Id, nil
}

func LoadStorage(ctx context.Context, storage model.Storage) error {
	storage.MountPath = utils.FixAndCleanPath(storage.MountPath)
	driverName := storage.Driver
	driverNew, err := GetDriver(driverName)
	if err != nil {
		return errors.WithMessage(err, "failed get driver new")
	}
	storageDriver := driverNew()

	err = initStorage(ctx, storage, storageDriver)
	go callStorageHooks("add", storageDriver)
	logrus.Debugf("storage %+v is created", storageDriver)
	return err
}

func DeleteStorage(ctx context.Context, storageID string) error {
	// 解析UUID
	id, err := uuid.Parse(storageID)
	if err != nil {
		return errors.WithMessage(err, "invalid storage ID format")
	}

	// 从数据库获取存储信息
	storage, err := service.GetStorageById(id)
	if err != nil {
		return errors.WithMessage(err, "failed to get storage from database")
	}
	if storage == nil {
		return errors.New("storage not found")
	}

	// 从内存中移除存储驱动
	mountPath := utils.FixAndCleanPath(storage.MountPath)
	if storagesMap.Has(mountPath) {
		storageDriver, _ := storagesMap.Load(mountPath)
		storagesMap.Delete(mountPath)
		go callStorageHooks("remove", storageDriver)
		logrus.Debugf("storage %s removed from memory", mountPath)
	}

	// 从数据库删除存储记录
	err = service.DeleteStorageById(id)
	if err != nil {
		return errors.WithMessage(err, "failed to delete storage from database")
	}

	logrus.Infof("storage %s deleted successfully", mountPath)
	return nil
}

func getCurrentGoroutineStack() string {
	buf := make([]byte, 1<<16)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}

func initStorage(ctx context.Context, storage model.Storage, storageDriver driver.Driver) (err error) {
	storageDriver.SetStorage(storage)
	driverStorage := storageDriver.GetStorage()
	defer func() {
		if err := recover(); err != nil {
			errInfo := fmt.Sprintf("[panic] err: %v\nstack: %s\n", err, getCurrentGoroutineStack())
			fmt.Errorf("panic init storage: %s", errInfo)
			driverStorage.SetStatus(errInfo)
			MustSaveDriverStorage(storageDriver)
			storagesMap.Store(driverStorage.MountPath, storageDriver)
		}
	}()
	// Json解析
	err = utils.Json.UnmarshalFromString(driverStorage.Addition, storageDriver.GetAddition())
	if err == nil {
		if ref, ok := storageDriver.(driver.Reference); ok {
			if strings.HasPrefix(driverStorage.Remark, "ref:/") {
				refMountPath := driverStorage.Remark
				i := strings.Index(refMountPath, "\n")
				if i > 0 {
					refMountPath = refMountPath[4:i]
				} else {
					refMountPath = refMountPath[4:]
				}
				var refStorage driver.Driver
				refStorage, err = GetStorageByMountPath(refMountPath)
				if err != nil {
					err = fmt.Errorf("ref: %w", err)
				} else {
					err = ref.InitReference(refStorage)
					if err != nil {
						err = fmt.Errorf("ref: storage is not %s", storageDriver.Config().Name)
					}
				}
			}
		}
	}
	if err == nil {
		err = storageDriver.Init(ctx)
	}
	storagesMap.Store(driverStorage.MountPath, storageDriver)
	if err != nil {
		err = errors.Wrap(err, "failed init storage")
	} else {
		driverStorage.SetStatus(configs.WORK)
	}
	MustSaveDriverStorage(storageDriver)
	return err
}

func MustSaveDriverStorage(driver driver.Driver) {
	err := saveDriverStorage(driver)
	if err != nil {
		fmt.Errorf("failed save driver storage: %s", err)
	}
}
func saveDriverStorage(driver driver.Driver) error {
	storage := driver.GetStorage()
	addition := driver.GetAddition()
	str, err := utils.Json.MarshalToString(addition)
	if err != nil {
		return errors.Wrap(err, "error while marshal addition")
	}
	storage.Addition = str
	err = service.UpdateStorage(storage)
	if err != nil {
		return errors.WithMessage(err, "failed update storage in database")
	}
	return nil
}

func getStoragesByPath(path string) []driver.Driver {
	storages := make([]driver.Driver, 0)
	curSlashCount := 0
	storagesMap.Range(func(mountPath string, value driver.Driver) bool {
		mountPath = utils.GetActualMountPath(mountPath)
		if utils.IsSubPath(mountPath, path) {
			slashCount := strings.Count(utils.PathAddSeparatorSuffix(mountPath), "/")
			if slashCount > curSlashCount {
				storages = storages[:0]
				curSlashCount = slashCount
			}
			if slashCount == curSlashCount {
				storages = append(storages, value)
			}
		}
		return true
	})
	sort.Slice(storages, func(i, j int) bool {
		return storages[i].GetStorage().MountPath < storages[j].GetStorage().MountPath
	})
	return storages
}

func GetStorageVirtualFilesByPath(prefix string) []model.Obj {
	files := make([]model.Obj, 0)
	storages := storagesMap.Values()
	sort.Slice(storages, func(i, j int) bool {
		if storages[i].GetStorage().Order == storages[j].GetStorage().Order {
			return storages[i].GetStorage().MountPath < storages[j].GetStorage().MountPath
		}
		return storages[i].GetStorage().Order < storages[j].GetStorage().Order
	})

	prefix = utils.FixAndCleanPath(prefix)
	set := mapset.NewSet[string]()
	for _, v := range storages {
		mountPath := utils.GetActualMountPath(v.GetStorage().MountPath)
		if len(prefix) >= len(mountPath) || !utils.IsSubPath(prefix, mountPath) {
			continue
		}
		name := strings.SplitN(strings.TrimPrefix(mountPath[len(prefix):], "/"), "/", 2)[0]
		if set.Add(name) {
			files = append(files, &model.Object{
				Name:         name,
				Size:         0,
				ModifiedTime: v.GetStorage().ModifiedTime,
				IsFolder:     true,
			})
		}
	}
	return files
}

func UpdateStorage(ctx context.Context, storage model.Storage) error {
	oldStorage, err := service.GetStorageById(storage.Id)
	if err != nil {
		return errors.WithMessage(err, "failed get old storage")
	}
	if oldStorage.Driver != storage.Driver {
		return errors.Errorf("driver cannot be changed")
	}
	storage.ModifiedTime = time.Now()
	storage.MountPath = utils.FixAndCleanPath(storage.MountPath)
	err = service.UpdateStorage(&storage)
	if err != nil {
		return errors.WithMessage(err, "failed update storage in database")
	}
	if storage.Disabled {
		return nil
	}
	storageDriver, err := GetStorageByMountPath(oldStorage.MountPath)
	if oldStorage.MountPath != storage.MountPath {
		storagesMap.Delete(oldStorage.MountPath)
	}
	if err != nil {
		return errors.WithMessage(err, "failed get storage driver")
	}
	err = storageDriver.Drop(ctx)
	if err != nil {
		return errors.Wrapf(err, "failed drop storage")
	}

	err = initStorage(ctx, storage, storageDriver)
	go callStorageHooks("update", storageDriver)
	logrus.Debugf("storage %+v is update", storageDriver)
	return err
}

// 启动时初始化读取所有挂载网盘
func LoadAllStorages(ctx context.Context) {
	storages, err := service.GetEnabledStorages()
	if err != nil {
		// fmt.Printf("[HelaList] Error getting enabled storages on startup: %+v\n", err)  // 注释掉以避免干扰 MCP 协议
		return
	}

	// fmt.Printf("[HelaList] Found %d storages to load...\n", len(storages))  // 注释掉以避免干扰 MCP 协议
	for _, storage := range storages {
		// fmt.Printf("[HelaList] Loading storage: %s\n", storage.MountPath)  // 注释掉以避免干扰 MCP 协议
		// 为每个存储创建一个新的驱动实例
		driverNew, err := GetDriver(storage.Driver)
		if err != nil {
			fmt.Printf("[HelaList] Error loading storage [%s]: failed to get driver new: %+v\n", storage.MountPath, err)
			continue
		}
		storageDriver := driverNew()
		// 初始化这个驱动
		err = initStorage(ctx, storage, storageDriver)
		if err != nil {
			fmt.Printf("[HelaList] Error initializing storage [%s]: %+v\n", storage.MountPath, err)
		}
	}
}
