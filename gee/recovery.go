package gee

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strings"
)

// trace 函数用于获取调用 trace 函数的堆栈信息。其中，
// 通过调用 runtime.Callers 函数获取当前 goroutine
// 的调用堆栈信息，第一个参数 3 表示跳过前三个调用者，
// 以避免输出一些不必要的信息。接下来，通过调用 runtime.FuncForPC 函数
// 获取函数指针所对应的函数的信息，例如文件名和行号。最后，将获取到的堆栈信息格式化为字符串并返回
func trace(message string) string {
	var pcs [32]uintptr
	n := runtime.Callers(3, pcs[:]) // skip first 3 caller

	var str strings.Builder
	str.WriteString(message + "\nTraceback:")
	for _, pc := range pcs[:n] {
		fn := runtime.FuncForPC(pc)
		file, line := fn.FileLine(pc)
		str.WriteString(fmt.Sprintf("\n\t%s:%d", file, line))
	}
	return str.String()
}

// Recovery 是一个中间件，主要用于在处理请求时捕获 panic，
// 并在捕获到 panic 后返回 500 状态码和错误信息，防止因为 panic 导致程序崩溃。
// Recovery 的实现方法是使用 defer 关键字捕获 panic，然
// 后将 panic 的信息转换成字符串并使用 log 输出错误日志，最后返回 500 状态码和错误信息。
// 需要注意的是，在处理 panic 后需要将中间件栈清空，否则有可能出现一些未知的问题
func Recovery() HandlerFunc {
	return func(c *Context) {
		defer func() {
			if err := recover(); err != nil {
				message := fmt.Sprintf("%s", err)
				log.Printf("%s\n\n", trace(message))
				c.Fail(http.StatusInternalServerError, "Internal Server Error")
			}
		}()
		c.Next()
	}
}
