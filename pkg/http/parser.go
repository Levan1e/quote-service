package http

import (
	"net/http"

	"quote-service/pkg/logger"
)

func ListenAndServe(addr string, handler http.Handler) error {
	logger.Infof("Starting HTTP server on %s", addr)
	return http.ListenAndServe(addr, handler)
}
