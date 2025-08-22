package bootstrap

// 用于初始化数据库
/*
	考虑到dsn等不能直接硬编码，后续肯定要修改为config设置，而且还会有对应的前端设置界面。不过现在启动就好
*/

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var Db *gorm.DB

func InitDB() {

	dsn := "host=localhost user=suzuki password=sande13784266678 dbname=hela port=5432 sslmode=disable TimeZone=Asia/Shanghai"

	var err error
	Db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Panicf("连接数据库失败: %v", err)
	}
	log.Println("数据库连接成功!")
}
