package dbf

import (
	"context"
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
		want1 error
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := tt.f.FindLink(context.Background(), tt.args.link, tt.args.byLink)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FMStorage.FindLink() got = %v, want %v", got, tt.want)
			}
			//if got1 != tt.want1 {
			//	t.Errorf("FMStorage.FindLink() got1 = %v, want %v", got1, tt.want1)
			//}
		})
	}
}
