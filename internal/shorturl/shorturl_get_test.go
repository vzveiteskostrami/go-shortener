package shorturl

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vzveiteskostrami/go-shortener/internal/auth"
)

func TestGetOwnerURLsListf(t *testing.T) {
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			GetOwnerURLsListf(tt.args.w, tt.args.r)
		})
	}
}

func BenchmarkGetOwnerURLsListf(b *testing.B) {
	for i := 0; i < b.N; i++ {
		r := httptest.NewRequest(http.MethodGet, "/status", nil)
		// создаём новый Recorder
		w := httptest.NewRecorder()
		c := context.WithValue(context.WithValue(r.Context(), auth.CPownerID, int64(555)), auth.CPownerValid, true)
		GetOwnerURLsListf(w, r.WithContext(c))
	}
}
