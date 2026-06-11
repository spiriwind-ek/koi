package gateway

import (
	"crypto/subtle"
	"net/http"
	"strings"
)

func (s *Server) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.APIKey == "" {
			next(w, r)
			return
		}

		// Check Authorization header
		if token, ok := bearerToken(r); ok {
			if subtle.ConstantTimeCompare([]byte(token), []byte(s.APIKey)) == 1 {
				next(w, r)
				return
			}
		}

		// Check query parameter
		if r.URL.Query().Get("api_key") == s.APIKey {
			next(w, r)
			return
		}

		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
	}
}

func bearerToken(r *http.Request) (string, bool) {
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer "), true
	}
	return "", false
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
