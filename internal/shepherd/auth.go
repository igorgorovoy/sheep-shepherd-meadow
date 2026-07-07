package shepherd

import (
	"net/http"
	"os"
	"strings"
)

// authToken returns the API bearer token from SHEPHERD_API_TOKEN. Empty means
// auth is disabled (backward compatible).
func authToken() string {
	return strings.TrimSpace(os.Getenv("SHEPHERD_API_TOKEN"))
}

// authRequireBearer rejects requests without a valid Bearer token when
// SHEPHERD_API_TOKEN is set. Unauthenticated paths: OPTIONS, /healthz,
// non-/api/ routes (embedded SPA), and GET /api/v1/auth/status.
func (api *APIServer) authRequireBearer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := api.apiToken
		if token == "" {
			next.ServeHTTP(w, r)
			return
		}

		if r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}
		if r.URL.Path == "/healthz" {
			next.ServeHTTP(w, r)
			return
		}
		if !strings.HasPrefix(r.URL.Path, "/api/") {
			next.ServeHTTP(w, r)
			return
		}
		if r.URL.Path == "/api/v1/auth/status" && r.Method == http.MethodGet {
			next.ServeHTTP(w, r)
			return
		}

		got := strings.TrimSpace(r.Header.Get("Authorization"))
		want := "Bearer " + token
		if got != want {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
