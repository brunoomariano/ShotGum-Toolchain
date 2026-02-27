package makefile

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// ImportMode controls which targets are imported.
type ImportMode string

const (
	ModeAll          ImportMode = "all"
	ModeIncludesOnly ImportMode = "includes_only"
)

// ImportSource controls how targets are discovered.
type ImportSource string

const (
	SourceParser ImportSource = "parser"
	SourceMakeQP ImportSource = "make_qp"
)

// Target represents a Makefile target with optional documentation.
type Target struct {
	Name string
	Doc  string
}

type targetMeta struct {
	Doc        string
	HasIgnore  bool
	HasInclude bool
}

var targetNamePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_./-]*$`)

// DiscoverMakefile looks for a Makefile in the given directory.
func DiscoverMakefile(dir string) (string, bool) {
	candidates := []string{"Makefile", "makefile", "GNUmakefile"}
	for _, name := range candidates {
		path := filepath.Join(dir, name)
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return path, true
		}
	}
	return "", false
}

// ParseTargets reads a Makefile (and its includes) and extracts targets.
func ParseTargets(path string, mode ImportMode, source ImportSource) ([]Target, error) {
	switch source {
	case SourceMakeQP:
		return parseTargetsWithMake(path, mode)
	default:
		return parseTargetsSimple(path, mode)
	}
}

func parseTargetsSimple(path string, mode ImportMode) ([]Target, error) {
	meta, orderedTargets, err := parseMeta(path)
	if err != nil {
		return nil, err
	}

	var result []Target
	seen := map[string]bool{}
	for _, name := range orderedTargets {
		if seen[name] {
			continue
		}
		seen[name] = true
		m := meta[name]
		if m.HasIgnore {
			continue
		}
		if mode == ModeIncludesOnly && !m.HasInclude {
			continue
		}
		result = append(result, Target{Name: name, Doc: m.Doc})
	}
	return result, nil
}

func parseTargetsWithMake(path string, mode ImportMode) ([]Target, error) {
	dir := filepath.Dir(path)
	cmd := exec.Command("make", "-qp", "-rR", "-f", path)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil && len(out) == 0 {
		return nil, fmt.Errorf("running make -qp: %w", err)
	}

	meta, _, metaErr := parseMeta(path)
	if metaErr != nil {
		return nil, metaErr
	}

	var result []Target
	seen := map[string]bool{}
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "\t") {
			continue
		}
		targets := parseTargetLine(line)
		if len(targets) == 0 {
			continue
		}
		for _, name := range targets {
			if seen[name] {
				continue
			}
			seen[name] = true
			m := meta[name]
			if m.HasIgnore {
				continue
			}
			if mode == ModeIncludesOnly && !m.HasInclude {
				continue
			}
			result = append(result, Target{Name: name, Doc: m.Doc})
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading make output: %w", err)
	}
	return result, nil
}

func parseMeta(path string) (map[string]targetMeta, []string, error) {
	meta := map[string]targetMeta{}
	var ordered []string
	visited := map[string]bool{}

	var parseFile func(string) error
	parseFile = func(p string) error {
		abs, err := filepath.Abs(p)
		if err != nil {
			return fmt.Errorf("resolving makefile path: %w", err)
		}
		if visited[abs] {
			return nil
		}
		visited[abs] = true

		f, err := os.Open(abs)
		if err != nil {
			return fmt.Errorf("opening makefile %s: %w", abs, err)
		}
		defer f.Close()

		dir := filepath.Dir(abs)
		scanner := bufio.NewScanner(f)

		var commentBlock []string
		for scanner.Scan() {
			line := scanner.Text()
			trimmed := strings.TrimSpace(line)

			if trimmed == "" {
				commentBlock = nil
				continue
			}

			if strings.HasPrefix(trimmed, "#") {
				comment := strings.TrimSpace(strings.TrimPrefix(trimmed, "#"))
				commentBlock = append(commentBlock, comment)
				continue
			}

			if includePaths, ok, missingOk := parseIncludeLine(trimmed); ok {
				for _, inc := range includePaths {
					if strings.Contains(inc, "$") {
						continue
					}
					incPath := inc
					if !filepath.IsAbs(incPath) {
						incPath = filepath.Join(dir, incPath)
					}
					if err := parseFile(incPath); err != nil {
						if missingOk {
							continue
						}
						return err
					}
				}
				commentBlock = nil
				continue
			}

			if targets := parseTargetLine(trimmed); len(targets) > 0 {
				hasIgnore, hasInclude := parseCommentDirectives(commentBlock)
				doc := buildDoc(commentBlock)
				for _, t := range targets {
					if _, ok := meta[t]; !ok {
						meta[t] = targetMeta{
							Doc:        doc,
							HasIgnore:  hasIgnore,
							HasInclude: hasInclude,
						}
					}
					ordered = append(ordered, t)
				}
				commentBlock = nil
				continue
			}

			commentBlock = nil
		}
		if err := scanner.Err(); err != nil {
			return fmt.Errorf("reading makefile %s: %w", abs, err)
		}
		return nil
	}

	if err := parseFile(path); err != nil {
		return nil, nil, err
	}
	return meta, ordered, nil
}

func parseIncludeLine(line string) ([]string, bool, bool) {
	// Strip inline comments
	if idx := strings.Index(line, "#"); idx >= 0 {
		line = strings.TrimSpace(line[:idx])
	}
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return nil, false, false
	}
	keyword := fields[0]
	switch keyword {
	case "include":
		return fields[1:], true, false
	case "-include", "sinclude":
		return fields[1:], true, true
	default:
		return nil, false, false
	}
}

func parseTargetLine(line string) []string {
	if strings.HasPrefix(line, "\t") {
		return nil
	}
	colon := strings.Index(line, ":")
	if colon <= 0 {
		return nil
	}
	left := strings.TrimSpace(line[:colon])
	if left == "" {
		return nil
	}
	if strings.Contains(left, "=") {
		return nil
	}
	if strings.HasSuffix(left, ":") {
		left = strings.TrimSuffix(left, ":")
		left = strings.TrimSpace(left)
	}
	fields := strings.Fields(left)
	var targets []string
	for _, f := range fields {
		if f == "" {
			continue
		}
		if !isValidTargetName(f) {
			continue
		}
		targets = append(targets, f)
	}
	return targets
}

func isValidTargetName(name string) bool {
	if strings.HasPrefix(name, ".") || strings.Contains(name, "%") {
		return false
	}
	return targetNamePattern.MatchString(name)
}

func parseCommentDirectives(lines []string) (hasIgnore, hasInclude bool) {
	for _, line := range lines {
		text := strings.ToLower(line)
		if strings.Contains(text, "shotgum:ignore") {
			hasIgnore = true
		}
		if strings.Contains(text, "shotgum:includes") {
			hasInclude = true
		}
	}
	return
}

func buildDoc(lines []string) string {
	var cleaned []string
	for _, line := range lines {
		text := strings.TrimSpace(line)
		if text == "" {
			continue
		}
		if strings.Contains(strings.ToLower(text), "shotgum:ignore") {
			continue
		}
		if strings.Contains(strings.ToLower(text), "shotgum:includes") {
			continue
		}
		cleaned = append(cleaned, text)
	}
	return strings.TrimSpace(strings.Join(cleaned, "\n"))
}
