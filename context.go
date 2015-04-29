package mux

import (
	"net/http"
)

type Context struct {
	Response http.ResponseWriter
	Request  *http.Request
	Vars     map[string]string
	Route    *Route
}

type HandlerFunc func(*Context)

func (f HandlerFunc) ServeHTTP(c *Context) {
	f(c)
}

type Handler interface {
	ServeHTTP(*Context)
}

type HandlerAdapter struct {
	handle http.Handler
}

func FromHttpHandler(handle http.Handler) *HandlerAdapter {
	return &HandlerAdapter{handle}
}

func (this *HandlerAdapter) ServeHTTP(c *Context) {
	this.handle.ServeHTTP(c.Response, c.Request)
}
