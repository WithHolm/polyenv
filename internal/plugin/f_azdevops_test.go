package plugin

import (
	"bytes"
	"testing"

	"github.com/withholm/polyenv/internal/model"
)

func TestAzDevopsFormatter_OutputFormat(t *testing.T) {
	tests := []struct {
		name    string
		input   []model.StoredEnv
		want    []byte
		wantErr bool
	}{
		{
			name: "non-secret variable",
			input: []model.StoredEnv{
				{Key: "KEY1", Value: "VALUE1", IsSecret: false},
			},
			want:    []byte("##vso[task.setvariable variable=KEY1;issecret=false]VALUE1\n"),
			wantErr: false,
		},
		{
			name: "secret variable",
			input: []model.StoredEnv{
				{Key: "SECRET_KEY", Value: "supersecret", IsSecret: true},
			},
			want:    []byte("##vso[task.setvariable variable=SECRET_KEY;issecret=true]supersecret\n"),
			wantErr: false,
		},
		{
			name: "multiple variables",
			input: []model.StoredEnv{
				{Key: "KEY1", Value: "VALUE1", IsSecret: false},
				{Key: "KEY2", Value: "VALUE2", IsSecret: true},
			},
			want: []byte("##vso[task.setvariable variable=KEY1;issecret=false]VALUE1\n" +
				"##vso[task.setvariable variable=KEY2;issecret=true]VALUE2\n"),
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
			f := &AzDevopsFormatter{}
			got, err := f.OutputFormat(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("AzDevopsFormatter.OutputFormat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !bytes.Equal(got, tt.want) {
				t.Errorf("AzDevopsFormatter.OutputFormat() = %s, want %s", string(got), string(tt.want))
			}
		})
	}
}
