package database

// 你应该放弃使用MongoDB和MySQL而选择这个更加专业化的PostgreSQL
/*
	数据库核心设计：
	在多线程环境确保数据库安全访问，故设计sync.Once；
	设计连接池，提高访问效率；
	设计Redis辅助数据库，提高访问效率；
*/

import (
	"sync"

	"gorm.io/gorm"
)

var (
	db   *gorm.DB
	once sync.Once
	err  error
)
