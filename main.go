package main

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func main() {
	var (
		intermediate string
		last         string
		prefix       string
		branch       string
		exclude      []string
		maxDepth     int
		verbose      bool
		showHidden   bool
		showSizes    bool
		showStats    bool
	)

	rootCmd := &cobra.Command{
		Use:   "dirscanner <directory> [output file] [flags]",
		Short: "Scan a directory and output its structure as Markdown",
		Long: `dirscanner recursively scans a directory and produces a Markdown file
containing the directory tree. Output defaults to stdout when no file is given
or when the output argument is "-".`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if verbose {
				logrus.SetLevel(logrus.DebugLevel)
			}

			dir := args[0]
			output := "-"
			if len(args) == 2 {
				output = ensureMarkdownExtension(args[1])
			}

			ignoredNames, err := readDirIgnore(dir)
			if err != nil {
				return fmt.Errorf("error reading .dirignore: %w", err)
			}

			compiledPatterns, err := compilePatterns(exclude)
			if err != nil {
				return fmt.Errorf("error compiling exclude patterns: %w", err)
			}

			opts := ScanOptions{
				Style: ConnectorStyle{
					Intermediate: intermediate,
					Last:         last,
					Prefix:       prefix,
					Branch:       branch,
				},
				ExcludePatterns: compiledPatterns,
				IgnoredNames:    ignoredNames,
				MaxDepth:        maxDepth,
				ShowHidden:      showHidden,
				ShowSizes:       showSizes,
			}

			result, err := scanDirectory(dir, "", opts, 0)
			if err != nil {
				return fmt.Errorf("error scanning directory: %w", err)
			}

			markdown := generateMarkdown(dir, result, showStats)

			if err := writeOutput(output, markdown); err != nil {
				return fmt.Errorf("error writing output: %w", err)
			}

			if output != "-" {
				fmt.Printf("Markdown directory hierarchy written to %s\n", output)
			}
			return nil
		},
	}

	f := rootCmd.Flags()
	f.StringVar(&intermediate, "intermediate", "├── ", "Symbol for intermediate nodes")
	f.StringVar(&last, "last", "└── ", "Symbol for the last node in a directory")
	f.StringVar(&prefix, "prefix", "    ", "Prefix for child nodes")
	f.StringVar(&branch, "branch", "│   ", "Branch connector for intermediate nodes")
	f.StringSliceVar(&exclude, "exclude", []string{}, "Exclude entries matching these glob patterns (e.g. '*.txt')")
	f.IntVar(&maxDepth, "depth", -1, "Maximum traversal depth (-1 for unlimited)")
	f.BoolVarP(&verbose, "verbose", "v", false, "Enable debug logging")
	f.BoolVar(&showHidden, "show-hidden", false, "Include hidden files and directories (dotfiles)")
	f.BoolVar(&showSizes, "sizes", false, "Show file sizes next to filenames")
	f.BoolVar(&showStats, "stats", false, "Append summary statistics to the output")

	if err := rootCmd.Execute(); err != nil {
		logrus.Fatal(err)
	}
}
