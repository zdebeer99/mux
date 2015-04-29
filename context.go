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

type MuxHandlerFunc func(*Context)

func (f MuxHandlerFunc) ServeHTTP(c *Context) {
	f(c)
}

type MuxHandler interface {
	ServeHTTP(*Context)
}

type MuxHandlerAdapter struct {
	handle http.Handler
}

func FromHttpHandler(handle http.Handler) *MuxHandlerAdapter {
	return &MuxHandlerAdapter{handle}
}

func (this *MuxHandlerAdapter) ServeHTTP(c *Context) {
	this.handle.ServeHTTP(c.Response, c.Request)
}
