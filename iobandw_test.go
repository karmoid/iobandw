package main

import "testing"

func TestIsWildcard(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"file.ext", false},
		{"file??.ext", true},
		{"file*.ext", true},
		{"file*.?xt", true},
		{"", false},
	}
	for _, c := range cases {
		got := isWildcard(c.in)
		if got != c.want {
			t.Errorf("isWildcard(%q) == %v, want %v", c.in, got, c.want)
		}
	}
}
