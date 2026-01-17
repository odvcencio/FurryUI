// Record demo screencasts for all FluffyUI examples
//
// Usage:
//
//	go run ./scripts/record-demos
//	go run ./scripts/record-demos --duration 5s --out docs/demos
//
// Requirements:
//   - Go 1.25+
//   - Optional: agg (for GIF conversion)
//   - Optional: ffmpeg (for MP4 conversion)
//
// This tool records terminal sessions as .cast files (Asciicast v2 format).
// These can be viewed with asciinema or converted to GIF/video.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorGreen  = "\033[0;32m"
	colorYellow = "\033[1;33m"
	colorRed    = "\033[0;31m"
)

var (
	outDir     = flag.String("out", "docs/demos", "output directory for recordings")
	duration   = flag.Duration("duration", 5*time.Second, "recording duration per demo")
	aggPath    = flag.String("agg", "", "path to agg binary (auto-detected if empty)")
	aggTheme   = flag.String("theme", "monokai", "agg theme for GIF conversion")
	skipGif    = flag.Bool("no-gif", false, "skip GIF conversion")
	verbose    = flag.Bool("v", false, "verbose output")
	filterDemo = flag.String("demo", "", "only record specific demo (comma-separated)")
)

type demo struct {
	name     string
	path     string
	category string
}

func main() {
	flag.Parse()

	// Find project root
	rootDir, err := findProjectRoot()
	if err != nil {
		fatal("failed to find project root: %v", err)
	}

	// Ensure output directory exists
	demosDir := filepath.Join(rootDir, *outDir)
	if err := os.MkdirAll(demosDir, 0o755); err != nil {
		fatal("failed to create output directory: %v", err)
	}

	// Find agg binary
	aggBin := findAgg()

	// Define all demos
	demos := []demo{
		// Simple examples
		{"quickstart", "examples/quickstart", "Simple Examples"},
		{"counter", "examples/counter", "Simple Examples"},

		// Feature examples
		{"todo-app", "examples/todo-app", "Feature Examples"},
		{"command-palette", "examples/command-palette", "Feature Examples"},
		{"settings-form", "examples/settings-form", "Feature Examples"},
		{"dashboard", "examples/dashboard", "Feature Examples"},

		// Showcase
		{"candy-wars", "examples/candy-wars", "Showcase"},

		// Widget galleries
		{"widgets-gallery", "examples/widgets/gallery", "Widget Galleries"},
		{"widgets-layout", "examples/widgets/layout", "Widget Galleries"},
		{"widgets-input", "examples/widgets/input", "Widget Galleries"},
		{"widgets-data", "examples/widgets/data", "Widget Galleries"},
		{"widgets-navigation", "examples/widgets/navigation", "Widget Galleries"},
		{"widgets-feedback", "examples/widgets/feedback", "Widget Galleries"},
	}

	// Filter demos if requested
	if *filterDemo != "" {
		filter := make(map[string]bool)
		for _, name := range strings.Split(*filterDemo, ",") {
			filter[strings.TrimSpace(name)] = true
		}
		var filtered []demo
		for _, d := range demos {
			if filter[d.name] {
				filtered = append(filtered, d)
			}
		}
		demos = filtered
	}

	// Print header
	fmt.Printf("%sFluffyUI Demo Recorder%s\n", colorGreen, colorReset)
	fmt.Printf("Recording demos to: %s\n", demosDir)
	fmt.Printf("Duration per demo: %s\n", *duration)
	if aggBin != "" && !*skipGif {
		fmt.Printf("GIF conversion: enabled (using %s)\n", aggBin)
	} else if *skipGif {
		fmt.Printf("GIF conversion: disabled\n")
	} else {
		fmt.Printf("GIF conversion: disabled (agg not found)\n")
	}
	fmt.Println()

	// Track results
	var succeeded, failed int
	currentCategory := ""

	for _, d := range demos {
		// Print category header
		if d.category != currentCategory {
			currentCategory = d.category
			fmt.Printf("=== Recording %s ===\n", currentCategory)
		}

		// Record the demo
		castFile := filepath.Join(demosDir, d.name+".cast")
		examplePath := filepath.Join(rootDir, d.path)

		fmt.Printf("%sRecording: %s%s\n", colorYellow, d.name, colorReset)

		if err := recordDemo(examplePath, castFile); err != nil {
			fmt.Printf("  %s-> ERROR: %v%s\n", colorRed, err, colorReset)
			failed++
			fmt.Println()
			continue
		}

		// Check if cast file was created
		if _, err := os.Stat(castFile); os.IsNotExist(err) {
			fmt.Printf("  %s-> Recording failed (no output file)%s\n", colorRed, colorReset)
			failed++
			fmt.Println()
			continue
		}

		fmt.Printf("  -> Created: %s\n", castFile)

		// Convert to GIF if agg is available
		if aggBin != "" && !*skipGif {
			gifFile := filepath.Join(demosDir, d.name+".gif")
			fmt.Printf("  -> Converting to GIF...")

			if err := convertToGif(aggBin, castFile, gifFile); err != nil {
				fmt.Printf(" %sfailed: %v%s\n", colorRed, err, colorReset)
			} else {
				fmt.Printf(" done\n")
				fmt.Printf("  -> Created: %s\n", gifFile)
			}
		}

		succeeded++
		fmt.Println()
	}

	// Print summary
	fmt.Printf("%sRecording complete!%s\n", colorGreen, colorReset)
	fmt.Printf("  Succeeded: %d\n", succeeded)
	if failed > 0 {
		fmt.Printf("  %sFailed: %d%s\n", colorRed, failed, colorReset)
	}
	fmt.Println()
	fmt.Printf("Files created in: %s\n", demosDir)
	fmt.Println()
	fmt.Println("To view recordings:")
	fmt.Printf("  asciinema play %s/quickstart.cast\n", demosDir)
	fmt.Println()
	fmt.Println("To convert to GIF (requires agg):")
	fmt.Printf("  agg --theme monokai %s/quickstart.cast %s/quickstart.gif\n", demosDir, demosDir)
	fmt.Println()
	fmt.Println("To convert to MP4 (requires agg + ffmpeg):")
	fmt.Printf("  agg %s/quickstart.cast /tmp/quickstart.webm\n", demosDir)
	fmt.Printf("  ffmpeg -i /tmp/quickstart.webm %s/quickstart.mp4\n", demosDir)
}

func recordDemo(examplePath, castFile string) error {
	ctx, cancel := context.WithTimeout(context.Background(), *duration)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "run", examplePath)
	cmd.Env = append(os.Environ(), "FLUFFYUI_RECORD="+castFile)

	if *verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	err := cmd.Run()

	// Context deadline exceeded is expected (we're timing out intentionally)
	if ctx.Err() == context.DeadlineExceeded {
		return nil
	}

	// Check if cast file was created - if so, consider it a success
	// (some examples exit with non-zero when terminated)
	if _, statErr := os.Stat(castFile); statErr == nil {
		return nil
	}

	// Exit status from interrupt is also expected
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			// Process was killed by timeout or signal - this is expected
			if exitErr.ExitCode() == -1 || exitErr.ExitCode() == 130 || exitErr.ExitCode() == 2 {
				return nil
			}
		}
	}

	return err
}

func convertToGif(aggBin, castFile, gifFile string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	args := []string{"--theme", *aggTheme, "--font-size", "14", castFile, gifFile}
	cmd := exec.CommandContext(ctx, aggBin, args...)

	if *verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	return cmd.Run()
}

func findAgg() string {
	if *aggPath != "" {
		return *aggPath
	}

	// Check common locations
	locations := []string{
		"agg",
		filepath.Join(os.Getenv("HOME"), ".cargo/bin/agg"),
		"/usr/local/bin/agg",
		"/usr/bin/agg",
	}

	for _, loc := range locations {
		if path, err := exec.LookPath(loc); err == nil {
			return path
		}
	}

	return ""
}

func findProjectRoot() (string, error) {
	// Start from current directory and walk up looking for go.mod
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("could not find go.mod in any parent directory")
		}
		dir = parent
	}
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, colorRed+"Error: "+format+colorReset+"\n", args...)
	os.Exit(1)
}
