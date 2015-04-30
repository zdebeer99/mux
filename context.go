package mux

import (
	"net/http"
)

func NewContext(w http.ResponseWriter, req *http.Request) *HandlerContext {
	return &HandlerContext{w, req, nil, nil}
}

type (
	HttpContext interface {
		Response() http.ResponseWriter
		Request() *http.Request
	}
	SupportVars interface {
		SetVars(map[string]string)
	}
	SupportRoute interface {
		SetRoute(*Route)
	}
)

type HandlerContext struct {
	response http.ResponseWriter
	request  *http.Request
	Vars     map[string]string
	Route    *Route
}

func (this *HandlerContext) SetVars(vars map[string]string) {
	this.Vars = vars
}

func (this *HandlerContext) SetRoute(route *Route) {
	this.Route = route
}

func (this *HandlerContext) Response() http.ResponseWriter {
	return this.response
}

func (this *HandlerContext) Request() *http.Request {
	return this.request
}

type HandlerFunc func(interface{})

func (f HandlerFunc) ServeHTTP(c interface{}) {
	f(c)
}

type Handler interface {
	ServeHTTP(interface{})
}

type HandlerAdapter struct {
	handle http.Handler
}

func FromHttpHandler(handle http.Handler) *HandlerAdapter {
	return &HandlerAdapter{handle}
}

func (this *HandlerAdapter) ServeHTTP(mc interface{}) {
	switch val1 := mc.(type) {
	case HttpContext:
		this.handle.ServeHTTP(val1.Response(), val1.Request())
	}
}
