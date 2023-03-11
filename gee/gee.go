package gee

import (
	"html/template"
	"net/http"
	"strings"
)

// HandlerFunc defines the request handler used by gee
type HandlerFunc func(*Context)

// Engine implement the interface of ServeHTTP
type Engine struct {
	router *router
	groups []*RouterGroup // store all groups
	*RouterGroup
}

// New is the constructor of gee.Engine
func New() *Engine {
	engine := &Engine{router: NewRouter()}
	engine.RouterGroup = &RouterGroup{engine: engine}
	engine.groups = []*RouterGroup{engine.RouterGroup}
	return engine
}

// Default use Logger() & Recovery middlewares
func Default() *Engine {
	engine := New()
	engine.Use(Logger(), Recovery())
	return engine
}

// SetFuncMap 方法是用来设置模板渲染时需要用到的自定义函数的FuncMap 是一个 map 类型，
// 其中 key 是函数名，value 是一个空接口，这个接口的实现可以是任何类型的函数。在模板渲染时，
// 我们可以通过函数名调用对应的自定义函数。这个方法的作用就是将这个 FuncMap
// 保存到 Engine 实例的 funcMap 属性中，以便在模板渲染时使用
func (engine *Engine) SetFuncMap(funcMap template.FuncMap) {
	engine.funcMap = funcMap
}

// LoadHTMLGlob 方法用于加载 HTML 模板文件，并将其解析成模板对象，
// 此方法接收一个文件路径模式作为参数，例如 views/*.html。
// 模板文件可以包含动态内容和控制结构，可以使用 Go 内置的模板语言进行定义和渲染。
// 模板语言是一种类似于 JSP 和 PHP 的模板技术，用于将模板和数据结合起来生成最终的 HTML 页面。
// template.Must 函数用于将模板对象与错误进行绑定，如果模板文件解析失败，则程序将抛出 panic 异常。
func (engine *Engine) LoadHTMLGlob(pattern string) {
	engine.htmlTemplates = template.Must(template.New("").Funcs(engine.funcMap).ParseGlob(pattern))
}

// Run defines the method to start a http server
func (engine *Engine) Run(addr string) (err error) {
	return http.ListenAndServe(addr, engine)
}

// ServeHTTP 实现Handler接口，自定义HTTP请求的处理方式
// 首先遍历所有的路由组，判断请求路径是否以路由组的前缀开头，
// 如果是，将路由组的中间件添加到中间件列表中。然后创建一个新的 Context 对象，
// 并将中间件列表赋值给 handlers 字段，将当前 Engine 对象赋值给 engine 字段。
// 最后调用 router 对象的 handle 方法处理请求
func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var middlewares []HandlerFunc
	for _, group := range engine.groups {
		if strings.HasPrefix(req.URL.Path, group.prefix) {
			middlewares = append(middlewares, group.middlewares...)
		}
	}
	c := newContext(w, req)
	c.handlers = middlewares
	c.engine = engine
	engine.router.handle(c)
}
