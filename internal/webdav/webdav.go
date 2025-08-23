package webdav

import (
	"HelaList/internal/model"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
)

/*
WebDAV部分其实先写的是锁系统，也就是该目录下的lock.go文件，
在锁系统完成的基础上，我们进一步开始考虑文件系统的设计。

别问我为什么这么爱写注释。这不是良好习惯，这是我写给自己看的。
*/

// 再次感慨interface真的是伟大发明，既打破了面向对象继承的桎梏，又开了函数式编程的先河 (指有时接口本身可以作为一种类型)
type FileSystem interface {
	Status(path string) (*model.Object, error)      // 获取文件/目录信息
	ReadDir(path string) ([]*model.Object, error)   // 列出目录内容
	Mkdir(path string) error                        // 创建目录
	Remove(path string) error                       // 删除文件/目录
	Rename(oldPath string, newPath string) error    // 移动/重命名
	Copy(sourcePath string, destiPath string) error // 复制文件
	OpenReader(path string) (io.ReadCloser, error)  // 读取文件内容
	OpenWriter(path string) (io.WriteCloser, error) // 写入文件内容
}

type Handler struct {
	fileSystem FileSystem  // 文件系统
	lockSystem *LockSystem // 你已经实现的锁结构
	logger     *log.Logger // 日志
}

// 创建Handler
func NewHandler(fs FileSystem, lockSys *LockSystem) *Handler {
	return &Handler{
		fileSystem: fs,
		lockSystem: lockSys,
		logger:     log.New(os.Stdout, "WebDAV: ", log.LstdFlags),
	}
}

// HTTP入口
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 如果都没连上，默认连接失败
	status, err := http.StatusBadRequest, errors.New("一开始就没连上")

	// 有一个缓存是好的，你需要设计一个缓存，于是就有了buffered_response

	if h.lockSystem == nil {
		status, err = http.StatusInternalServerError, errors.New("你的锁坏了导致没连上")
	} else {
		switch r.Method {
		case "OPTIONS":
		// 执行OPTIONS操作
		case "GET", "HEAD", "POST":
		// 执行GET、HEAD、POST操作
		case "DELETE":
			// 执行DELETE操作handle
		case "PUT":
			// 执行PUT操作handle
		case "MKCOL":
			// 执行MKCOL操作handle
		case "COPY", "MOVE":
			// 执行COPY、MOVE操作handle
		case "LOCK":
			// 执行LOCK操作handle
		case "UNLOCK":
			// 执行UNLOCK操作handle
		case "PROPFIND":
			// 执行PROPFIND操作handle
		case "PROPPATCH":
			// 执行PROPPATCH操作handle

		}
	}
}
