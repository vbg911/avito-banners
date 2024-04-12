package middleware

import (
	"net/http"
)

var (
	userUrls = map[string]struct{}{
		"/user_banner": struct{}{},
	}
	userToken  = "user"
	adminToken = "admin"
)

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := userUrls[r.URL.Path]; ok {
			token := r.Header.Get("token")
			if token == userToken || token == adminToken {
				next.ServeHTTP(w, r)
			} else {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
		} else {
			token := r.Header.Get("token")
			if token == userToken {
				w.WriteHeader(http.StatusForbidden)
				return
			} else if token == adminToken {
				next.ServeHTTP(w, r)
			} else {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
		}
	})
}
