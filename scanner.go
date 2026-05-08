package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
)

// ConnectorStyle defines the symbols used when drawing the directory tree.
type ConnectorStyle struct {
	Intermediate string
	Last         string
	Prefix       string
	Branch       string
}

// ScanOptions holds all parameters that control directory scanning.
type ScanOptions struct {
	Style           ConnectorStyle
	ExcludePatterns []*regexp.Regexp
	IgnoredNames    map[string]struct{}
	MaxDepth        int
	ShowHidden      bool
	ShowSizes       bool
}

// ScanResult holds the rendered tree and aggregate statistics.
type ScanResult struct {
	Tree      string
	FileCount int
	DirCount  int
	TotalSize int64
}

func compilePatterns(patterns []string) ([]*regexp.Regexp, error) {
	compiled := make([]*regexp.Regexp, 0, len(patterns))
	for _, p := range patterns {
		regexStr, err := patternToRegex(p)
		if err != nil {
			return nil, fmt.Errorf("invalid pattern %q: %w", p, err)
		}
		re, err := regexp.Compile(regexStr)
		if err != nil {
			return nil, fmt.Errorf("failed to compile pattern %q: %w", p, err)
		}
		compiled = append(compiled, re)
	}
	return compiled, nil
}

func patternToRegex(pattern string) (string, error) {
	regexPattern := regexp.QuoteMeta(pattern)
	regexPattern = strings.ReplaceAll(regexPattern, `\*`, `.*`)
	regexPattern = strings.ReplaceAll(regexPattern, `\?`, `.`)
	regexPattern = "^" + regexPattern + "$"
	if _, err := regexp.Compile(regexPattern); err != nil {
		return "", fmt.Errorf("error compiling regex: %w", err)
	}
	return regexPattern, nil
}

func isExcluded(name string, opts ScanOptions) bool {
	if _, ok := opts.IgnoredNames[name]; ok {
		return true
	}
	for _, re := range opts.ExcludePatterns {
		if re.MatchString(name) {
			return true
		}
	}
	return false
}

func scanDirectory(root, prefix string, opts ScanOptions, currentDepth int) (ScanResult, error) {
	logrus.Debugf("Scanning %q at depth %d", root, currentDepth)

	var result ScanResult

	if opts.MaxDepth != -1 && currentDepth > opts.MaxDepth {
		return result, nil
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		return result, fmt.Errorf("error reading directory %q: %w", root, err)
	}

	var filtered []os.DirEntry
	for _, entry := range entries {
		if !opts.ShowHidden && strings.HasPrefix(entry.Name(), ".") {
			logrus.Debugf("Skipping hidden entry: %s", entry.Name())
			continue
		}
		if isExcluded(entry.Name(), opts) {
			logrus.Debugf("Excluding entry: %s", entry.Name())
			continue
		}
		filtered = append(filtered, entry)
	}

	var tree strings.Builder
	for i, entry := range filtered {
		connector := opts.Style.Intermediate
		newPrefix := prefix + opts.Style.Branch
		if i == len(filtered)-1 {
			connector = opts.Style.Last
			newPrefix = prefix + opts.Style.Prefix
		}

		label := entry.Name()
		if entry.IsDir() {
			result.DirCount++
		} else {
			result.FileCount++
			if info, err := entry.Info(); err == nil {
				result.TotalSize += info.Size()
				if opts.ShowSizes {
					label = fmt.Sprintf("%s (%s)", label, formatSize(info.Size()))
				}
			}
		}

		tree.WriteString(fmt.Sprintf("%s%s%s\n", prefix, connector, label))

		if entry.IsDir() {
			sub, err := scanDirectory(filepath.Join(root, entry.Name()), newPrefix, opts, currentDepth+1)
			if err != nil {
				return result, err
			}
			tree.WriteString(sub.Tree)
			result.FileCount += sub.FileCount
			result.DirCount += sub.DirCount
			result.TotalSize += sub.TotalSize
		}
	}

	result.Tree = tree.String()
	return result, nil
}

// formatSize converts a byte count to a human-readable string (e.g. "1.4 KB").
func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// readDirIgnore reads the .dirignore file and returns a set of names to skip.
// Lines starting with '#' are treated as comments and ignored.
func readDirIgnore(root string) (map[string]struct{}, error) {
	ignoredNames := make(map[string]struct{})
	path := filepath.Join(root, ".dirignore")

	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return ignoredNames, nil
		}
		return nil, fmt.Errorf("error opening .dirignore: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		ignoredNames[line] = struct{}{}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading .dirignore: %w", err)
	}
	return ignoredNames, nil
}
