package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Storage表示系统中一个存储后端的完整配置信息，你可以理解为一个被挂载的网盘
type Storage struct {
	Id              uuid.UUID `gorm:"primaryKey" json:"id"`
	MountPath       string    `gorm:"unique" json:"mount_path" binding:"required"` // 虚拟路径
	Order           int       `json:"order"`                                       // 用于排序
	Driver          string    `json:"driver"`                                      // 说明是哪一种网盘(项目初期，默认最常见的WebDAV)
	CacheExpiration int       `json:"cache_expiration"`                            // 文件缓存过期时间(其实我想用time.Duration)
	Status          string    `json:"status"`                                      // 文件状态
	Addition        string    `gorm:"type:text" json:"addition"`                   // 附加信息
	Remark          string    `json:"remark"`                                      // 文件备注
	ModifiedTime    time.Time `json:"modified_time"`                               // 修改时间
	Disabled        bool      `json:"disabled"`                                    // 该存储是否被禁用
	Sort                      // 排序用
	// 代理配置
	WebProxy         bool   `json:"web_proxy"`          // 是否启用Web代理
	WebdavPolicy     string `json:"webdav_policy"`      // WebDAV策略
	ProxyRange       bool   `json:"proxy_range"`        // 是否支持范围请求代理
	DownProxyURL     string `json:"down_proxy_url"`     // 下载代理URL
	DisableProxySign bool   `json:"disable_proxy_sign"` // 禁用代理签名
}

// 文件的默认排序
type Sort struct {
	OrderBy        string `json:"order_by"`        // 比如"ModifiedTime"，就是按修改时间排序，"order"就是按Order排序
	OrderDirection string `json:"order_direction"` // 升序和降序(其实我想用bool然后换个名字)
	ExtractFolder  string `json:"extract_folder"`  // 暂定
}

func (Storage) TableName() string {
	return "storages"
}

// BeforeCreate 钩子：在创建前生成 UUID v7
func (s *Storage) BeforeCreate(tx *gorm.DB) error {
	if s.Id == uuid.Nil {
		s.Id = uuid.Must(uuid.NewV7())
	}
	return nil
}

func (s *Storage) GetStorage() *Storage {
	return s
}

func (s *Storage) SetStorage(storage Storage) {
	*s = storage
}

func (s *Storage) SetStatus(status string) {
	s.Status = status
}

// WebDAV策略相关方法
func (s *Storage) Webdav302() bool {
	return s.WebdavPolicy == "302_redirect"
}

func (s *Storage) WebdavProxyURL() bool {
	return s.WebdavPolicy == "use_proxy_url"
}
