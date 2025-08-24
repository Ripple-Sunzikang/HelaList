package webdav

import "net/http"

// 由服务器向客户端写回时使用的缓冲
/*
我知道你要问为什么buffer是小写字母开头。别忘了大写开头表示公开,小写开头不就表示私有吗
公开和私有什么区别？很简单，公开的类型也好函数也好，在自己的package以外也是可以使用的。
由于brw只会在webdav内部使用，因此出于代码规范性考虑，把brw设置为package内私有。
*/
type bufferedResponseWriter struct {
	status int         // 状态码
	data   []byte      // 数据
	header http.Header // http表头
}

// 构造函数
/*
考考你的呀，如果在这里把函数名首字母大写，改成NewBufferedResponseWriter会怎么样？
答案是，这个函数确实会对外公开了，但是bufferedResponseWriter并没有对外公开。
于是这个函数根本无法使用。
*/
func newBufferedResponseWriter() *bufferedResponseWriter {
	return &bufferedResponseWriter{
		status: 0,
	}
}

// 获取brw的表头
func (brw *bufferedResponseWriter) GetHeader() http.Header {
	return brw.header
}

// 向brw写入数据，并返回写入数据的长度
func (brw *bufferedResponseWriter) Write(bytes []byte) int {
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
func (brw *bufferedResponseWriter) WriteHeader(status int) {
	if brw.status == 0 {
		brw.status = status
	}
}

// 将缓冲中的数据写入ResponseWriter
func (brw *bufferedResponseWriter) WriteToResponse(rw http.ResponseWriter) (int, error) {
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
