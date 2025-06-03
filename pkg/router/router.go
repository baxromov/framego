package router

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

// Route represents a route in the router
type Route struct {
	Pattern     string
	Method      string
	Handler     http.HandlerFunc
	Middleware  []Middleware
	PathParams  []string
	RegexPattern *regexp.Regexp
}

// Middleware represents a middleware function
type Middleware func(http.HandlerFunc) http.HandlerFunc

// Router represents a HTTP router
type Router struct {
	Routes      []Route
	Middleware  []Middleware
	NotFound    http.HandlerFunc
	MethodNotAllowed http.HandlerFunc
}

// New creates a new router
func New() *Router {
	return &Router{
		Routes:     []Route{},
		Middleware: []Middleware{},
		NotFound:   defaultNotFound,
		MethodNotAllowed: defaultMethodNotAllowed,
	}
}

// defaultNotFound is the default handler for 404 Not Found errors
func defaultNotFound(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "404 Not Found", http.StatusNotFound)
}

// defaultMethodNotAllowed is the default handler for 405 Method Not Allowed errors
func defaultMethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "405 Method Not Allowed", http.StatusMethodNotAllowed)
}

// Use adds middleware to the router
func (r *Router) Use(middleware ...Middleware) {
	r.Middleware = append(r.Middleware, middleware...)
}

// Handle registers a new route with the router
func (r *Router) Handle(method, pattern string, handler http.HandlerFunc, middleware ...Middleware) {
	// Extract path parameters from pattern
	pathParams := extractPathParams(pattern)

	// Convert pattern to regex
	regexPattern := patternToRegex(pattern)

	route := Route{
		Pattern:     pattern,
		Method:      method,
		Handler:     handler,
		Middleware:  middleware,
		PathParams:  pathParams,
		RegexPattern: regexp.MustCompile(regexPattern),
	}

	r.Routes = append(r.Routes, route)
}

// GET registers a new GET route
func (r *Router) GET(pattern string, handler http.HandlerFunc, middleware ...Middleware) {
	r.Handle(http.MethodGet, pattern, handler, middleware...)
}

// POST registers a new POST route
func (r *Router) POST(pattern string, handler http.HandlerFunc, middleware ...Middleware) {
	r.Handle(http.MethodPost, pattern, handler, middleware...)
}

// PUT registers a new PUT route
func (r *Router) PUT(pattern string, handler http.HandlerFunc, middleware ...Middleware) {
	r.Handle(http.MethodPut, pattern, handler, middleware...)
}

// DELETE registers a new DELETE route
func (r *Router) DELETE(pattern string, handler http.HandlerFunc, middleware ...Middleware) {
	r.Handle(http.MethodDelete, pattern, handler, middleware...)
}

// PATCH registers a new PATCH route
func (r *Router) PATCH(pattern string, handler http.HandlerFunc, middleware ...Middleware) {
	r.Handle(http.MethodPatch, pattern, handler, middleware...)
}

// OPTIONS registers a new OPTIONS route
func (r *Router) OPTIONS(pattern string, handler http.HandlerFunc, middleware ...Middleware) {
	r.Handle(http.MethodOptions, pattern, handler, middleware...)
}

// HEAD registers a new HEAD route
func (r *Router) HEAD(pattern string, handler http.HandlerFunc, middleware ...Middleware) {
	r.Handle(http.MethodHead, pattern, handler, middleware...)
}

// ServeHTTP implements the http.Handler interface
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Find matching route
	var allow []string
	for _, route := range r.Routes {
		// Check if path matches
		matches := route.RegexPattern.FindStringSubmatch(req.URL.Path)
		if len(matches) > 0 {
			// Check if method matches
			if route.Method != req.Method {
				allow = append(allow, route.Method)
				continue
			}

			// Extract path parameters
			params := make(map[string]string)
			for i, param := range route.PathParams {
				params[param] = matches[i+1]
			}

			// Store path parameters in request context
			ctx := req.Context()
			ctx = context.WithValue(ctx, paramsKey, params)
			req = req.WithContext(ctx)

			// Apply router middleware
			handler := route.Handler
			for i := len(r.Middleware) - 1; i >= 0; i-- {
				handler = r.Middleware[i](handler)
			}

			// Apply route middleware
			for i := len(route.Middleware) - 1; i >= 0; i-- {
				handler = route.Middleware[i](handler)
			}

			// Call handler
			handler(w, req)
			return
		}
	}

	// If we get here, no route matched
	if len(allow) > 0 {
		// If methods are allowed, return 405 Method Not Allowed
		w.Header().Set("Allow", strings.Join(allow, ", "))
		r.MethodNotAllowed(w, req)
	} else {
		// Otherwise, return 404 Not Found
		r.NotFound(w, req)
	}
}

// extractPathParams extracts path parameters from a pattern
func extractPathParams(pattern string) []string {
	var params []string
	parts := strings.Split(pattern, "/")

	for _, part := range parts {
		if strings.HasPrefix(part, ":") {
			params = append(params, part[1:])
		}
	}

	return params
}

// patternToRegex converts a pattern to a regex
func patternToRegex(pattern string) string {
	parts := strings.Split(pattern, "/")

	for i, part := range parts {
		if strings.HasPrefix(part, ":") {
			// Replace :param with named capture group
			parts[i] = fmt.Sprintf("([^/]+)")
		}
	}

	// Ensure the pattern matches the entire path
	return "^" + strings.Join(parts, "/") + "$"
}

// pathParamsKey is the key used to store path parameters in the request context
type pathParamsKey struct{}

// paramsKey is the context key for path parameters
var paramsKey = pathParamsKey{}

// GetPathParams gets path parameters from the request context
func GetPathParams(r *http.Request) map[string]string {
	params, _ := r.Context().Value(paramsKey).(map[string]string)
	return params
}

// GetPathParam gets a path parameter from the request context
func GetPathParam(r *http.Request, name string) string {
	params := GetPathParams(r)
	return params[name]
}

// Group creates a new router group
func (r *Router) Group(prefix string, middleware ...Middleware) *Group {
	return &Group{
		Router:     r,
		Prefix:     prefix,
		Middleware: middleware,
	}
}

// Group represents a router group
type Group struct {
	Router     *Router
	Prefix     string
	Middleware []Middleware
}

// Use adds middleware to the group
func (g *Group) Use(middleware ...Middleware) {
	g.Middleware = append(g.Middleware, middleware...)
}

// Handle registers a new route with the group
func (g *Group) Handle(method, pattern string, handler http.HandlerFunc, middleware ...Middleware) {
	// Combine group and route middleware
	allMiddleware := append(g.Middleware, middleware...)

	// Combine group prefix and route pattern
	fullPattern := g.Prefix
	if !strings.HasSuffix(fullPattern, "/") {
		fullPattern += "/"
	}
	if strings.HasPrefix(pattern, "/") {
		pattern = pattern[1:]
	}
	fullPattern += pattern

	// Register route with router
	g.Router.Handle(method, fullPattern, handler, allMiddleware...)
}

// GET registers a new GET route with the group
func (g *Group) GET(pattern string, handler http.HandlerFunc, middleware ...Middleware) {
	g.Handle(http.MethodGet, pattern, handler, middleware...)
}

// POST registers a new POST route with the group
func (g *Group) POST(pattern string, handler http.HandlerFunc, middleware ...Middleware) {
	g.Handle(http.MethodPost, pattern, handler, middleware...)
}

// PUT registers a new PUT route with the group
func (g *Group) PUT(pattern string, handler http.HandlerFunc, middleware ...Middleware) {
	g.Handle(http.MethodPut, pattern, handler, middleware...)
}

// DELETE registers a new DELETE route with the group
func (g *Group) DELETE(pattern string, handler http.HandlerFunc, middleware ...Middleware) {
	g.Handle(http.MethodDelete, pattern, handler, middleware...)
}

// PATCH registers a new PATCH route with the group
func (g *Group) PATCH(pattern string, handler http.HandlerFunc, middleware ...Middleware) {
	g.Handle(http.MethodPatch, pattern, handler, middleware...)
}

// OPTIONS registers a new OPTIONS route with the group
func (g *Group) OPTIONS(pattern string, handler http.HandlerFunc, middleware ...Middleware) {
	g.Handle(http.MethodOptions, pattern, handler, middleware...)
}

// HEAD registers a new HEAD route with the group
func (g *Group) HEAD(pattern string, handler http.HandlerFunc, middleware ...Middleware) {
	g.Handle(http.MethodHead, pattern, handler, middleware...)
}
