package gee

import (
	"html/template"
	"log"
	"net/http"
	"path"
)

// RouterGroup 用于实现路由分组和中间件功能
type RouterGroup struct {
	prefix        string             // 前缀，用于给分组内的所有路由统一添加前缀
	middlewares   []HandlerFunc      // 中间件列表，用于在路由处理函数执行前或执行后进行操作
	parent        *RouterGroup       // 父级分组，支持嵌套分组
	htmlTemplates *template.Template // HTML 模板，用于渲染 HTML 页面
	funcMap       template.FuncMap   // 函数映射，用于在 HTML 模板中使用自定义函数
	engine        *Engine            // Engine 实例，用于所有分组共享 Engine 实例的功能
}

// Group is defined to create a new RouterGroup
// remember all groups share the same Engine instance
func (group *RouterGroup) Group(prefix string) *RouterGroup {
	engine := group.engine
	newGroup := &RouterGroup{
		prefix: group.prefix + prefix,
		parent: group,
		engine: engine,
	}
	engine.groups = append(engine.groups, newGroup)
	return newGroup
}

// addRoute
// 可以仔细观察下addRoute函数，调用了group.engine.router.addRoute来实现了路由的映射。
// 由于Engine从某种意义上继承了RouterGroup的所有属性和方法，因为 (*Engine).engine 是指向自己的。
// 这样实现，我们既可以像原来一样添加路由，也可以通过分组添加路由。
func (group *RouterGroup) addRoute(method string, comp string, handler HandlerFunc) {
	pattern := group.prefix + comp
	log.Printf("Route %4s - %s", method, pattern)
	group.engine.router.addRoute(method, pattern, handler)
}

// GET defines the method to add GET request
func (group *RouterGroup) GET(pattern string, handler HandlerFunc) {
	group.addRoute("GET", pattern, handler)
}

// POST defines the method to add POST request
func (group *RouterGroup) POST(pattern string, handler HandlerFunc) {
	group.addRoute("POST", pattern, handler)
}

// Use 方法用于为该组添加中间件。在 Gin 框架中，中间件是对于 HTTP 请求处理流程的一些拦截器，
// 能够在请求前或请求后执行一些自定义操作。在该方法中，首先获取到该组中已有的中间件列表，
// 然后将新传入的中间件列表追加到原有中间件列表后面。这样，该组中所有的路由请求都会按顺序依次执行这些中间件
func (group *RouterGroup) Use(middlewares ...HandlerFunc) {
	group.middlewares = append(group.middlewares, middlewares...)
}

// create static handler
func (group *RouterGroup) createStaticHandler(relativePath string, fs http.FileSystem) HandlerFunc {
	// 计算绝对路径
	absolutePath := path.Join(group.prefix, relativePath)
	// 创建文件服务器的处理器
	fileServer := http.StripPrefix(absolutePath, http.FileServer(fs))
	// 返回静态文件处理器
	return func(c *Context) {
		// 获取 URL 中的文件路径参数
		file := c.Param("filepath")
		// 检查文件是否存在并且我们是否有访问权限
		if _, err := fs.Open(file); err != nil {
			// 文件不存在或者没有访问权限，返回 404 Not Found
			c.Status(http.StatusNotFound)
			return
		}

		// 文件存在并且有访问权限，调用文件服务器的处理器处理请求
		fileServer.ServeHTTP(c.Writer, c.Req)
	}
}

// Static 方法将用于向路由器组添加一个静态文件服务器的处理函数。
// 它将接收两个参数，一个是 URL 的相对路径，一个是文件系统的根目录。
// 通过调用 createStaticHandler 函数，创建了一个处理函数，
// 该处理函数使用 http.FileServer 和指定的文件系统（在这种情况下是本地文件系统），
// 并返回一个 http.HandlerFunc 对象。
// 在内部，该处理函数检查 URL 中的文件路径，确保文件存在，并根据需要服务文件。
// 然后将这个处理函数绑定到路由器组的 GET 方法上，使用由相对路径和 /*filepath 组成的 URL 模式
func (group *RouterGroup) Static(relativePath string, root string) {
	handler := group.createStaticHandler(relativePath, http.Dir(root))
	urlPattern := path.Join(relativePath, "/*filepath")
	// Register GET handlers
	group.GET(urlPattern, handler)
}
