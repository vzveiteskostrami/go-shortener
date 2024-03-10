package auth

import (
	"testing"

	"github.com/vzveiteskostrami/go-shortener/internal/logging"
)

func Test_getOwnerID(t *testing.T) {
	tests := []struct {
		name        string
		tokenString string
		want1       bool
	}{
		{"Wrong", "aaaaaaaaaaaaaaaaaaaaaaa", false},
		{"Right", "aaaaaaaaaaaaaaaaaaaaaaa", true},
	}
	tests[1].tokenString, _, _ = TokenService.MakeToken()
	logging.LoggingInit()
	TokenService.NewOWNERID = 0
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, got1 := GetOwnerID(tt.tokenString)
			if ((got1 != nil) && tt.want1) || ((got1 == nil) && !tt.want1) {
				t.Errorf("getOwnerID() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
