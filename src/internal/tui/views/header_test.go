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
	h := NewHeaderModel()
	if h.width != 0 {
		t.Errorf("width = %d, want 0", h.width)
	}
}

func TestHeaderModel_Init(t *testing.T) {
	h := NewHeaderModel()
	cmd := h.Init()
	if cmd != nil {
		t.Error("Init() should return nil (no animation)")
	}
}

func TestHeaderModel_UpdateWindowSize(t *testing.T) {
	h := NewHeaderModel()
	h2, _ := h.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	if h2.width != 120 {
		t.Errorf("width = %d, want 120", h2.width)
	}
}

func TestHeaderModel_View_ContainsMetadata(t *testing.T) {
	h := NewHeaderModel()
	h.width = 160
	view := h.View()
	stripped := stripANSI(view)
	if !strings.Contains(stripped, "Script Manager") {
		t.Errorf("View() should contain 'Script Manager': %q", stripped)
	}
	if !strings.Contains(stripped, "v0.4.0") {
		t.Errorf("View() should contain version 0.4.0: %q", stripped)
	}
	if strings.Contains(stripped, "https://github.com/brunoomariano/ShotGum-Toolchain") {
		t.Errorf("View() should not contain repo URL: %q", stripped)
	}
}

func TestHeaderModel_View_ZeroWidth_NoPanic(t *testing.T) {
	h := NewHeaderModel()
	// width = 0 → innerW clamped to 20 minimum; should not panic
	view := h.View()
	if view == "" {
		t.Error("View() with zero width returned empty string")
	}
}

func TestHeaderModel_View_NarrowWidth_UsesCompactLayout(t *testing.T) {
	h := NewHeaderModel()
	h.width = 30
	view := h.View()
	stripped := stripANSI(view)

	if !strings.Contains(stripped, "ShotGum") {
		t.Errorf("compact View() should contain ShotGum title: %q", stripped)
	}
	if !strings.Contains(stripped, "Script Manager") {
		t.Errorf("compact View() should contain subtitle: %q", stripped)
	}
	if strings.Contains(stripped, artRow1) {
		t.Errorf("compact View() should not render wide ASCII art row: %q", stripped)
	}
}
