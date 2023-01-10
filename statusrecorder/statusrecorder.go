// Copyright 2022 Outreach Corporation. All Rights Reserved.

// Description: This file defines the status recorder.
// Managed: false

// Package statusrecorder provides the implementation of a status recorder for apiproxy
package statusrecorder

import (
	"net/http"
)

// StatusRecorder is a wrapper for http.ResponseWriter that keeps track of the response size and status code
type StatusRecorder struct {
	http.ResponseWriter
	StatusCode int
	Size       int
	Written    bool
}

func (r *StatusRecorder) WriteHeader(code int) {
	r.Written = true
	r.StatusCode = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *StatusRecorder) Write(b []byte) (int, error) {
	r.Written = true
	size, err := r.ResponseWriter.Write(b)
	r.Size += size
	return size, err
}
