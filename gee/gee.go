package gee

import (
	"net/http"
	"strings"
	"sync"
)

// HandlerFunc defines the request handler used by gee
type HandlerFunc func(*Context)

// Engine implement the interface of ServeHTTP
type Engine struct {
	router *router
}

// New is the constructor of gee.Engine
func New() *Engine {
	return &Engine{router: NewRouter()}
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

func (engine *Engine) addRoute(method string, pattern string, handler HandlerFunc) {
	engine.router.addRoute(method, pattern, handler)
}

// GET defines the method to add GET request
func (engine *Engine) GET(pattern string, handler HandlerFunc) {
	engine.addRoute("GET", pattern, handler)
}

// POST defines the method to add POST request
func (engine *Engine) POST(pattern string, handler HandlerFunc) {
	engine.addRoute("POST", pattern, handler)
}

// Run defines the method to start a http server
func (engine *Engine) Run(addr string) (err error) {
	return http.ListenAndServe(addr, engine)
}

func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c := newContext(w, req)
	engine.router.handle(c)
}
