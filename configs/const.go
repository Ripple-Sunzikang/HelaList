package configs

// 用于存储一些常量配置信息

const (
	TypeString = "string"
	TypeSelect = "select"
	TypeBool   = "bool"
	TypeText   = "text"
	TypeNumber = "number"
)

type ContextKey int

const (
	_ ContextKey = iota
	MetaKey
	UserKey
	NoTaskKey
	ApiUrlKey
)

const (
	WORK     = "work"
	DISABLED = "disabled"
	RootName = "root"
)
