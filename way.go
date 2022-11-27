package way

import (
	"net/http"
	"strings"
)

// Router routes HTTP requests.
type Router struct {
	tree *node
	// NotFound is the http.Handler to call when no routes
	// match. By default uses http.NotFoundHandler().
	NotFound http.Handler
}

// NewRouter makes a new Router.
func NewRouter() *Router {
	return &Router{
		tree:     new(node),
		NotFound: http.NotFoundHandler(),
	}
}

// Handle adds a handler with the specified method and pattern.
// Method can be any HTTP method string or "*" to match all methods.
// Pattern can contain path segments such as: /item/:id which is
// accessible via the Param function.
// If pattern ends with trailing /, it acts as a prefix.
func (r *Router) Handle(method, pattern string, handler http.Handler) {
	mh := &methodBoundHandler{strings.ToUpper(method), handler}
	pattern = withSlashPrefix(pattern)
	if strings.HasSuffix(pattern, "...") {
		pattern = pattern[:len(pattern)-3] + "/" // transform prefix dots into rooted subhandler.
	}
	if err := r.tree.Insert(withSlashPrefix(pattern), mh); err != nil {
		panic(err) // todo panic with more context?
	}
}

// HandleFunc is the http.HandlerFunc alternative to http.Handle.
func (r *Router) HandleFunc(method, pattern string, fn http.HandlerFunc) {
	r.Handle(method, pattern, fn)
}

// ServeHTTP routes the incoming http.Request based on method and path
// extracting path parameters as it goes.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	params := parpool.Get()
	defer parpool.Put(params)

	handlerMap := r.tree.Lookup(withSlashPrefix(req.URL.Path), params)
	if handlerMap == nil {
		r.NotFound.ServeHTTP(w, req)
		return
	}

	handler := handlerMap[strings.ToUpper(req.Method)]
	if handler == nil {
		handler = handlerMap["*"]
		if handler == nil {
			r.NotFound.ServeHTTP(w, req)
			return
		}
	}

	if len(*params) > 0 {
		req = req.WithContext(toParamContext(req.Context(), params))
	}

	handler.ServeHTTP(w, req)
}

func withSlashPrefix(value string) string {
	if strings.HasPrefix(value, "/") {
		return value
	}
	return "/" + value
}
