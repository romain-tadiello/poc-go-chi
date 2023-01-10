package apiproxy

import (
	"fmt"
	"net/http"

	"github.com/getoutreach/gobox/pkg/log"
	"github.com/google/uuid"
	"github.com/romain-tadiello/poc-go-chi/statusrecorder"
)

func LoggingNonAPIv2MW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Warn(r.Context(), "Route is not under /api/v2 path: HTTP 404 Status")
		next.ServeHTTP(w, r)
	})
}

func logRequestMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		wx := &statusrecorder.StatusRecorder{
			ResponseWriter: w,
			StatusCode:     200, //Default value if w.WriteHeader() is not called
		}
		defer func() {
			endpoint := "server-apiv2"
			if v, ok := req.Context().Value(RoutingInfoKey).(bool); v && ok {
				endpoint = "giraffe"
			}
			info := log.F{
				"http.status": wx.StatusCode,
				"endpoint":    endpoint,
			}
			log.Info(req.Context(), "Handled Request", info)
		}()
		next.ServeHTTP(wx, req)
	})
}

// Middleware that sets X-Request-Id header to a random GUID if not present
// Helps to trace request across apiproxy, giraffe and flagship
func requestIDFixture(next http.Handler) http.Handler {
	f := func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Request-Id") == "" {
			r.Header.Set("X-Request-Id", uuid.New().String())
		}
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(f)
}

// MW for rate limiting example. To simulate we use X-RateLimit-Me header:
// if it is in the request and not empty it will "rate limit", otherwise it doesn't
func rateLimiterMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-RateLimit-Me") != "" {
			w.WriteHeader(429)
			log.Warn(r.Context(), fmt.Sprintf("Request#%s rate limited", r.Header.Get("X-Request-Id")))
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

// Composer of MWs that apply to all requests on their way to be sent to Giraffe
func makeGiraffeRouteMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Changing request context with GQL data (sort, filter...)")
		next.ServeHTTP(w, r)
	})
}
