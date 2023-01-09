package apiproxy

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
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
// This should be called after setting all the previous MW
func (r *ApiProxyRouter) buildRoutingRules() {
	r.rootRouter.Route("/api/v2", func(router chi.Router) {
		router.Use(func(h http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				goToGiraffe := req.Header.Get("disable-route") == ""
				fmt.Println("Routing MW Called")
				if goToGiraffe {
					r.giraffeRouter.ServeHTTP(w, req)
				} else {
					r.flagshipRouter.ServeHTTP(w, req)
				}
			})
		})
		//Mandatory so we go to the MW
		router.HandleFunc("/*", func(w http.ResponseWriter, r *http.Request) {
		})
	})
}

func newRootRouter() chi.Router {
	return chi.NewRouter()
}

func newGiraffeRouter() chi.Router {
	return chi.NewRouter().Route("/", func(r chi.Router) {
		r.Route("/calls", func(r chi.Router) {
			r.MethodFunc("PATCH", "/{id}", func(w http.ResponseWriter, r *http.Request) {
				fmt.Println("calls Patch Giraffe")
			})
			r.MethodFunc("POST", "/", func(w http.ResponseWriter, r *http.Request) {
				fmt.Println("calls Post Giraffe")
			})
		})
		r.HandleFunc("/webhooks", func(w http.ResponseWriter, r *http.Request) {
			fmt.Println("webhooks Giraffe")
		})
		r.HandleFunc("/tasks", func(w http.ResponseWriter, r *http.Request) {
			fmt.Println("tasks Giraffe")
		})
	})
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
