package redis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

// Redis服务管理器
type Service struct {
	client *redis.ClusterClient
	ctx    context.Context
}

// 全局Redis服务实例
var RedisService *Service

// 初始化Redis服务
func InitRedisService(cfg *Config) error {
	client := NewClusterClient(cfg)

	// 测试连接
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return err
	}

	RedisService = &Service{
		client: client,
		ctx:    ctx,
	}

	return nil
}

// 缓存接口
type CacheService interface {
	Set(key string, value interface{}, expiration time.Duration) error
	Get(key string, dest interface{}) error
	Del(keys ...string) error
	Exists(keys ...string) (int64, error)
	SetNX(key string, value interface{}, expiration time.Duration) (bool, error)
}

// 实现缓存接口
func (s *Service) Set(key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return s.client.Set(s.ctx, key, data, expiration).Err()
}

func (s *Service) Get(key string, dest interface{}) error {
	data, err := s.client.Get(s.ctx, key).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(data), dest)
}

func (s *Service) Del(keys ...string) error {
	return s.client.Del(s.ctx, keys...).Err()
}

func (s *Service) Exists(keys ...string) (int64, error) {
	return s.client.Exists(s.ctx, keys...).Result()
}

func (s *Service) SetNX(key string, value interface{}, expiration time.Duration) (bool, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return false, err
	}
	return s.client.SetNX(s.ctx, key, data, expiration).Result()
}

// 分布式锁
func (s *Service) Lock(key string, expiration time.Duration) (bool, error) {
	return s.client.SetNX(s.ctx, "lock:"+key, "locked", expiration).Result()
}

func (s *Service) Unlock(key string) error {
	return s.client.Del(s.ctx, "lock:"+key).Err()
}

// 用户缓存相关
func (s *Service) SetUserCache(username string, user interface{}, expiration time.Duration) error {
	return s.Set("user:info:"+username, user, expiration)
}

func (s *Service) GetUserCache(username string, dest interface{}) error {
	return s.Get("user:info:"+username, dest)
}

func (s *Service) DelUserCache(username string) error {
	return s.Del("user:info:" + username)
}

// Token缓存相关
func (s *Service) SetTokenCache(token string, valid bool, expiration time.Duration) error {
	return s.Set("token:"+token, valid, expiration)
}

func (s *Service) GetTokenCache(token string) (bool, error) {
	var valid bool
	err := s.Get("token:"+token, &valid)
	if err != nil {
		return false, err
	}
	return valid, nil
}

func (s *Service) DelTokenCache(token string) error {
	return s.Del("token:" + token)
}

// 文件缓存相关
func (s *Service) SetFileListCache(path string, files interface{}, expiration time.Duration) error {
	return s.Set("file:list:"+path, files, expiration)
}

func (s *Service) GetFileListCache(path string, dest interface{}) error {
	return s.Get("file:list:"+path, dest)
}

func (s *Service) DelFileListCache(path string) error {
	return s.Del("file:list:" + path)
}

// 存储配置缓存
func (s *Service) SetStorageCache(storageId string, storage interface{}, expiration time.Duration) error {
	return s.Set("storage:config:"+storageId, storage, expiration)
}

func (s *Service) GetStorageCache(storageId string, dest interface{}) error {
	return s.Get("storage:config:"+storageId, dest)
}

func (s *Service) DelStorageCache(storageId string) error {
	return s.Del("storage:config:" + storageId)
}

// 关闭连接
func (s *Service) Close() error {
	return s.client.Close()
}
