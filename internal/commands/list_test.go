package commands

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/brunoomariano/ShotGum-Toolchain/internal/config"
	"github.com/brunoomariano/ShotGum-Toolchain/internal/registry"
)

// captureStdout runs f and returns whatever was written to os.Stdout.
func captureStdout(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	f()
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	os.Stdout = old
	return buf.String()
}

// setupHome creates a temp dir, sets HOME, writes the given config to the
// global config path, and changes CWD to tmpdir so LoadLocal finds nothing.
func setupHome(t *testing.T, cfg *config.Config) string {
	t.Helper()
	tmpdir := t.TempDir()
	t.Setenv("HOME", tmpdir)
	if err := config.Save(cfg, config.GlobalConfigPath()); err != nil {
		t.Fatalf("writing global config: %v", err)
	}
	origDir, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(origDir) })
	os.Chdir(tmpdir)
	return tmpdir
}

func loadTestRegistry(t *testing.T) *registry.Registry {
	t.Helper()
	reg, err := registry.Load()
	if err != nil {
		t.Fatalf("registry.Load(): %v", err)
	}
	return reg
}

func TestListCategories_Empty(t *testing.T) {
	setupHome(t, &config.Config{Version: "1"})
	reg := loadTestRegistry(t)

	out := captureStdout(func() { listCategories(reg) })
	if !strings.Contains(out, "No categories") {
		t.Errorf("output = %q, want 'No categories'", out)
	}
}

func TestListCategories_WithCategories(t *testing.T) {
	cfg := &config.Config{
		Version: "1",
		Categories: []config.Category{
			{Name: "tools", Description: "My tools"},
			{Name: "ops", Description: "Ops scripts"},
		},
	}
	setupHome(t, cfg)
	reg := loadTestRegistry(t)

	out := captureStdout(func() { listCategories(reg) })
	stripped := stripCmdsANSI(out)
	if !strings.Contains(stripped, "tools") {
		t.Errorf("output = %q, want 'tools'", stripped)
	}
	if !strings.Contains(stripped, "My tools") {
		t.Errorf("output = %q, want 'My tools'", stripped)
	}
}

func TestListScripts_NoScripts(t *testing.T) {
	cfg := &config.Config{
		Version: "1",
		Categories: []config.Category{
			{Name: "tools"},
		},
	}
	setupHome(t, cfg)
	reg := loadTestRegistry(t)

	out := captureStdout(func() { listScripts(reg, "tools") })
	if !strings.Contains(out, "No scripts found") {
		t.Errorf("output = %q, want 'No scripts found'", out)
	}
}

func TestListScripts_WithScripts(t *testing.T) {
	cfg := &config.Config{
		Version: "1",
		Categories: []config.Category{
			{Name: "tools"},
		},
		Scripts: []config.Script{
			{Name: "build", Category: "tools", Executable: "/bin/sh", Description: "Build it"},
		},
	}
	setupHome(t, cfg)
	reg := loadTestRegistry(t)

	out := captureStdout(func() { listScripts(reg, "tools") })
	stripped := stripCmdsANSI(out)
	if !strings.Contains(stripped, "build") {
		t.Errorf("output = %q, want 'build'", stripped)
	}
	if !strings.Contains(stripped, "Build it") {
		t.Errorf("output = %q, want 'Build it'", stripped)
	}
}

// stripCmdsANSI removes ANSI escape codes for plain-text assertions.
func stripCmdsANSI(s string) string {
	var result strings.Builder
	i := 0
	for i < len(s) {
		if s[i] == '\x1b' && i+1 < len(s) && s[i+1] == '[' {
			i += 2
			for i < len(s) && s[i] != 'm' {
				i++
			}
			i++
		} else {
			result.WriteByte(s[i])
			i++
		}
	}
	return result.String()
}
