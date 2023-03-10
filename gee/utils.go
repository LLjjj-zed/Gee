package gee

import (
	"strings"
	"sync"
)

// BufferPool  全局buffer池，复用对象，提高性能
var BufferPool = sync.Pool{
	New: func() interface{} {
		return new(strings.Builder)
	},
}

// Concat 使用strings.Builder.WriteString（）拼接字符串提高性能
func Concat(s ...string) string {
	buf := BufferPool.Get().(*strings.Builder)
	defer BufferPool.Put(buf)
	for i := 0; i < len(s); i++ {
		buf.WriteString(s[i])
	}
	//这样写会产生内存逃逸，所以使用defer方式
	//str := buf.String()
	//return str
	defer buf.Reset()
	return buf.String()
}
