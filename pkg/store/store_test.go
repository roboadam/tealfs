package store

import (
	"github.com/google/uuid"
	"os"
	"testing"
)

func TestPath_Save(t *testing.T) {
	tempDir, _ := os.MkdirTemp("", "*-test-save")
	defer os.RemoveAll(tempDir) // Clean up the temporary directory when the test is done

	type fields struct {
		id  uuid.UUID
		raw string
	}
	type args struct {
		hash []byte
		data []byte
	}
	tests := []struct {
		name    string
		path    Path
		args    args
		wantErr bool
	}{
		{
			name: "blah",
			path: NewPath(tempDir),
			args: args{
				hash: []byte{0x01, 0x0F, 0x5B},
				data: []byte{0x02, 0x10, 0x5C},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.path.Save(tt.args.hash, tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("Save() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
