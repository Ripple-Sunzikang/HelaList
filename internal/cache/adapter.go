package cache

import (
	"HelaList/internal/redis"
	"errors"
	"time"

	"github.com/OpenListTeam/go-cache"
)

var ErrCacheMiss = errors.New("cache miss")

// 缓存适配器 - 支持Redis和内存缓存的fallback机制
type CacheAdapter struct {
	redisEnabled bool
	memCache     cache.ICache[interface{}]
}

// 全局缓存实例
var GlobalCache *CacheAdapter

// 初始化缓存适配器
func InitCacheAdapter() {
	GlobalCache = &CacheAdapter{
		redisEnabled: redis.RedisService != nil,
		memCache:     cache.NewMemCache[interface{}](),
	}
}

// Set 设置缓存
func (c *CacheAdapter) Set(key string, value interface{}, expiration time.Duration) error {
	// 优先使用Redis
	if c.redisEnabled {
		if err := redis.RedisService.Set(key, value, expiration); err == nil {
			return nil
		}
		// Redis失败时记录日志并fallback到内存缓存
	}

	// fallback到内存缓存
	c.memCache.Set(key, value, cache.WithEx[interface{}](expiration))
	return nil
}

// Get 获取缓存
func (c *CacheAdapter) Get(key string, dest interface{}) error {
	// 优先从Redis获取
	if c.redisEnabled {
		if err := redis.RedisService.Get(key, dest); err == nil {
			return nil
		}
	}

	// fallback到内存缓存
	if value, ok := c.memCache.Get(key); ok {
		// 这里需要类型断言或序列化处理
		if destPtr, ok := dest.(*interface{}); ok {
			*destPtr = value
			return nil
		}
	}

	return ErrCacheMiss
}

// Del 删除缓存
func (c *CacheAdapter) Del(keys ...string) error {
	// 从Redis删除
	if c.redisEnabled {
		redis.RedisService.Del(keys...)
	}

	// 从内存缓存删除
	for _, key := range keys {
		c.memCache.Del(key)
	}

	return nil
}

// 便捷方法
func Set(key string, value interface{}, expiration time.Duration) error {
	if GlobalCache == nil {
		InitCacheAdapter()
	}
	return GlobalCache.Set(key, value, expiration)
}

func Get(key string, dest interface{}) error {
	if GlobalCache == nil {
		InitCacheAdapter()
	}
	return GlobalCache.Get(key, dest)
}

func Del(keys ...string) error {
	if GlobalCache == nil {
		InitCacheAdapter()
	}
	return GlobalCache.Del(keys...)
}
