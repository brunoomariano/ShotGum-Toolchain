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
		{"global badge", "global", "[global]"},
		{"unknown defaults to global", "", "[global]"},
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

func TestTypeTag(t *testing.T) {
	tests := []struct {
		name string
		typ  string
		want string
	}{
		{"executable tag", "executable", "[bin]"},
		{"script tag", "script", "[sh]"},
		{"unknown defaults to sh", "", "[sh]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TypeTag(tt.typ)
			if !strings.Contains(got, tt.want) {
				t.Errorf("TypeTag(%q) = %q, want it to contain %q", tt.typ, got, tt.want)
			}
		})
	}
}
