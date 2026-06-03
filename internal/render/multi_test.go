package render

import "testing"

func TestSanitiseName(t *testing.T) {
	cases := []struct{ in, want string }{
		{"server-01", "server-01"},
		{"my profile", "my_profile"},
		{"node/edge", "node_edge"},
	}
	for _, tc := range cases {
		if got := sanitiseName(tc.in); got != tc.want {
			t.Errorf("sanitiseName(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
