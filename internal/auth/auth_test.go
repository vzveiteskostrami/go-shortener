package auth

import (
	"testing"

	"github.com/vzveiteskostrami/go-shortener/internal/logging"
)

func Test_getOwnerID(t *testing.T) {
	tests := []struct {
		name        string
		tokenString string
		want        int64
		want1       bool
	}{
		{"Wrong", "aaaaaaaaaaaaaaaaaaaaaaa", -1, false},
		{"Right", "aaaaaaaaaaaaaaaaaaaaaaa", 1, true},
	}
	tests[1].tokenString, tests[1].want, _ = tokenService.makeToken()
	logging.LoggingInit()
	tokenService.NewOWNERID = 0
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := getOwnerID(tt.tokenString)
			if got != tt.want {
				t.Errorf("getOwnerID() got = %v, want %v", got, tt.want)
			}
			if (got1 == nil) == tt.want1 {
				t.Errorf("getOwnerID() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
