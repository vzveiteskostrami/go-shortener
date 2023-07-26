package dbf

import (
	"reflect"
	"testing"
)

func TestFMStorage_FindLink(t *testing.T) {
	type args struct {
		link   string
		byLink bool
	}
	tests := []struct {
		name  string
		f     *FMStorage
		args  args
		want  StorageURL
		want1 bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := tt.f.FindLink(tt.args.link, tt.args.byLink)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FMStorage.FindLink() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("FMStorage.FindLink() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
