package gee

import (
	"log"
	"net/http"
)

// router route struct
type router struct {
	handlers map[string]HandlerFunc
}

// NewRouter Create New Router object
func NewRouter() *router {
	return &router{handlers: make(map[string]HandlerFunc)}
}

// addRoute register handler
func (r *router) addRoute(method string, pattern string, handler HandlerFunc) {
	log.Printf("Route %4s - %s", method, pattern)
	key := Concat(method, "-", pattern)
	r.handlers[key] = handler
}

// handle  HTTP request  process
func (r *router) handle(c *Context) {
	key := Concat(c.Method, "-", c.Path)
	if handler, ok := r.handlers[key]; ok {
		handler(c)
	} else {
		c.String(http.StatusNotFound, "404 NOT FOUND: %s\n", c.Path)
	}
}
