package network

import (
	"net/http"
	"time"
)

var (
	maxHeaderBytes = 1 << 20
	readTimeout    = 60 * time.Second
	writeTimeout   = 60 * time.Second
)

func HttpServer(handler http.Handler) *http.Server {
	return &http.Server{
		Handler:        handler,
		MaxHeaderBytes: maxHeaderBytes,
		ReadTimeout:    readTimeout,
		WriteTimeout:   writeTimeout,
	}
}
