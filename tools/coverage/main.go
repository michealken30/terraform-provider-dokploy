// Package main provides a coverage analysis tool that recognizes //coverage:ignore comments.
//
// Usage:
//
//	go run ./tools/coverage [options] [directories...]
//
// Options:
//
//	-coverprofile string   Path to coverage profile (default "coverage.out")
//	-threshold float       Minimum adjusted coverage percentage (default 0, disabled)
//	-v                     Verbose output showing ignored lines
//
// If no directories are specified, defaults to ./internal/...
package main

import (
	"bufio"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

const ignoreMarker = "//coverage:ignore"

type coverageBlock struct {
	file       string
	startLine  int
	startCol   int
	endLine    int
	endCol     int
	statements int
	count      int
}

type ignoredLine struct {
	file string
	line int
}

func main() {
	coverProfile := flag.String("coverprofile", "coverage.out", "Path to coverage profile")
	threshold := flag.Float64("threshold", 0, "Minimum adjusted coverage percentage (0 = disabled)")
	verbose := flag.Bool("v", false, "Verbose output showing ignored lines")
	flag.Parse()

	// Get directories to scan
	dirs := flag.Args()
	if len(dirs) == 0 {
		dirs = []string{"./internal"}
	}

	// Find all ignored lines in source files
	var ignoredLines []ignoredLine
	for _, dir := range dirs {
		lines, err := findIgnoredLines(dir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error finding ignored lines in %s: %v\n", dir, err)
			os.Exit(1)
		}
		ignoredLines = append(ignoredLines, lines...)
	}

	if *verbose && len(ignoredLines) > 0 {
		fmt.Println("Ignored lines (marked with //coverage:ignore):")
		for _, il := range ignoredLines {
			fmt.Printf("  %s:%d\n", il.file, il.line)
		}
		fmt.Println()
	}

	// Parse coverage profile
	blocks, err := parseCoverageProfile(*coverProfile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing coverage profile: %v\n", err)
		os.Exit(1)
	}

	// Calculate raw coverage
	rawCovered, rawTotal := calculateCoverage(blocks)
	rawPct := 0.0
	if rawTotal > 0 {
		rawPct = float64(rawCovered) / float64(rawTotal) * 100
	}

	// Adjust coverage by treating ignored lines as covered
	adjustedBlocks := adjustCoverage(blocks, ignoredLines)
	adjCovered, adjTotal := calculateCoverage(adjustedBlocks)
	adjPct := 0.0
	if adjTotal > 0 {
		adjPct = float64(adjCovered) / float64(adjTotal) * 100
	}

	// Count ignored statements
	ignoredStatements := adjCovered - rawCovered

	// Output results
	fmt.Printf("Coverage Report\n")
	fmt.Printf("===============\n")
	fmt.Printf("Raw coverage:      %.1f%% (%d/%d statements)\n", rawPct, rawCovered, rawTotal)
	fmt.Printf("Ignored lines:     %d lines (%d statements)\n", len(ignoredLines), ignoredStatements)
	fmt.Printf("Adjusted coverage: %.1f%% (%d/%d statements)\n", adjPct, adjCovered, adjTotal)

	// Check threshold
	if *threshold > 0 && adjPct < *threshold {
		fmt.Printf("\nFAIL: Adjusted coverage %.1f%% is below threshold %.1f%%\n", adjPct, *threshold)
		os.Exit(1)
	}

	if *threshold > 0 {
		fmt.Printf("\nPASS: Adjusted coverage %.1f%% meets threshold %.1f%%\n", adjPct, *threshold)
	}
}

func findIgnoredLines(root string) ([]ignoredLine, error) {
	var ignored []ignoredLine

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip vendor, .git, and other non-source directories
		if info.IsDir() {
			base := filepath.Base(path)
			if base == "vendor" || base == ".git" || base == "testdata" {
				return filepath.SkipDir
			}
			return nil
		}

		// Only process Go files (excluding test files for ignore markers)
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		fileIgnored, err := findIgnoredLinesInFile(path)
		if err != nil {
			return fmt.Errorf("processing %s: %w", path, err)
		}

		ignored = append(ignored, fileIgnored...)
		return nil
	})

	return ignored, err
}

func findIgnoredLinesInFile(path string) ([]ignoredLine, error) {
	var ignored []ignoredLine

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	// Also scan the raw file for inline comments that parser might not associate
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(content), "\n")
	for i, line := range lines {
		if strings.Contains(line, ignoreMarker) {
			ignored = append(ignored, ignoredLine{
				file: path,
				line: i + 1,
			})
		}
	}

	// Also check AST comments for block-style ignores
	for _, cg := range file.Comments {
		for _, c := range cg.List {
			if strings.Contains(c.Text, ignoreMarker) {
				pos := fset.Position(c.Pos())
				// Avoid duplicates
				found := false
				for _, il := range ignored {
					if il.line == pos.Line {
						found = true
						break
					}
				}
				if !found {
					ignored = append(ignored, ignoredLine{
						file: path,
						line: pos.Line,
					})
				}
			}
		}
	}

	// Check for function-level ignores
	ast.Inspect(file, func(n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok {
			if fn.Doc != nil {
				for _, c := range fn.Doc.List {
					if strings.Contains(c.Text, ignoreMarker) {
						// Mark all lines in the function as ignored
						start := fset.Position(fn.Body.Lbrace).Line
						end := fset.Position(fn.Body.Rbrace).Line
						for line := start; line <= end; line++ {
							ignored = append(ignored, ignoredLine{
								file: path,
								line: line,
							})
						}
					}
				}
			}
		}
		return true
	})

	return ignored, nil
}

func parseCoverageProfile(path string) ([]coverageBlock, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var blocks []coverageBlock
	scanner := bufio.NewScanner(file)

	// Skip mode line
	if scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "mode:") {
			return nil, fmt.Errorf("invalid coverage profile: missing mode line")
		}
	}

	// Parse coverage blocks
	// Format: file:startLine.startCol,endLine.endCol statements count
	re := regexp.MustCompile(`^(.+):(\d+)\.(\d+),(\d+)\.(\d+) (\d+) (\d+)$`)

	for scanner.Scan() {
		line := scanner.Text()
		matches := re.FindStringSubmatch(line)
		if matches == nil {
			continue
		}

		startLine, _ := strconv.Atoi(matches[2])
		startCol, _ := strconv.Atoi(matches[3])
		endLine, _ := strconv.Atoi(matches[4])
		endCol, _ := strconv.Atoi(matches[5])
		statements, _ := strconv.Atoi(matches[6])
		count, _ := strconv.Atoi(matches[7])

		blocks = append(blocks, coverageBlock{
			file:       matches[1],
			startLine:  startLine,
			startCol:   startCol,
			endLine:    endLine,
			endCol:     endCol,
			statements: statements,
			count:      count,
		})
	}

	return blocks, scanner.Err()
}

func calculateCoverage(blocks []coverageBlock) (covered, total int) {
	for _, b := range blocks {
		total += b.statements
		if b.count > 0 {
			covered += b.statements
		}
	}
	return
}

func adjustCoverage(blocks []coverageBlock, ignored []ignoredLine) []coverageBlock {
	// Build a map for quick lookup
	ignoredMap := make(map[string]map[int]bool)
	for _, il := range ignored {
		if ignoredMap[il.file] == nil {
			ignoredMap[il.file] = make(map[int]bool)
		}
		ignoredMap[il.file][il.line] = true
	}

	// Create adjusted blocks
	adjusted := make([]coverageBlock, len(blocks))
	for i, b := range blocks {
		adjusted[i] = b

		// Check if any line in this block is ignored
		fileIgnored := ignoredMap[b.file]
		if fileIgnored == nil {
			// Try with just the base path (coverage uses full module paths)
			for f, m := range ignoredMap {
				if strings.HasSuffix(b.file, f) {
					fileIgnored = m
					break
				}
			}
		}

		if fileIgnored != nil {
			for line := b.startLine; line <= b.endLine; line++ {
				if fileIgnored[line] {
					// Treat this block as covered
					adjusted[i].count = 1
					break
				}
			}
		}
	}

	return adjusted
}
