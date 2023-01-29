package core

import "net/http"

type middleware func(http.Handler) http.Handler
type ConverterFunc func(wr http.ResponseWriter, r *http.Request)
type Router struct {
	middlewareChain []middleware
	mux             map[string]http.Handler
}

func NewRouter() *Router {
	return &Router{}
}

func (r *Router) Use(m middleware) {
	r.middlewareChain = append(r.middlewareChain, m)
}

func (r *Router) Add(route string, h http.Handler) {
	r.mux = make(map[string]http.Handler)
	var mergedHandler = h
	for i := len(r.middlewareChain) - 1; i >= 0; i-- {
		// 反着把中间件函数放入
		mergedHandler = r.middlewareChain[i](mergedHandler)
	}
	r.mux[route] = mergedHandler
	http.Handle(route, mergedHandler)
}

// HandleChain  反着执行，才是正确的顺序
func (r *Router) HandleChain(router string, f ConverterFunc) {
	r.Add(router, http.HandlerFunc(f))
}
