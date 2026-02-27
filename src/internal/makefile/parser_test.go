package makefile

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func writeFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
	return path
}

func TestDiscoverMakefile(t *testing.T) {
	tmpdir := t.TempDir()
	if _, ok := DiscoverMakefile(tmpdir); ok {
		t.Fatal("expected no Makefile")
	}
	writeFile(t, tmpdir, "Makefile", "all:\n\t@echo ok\n")
	path, ok := DiscoverMakefile(tmpdir)
	if !ok {
		t.Fatal("expected Makefile to be discovered")
	}
	if filepath.Base(path) != "Makefile" {
		t.Fatalf("unexpected path: %s", path)
	}
}

func TestParseTargets_Parser_AllAndIncludes(t *testing.T) {
	tmpdir := t.TempDir()
	writeFile(t, tmpdir, "included.mk", "# from include\n#shotgum:includes\nrelease:\n\t@echo release\n")
	writeFile(t, tmpdir, "Makefile", `
include included.mk

# Build all
build:
	@echo build

#shotgum:ignore
secret:
	@echo secret

#shotgum:includes
pack:
	@echo pack

phony: deps
	@echo ok

%.o: %.c
	@echo pattern
`)

	path := filepath.Join(tmpdir, "Makefile")

	all, err := ParseTargets(path, ModeAll, SourceParser)
	if err != nil {
		t.Fatalf("ParseTargets(all) error: %v", err)
	}
	assertHasTargets(t, all, []string{"build", "pack", "phony", "release"})
	assertNotHasTargets(t, all, []string{"secret"})

	only, err := ParseTargets(path, ModeIncludesOnly, SourceParser)
	if err != nil {
		t.Fatalf("ParseTargets(includes_only) error: %v", err)
	}
	assertHasTargets(t, only, []string{"pack", "release"})
	assertNotHasTargets(t, only, []string{"build", "secret", "phony"})
}

func TestParseTargets_Parser_IncludeMissingAndVars(t *testing.T) {
	tmpdir := t.TempDir()
	writeFile(t, tmpdir, "Makefile", `
-include missing.mk
include $(SOME_VAR).mk

ok:
	@echo ok
`)
	path := filepath.Join(tmpdir, "Makefile")
	targets, err := ParseTargets(path, ModeAll, SourceParser)
	if err != nil {
		t.Fatalf("ParseTargets error: %v", err)
	}
	assertHasTargets(t, targets, []string{"ok"})
}

func TestParseTargets_MakeQP(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("make -qp test only on linux")
	}
	tmpdir := t.TempDir()
	writeFile(t, tmpdir, "Makefile", `
.PHONY: all
all: build
	@echo all

build:
	@echo build

#shotgum:ignore
secret:
	@echo secret
`)
	path := filepath.Join(tmpdir, "Makefile")
	targets, err := ParseTargets(path, ModeAll, SourceMakeQP)
	if err != nil {
		t.Fatalf("ParseTargets(make_qp) error: %v", err)
	}
	assertHasTargets(t, targets, []string{"all", "build"})
	assertNotHasTargets(t, targets, []string{"secret"})
}

func assertHasTargets(t *testing.T, targets []Target, want []string) {
	t.Helper()
	seen := map[string]bool{}
	for _, t := range targets {
		seen[t.Name] = true
	}
	for _, name := range want {
		if !seen[name] {
			t.Errorf("expected target %q", name)
		}
	}
}

func assertNotHasTargets(t *testing.T, targets []Target, want []string) {
	t.Helper()
	seen := map[string]bool{}
	for _, t := range targets {
		seen[t.Name] = true
	}
	for _, name := range want {
		if seen[name] {
			t.Errorf("did not expect target %q", name)
		}
	}
}
