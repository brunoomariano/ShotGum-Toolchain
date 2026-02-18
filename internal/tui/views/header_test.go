package views

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// stripANSI removes ANSI escape codes from a string for plain-text assertions.
func stripANSI(s string) string {
	var result strings.Builder
	i := 0
	for i < len(s) {
		if s[i] == '\x1b' && i+1 < len(s) && s[i+1] == '[' {
			i += 2
			for i < len(s) && s[i] != 'm' {
				i++
			}
			i++ // skip 'm'
		} else {
			result.WriteByte(s[i])
			i++
		}
	}
	return result.String()
}

func TestNewHeaderModel(t *testing.T) {
	h := NewHeaderModel("1.2.3", "github.com/user/repo/issues")
	if h.version != "1.2.3" {
		t.Errorf("version = %q, want 1.2.3", h.version)
	}
	if h.issues != "github.com/user/repo/issues" {
		t.Errorf("issues = %q, unexpected", h.issues)
	}
}

func TestHeaderModel_Init(t *testing.T) {
	h := NewHeaderModel("1.0", "issues")
	cmd := h.Init()
	if cmd != nil {
		t.Error("Init() should return nil")
	}
}

func TestHeaderModel_UpdateWindowSize(t *testing.T) {
	h := NewHeaderModel("1.0", "issues")
	h2, _ := h.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	if h2.width != 120 {
		t.Errorf("width = %d, want 120", h2.width)
	}
}

func TestHeaderModel_View_ContainsMetadata(t *testing.T) {
	h := NewHeaderModel("2.0.1", "github.com/org/repo/issues")
	h.width = 100
	view := h.View()
	stripped := stripANSI(view)
	if !strings.Contains(stripped, "SHOTGUM") {
		t.Errorf("View() should contain SHOTGUM: %q", stripped)
	}
	if !strings.Contains(stripped, "2.0.1") {
		t.Errorf("View() should contain version 2.0.1: %q", stripped)
	}
	if !strings.Contains(stripped, "github.com/org/repo/issues") {
		t.Errorf("View() should contain issues URL: %q", stripped)
	}
}

func TestHeaderModel_View_ZeroWidth_NoPanic(t *testing.T) {
	h := NewHeaderModel("1.0", "issues")
	// width = 0 → innerW clamped to 20 minimum; should not panic
	view := h.View()
	if view == "" {
		t.Error("View() with zero width returned empty string")
	}
}
