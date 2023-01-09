package apiproxy

import (
	"net/http"
)

func Handler() http.Handler {
	router := NewApiProxyRouter()
	router.buildRoutingRules()
	return router
}
