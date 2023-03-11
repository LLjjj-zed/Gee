package gee

import (
	"net/http"
	"strings"
)

// router route struct
type router struct {
	roots    map[string]*node       //路由树，以不同的 HTTP 方法作为键（key），对应的值是路由树的根节点
	handlers map[string]HandlerFunc //处理函数（handler）的字典，以路径作为键，对应的值是处理该路径的处理函数
}

// NewRouter Create New Router object
// 使用 roots 来存储每种请求方式的Trie 树根节点
// 使用 handlers 存储每种请求方式的 HandlerFunc
func NewRouter() *router {
	return &router{
		roots:    make(map[string]*node),
		handlers: make(map[string]HandlerFunc),
	}
}

// parsePattern 该函数接收一个参数：
//
//	pattern		表示待解析的路由模式。
//
// 首先，将路由模式字符串按照 / 字符进行分割，得到一个字符串切片 vs。
// 然后，创建一个空的字符串切片 parts，用于存储解析后的路由部分。
// 接着，遍历字符串切片 vs，将非空的字符串项加入 parts 切片中，并判断当前字符串项是否以 * 字符开头，
// 如果是，则表示已经匹配到了通配符部分，不需要再继续解析后面的部分，因此退出循环。最后返回解析后的字符串切片 parts。
func parsePattern(pattern string) []string {
	vs := strings.Split(pattern, "/")

	parts := make([]string, 0)
	for _, item := range vs {
		if item != "" {
			parts = append(parts, item)
			if item[0] == '*' {
				break
			}
		}
	}
	return parts
}

// addRoute 在路由中注册一个路由和其对应的处理函数，我们需要完成以下步骤：
//
//	解析路由模式，获取其中的所有部分。
//	以请求方法作为键，将节点树添加到 roots 映射中。
//	使用节点树将路由模式添加到路由树中。
//	以 请求方法 + 路由模式 作为键，将处理函数添加到处理函数映射中
//
// 我们首先通过 parsePattern 函数解析出路由模式中的所有部分，然后将请求方法作为键，
// 将节点树添加到 roots 映射中。如果 roots 映射尚不存在，则创建一个新的 roots 映射。
// 我们在 handlers 映射中存储处理函数，以 请求方法 + 路由模式 作为键。
// 现在我们可以通过以下代码调用 addRoute 方法，将路由和其处理函数添加到路由树中
func (r *router) addRoute(method string, pattern string, handler HandlerFunc) {
	parts := parsePattern(pattern)

	key := Concat(method, "-", pattern)
	_, ok := r.roots[method]
	if !ok {
		r.roots[method] = &node{}
	}
	r.roots[method].insert(pattern, parts, 0)
	r.handlers[key] = handler
}

func (r *router) getRoute(method string, path string) (*node, map[string]string) {
	searchParts := parsePattern(path)
	params := make(map[string]string)
	root, ok := r.roots[method]

	if !ok {
		return nil, nil
	}

	n := root.search(searchParts, 0)

	if n != nil {
		parts := parsePattern(n.pattern)
		for index, part := range parts {
			if part[0] == ':' {
				params[part[1:]] = searchParts[index]
			}
			if part[0] == '*' && len(part) > 1 {
				params[part[1:]] = strings.Join(searchParts[index:], "/")
				break
			}
		}
		return n, params
	}

	return nil, nil
}

// handle  HTTP request  process
// 在调用匹配到的handler前，将解析出来的路由参数赋值给了c.Params。
// 这样就能够在handler中，通过Context对象访问到具体的值了。
func (r *router) handle(c *Context) {
	n, params := r.getRoute(c.Method, c.Path)
	if n != nil {
		c.Params = params
		key := Concat(c.Method, "-", n.pattern)
		//r.handlers[key](c)
		c.handlers = append(c.handlers, r.handlers[key])
	} else {
		c.handlers = append(c.handlers, func(c *Context) {
			c.String(http.StatusNotFound, "404 NOT FOUND: %s\n", c.Path)
		})
	}
	c.Next()
}
