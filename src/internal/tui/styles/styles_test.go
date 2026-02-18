package styles

import (
	"strings"
	"testing"
)

func TestBadge(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   string
	}{
		{"local badge", "local", "[local]"},
		{"user badge", "user", "[user]"},
		{"unknown defaults to user", "", "[user]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Badge(tt.source)
			if !strings.Contains(got, tt.want) {
				t.Errorf("Badge(%q) = %q, want it to contain %q", tt.source, got, tt.want)
			}
		})
	}
}
