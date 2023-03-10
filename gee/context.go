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

// Context struct
type Context struct {
	// origin objects
	Writer http.ResponseWriter
	Req    *http.Request
	// request info
	Path   string
	Method string
	// response info
	StautsCode int
}

// newContext 创建一个新的context对象
func newContext(w http.ResponseWriter, req *http.Request) *Context {
	return &Context{
		Writer: w,
		Req:    req,
		Path:   req.URL.Path,
		Method: req.Method,
	}
}

// PostForm  获取URL中的参数  如http://gee.com?key=value/
func (c *Context) PostForm(key string) string {
	return c.Req.FormValue(key)
}

// Query 获取query value
func (c *Context) Query(key string) string {
	//Query 解析 RawQuery 并返回相应的值,返回的Values实际上是一个map对象。Get 获取第一个value
	return c.Req.URL.Query().Get(key)
}

// Status 设置HTTP StautsCode 并写入Header
func (c *Context) Status(code int) {
	c.StautsCode = code
	//使用提供的状态代码发送 HTTP 响应头
	c.Writer.WriteHeader(code)
}

// SetHeader 设置HTTP响应头
func (c *Context) SetHeader(key string, value string) {
	//Set 将与键关联的标头条目设置为单个元素value
	c.Writer.Header().Set(key, value)
}

// String 写入响应数据
func (c *Context) String(code int, format string, values ...interface{}) {
	c.SetHeader("Content-Type", "text/plain")
	c.Status(code)
	//Writer.Write() 在响应头部写入数据之后发送数据,通常被用来向HTTP客户端发送响应数据。
	c.Writer.Write([]byte(fmt.Sprintf(format, values...)))
}

// JSON 将json写入Writer中
func (c *Context) JSON(code int, object interface{}) {
	c.SetHeader("Content-Type", "application/json")
	c.StautsCode = code
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
func (c *Context) HTML(code int, html string) {
	c.SetHeader("Content-Type", "text/html")
	c.Status(code)
	c.Writer.Write([]byte(html))
}
