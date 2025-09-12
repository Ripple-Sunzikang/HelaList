package bootstrap

// 用于初始化数据库和Redis
/*
	考虑到dsn等不能直接硬编码，后续肯定要修改为config设置，而且还会有对应的前端设置界面。不过现在启动就好
*/

import (
	"HelaList/configs"
	"HelaList/internal/redis"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var Db *gorm.DB

func InitDB() {
	// dsn := "host=frp-oil.com user=suzuki password=suzuki dbname=hela port=37260 sslmode=disable TimeZone=Asia/Shanghai"
	dsn := configs.Conf.Database.DSN

	var err error
	Db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Panicf("Database connection failed: %v", err)
	}
	// log.Println("Database connected!")  // 注释掉以避免干扰 MCP 协议
}

func InitRedis() {
	// 初始化Redis服务
	if err := redis.InitRedisService(&configs.Conf.Redis); err != nil {
		log.Printf("Redis connection failed: %v, continuing without Redis cache", err)
		return
	}
	log.Println("Redis connected!")
}
