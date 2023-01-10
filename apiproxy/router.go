package apiproxy

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
)

type contextKey int

const (
	RoutingInfoKey contextKey = iota
)

type ApiProxyRouter struct {
	rootRouter     chi.Router
	giraffeRouter  chi.Router
	flagshipRouter chi.Router
}

func NewApiProxyRouter() *ApiProxyRouter {
	rootRouter := newRootRouter()
	giraffeRouter := newGiraffeRouter()
	flagshipRouter := newFlagshipRouter()
	return &ApiProxyRouter{
		rootRouter:     rootRouter,
		giraffeRouter:  giraffeRouter,
		flagshipRouter: flagshipRouter,
	}
}

func (r *ApiProxyRouter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.rootRouter.ServeHTTP(w, req)
}

func (r *ApiProxyRouter) Use(middlewares ...func(http.Handler) http.Handler) {
	r.rootRouter.Use(middlewares...)
}

// buildRoutingRules builds the Rules for routing between Giraffe and Flagship
// using the middlewares passed in params for requests that match prefix /api/v2
func (r *ApiProxyRouter) BuildRoutingRules(middlewares ...func(http.Handler) http.Handler) {
	r.rootRouter.Route("/api/v2", func(router chi.Router) {
		router.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				routeTo := req.Header.Get("X-Route-To")
				rctx := req.Context()
				if strings.EqualFold(routeTo, "GQL") {
					rctx = context.WithValue(rctx, RoutingInfoKey, true)
				} else {
					rctx = context.WithValue(rctx, RoutingInfoKey, false)
				}
				next.ServeHTTP(w, req.WithContext(rctx))
			})
		})
		router.Use(middlewares...)
		router.HandleFunc("/*", func(w http.ResponseWriter, req *http.Request) {
			if v, ok := req.Context().Value(RoutingInfoKey).(bool); v && ok {
				r.giraffeRouter.ServeHTTP(w, req)
			} else {
				r.flagshipRouter.ServeHTTP(w, req)
			}
		})
	})
	// All request not matching /api/v2/* get 404 response (cf. design)
	r.rootRouter.With(LoggingNonAPIv2MW).HandleFunc("/*", chi.NewMux().NotFoundHandler())

}

func newRootRouter() chi.Router {
	return chi.NewRouter()
}

func newGiraffeRouter() chi.Router {
	gqlRouter := chi.NewRouter().Route("/", func(r chi.Router) {
		r.Route("/calls", func(r chi.Router) {
			r.Get("/", func(w http.ResponseWriter, r *http.Request) {
				fmt.Println("calls collection Giraffe")
			})
			r.Post("/", func(w http.ResponseWriter, r *http.Request) {
				fmt.Println("calls wrong handler Giraffe")
			})
		})
		r.Route("/webhooks", func(r chi.Router) {
			r.Get("/", func(w http.ResponseWriter, r *http.Request) {
				fmt.Println("webhooks collection Giraffe")
			})
			r.Route("/{id:[0-9]+}", func(r chi.Router) {
				r.Get("/", func(w http.ResponseWriter, r *http.Request) {
					fmt.Printf("Get webhooks id#%s Giraffe\n", chi.URLParam(r, "id"))
				})
			})
		})
	})
	// Group handlers with a specific router.
	// This overwrite previous definitions
	gqlRouter.Group(func(r chi.Router) {
		r.Use(makeGiraffeRouteMW)
		r.Patch("/calls/{id}", func(w http.ResponseWriter, r *http.Request) {
			fmt.Println("calls Patch Giraffe")
		})
		r.Post("/calls", func(w http.ResponseWriter, r *http.Request) {
			fmt.Println("calls Post Giraffe")
		})
	})
	return gqlRouter
}

func newFlagshipRouter() chi.Router {
	return chi.NewRouter().Route("/", func(r chi.Router) {
		r.HandleFunc("/webhooks", func(w http.ResponseWriter, r *http.Request) {
			fmt.Println("webhooks Flagship")
		})
		r.HandleFunc("/tasks", func(w http.ResponseWriter, r *http.Request) {
			fmt.Println("tasks Flagship")
		})
	})
}
