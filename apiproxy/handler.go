package apiproxy

import (
	"net/http"
)

func Handler() http.Handler {
	router := NewApiProxyRouter()
	router.BuildRoutingRules(
		requestIDFixture,
		logRequestMW,
		rateLimiterMW,
	)
	return router
}
