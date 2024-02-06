package logging

import (
	"reflect"
	"testing"

	"go.uber.org/zap"
)

func TestS(t *testing.T) {
	tests := []struct {
		name string
		want *zap.SugaredLogger
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := S(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("S() = %v, want %v", got, tt.want)
			}
		})
	}
}
