package model

// 存放一堆参数用的文件

type ListArgs struct {
	ReqPath string
	Refresh bool
}

type FsOtherArgs struct {
	Path   string      `json:"path" form:"path"`
	Method string      `json:"method" form:"method"`
	Data   interface{} `json:"data" form:"data"`
}

type OtherArgs struct {
	Obj    Obj
	Method string
	Data   interface{}
}
