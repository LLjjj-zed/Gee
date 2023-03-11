package gee

import (
	"log"
	"time"
)

// Logger 函数返回一个 HandlerFunc 类型的函数，用于记录请求的处理时间和状态码等信息
func Logger() HandlerFunc {
	return func(c *Context) {
		// Start timer
		t := time.Now()
		// Process request
		c.Next()
		// Calculate resolution time
		log.Printf("[%d] %s in %v", c.StatusCode, c.Req.RequestURI, time.Since(t))
	}
}
