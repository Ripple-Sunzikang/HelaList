package op

import (
	"HelaList/configs"
	"HelaList/internal/driver"
	"reflect"
	"strings"

	"github.com/pkg/errors"
)

type DriverConstructor func() driver.Driver

var driverMap = map[string]DriverConstructor{} // 在内存中记录总共挂载了哪些网盘
var driverInfoMap = map[string]driver.Info{}

// 下面所有的内容只专注于一件事：把数据库的网盘数据写到内存里

func RegisterDriver(driver DriverConstructor) {
	// log.Infof("register driver: [%s]", config.Name)
	tempDriver := driver()
	tempConfig := tempDriver.Config()
	registerDriverItems(tempConfig, tempDriver.GetAddition())
	driverMap[tempConfig.Name] = driver
}

func registerDriverItems(config driver.Config, addition driver.Additional) {
	// log.Debugf("addition of %s: %+v", config.Name, addition)
	tAddition := reflect.TypeOf(addition)
	for tAddition.Kind() == reflect.Pointer {
		tAddition = tAddition.Elem()
	}
	mainItems := getMainItems(config)
	additionalItems := getAdditionalItems(tAddition, config.DefaultRoot)
	driverInfoMap[config.Name] = driver.Info{
		Common:     mainItems,
		Additional: additionalItems,
		Config:     config,
	}
}

func GetDriver(name string) (DriverConstructor, error) {
	n, ok := driverMap[name]
	if !ok {
		return nil, errors.Errorf("no driver named: %s", name)
	}
	return n, nil
}

func GetDriverNames() []string {
	var driverNames []string
	for k := range driverInfoMap {
		driverNames = append(driverNames, k)
	}
	return driverNames
}

func GetDriverInfoMap() map[string]driver.Info {
	return driverInfoMap
}

func getMainItems(config driver.Config) []driver.Item {
	items := []driver.Item{{
		Name:     "mount_path",
		Type:     configs.TypeString,
		Required: true,
		Help:     "The path you want to mount to, it is unique and cannot be repeated",
	}, {
		Name: "order",
		Type: configs.TypeNumber,
		Help: "use to sort",
	}, {
		Name: "remark",
		Type: configs.TypeText,
	}}
	if !config.NoCache {
		items = append(items, driver.Item{
			Name:     "cache_expiration",
			Type:     configs.TypeNumber,
			Default:  "30",
			Required: true,
			Help:     "The cache expiration time for this storage",
		})
	}
	items = append(items, driver.Item{
		Name: "down_proxy_url",
		Type: configs.TypeText,
	})
	items = append(items, driver.Item{
		Name:    "disable_proxy_sign",
		Type:    configs.TypeBool,
		Default: "false",
		Help:    "Disable sign for Download proxy URL",
	})
	if config.LocalSort {
		items = append(items, []driver.Item{{
			Name:    "order_by",
			Type:    configs.TypeSelect,
			Options: "name,size,modified",
		}, {
			Name:    "order_direction",
			Type:    configs.TypeSelect,
			Options: "asc,desc",
		}}...)
	}
	items = append(items, driver.Item{
		Name:    "extract_folder",
		Type:    configs.TypeSelect,
		Options: "front,back",
	})
	items = append(items, driver.Item{
		Name:     "disable_index",
		Type:     configs.TypeBool,
		Default:  "false",
		Required: true,
	})
	items = append(items, driver.Item{
		Name:     "enable_sign",
		Type:     configs.TypeBool,
		Default:  "false",
		Required: true,
	})
	return items
}
func getAdditionalItems(t reflect.Type, defaultRoot string) []driver.Item {
	var items []driver.Item
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Type.Kind() == reflect.Struct {
			items = append(items, getAdditionalItems(field.Type, defaultRoot)...)
			continue
		}
		tag := field.Tag
		ignore, ok1 := tag.Lookup("ignore")
		name, ok2 := tag.Lookup("json")
		if (ok1 && ignore == "true") || !ok2 {
			continue
		}
		item := driver.Item{
			Name:     name,
			Type:     strings.ToLower(field.Type.Name()),
			Default:  tag.Get("default"),
			Options:  tag.Get("options"),
			Required: tag.Get("required") == "true",
			Help:     tag.Get("help"),
		}
		if tag.Get("type") != "" {
			item.Type = tag.Get("type")
		}
		if item.Name == "root_folder_id" || item.Name == "root_folder_path" {
			if item.Default == "" {
				item.Default = defaultRoot
			}
			item.Required = item.Default != ""
		}
		// set default type to string
		if item.Type == "" {
			item.Type = "string"
		}
		items = append(items, item)
	}
	return items
}
