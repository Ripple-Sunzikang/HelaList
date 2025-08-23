package webdav

import "net/http"

// 由服务器向客户端写回时使用的缓冲
type BufferedResponseWriter struct {
	status int         // 状态码
	data   []byte      // 数据
	header http.Header // http表头
}

// 构造函数
func newBufferedResponseWriter() *BufferedResponseWriter {
	return &BufferedResponseWriter{
		status: 0,
	}
}

// 获取brw的表头
func (brw *BufferedResponseWriter) GetHeader() http.Header {
	return brw.header
}

// 向brw写入数据，并返回写入数据的长度
func (brw *BufferedResponseWriter) Write(bytes []byte) int {
	brw.data = append(brw.data, bytes...)
	return len(bytes)
}

/*
我知道你很好奇这个名字是怎么回事，明明是写status却要起个WriteHeader的名字。
其实很好理解，status往往是http的header的第一行。我们设WriteHeader(200)，
相当于在告诉代码，连接顺利，开始传输。
实际上，你在接下来的函数就会发现，哪怕是go语言内置的ResponseWriter，他用的
也是WriteHeader(status int)
*/
func (brw *BufferedResponseWriter) WriteHeader(status int) {
	if brw.status == 0 {
		brw.status = status
	}
}

// 将缓冲中的数据写入ResponseWriter
func (brw *BufferedResponseWriter) WriteToResponse(rw http.ResponseWriter) (int, error) {
	header := rw.Header() // 要写入的ResponseWriter的header

	// 将缓冲内所有header的所有键值都传输到ResponseWriter中
	for key, values := range brw.header {
		for _, value := range values {
			header.Add(key, value)
		}
	}

	// 写入status,写入data
	rw.WriteHeader(brw.status)
	return rw.Write(brw.data)
}
