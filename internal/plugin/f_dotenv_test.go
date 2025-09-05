package plugin

import (
	"bytes"
	"testing"

	"github.com/withholm/polyenv/internal/model"
)

func TestDotenvFormatter_OutputFormat(t *testing.T) {
	tests := []struct {
		name    string
		input   []model.StoredEnv
		want    []byte
		wantErr bool
	}{
		{
			name: "basic key-value pairs",
			input: []model.StoredEnv{
				{Key: "KEY1", Value: "VALUE1"},
				{Key: "KEY2", Value: "VALUE2"},
			},
			want:    []byte("KEY1=VALUE1\nKEY2=VALUE2\n"),
			wantErr: false,
		},
		{
			name: "value with spaces",
			input: []model.StoredEnv{
				{Key: "KEY_WITH_SPACE", Value: "value with space"},
			},
			want:    []byte("KEY_WITH_SPACE=\"value with space\"\n"),
			wantErr: false,
		},
		{
			name: "value with special characters",
			input: []model.StoredEnv{
				{Key: "SPECIAL", Value: "value#with#comment"},
			},
			want:    []byte("SPECIAL=\"value#with#comment\"\n"),
			wantErr: false,
		},
		{
			name: "empty input",
			input:   []model.StoredEnv{},
			want:    []byte(""),
			wantErr: false,
		},
		{
			name: "multiple values sorted by key",
			input: []model.StoredEnv{
				{Key: "B", Value: "2"},
				{Key: "A", Value: "1"},
			},
			// godotenv.Marshal sorts keys alphabetically
			want:    []byte("A=1\nB=2\n"),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &DotenvFormatter{}
			got, err := f.OutputFormat(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("DotenvFormatter.OutputFormat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// Marshal can produce different orderings, so we compare line by line
			if !bytes.Equal(got, tt.want) {
				t.Errorf("DotenvFormatter.OutputFormat() = %v, want %v", string(got), string(tt.want))
			}
		})
	}
}
