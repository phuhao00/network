package network

import (
	"github.com/gorilla/mux"
	"net/http"
	"sync/atomic"
)

type HttpRouter struct {
	mux      *mux.Router
	pageView int64 // requests process by multi goroutine，need atomic
}

func NewHttpRouter() *HttpRouter {
	return &HttpRouter{mux: mux.NewRouter().StrictSlash(true), pageView: 0}
}

func (g *HttpRouter) Handle(method, path string, h http.Handler) {
	g.mux.Handle(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&g.pageView, 1)
		SetRouteVars(r, mux.Vars(r))
		h.ServeHTTP(w, r)
	})).Methods(method)
}

func (g *HttpRouter) GetPageView() int64 {
	return atomic.LoadInt64(&g.pageView)
}

func (g *HttpRouter) HandleFunc(method, path string, h func(http.ResponseWriter, *http.Request)) {
	g.Handle(method, path, http.HandlerFunc(h))
}

func (g *HttpRouter) HandleStaticFile(method, path, dir string) {
	g.mux.PathPrefix(path).Handler(http.StripPrefix(path, http.FileServer(http.Dir(dir))))
}

func (g *HttpRouter) SetNotFoundHandler(h http.Handler) {
	g.mux.NotFoundHandler = h
}

func (g *HttpRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	g.mux.ServeHTTP(w, r)
}
