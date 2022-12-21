package core

import "net/http"

// 暂时玩不明白
type middleware func(http.Handler) http.Handler
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

//func (r *Router) Add(route string, h http.Handler) {
//	r.mux = make(map[string]http.Handler)
//	var mergedHandler = h
//
//	for i := len(r.middlewareChain) - 1; i >= 0; i-- {
//		mergedHandler = r.middlewareChain[i](mergedHandler)
//	}
//
//	r.mux[route] = mergedHandler
//}

func (r *Router) Chain(f http.Handler) http.Handler {
	for _, m := range r.middlewareChain {
		f = m(f)
	}
	return f
}
