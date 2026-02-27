package registry

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/brunoomariano/ShotGum-Toolchain/internal/config"
)

func TestRegistry_LoadsMakefileTargets(t *testing.T) {
	tmpdir := t.TempDir()
	t.Setenv("HOME", tmpdir)
	origDir, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(origDir) })
	os.Chdir(tmpdir)

	makefile := []byte("build:\n\t@echo build\n")
	if err := os.WriteFile(filepath.Join(tmpdir, "Makefile"), makefile, 0644); err != nil {
		t.Fatalf("write Makefile: %v", err)
	}

	cfg := &config.Config{
		Version:              "1",
		MakefileImport:       ptr(true),
		MakefileImportMode:   "all",
		MakefileImportSource: "parser",
	}
	cfgPath := filepath.Join(tmpdir, ".config", "shotgum", "config.yaml")
	if err := os.MkdirAll(filepath.Dir(cfgPath), 0755); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}
	if err := config.Save(cfg, cfgPath); err != nil {
		t.Fatalf("save config: %v", err)
	}

	reg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	cats := reg.GetCategories()
	var found bool
	for _, c := range cats {
		if c.Name == "make" && c.Source == "make" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected make category from Makefile")
	}

	scripts := reg.GetScripts("make")
	if len(scripts) == 0 || scripts[0].Name != "build" {
		t.Fatalf("expected build script from Makefile, got %+v", scripts)
	}

	entry := scripts[0]
	exe, args := reg.ResolveInvocation(entry)
	if exe != "make" || len(args) != 1 || args[0] != "build" {
		t.Fatalf("ResolveInvocation() = %q %v, want make build", exe, args)
	}
	if reg.ResolveHelpFlag(entry) != "" {
		t.Fatalf("ResolveHelpFlag() should be empty for make targets")
	}
}

func ptr(v bool) *bool {
	return &v
}
