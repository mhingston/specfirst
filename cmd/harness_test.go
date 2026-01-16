package cmd

import (
	"reflect"
	"testing"
)

func TestSplitArgs(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []string
		wantErr bool
	}{
		{
			name:  "simple",
			input: "--foo bar",
			want:  []string{"--foo", "bar"},
		},
		{
			name:  "quoted",
			input: "--name \"hello world\" --path='a b'",
			want:  []string{"--name", "hello world", "--path=a b"},
		},
		{
			name:  "escaped",
			input: "--foo\\ bar",
			want:  []string{"--foo bar"},
		},
		{
			name:    "unterminated quote",
			input:   "--foo \"bar",
			wantErr: true,
		},
		{
			name:    "unfinished escape",
			input:   "--foo \\",
			wantErr: true,
		},
		{
			name:  "empty",
			input: "  ",
			want:  nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := splitArgs(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("splitArgs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("splitArgs() = %v, want %v", got, tt.want)
			}
		})
	}
}
