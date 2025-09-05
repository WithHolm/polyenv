package plugin

import (
	"bytes"
	"testing"

	"github.com/withholm/polyenv/internal/model"
)

func TestPwshFormatter_OutputFormat(t *testing.T) {
	tests := []struct {
		name    string
		input   []model.StoredEnv
		want    []byte
		wantErr bool
	}{
		{
			name: "single key-value pair",
			input: []model.StoredEnv{
				{Key: "KEY1", Value: "VALUE1"},
			},
			want:    []byte("KEY1='VALUE1'"),
			wantErr: false,
		},
		{
			name: "multiple key-value pairs",
			input: []model.StoredEnv{
				{Key: "KEY1", Value: "VALUE1"},
				{Key: "KEY2", Value: "VALUE2"},
			},
			want:    []byte("KEY1='VALUE1';KEY2='VALUE2'"),
			wantErr: false,
		},
		{
			name: "value with single quote (current behavior)",
			input: []model.StoredEnv{
				{Key: "QUOTED", Value: "it's a value"},
			},
			// This is not valid PowerShell, but it is the expected output of the current implementation.
			want:    []byte("QUOTED='it's a value'"),
			wantErr: false,
		},
		{
			name:    "empty input",
			input:   []model.StoredEnv{},
			want:    []byte(""),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &PwshFormatter{}
			got, err := f.OutputFormat(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("PwshFormatter.OutputFormat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !bytes.Equal(got, tt.want) {
				t.Errorf("PwshFormatter.OutputFormat() = %s, want %s", string(got), string(tt.want))
			}
		})
	}
}
