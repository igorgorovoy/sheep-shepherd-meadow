package registry

import (
	"net/http"
	"os"
	"strings"
)

func meadowAuthToken() string {
	return strings.TrimSpace(os.Getenv("MEADOW_API_TOKEN"))
}

func (s *Server) cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) authRequireBearer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := s.apiToken
		if token == "" {
			next.ServeHTTP(w, r)
			return
		}

		if r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}
		if r.URL.Path == "/meadow/auth/status" && r.Method == http.MethodGet {
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

func (s *Server) handleMeadowAuthStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	respondJSON2(w, http.StatusOK, map[string]any{
		"auth_required": s.apiToken != "",
	})
}
