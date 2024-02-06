package dbf

import (
	"testing"

	_ "github.com/lib/pq"
)

func TestPGStorage_DBFSaveLink(t *testing.T) {
	type args struct {
		storageURLItem *StorageURL
	}
	tests := []struct {
		name string
		d    *PGStorage
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.d.DBFSaveLink(tt.args.storageURLItem)
		})
	}
}
