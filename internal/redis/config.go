package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// 用于redis集群相关配置
/*
	3个Redis集群，1个主Cluster,2个从Cluster
	每一个服务器分别配置3个哨兵
*/

type Config struct {
	// Cluster节点
	ClusterAddrs []string `json:"cluster_addrs"`

	// Sentinel配置
	SentinelAddrs []string `json:"sentinel_addrs"`
	MasterNames   []string `json:"master_names"`

	// 连接配置
	Password     string `json:"password"`
	PoolSize     int    `json:"pool_size"`
	MinIdleConns int    `json:"min_idle_conns"`

	// 超时配置
	DialTimeout  time.Duration `json:"dial_timeout"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
}

// 默认配置
func DefaultConfig() *Config {
	return &Config{
		ClusterAddrs: []string{
			"127.0.0.1:7001",
			"127.0.0.1:7002",
			"127.0.0.1:7003",
			"127.0.0.1:7004",
			"127.0.0.1:7005",
			"127.0.0.1:7006",
		},
		SentinelAddrs: []string{
			"127.0.0.1:27001",
			"127.0.0.1:27002",
			"127.0.0.1:27003",
		},
		MasterNames: []string{
			"master1",
			"master2",
			"master3",
		},
		Password:     "",
		PoolSize:     100,
		MinIdleConns: 10,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}
}

func NewClusterClient(cfg *Config) *redis.ClusterClient {
	return redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:        cfg.ClusterAddrs,
		Password:     cfg.Password,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,

		// 超时配置
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,

		// 故障转移配置
		MaxRedirects:   3,
		ReadOnly:       true,
		RouteByLatency: true,
	})
}

// 带Sentinel监控的Cluster客户端
func NewSentinelClusterClient(ctx context.Context, cfg *Config) (*redis.ClusterClient, error) {
	if len(cfg.SentinelAddrs) == 0 {
		return nil, redis.Nil
	}

	// 通过Sentinel获取Cluster节点信息
	sentinelClient := redis.NewSentinelClient(&redis.Options{
		Addr:         cfg.SentinelAddrs[0],
		Password:     cfg.Password,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	})
	defer sentinelClient.Close()

	// 获取所有master地址
	var clusterAddrs []string
	for _, masterName := range cfg.MasterNames {
		addr, err := sentinelClient.GetMasterAddrByName(ctx, masterName).Result()
		if err == nil && len(addr) >= 2 {
			clusterAddrs = append(clusterAddrs, addr[0]+":"+addr[1])
		}
	}

	if len(clusterAddrs) == 0 {
		return nil, redis.Nil
	}

	return redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:        clusterAddrs,
		Password:     cfg.Password,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}), nil
}

// 健康检查
func (cfg *Config) HealthCheck(ctx context.Context) error {
	client := NewClusterClient(cfg)
	defer client.Close()

	return client.Ping(ctx).Err()
}
