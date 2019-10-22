package main

import (
	log "github.com/cihub/seelog"
	"net/http"
	"time"
)

type Middleware struct {
}

// LoggingHandler log request with time-out
func (m Middleware) LoggingHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		timeStart := time.Now()
		next.ServeHTTP(w, r)
		timeEnd := time.Now()
		log.Infof("[%s] %s %v", r.Method, r.URL.String(), timeEnd.Sub(timeStart))
	}
	return http.HandlerFunc(fn)
}

// RecoverHandler recover from panic and return 500
func (m Middleware) RecoverHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Infof("recover from panic")
				http.Error(w, http.StatusText(500), 500)
			}
		}()
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
