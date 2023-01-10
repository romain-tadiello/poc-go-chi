package apiproxy

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/getoutreach/gobox/pkg/log"
	"github.com/go-chi/chi"
	"github.com/romain-tadiello/poc-go-chi/statusrecorder"
)

type ApiProxyRouter struct {
	rootRouter     chi.Router
	giraffeRouter  chi.Router
	flagshipRouter chi.Router
	routingInfo    string
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

func (r *ApiProxyRouter) LogRequestMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		wx := &statusrecorder.StatusRecorder{
			ResponseWriter: w,
			StatusCode:     200, //Default value if w.WriteHeader() is not called
		}
		defer func() {
			info := log.F{
				"http.status": wx.StatusCode,
				"endpoint":    r.routingInfo,
			}
			log.Info(req.Context(), "Handled Request", info)
		}()
		next.ServeHTTP(wx, req)
	})
}

// buildRoutingRules builds the Rules for routing between Giraffe and Flagship
// using the middlewares passed in params for requests that match prefix /api/v2
func (r *ApiProxyRouter) BuildRoutingRules(middlewares ...func(http.Handler) http.Handler) {
	r.rootRouter.Route("/api/v2", func(router chi.Router) {
		router.Use(middlewares...)
		router.Use(func(h http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				routeTo := req.Header.Get("X-Route-To")
				if strings.EqualFold(routeTo, "GQL") {
					r.routingInfo = "giraffe"
					r.giraffeRouter.ServeHTTP(w, req)
				} else {
					r.routingInfo = "server-apiv2"
					r.flagshipRouter.ServeHTTP(w, req)
				}
			})
		})
		//Mandatory so we go to the MW
		router.HandleFunc("/*", func(w http.ResponseWriter, r *http.Request) {
		})
	})
	// All request not matching /api/v2/* get 404 response (cf. design)
	r.rootRouter.Route("/", func(router chi.Router) {
		router.Use(LoggingNonAPIv2MW)
		router.HandleFunc("/*", chi.NewMux().NotFoundHandler())
	})
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
