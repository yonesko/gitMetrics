package main

import "testing"

func Test_extractExtension(t *testing.T) {
	tests := []struct {
		args string
		want string
	}{
		{"", ""},
	}
	for _, tt := range tests {
		t.Run(tt.args, func(t *testing.T) {
			if got := extractExtension(tt.args); got != tt.want {
				t.Errorf("extractExtension() = %v, want %v", got, tt.want)
			}
		})
	}
}
