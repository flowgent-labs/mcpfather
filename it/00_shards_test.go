package tests

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestConfig_ITShardPatternsCoverAllTests(t *testing.T) {
	root := repoRoot(t)
	makefile, err := os.ReadFile(filepath.Join(root, "Makefile"))
	if err != nil {
		t.Fatalf("read Makefile: %v", err)
	}

	shards := parseITShardPatterns(t, string(makefile))
	if len(shards) == 0 {
		t.Fatal("no IT shard patterns found in Makefile")
	}

	tests := listITTestFunctions(t, filepath.Join(root, "it"))
	var missing []string
	var duplicate []string
	for _, name := range tests {
		var hits []string
		for shard, pattern := range shards {
			if pattern.MatchString(name) {
				hits = append(hits, shard)
			}
		}
		switch len(hits) {
		case 0:
			missing = append(missing, name)
		case 1:
		default:
			duplicate = append(duplicate, name+" => "+strings.Join(hits, ","))
		}
	}

	if len(missing) > 0 || len(duplicate) > 0 {
		t.Fatalf("IT shard coverage mismatch:\nmissing=%v\nduplicate=%v", missing, duplicate)
	}
}

func parseITShardPatterns(t *testing.T, makefile string) map[string]*regexp.Regexp {
	t.Helper()

	assignRE := regexp.MustCompile(`(?m)^(IT_[A-Z0-9_]+_RE) := (.+)$`)
	shards := map[string]*regexp.Regexp{}
	for _, match := range assignRE.FindAllStringSubmatch(makefile, -1) {
		pattern := strings.ReplaceAll(match[2], "$$", "$")
		compiled, err := regexp.Compile(pattern)
		if err != nil {
			t.Fatalf("compile %s=%q: %v", match[1], pattern, err)
		}
		shards[match[1]] = compiled
	}
	return shards
}

func listITTestFunctions(t *testing.T, dir string) []string {
	t.Helper()

	files, err := filepath.Glob(filepath.Join(dir, "*_test.go"))
	if err != nil {
		t.Fatalf("glob IT test files: %v", err)
	}
	funcRE := regexp.MustCompile(`(?m)^func (Test[A-Za-z0-9_]*)\(`)
	var tests []string
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		for _, match := range funcRE.FindAllStringSubmatch(string(data), -1) {
			if match[1] == "TestMain" {
				continue
			}
			tests = append(tests, match[1])
		}
	}
	return tests
}
