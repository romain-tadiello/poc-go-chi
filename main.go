package main

import (
	"net/http"

	"github.com/romain-tadiello/poc-go-chi/apiproxy"
)

func main() {
	http.ListenAndServe(":3000", apiproxy.Handler())
}
