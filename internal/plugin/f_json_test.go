package plugin

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/withholm/polyenv/internal/model"
)

func TestJsonFormatter_OutputFormat(t *testing.T) {
	tests := []struct {
		name    string
		asArray bool
		input   []model.StoredEnv
		want    string
		wantErr bool
	}{
		{
			name:    "map output - basic",
			asArray: false,
			input: []model.StoredEnv{
				{Key: "KEY1", Value: "VALUE1"},
				{Key: "KEY2", Value: "VALUE2"},
			},
			want:    `{
  "KEY1": "VALUE1",
  "KEY2": "VALUE2"
}`,
			wantErr: false,
		},
		{
			name:    "map output - empty",
			asArray: false,
			input:   []model.StoredEnv{},
			want:    `{}`,
			wantErr: false,
		},
		{
			name:    "array output - basic",
			asArray: true,
			input: []model.StoredEnv{
				{Key: "KEY1", Value: "VALUE1"},
				{Key: "KEY2", Value: "VALUE2"},
			},
			want: `[
  {
    "key": "KEY1",
    "value": "VALUE1"
  },
  {
    "key": "KEY2",
    "value": "VALUE2"
  }
]`,
			wantErr: false,
		},
		{
			name:    "array output - empty",
			asArray: true,
			input:   []model.StoredEnv{},
			want:    `[]`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &JsonFormatter{AsArray: tt.asArray}
			got, err := f.OutputFormat(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("JsonFormatter.OutputFormat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Unmarshal both to compare content, ignoring formatting differences
			var gotInterface, wantInterface interface{}/

			if err := json.Unmarshal(got, &gotInterface); err != nil {
				t.Fatalf("Failed to unmarshal actual output: %v\nOutput: %s", err, string(got))
			}
			if err := json.Unmarshal([]byte(tt.want), &wantInterface); err != nil {
				t.Fatalf("Failed to unmarshal expected output: %v", err)
			}

			if !reflect.DeepEqual(gotInterface, wantInterface) {
				t.Errorf("JsonFormatter.OutputFormat() got = %s, want %s", string(got), tt.want)
			}
		})
	}
}
