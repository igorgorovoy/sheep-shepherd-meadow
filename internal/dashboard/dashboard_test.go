package dashboard

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandler(t *testing.T) {
	h := Handler()

	tests := []struct {
		name       string
		path       string
		wantStatus int
		wantBody   string // substring expected in body
	}{
		{name: "root serves index", path: "/", wantStatus: http.StatusOK, wantBody: "<!DOCTYPE html>"},
		{name: "spa fallback for unknown route", path: "/pods/some-pod", wantStatus: http.StatusOK, wantBody: "<!DOCTYPE html>"},
		{name: "spa fallback for nested route", path: "/nodes", wantStatus: http.StatusOK, wantBody: "<!DOCTYPE html>"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()

			h.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if !strings.Contains(rec.Body.String(), tt.wantBody) {
				t.Fatalf("body = %q, want to contain %q", rec.Body.String(), tt.wantBody)
			}
		})
	}
}
