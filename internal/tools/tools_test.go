// This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
// If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.

package tools

import "testing"

func TestCheckDoubleDashS(t *testing.T) {
	type args struct {
		s    string
		name string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{

		{
			name: "should pass",
			args: args{
				s:    "other",
				name: "test",
			},
			wantErr: false,
		},
		{
			name: "should fail",
			args: args{
				s:    "est",
				name: "test",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckDoubleDashS(tt.args.s, tt.args.name)
			if tt.wantErr {
				if err == nil {
					t.Errorf("CheckDoubleDashS() error = %v, wantErr %v", err, tt.wantErr)
				}
			} else if err != nil {
				t.Errorf("CheckDoubleDashS() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestToIndentedJson(t *testing.T) {
	type args struct {
		data any
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "should pass",
			args: args{
				data: map[string]string{
					"key": "value",
				},
			},
			want: `{
  "key": "value"
}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToIndentedJSON(tt.args.data); got != tt.want {
				t.Errorf("ToIndentedJson() = %v, want %v", got, tt.want)
			}
		})
	}
}
