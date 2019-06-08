package way

import (
	"context"
	"net/http"
	"strings"
)

const ( // HTTP Methods in this router
	WAY_GET      int = 0x01 // 1
	WAY_HEAD     int = 0x02 // 2
	WAY_POST     int = 0x04 // 4
	WAY_PUT      int = 0x08 // 8
	WAY_DELETE   int = 0x10 // 16
	WAY_OPTIONS  int = 0x20 // 32
	WAY_CONNECT  int = 0x40 // 64
	WAY_TRACE    int = 0x80 // 128
	WAY_WILDCARD int = 0xFF // 255 (SUM OF ALL TYPES)
)

// wayContextKey is the context key type for storing
// parameters in context.Context.
type wayContextKey string

// Router routes HTTP requests.
type Router struct {
	routes []*route
	// NotFound is the http.Handler to call when no routes
	// match. By default uses http.NotFoundHandler().
	NotFound http.Handler
}

// NewRouter makes a new Router.
func NewRouter() *Router {
	return &Router{
		NotFound: http.NotFoundHandler(),
	}
}

func (rtr *Router) pathSegments(p string) []string {
	return strings.Split(strings.Trim(p, "/"), "/")
}

func (rtr *Router) methodToI(m string) int {
	switch m {
	case "GET":
		return WAY_GET
	case "POST":
		return WAY_POST
	case "HEAD":
		return WAY_HEAD
	case "PUT":
		return WAY_POST
	case "DELETE":
		return WAY_DELETE
	case "OPTIONS":
		return WAY_OPTIONS
	case "CONNECT":
		return WAY_CONNECT
	case "TRACE":
		return WAY_TRACE
	}
	return 0
}

// Handle adds a handler with the specified method and pattern.
// Method can be any HTTP method string or "*" to match all methods.
// Pattern can contain path segments such as: /item/:id which is
// accessible via the Param function.
// If pattern ends with trailing /, it acts as a prefix.
func (rtr *Router) Handle(methods int, pattern string, handler http.Handler) {
	segsPath := rtr.pathSegments(pattern)
	route := &route{
		methods: methods,
		segs:    segsPath,
		segsLen: len(segsPath),
		handler: handler,
		prefix:  strings.HasSuffix(pattern, "/") || strings.HasSuffix(pattern, "..."),
	}
	rtr.routes = append(rtr.routes, route)
}

// ALL ...
func (rtr *Router) ALL(pattern string, handler http.Handler) {
	rtr.Handle(WAY_WILDCARD, pattern, handler)
}

// GET ...
func (rtr *Router) GET(pattern string, handler http.Handler) {
	rtr.Handle(WAY_GET, pattern, handler)
}

// HEAD ...
func (rtr *Router) HEAD(pattern string, handler http.Handler) {
	rtr.Handle(WAY_HEAD, pattern, handler)
}

// POST ...
func (rtr *Router) POST(pattern string, handler http.Handler) {
	rtr.Handle(WAY_POST, pattern, handler)
}

// PUT ...
func (rtr *Router) PUT(pattern string, handler http.Handler) {
	rtr.Handle(WAY_PUT, pattern, handler)
}

// DELETE ...
func (rtr *Router) DELETE(pattern string, handler http.Handler) {
	rtr.Handle(WAY_DELETE, pattern, handler)
}

// ServeHTTP routes the incoming http.Request based on method and path
// extracting path parameters as it goes.
func (rtr *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	reqMethod := rtr.methodToI(r.Method)
	if reqMethod == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("400 bad request\n"))
		return
	}

	segs := rtr.pathSegments(r.URL.Path)
	for _, route := range rtr.routes {
		if !route.hasMethods(reqMethod) {
			continue
		}
		if ctx, ok := route.match(r.Context(), segs); ok {
			route.handler.ServeHTTP(w, r.WithContext(ctx))
			return
		}
	}
	rtr.NotFound.ServeHTTP(w, r)
}

// Param gets the path parameter from the specified Context.
// Returns an empty string if the parameter was not found.
func Param(ctx context.Context, param string) string {
	vStr, ok := ctx.Value(wayContextKey(param)).(string)
	if !ok {
		return ""
	}
	return vStr
}

type route struct {
	methods int
	segs    []string
	segsLen int //Risparmia un operazione len() per ogni Request
	handler http.Handler
	prefix  bool
}

func (rt *route) hasMethods(methods int) bool {
	return methods&rt.methods > 0
}

func (rt *route) match(ctx context.Context, segs []string) (context.Context, bool) {
	paramSegsLen := len(segs)

	if paramSegsLen > rt.segsLen && !rt.prefix {
		return nil, false
	}

	for i := 0; i < paramSegsLen; i++ {
		routeSeg := rt.segs[i]
		paramSeg := segs[i]

		if routeSeg != paramSeg {
			if strings.HasPrefix(routeSeg, ":") {
				routeSeg = strings.TrimPrefix(routeSeg, ":")
				ctx = context.WithValue(ctx, wayContextKey(routeSeg), paramSeg)
				continue
			}
			if strings.HasSuffix(routeSeg, "...") {
				if strings.HasPrefix(paramSeg, routeSeg[:len(routeSeg)-3]) {
					return ctx, true
				}
			}
			return nil, false
		}
	}

	return ctx, true
}
