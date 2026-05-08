package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --- patternToRegex ---

func TestPatternToRegex(t *testing.T) {
	cases := []struct {
		pattern string
		match   string
		want    bool
	}{
		{"*.txt", "file.txt", true},
		{"*.txt", "file.go", false},
		{"file?", "file1", true},
		{"file?", "file12", false},
		{"node_modules", "node_modules", true},
		{"node_modules", "node_modules_extra", false},
	}
	for _, tc := range cases {
		re, err := patternToRegex(tc.pattern)
		if err != nil {
			t.Fatalf("patternToRegex(%q) error: %v", tc.pattern, err)
		}
		compiled, _ := compilePatterns([]string{tc.pattern})
		got := compiled[0].MatchString(tc.match)
		if got != tc.want {
			t.Errorf("pattern=%q input=%q: got %v, want %v (regex=%s)", tc.pattern, tc.match, got, tc.want, re)
		}
	}
}

// --- formatSize ---

func TestFormatSize(t *testing.T) {
	cases := []struct {
		bytes int64
		want  string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
	}
	for _, tc := range cases {
		got := formatSize(tc.bytes)
		if got != tc.want {
			t.Errorf("formatSize(%d) = %q, want %q", tc.bytes, got, tc.want)
		}
	}
}

// --- ensureMarkdownExtension ---

func TestEnsureMarkdownExtension(t *testing.T) {
	if got := ensureMarkdownExtension("output"); got != "output.md" {
		t.Errorf("got %q, want output.md", got)
	}
	if got := ensureMarkdownExtension("output.md"); got != "output.md" {
		t.Errorf("got %q, want output.md", got)
	}
	if got := ensureMarkdownExtension("-"); got != "-" {
		t.Errorf("stdout sentinel should be unchanged, got %q", got)
	}
}

// --- readDirIgnore ---

func TestReadDirIgnore(t *testing.T) {
	dir := t.TempDir()
	content := "# comment\nnode_modules\n\n.git\n"
	if err := os.WriteFile(filepath.Join(dir, ".dirignore"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	ignored, err := readDirIgnore(dir)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := ignored["node_modules"]; !ok {
		t.Error("expected node_modules to be ignored")
	}
	if _, ok := ignored[".git"]; !ok {
		t.Error("expected .git to be ignored")
	}
	if _, ok := ignored["# comment"]; ok {
		t.Error("comment lines should not be in ignored set")
	}
}

func TestReadDirIgnoreMissing(t *testing.T) {
	dir := t.TempDir()
	ignored, err := readDirIgnore(dir)
	if err != nil {
		t.Fatalf("expected no error for missing .dirignore, got: %v", err)
	}
	if len(ignored) != 0 {
		t.Errorf("expected empty map, got %v", ignored)
	}
}

// --- scanDirectory ---

func setupTestDir(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	dirs := []string{"a", "b", ".hidden"}
	for _, d := range dirs {
		if err := os.Mkdir(filepath.Join(root, d), 0755); err != nil {
			t.Fatal(err)
		}
	}
	files := map[string]string{
		"a/file1.txt": "hello",
		"a/file2.go":  "package main",
		"b/data.csv":  "1,2,3",
		".hidden/sec": "secret",
		"root.md":     "root file",
	}
	for path, content := range files {
		if err := os.WriteFile(filepath.Join(root, path), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func defaultOpts() ScanOptions {
	return ScanOptions{
		Style: ConnectorStyle{
			Intermediate: "├── ",
			Last:         "└── ",
			Prefix:       "    ",
			Branch:       "│   ",
		},
		MaxDepth: -1,
	}
}

func TestScanDirectoryBasic(t *testing.T) {
	root := setupTestDir(t)
	opts := defaultOpts()
	result, err := scanDirectory(root, "", opts, 0)
	if err != nil {
		t.Fatal(err)
	}
	// files: a/file1.txt, a/file2.go, b/data.csv, root.md = 4
	if result.FileCount != 4 {
		t.Errorf("file count: got %d, want 4", result.FileCount)
	}
	// dirs: a, b = 2
	if result.DirCount != 2 {
		t.Errorf("dir count: got %d, want 2", result.DirCount)
	}
}

func TestScanDirectoryShowHidden(t *testing.T) {
	root := setupTestDir(t)
	opts := defaultOpts()
	opts.ShowHidden = true
	result, err := scanDirectory(root, "", opts, 0)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result.Tree, ".hidden") {
		t.Error("expected .hidden directory in tree when ShowHidden=true")
	}
	// files: a/file1.txt, a/file2.go, b/data.csv, root.md, .hidden/sec = 5
	if result.FileCount != 5 {
		t.Errorf("file count with hidden: got %d, want 5", result.FileCount)
	}
}

func TestScanDirectoryExcludePattern(t *testing.T) {
	root := setupTestDir(t)
	opts := defaultOpts()
	patterns, err := compilePatterns([]string{"*.txt"})
	if err != nil {
		t.Fatal(err)
	}
	opts.ExcludePatterns = patterns
	result, err := scanDirectory(root, "", opts, 0)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(result.Tree, "file1.txt") {
		t.Error("file1.txt should be excluded by *.txt pattern")
	}
}

func TestScanDirectoryMaxDepth(t *testing.T) {
	root := setupTestDir(t)
	opts := defaultOpts()
	opts.MaxDepth = 0
	result, err := scanDirectory(root, "", opts, 0)
	if err != nil {
		t.Fatal(err)
	}
	// At depth 0, only top-level entries are shown; no recursion into subdirs.
	if strings.Contains(result.Tree, "file1.txt") {
		t.Error("file1.txt should not appear at depth 0")
	}
}

func TestScanDirectorySizes(t *testing.T) {
	root := setupTestDir(t)
	opts := defaultOpts()
	opts.ShowSizes = true
	result, err := scanDirectory(root, "", opts, 0)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result.Tree, "root.md (") {
		t.Error("expected file sizes in tree output")
	}
	if result.TotalSize == 0 {
		t.Error("expected non-zero total size")
	}
}

// --- generateMarkdown ---

func TestGenerateMarkdown(t *testing.T) {
	result := ScanResult{Tree: "a\n", FileCount: 1, DirCount: 0, TotalSize: 100}
	md := generateMarkdown("/tmp/foo", result, false)
	if !strings.Contains(md, "# Directory Structure for /tmp/foo") {
		t.Error("missing title in markdown")
	}
	if strings.Contains(md, "Summary") {
		t.Error("summary should not appear when showStats=false")
	}

	md = generateMarkdown("/tmp/foo", result, true)
	if !strings.Contains(md, "Summary") {
		t.Error("expected summary when showStats=true")
	}
}
