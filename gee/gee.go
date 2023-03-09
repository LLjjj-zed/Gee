package gee

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
)

//实现了路由映射表，提供了用户注册静态路由的方法，包装了启动服务的函数

// HandlerFunc defines the request handler used by gee
type HandleFunc func(http.ResponseWriter, *http.Request)

// Engine implement the interface of ServeHTTP
type Engine struct {
	router map[string]HandleFunc
}

// New is the constructor of gee.Engine
func New() *Engine {
	return &Engine{router: make(map[string]HandleFunc)}
}

var BufferPool = sync.Pool{
	New: func() interface{} {
		return new(strings.Builder)
	},
}

func Concate(s ...string) string {
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

func (engine *Engine) addRoute(method string, parttern string, handler HandleFunc) {
	key := Concate(method, "-", parttern)
	engine.router[key] = handler
}

// GET defines the method to add GET request
func (engine *Engine) GET(pattern string, handler HandleFunc) {
	engine.addRoute("GET", pattern, handler)
}

// POST defines the method to add POST request
func (engine *Engine) POST(pattern string, handler HandleFunc) {
	engine.addRoute("POST", pattern, handler)
}

// Run defines the method to start a http server
func (engine *Engine) Run(addr string) (err error) {
	return http.ListenAndServe(addr, engine)
}

func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	key := Concate(req.Method, "-", req.URL.Path)
	if handler, ok := engine.router[key]; ok {
		handler(w, req)
	} else {
		fmt.Fprintf(w, "404 NOT FOUND: %s\\n", req.URL)
	}
}
