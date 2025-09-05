package plugin

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/withholm/polyenv/internal/model"
)

func TestPassthroughFormatter_OutputFormat(t *testing.T) {
	tests := []struct {
		name    string
		input   []model.StoredEnv
		want    string
		wantErr bool
	}{
		{
			name: "single item",
			input: []model.StoredEnv{
				{Key: "KEY1", Value: "VALUE1", File: "file1.env", IsSecret: false},
			},
			// Note: The JSON keys are capitalized because the struct fields are exported.
			want:    `[{"Value":"VALUE1","Key":"KEY1","File":"file1.env","IsSecret":false}]`,
			wantErr: false,
		},
		{
			name: "multiple items",
			input: []model.StoredEnv{
				{Key: "KEY1", Value: "VALUE1", File: "file1.env", IsSecret: false},
				{Key: "KEY2", Value: "VALUE2", File: "file2.env", IsSecret: true},
			},
			want: `[{"Value":"VALUE1","Key":"KEY1","File":"file1.env","IsSecret":false},` +
				`{"Value":"VALUE2","Key":"KEY2","File":"file2.env","IsSecret":true}]`,
			wantErr: false,
		},
		{
			name:    "empty input",
			input:   []model.StoredEnv{},
			want:    `[]`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &PassthroughFormatter{}
			got, err := f.OutputFormat(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("PassthroughFormatter.OutputFormat() error = %v, wantErr %v", err, tt.wantErr)
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
				t.Errorf("PassthroughFormatter.OutputFormat() got = %s, want %s", string(got), tt.want)
			}
		})
	}
}
