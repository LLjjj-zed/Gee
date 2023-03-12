package gee

//代码最开头，给map[string]interface{}起了一个别名gee.H，构建JSON数据时，显得更简洁。
//Context目前只包含了http.ResponseWriter和*http.Request，另外提供了对 Method 和 Path 这两个常用属性的直接访问。
//提供了访问Query和PostForm参数的方法。
//提供了快速构造String/Data/JSON/HTML响应的方法。

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type H map[string]interface{}

// Context struct 用于封装 HTTP 请求和响应的相关信息，以及相关的处理函数
type Context struct {
	// origin objects
	Writer http.ResponseWriter //HTTP 响应的写入器，用于向客户端发送响应数据
	Req    *http.Request       //HTTP 请求的指针，用于获取客户端发送的请求信息，例如请求方法、请求头、请求体等
	// request info
	Path   string            //请求的路径，即 URL 中的路径部分
	Method string            //请求的方法，例如 GET、POST 等
	Params map[string]string //请求中的路由参数，由路由器解析后存储在此字段中
	// response info
	StatusCode int //响应的状态码，例如 200、404 等
	// middleware
	handlers []HandlerFunc //处理函数的切片，用于存储当前请求所需要执行的所有处理函数
	index    int           //当前请求需要执行的处理函数在 handlers 切片中的索引
	// engine pointer
	engine *Engine //指向引擎的指针，用于访问引擎中的一些全局配置和方法
}

// newContext 创建一个新的context对象
func newContext(w http.ResponseWriter, req *http.Request) *Context {
	return &Context{
		Writer: w,
		Req:    req,
		Path:   req.URL.Path,
		Method: req.Method,
		index:  -1,
	}
}

// Next index是记录当前执行到第几个中间件，当在中间件中调用Next方法时，
// 控制权交给了下一个中间件，直到调用到最后一个中间件，然后再从后往前，
// 调用每个中间件在Next方法之后定义的部分
func (c *Context) Next() {
	c.index++
	s := len(c.handlers)
	for ; c.index < s; c.index++ {
		c.handlers[c.index](c)
		DPrintf("[Context]index:%d\n", c.index)
	}
}

// Param  我们将解析后的参数存储到Params中，通过c.Param("lang")的方式获取到对应的值
func (c *Context) Param(key string) string {
	value := c.Params[key]
	DPrintf("[Context]Params:%s\n", value)
	return value
}

// PostForm  获取URL中的参数  如http://gee.com?key=value/
func (c *Context) PostForm(key string) string {
	DPrintf("[Context]PostFormValue:%s\n", c.Req.FormValue(key))
	return c.Req.FormValue(key)
}

// Query 获取query value
func (c *Context) Query(key string) string {
	//Query 解析 RawQuery 并返回相应的值,返回的Values实际上是一个map对象。Get 获取第一个value
	DPrintf("[Context]QueryValue:%s\n", c.Req.URL.Query().Get(key))
	return c.Req.URL.Query().Get(key)
}

// Status 设置HTTP StatusCode 并写入Header
func (c *Context) Status(code int) {
	c.StatusCode = code
	DPrintf("[Context]StatusCode:%d\n", c.StatusCode)
	//使用提供的状态代码发送 HTTP 响应头
	c.Writer.WriteHeader(code)
}

// SetHeader 设置HTTP响应头
func (c *Context) SetHeader(key string, value string) {
	//Set 将与键关联的标头条目设置为单个元素value
	DPrintf("[Context]SetKey:%s\tSetValue:%s\n", key, value)
	c.Writer.Header().Set(key, value)
}

// String 写入响应数据
func (c *Context) String(code int, format string, values ...interface{}) {
	c.SetHeader("Content-Type", "text/plain")
	c.Status(code)
	//Writer.Write() 在响应头部写入数据之后发送数据,通常被用来向HTTP客户端发送响应数据。
	c.Writer.Write([]byte(fmt.Sprintf(format, values...)))
}

// JSON 将 HTTP 响应状态码设置为指定的 code，将 Content-Type 头设置为 "application/json"。
// 然后，它使用 Go 标准库中的 json.NewEncoder 函数将给定的 object 编码为 JSON 字符串，
// 并将其写入 HTTP 响应正文中。如果在编码期间出现错误，则返回 HTTP 500 内部服务器错误，并在响应正文中包含错误消息
func (c *Context) JSON(code int, object interface{}) {
	c.SetHeader("Content-Type", "application/json")
	c.StatusCode = code
	encoder := json.NewEncoder(c.Writer)
	if err := encoder.Encode(object); err != nil {
		http.Error(c.Writer, err.Error(), 500)
	}
}

// Data 写入响应数据
func (c *Context) Data(code int, data []byte) {
	c.Status(code)
	c.Writer.Write(data)
}

// HTML 将html写入Writer中
func (c *Context) HTML(code int, name string, data interface{}) {
	c.SetHeader("Content-Type", "text/html")
	c.Status(code)
	if err := c.engine.htmlTemplates.ExecuteTemplate(c.Writer, name, data); err != nil {
		c.Fail(500, err.Error())
	}
}

// Fail 将 HTTP 响应状态码设置为指定的 code，并将一个包含错误信息的 JSON 响应发送给客户端，
// 然后将当前处理程序的索引设置为 handlers 切片的末尾，以确保在 handlers 切片中的后续处理程序不会被执行。
// 这个方法通常在处理请求时遇到错误时被调用，以及在中间件中进行错误处理时使用。
func (c *Context) Fail(code int, err string) {
	c.index = len(c.handlers)
	c.JSON(code, H{"message": err})
}
