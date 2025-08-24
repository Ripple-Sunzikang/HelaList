package webdav

import (
	"HelaList/internal/model"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
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

// 别忘了属性命名大写表示public
/*
依旧是课堂时间，Prefix在WebDAV中至关重要。
假如用户的请求是GET http:localhost:8080/webdav/files/test.txt。
go语言的http功能会帮我们自动解析一部分，给我们返回的URL是/webdav/files/test.txt。
但是我们只需要/webdav后面，不需要"/webdav"这个多余的前缀
即, /files/test.txt部分，才是我们服务器内处理的部分。这也是为什么我们会设计一个StripPrefix的函数，用来分割用户请求的http链接
那么如何划分呢？依据的是Handler中的Prefix属性
*/
type Handler struct {
	Prefix     string      // Handler的前缀，此处应默认为/webdav
	LockSystem *LockSystem // 你已经实现的锁结构
	Logger     *log.Logger // 日志
}

// 创建Handler
func newHandler(lockSys *LockSystem) *Handler {
	return &Handler{
		Prefix:     "/webdav",
		LockSystem: lockSys,
		Logger:     log.New(os.Stdout, "WebDAV: ", log.LstdFlags),
	}
}

// 处理客户端请求前缀
func (h *Handler) stripPrefix(p string) (string, int, error) {
	if h.Prefix == "" {
		return p, http.StatusOK, nil // 错误检测，规范用，实际上根本不会有这种错误。
	}

	if r := strings.TrimPrefix(p, h.Prefix); len(r) < len(p) {
		return r, http.StatusOK, nil
	}
	return p, http.StatusNotFound, errors.New("oops, 你webdav的前缀不匹配")
}

// HTTP入口
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 如果都没连上，默认连接失败
	status, err := http.StatusBadRequest, errors.New("一开始就没连上")

	brw := newBufferedResponseWriter() // 写回客户端的缓冲
	useBrw := true                     // 在处理GET等请求时，若待处理的文件过大，使用brw的话会给服务器内存带来巨大压力。所以有时需要禁止使用brw

	if h.LockSystem == nil {
		status, err = http.StatusInternalServerError, errors.New("你的锁坏了导致没连上")
	} else {
		switch r.Method {
		case "OPTIONS":
		// 执行OPTIONS操作
		case "GET", "HEAD", "POST":
		// 执行GET、HEAD、POST操作
		case "DELETE":
			status, err := h.handleDelete(brw, r)
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

	// 善后操作
	if status != 0 {
		w.WriteHeader(status)
		if status != http.StatusNoContent {
			w.Write([]byte(http.StatusText(status))) // 若响应到了内容，则直接写入ResponseWriter
		}
	} else if useBrw {
		brw.WriteToResponse(w) // 若使用缓冲，直接将缓冲导入ResponseWriter
	}
	if h.Logger != nil && err != nil {
		h.Logger.Printf("请求失败: %s %s, 发生错误: %v", r.Method, r.URL.Path, err) // 我知道我写的error和log依托。后面再改！
	}
}

func (h *Handler) handleDelete(brw bufferedResponseWriter, w http.ResponseWriter) (status int, err error) {
	/*
		也是实际写到这里了才意识到fileSystem根本没做。
		于是快马加鞭回去赶工出一个fileSystem出来。
	*/

	/*
		当webdav.go调用handleDelete时，流程如下：
		1. 用stripPrefix()切除/webdav前缀，得到/file/text.txt这样的字符串，命名为reqPath (已实现)
		2. 用confirmLocks()和defer进行上锁关锁
		3. 从context中读取user的配置文件，新建user
		4. 根据user信息，将reqPath合并到user的请求中，成为新的reqPath
		5. 调用fileSystem的Get()方法，传入新的reqPaht，查看文件是否存在
		6. 调用fileSystem的Delete()方法，删除对应文件
		前几个配置文件也好，上锁也好，都是小问题。大问题在fileSystem你根本没设计啊草。
	*/
}
