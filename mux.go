// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mux

import (
	"fmt"
	"net/http"
	"path"
)

// NewRouter returns a new router instance.
func NewRouter() *Router {
	return &Router{namedRoutes: make(map[string]*Route)}
}

// Router registers routes to be matched and dispatches a handler.
//
// It implements the http.Handler interface, so it can be registered to serve
// requests:
//
//     var router = mux.NewRouter()
//
//     func main() {
//         http.Handle("/", router)
//     }
//
// Or, for Google App Engine, register it in a init() function:
//
//     func init() {
//         http.Handle("/", router)
//     }
//
// This will send all incoming requests to the router.
type Router struct {
	// Configurable Handler to be used when no route matches.
	NotFoundHandler Handler
	// Parent route, if this is a subrouter.
	parent parentRoute
	// Routes to be matched, in order.
	routes []*Route
	// Routes by name for URL building.
	namedRoutes map[string]*Route
	// See Router.StrictSlash(). This defines the flag for new routes.
	strictSlash bool
	// create a custom context passed to the handler function
	contextFactory func(w http.ResponseWriter, req *http.Request) interface{}

	handleApapter func(ctx interface{})
}

// Match matches registered routes against the request.
func (r *Router) Match(req *http.Request, match *RouteMatch) bool {
	for _, route := range r.routes {
		if route.Match(req, match) {
			return true
		}
	}
	return false
}

// ServeHTTP dispatches the handler registered in the matched route.
//
// When there is a match, the route variables can be retrieved calling
// mux.Vars(request).
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Clean path to canonical form and redirect.
	if p := cleanPath(req.URL.Path); p != req.URL.Path {

		// Added 3 lines (Philip Schlump) - It was droping the query string and #whatever from query.
		// This matches with fix in go 1.2 r.c. 4 for same problem.  Go Issue:
		// http://code.google.com/p/go/issues/detail?id=5252
		url := *req.URL
		url.Path = p
		p = url.String()

		w.Header().Set("Location", p)
		w.WriteHeader(http.StatusMovedPermanently)
		return
	}
	var match RouteMatch
	var handler Handler
	var ctx interface{}
	if r.contextFactory != nil {
		ctx = r.contextFactory(w, req)
	} else {
		ctx = NewContext(w, req)
	}
	if r.Match(req, &match) {
		handler = match.Handler
		if vcx, ok := ctx.(SupportVars); ok {
			vcx.SetVars(match.Vars)
		}
		if rcx, ok := ctx.(SupportRoute); ok {
			rcx.SetRoute(match.Route)
		}
	}
	if handler == nil {
		handler = r.NotFoundHandler
		if handler == nil {
			handler = FromHttpHandler(http.NotFoundHandler())
		}
	}

	if r.handleApapter != nil {
		r.handleApapter(handler.ServeHTTP)
	} else {
		handler.ServeHTTP(ctx)
	}
}

// ServeHTTPContext Can be used when the context class is already created by your middleware.
func (r *Router) ServeHTTPContext(c interface{}) {
	ctx := c.(HttpContext)
	req := ctx.Request()
	w := ctx.Response()

	// Clean path to canonical form and redirect.
	if p := cleanPath(req.URL.Path); p != req.URL.Path {

		// Added 3 lines (Philip Schlump) - It was droping the query string and #whatever from query.
		// This matches with fix in go 1.2 r.c. 4 for same problem.  Go Issue:
		// http://code.google.com/p/go/issues/detail?id=5252
		url := *req.URL
		url.Path = p
		p = url.String()

		w.Header().Set("Location", p)
		w.WriteHeader(http.StatusMovedPermanently)
		return
	}
	var match RouteMatch
	var handler Handler

	if r.Match(req, &match) {
		handler = match.Handler
		if vcx, ok := ctx.(SupportVars); ok {
			vcx.SetVars(match.Vars)
		}
		if rcx, ok := ctx.(SupportRoute); ok {
			rcx.SetRoute(match.Route)
		}
	}
	if handler == nil {
		handler = r.NotFoundHandler
		if handler == nil {
			handler = FromHttpHandler(http.NotFoundHandler())
		}
	}

	if r.handleApapter != nil {
		r.handleApapter(handler.ServeHTTP)
	} else {
		handler.ServeHTTP(ctx)
	}
}

// Get returns a route registered with the given name.
func (r *Router) Get(name string) *Route {
	return r.getNamedRoutes()[name]
}

// StrictSlash defines the trailing slash behavior for new routes. The initial
// value is false.
//
// When true, if the route path is "/path/", accessing "/path" will redirect
// to the former and vice versa. In other words, your application will always
// see the path as specified in the route.
//
// When false, if the route path is "/path", accessing "/path/" will not match
// this route and vice versa.
//
// Special case: when a route sets a path prefix using the PathPrefix() method,
// strict slash is ignored for that route because the redirect behavior can't
// be determined from a prefix alone. However, any subrouters created from that
// route inherit the original StrictSlash setting.
func (r *Router) StrictSlash(value bool) *Router {
	r.strictSlash = value
	return r
}

// ----------------------------------------------------------------------------
// parentRoute
// ----------------------------------------------------------------------------

// getNamedRoutes returns the map where named routes are registered.
func (r *Router) getNamedRoutes() map[string]*Route {
	if r.namedRoutes == nil {
		if r.parent != nil {
			r.namedRoutes = r.parent.getNamedRoutes()
		} else {
			r.namedRoutes = make(map[string]*Route)
		}
	}
	return r.namedRoutes
}

// getRegexpGroup returns regexp definitions from the parent route, if any.
func (r *Router) getRegexpGroup() *routeRegexpGroup {
	if r.parent != nil {
		return r.parent.getRegexpGroup()
	}
	return nil
}

func (r *Router) buildVars(m map[string]string) map[string]string {
	if r.parent != nil {
		m = r.parent.buildVars(m)
	}
	return m
}

// ----------------------------------------------------------------------------
// Route factories
// ----------------------------------------------------------------------------

// NewRoute registers an empty route.
func (r *Router) NewRoute() *Route {
	route := &Route{parent: r, strictSlash: r.strictSlash}
	r.routes = append(r.routes, route)
	return route
}

// Handle registers a new route with a matcher for the URL path.
// See Route.Path() and Route.Handler().
func (r *Router) Handle(path string, handler Handler) *Route {
	return r.NewRoute().Path(path).Handler(handler)
}

// Handle registers a new route with a matcher for the URL path.
// See Route.Path() and Route.Handler().
func (r *Router) HandleHttp(path string, handler http.Handler) *Route {
	return r.NewRoute().Path(path).Handler(FromHttpHandler(handler))
}

// HandleFunc registers a new route with a matcher for the URL path.
// See Route.Path() and Route.HandlerFunc().
func (r *Router) HandleFunc(path string, f func(interface{})) *Route {
	return r.NewRoute().Path(path).HandlerFunc(f)
}

// Headers registers a new route with a matcher for request header values.
// See Route.Headers().
func (r *Router) Headers(pairs ...string) *Route {
	return r.NewRoute().Headers(pairs...)
}

// Host registers a new route with a matcher for the URL host.
// See Route.Host().
func (r *Router) Host(tpl string) *Route {
	return r.NewRoute().Host(tpl)
}

// MatcherFunc registers a new route with a custom matcher function.
// See Route.MatcherFunc().
func (r *Router) MatcherFunc(f MatcherFunc) *Route {
	return r.NewRoute().MatcherFunc(f)
}

// Methods registers a new route with a matcher for HTTP methods.
// See Route.Methods().
func (r *Router) Methods(methods ...string) *Route {
	return r.NewRoute().Methods(methods...)
}

// Path registers a new route with a matcher for the URL path.
// See Route.Path().
func (r *Router) Path(tpl string) *Route {
	return r.NewRoute().Path(tpl)
}

// PathPrefix registers a new route with a matcher for the URL path prefix.
// See Route.PathPrefix().
func (r *Router) PathPrefix(tpl string) *Route {
	return r.NewRoute().PathPrefix(tpl)
}

// Queries registers a new route with a matcher for URL query values.
// See Route.Queries().
func (r *Router) Queries(pairs ...string) *Route {
	return r.NewRoute().Queries(pairs...)
}

// Schemes registers a new route with a matcher for URL schemes.
// See Route.Schemes().
func (r *Router) Schemes(schemes ...string) *Route {
	return r.NewRoute().Schemes(schemes...)
}

// BuildVars registers a new route with a custom function for modifying
// route variables before building a URL.
func (r *Router) BuildVarsFunc(f BuildVarsFunc) *Route {
	return r.NewRoute().BuildVarsFunc(f)
}

// SetContextFactory can be used to create custom context objects, the object must contain
// mux.Context as a anonmys type.
// Casting is left to the user.
// Default context factory returns &mux.Context{w, req, nil, nil}
func (r *Router) SetContextFactory(f func(w http.ResponseWriter, req *http.Request) interface{}) {
	r.contextFactory = f
}

func (r *Router) SetHandleAdapter(f func(ctx interface{})) {
	r.handleApapter = f
	panic("Not Implementet")
}

// ----------------------------------------------------------------------------
// Context
// ----------------------------------------------------------------------------

// RouteMatch stores information about a matched route.
type RouteMatch struct {
	Route   *Route
	Handler Handler
	Vars    map[string]string
}

type contextKey int

const (
	varsKey contextKey = iota
	routeKey
)

// ----------------------------------------------------------------------------
// Helpers
// ----------------------------------------------------------------------------

// cleanPath returns the canonical path for p, eliminating . and .. elements.
// Borrowed from the net/http package.
func cleanPath(p string) string {
	if p == "" {
		return "/"
	}
	if p[0] != '/' {
		p = "/" + p
	}
	np := path.Clean(p)
	// path.Clean removes trailing slash except for root;
	// put the trailing slash back if necessary.
	if p[len(p)-1] == '/' && np != "/" {
		np += "/"
	}
	return np
}

// uniqueVars returns an error if two slices contain duplicated strings.
func uniqueVars(s1, s2 []string) error {
	for _, v1 := range s1 {
		for _, v2 := range s2 {
			if v1 == v2 {
				return fmt.Errorf("mux: duplicated route variable %q", v2)
			}
		}
	}
	return nil
}

// mapFromPairs converts variadic string parameters to a string map.
func mapFromPairs(pairs ...string) (map[string]string, error) {
	length := len(pairs)
	if length%2 != 0 {
		return nil, fmt.Errorf(
			"mux: number of parameters must be multiple of 2, got %v", pairs)
	}
	m := make(map[string]string, length/2)
	for i := 0; i < length; i += 2 {
		m[pairs[i]] = pairs[i+1]
	}
	return m, nil
}

// matchInArray returns true if the given string value is in the array.
func matchInArray(arr []string, value string) bool {
	for _, v := range arr {
		if v == value {
			return true
		}
	}
	return false
}

// matchMap returns true if the given key/value pairs exist in a given map.
func matchMap(toCheck map[string]string, toMatch map[string][]string,
	canonicalKey bool) bool {
	for k, v := range toCheck {
		// Check if key exists.
		if canonicalKey {
			k = http.CanonicalHeaderKey(k)
		}
		if values := toMatch[k]; values == nil {
			return false
		} else if v != "" {
			// If value was defined as an empty string we only check that the
			// key exists. Otherwise we also check for equality.
			valueExists := false
			for _, value := range values {
				if v == value {
					valueExists = true
					break
				}
			}
			if !valueExists {
				return false
			}
		}
	}
	return true
}
