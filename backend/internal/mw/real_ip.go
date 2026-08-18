package mw

import (
	"net/http"
)

func RealIPMiddleware(trustX bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if ip := r.Header.Get("X-Forwarded-For"); ip != "" && trustX {
				r.RemoteAddr = ip
			}

			next.ServeHTTP(w, r)
		})
	}
}
