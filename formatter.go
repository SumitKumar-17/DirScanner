package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

func generateMarkdown(root string, result ScanResult, showStats bool) string {
	content := fmt.Sprintf("# Directory Structure for %s\n\n```\n%s```\n", root, result.Tree)
	if showStats {
		content += fmt.Sprintf(
			"\n---\n**Summary:** %d files, %d directories, %s total size\n",
			result.FileCount, result.DirCount, formatSize(result.TotalSize),
		)
	}
	return content
}

// writeOutput writes content to a file, or to stdout when filename is "-".
func writeOutput(filename, content string) error {
	if filename == "-" {
		_, err := fmt.Print(content)
		return err
	}
	logrus.Debugf("Writing to file: %s", filename)
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("error creating file %q: %w", filename, err)
	}
	defer f.Close()
	if _, err = f.WriteString(content); err != nil {
		return fmt.Errorf("error writing to file %q: %w", filename, err)
	}
	return nil
}

func ensureMarkdownExtension(filename string) string {
	if filename == "-" {
		return filename
	}
	if filepath.Ext(filename) != ".md" {
		return filename + ".md"
	}
	return filename
}
