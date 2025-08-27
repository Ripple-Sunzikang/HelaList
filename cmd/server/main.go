package main

import (
	"HelaList/internal/bootstrap"
	"HelaList/internal/model"
	"HelaList/internal/server"
	"log"
)

func main() {
	bootstrap.InitDB()

	err := bootstrap.Db.AutoMigrate(&model.User{})
	if err != nil {
		log.Fatal("数据库迁移失败: %v", err)
	}
	log.Println("数据库迁移成功！")

	r := server.Init()
	if err := r.Run(); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
	}

	// newUser := &model.User{
	// 	Username: "suzuki",
	// 	Email:    "1063046101@qq.com",
	// 	Password: "suzuki",
	// 	Identity: model.GUEST,
	// }
	// err = database.CreateUser(newUser)
	// if err != nil {
	// 	log.Fatalf("创建用户失败: %v", err)
	// }

	// newMount := &model.Mount{
	// 	MountPath: "/",
	// 	Driver:    "WebDAV",
	// }
	// repository.CreateMount(newMount)
}
