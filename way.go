package way

import (
	"context"
	"net/http"
	"strings"
)

// WayRouter routes HTTP requests.
type WayRouter struct {
	routes []*route
	// NotFound is the http.Handler to call when no routes
	// match. By default uses http.NotFoundHandler().
	NotFound http.Handler
	// WithValue puts the path parameter into the context.
	WithValue func(context.Context, string, string) context.Context
}

// NewWayRouter makes a new WayRouter.
func NewWayRouter() *WayRouter {
	return &WayRouter{
		NotFound: http.NotFoundHandler(),
		WithValue: func(ctx context.Context, key string, value string) context.Context {
			return context.WithValue(ctx, key, value)
		},
	}
}

func (r *WayRouter) pathSegments(p string) []string {
	return strings.Split(strings.Trim(p, "/"), "/")
}

// Handle adds a handler with the specified method and pattern.
// Method can be any HTTP method string or "*" to match all methods.
// Pattern can contain path segments such as: /item/:id which is
// accessible via context.Value("id").
func (r *WayRouter) Handle(method, pattern string, handler http.Handler) {
	route := &route{
		method:  strings.ToLower(method),
		segs:    r.pathSegments(pattern),
		handler: handler,
		prefix:  strings.HasSuffix(pattern, "/"),
	}
	r.routes = append(r.routes, route)
}

// ServeHTTP routes the incoming http.Request based on method and path
// extracting path parameters as it goes.
func (r *WayRouter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	method := strings.ToLower(req.Method)
	segs := r.pathSegments(req.URL.Path)
	for _, route := range r.routes {
		if route.method != method && route.method != "*" {
			continue
		}
		if ctx, ok := route.match(req.Context(), r, segs); ok {
			route.handler.ServeHTTP(w, req.WithContext(ctx))
			return
		}
	}
	r.NotFound.ServeHTTP(w, req)
}

type route struct {
	method  string
	segs    []string
	handler http.Handler
	prefix  bool
}

func (r *route) match(ctx context.Context, router *WayRouter, segs []string) (context.Context, bool) {
	if len(segs) > len(r.segs) && !r.prefix {
		return nil, false
	}
	for i, seg := range r.segs {
		if i > len(segs)-1 {
			return nil, false
		}
		isParam := false
		if strings.HasPrefix(seg, ":") {
			isParam = true
			seg = strings.TrimPrefix(seg, ":")
		}
		if !isParam { // verbatim check
			if seg != segs[i] {
				return nil, false
			}
		}
		if isParam {
			ctx = router.WithValue(ctx, seg, segs[i])
		}
	}
	return ctx, true
}
