package compressing

import (
	"testing"
)

func Test_gzipWriter_Write(t *testing.T) {
	type args struct {
		b []byte
	}
	tests := []struct {
		name    string
		gw      gzipWriter
		args    args
		want    int
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.gw.Write(tt.args.b)
			if (err != nil) != tt.wantErr {
				t.Errorf("gzipWriter.Write() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("gzipWriter.Write() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_gzipWriter_WriteHeader(t *testing.T) {
	type args struct {
		statusCode int
	}
	tests := []struct {
		name string
		gw   gzipWriter
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.gw.WriteHeader(tt.args.statusCode)
		})
	}
}
