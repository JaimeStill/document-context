package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/JaimeStill/document-context/pkg/cache"
	"github.com/JaimeStill/document-context/pkg/config"
	"github.com/JaimeStill/document-context/pkg/document"
	"github.com/JaimeStill/document-context/pkg/encoding"
	"github.com/JaimeStill/document-context/pkg/image"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "convert":
		if err := runConvert(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "cache":
		if err := runCache(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`document-converter - Convert PDF documents to images

Usage:
  document-converter convert [flags]    Convert PDF pages to images
  document-converter cache <command>    Manage cache

Cache Commands:
  clear      Clear all cache entries
  inspect    Show cache directory structure
  stats      Show cache statistics

Convert Flags:
  -input <path>        PDF path (default: vim-cheatsheet.pdf)
  -output <dir>        Output directory (default: output/)
  -page <spec>         Page selection (default: all pages)
  -format <fmt>        Output format: png or jpg (default: png)
  -dpi <int>           Rendering DPI (default: 300)
  -quality <int>       JPEG quality 1-100 (default: 90)
  -cache-dir <path>    Cache directory (default: /tmp/document-context-cache)
  -no-cache            Disable caching
  -base64              Include base64 data URI files
  -brightness <int>    Brightness 0-200 (100=neutral)
  -contrast <int>      Contrast -100 to +100 (0=neutral)
  -saturation <int>    Saturation 0-200 (100=neutral)
  -rotation <int>      Rotation 0-360 degrees
  -background <color>  Background color (default: white)

Page Selection Syntax:
  3          Single page (page 3)
  1,3,5      Specific pages (pages 1, 3, and 5)
  2:5        Range (pages 2-5)
  2:         From page to end (page 2 to last page)
  :3         Up to page (pages 1-3)
  (empty)    All pages

Examples:
  document-converter convert
  document-converter convert -input report.pdf -format jpg -quality 85
  document-converter convert -page 1 -brightness 110 -contrast 5
  document-converter convert -base64
  document-converter cache clear`)
}

func runConvert(args []string) error {
	fs := flag.NewFlagSet("convert", flag.ExitOnError)

	input := fs.String("input", "vim-cheatsheet.pdf", "PDF path")
	output := fs.String("output", "output", "Output directory")
	pageSpec := fs.String("page", "", "Page selection")
	format := fs.String("format", "png", "Output format (png or jpg)")
	dpi := fs.Int("dpi", 300, "Rendering DPI")
	quality := fs.Int("quality", 90, "JPEG quality")
	cacheDir := fs.String("cache-dir", "/tmp/document-context-cache", "Cache directory")
	noCache := fs.Bool("no-cache", false, "Disable caching")
	includeBase64 := fs.Bool("base64", false, "Include base64 data URI files")

	brightness := fs.Int("brightness", 0, "Brightness 0-200 (0=not set)")
	contrast := fs.Int("contrast", 0, "Contrast -100 to +100 (0=not set)")
	saturation := fs.Int("saturation", 0, "Saturation 0-200 (0=not set)")
	rotation := fs.Int("rotation", 0, "Rotation 0-360 degrees (0=not set)")
	background := fs.String("background", "white", "Background color")

	if err := fs.Parse(args); err != nil {
		return err
	}

	doc, err := document.OpenPDF(*input)
	if err != nil {
		return fmt.Errorf("failed to open PDF: %w", err)
	}
	defer doc.Close()

	pages, err := parsePageSpec(*pageSpec, doc.PageCount())
	if err != nil {
		return fmt.Errorf("invalid page spec: %w", err)
	}

	cfg := config.ImageConfig{
		Format:  *format,
		DPI:     *dpi,
		Quality: *quality,
		Options: make(map[string]any),
	}

	if *brightness != 0 {
		cfg.Options["brightness"] = *brightness
	}
	if *contrast != 0 {
		cfg.Options["contrast"] = *contrast
	}
	if *saturation != 0 {
		cfg.Options["saturation"] = *saturation
	}
	if *rotation != 0 {
		cfg.Options["rotation"] = *rotation
	}
	cfg.Options["background"] = *background

	renderer, err := image.NewImageMagickRenderer(cfg)
	if err != nil {
		return fmt.Errorf("failed to create renderer: %w", err)
	}

	var c cache.Cache
	if !*noCache {
		cacheCfg := &config.CacheConfig{
			Name: "filesystem",
			Logger: config.LoggerConfig{
				Level:  config.LogLevelDisabled,
				Output: config.LoggerOutputDiscard,
			},
			Options: map[string]any{
				"directory": *cacheDir,
			},
		}
		c, err = cache.Create(cacheCfg)
		if err != nil {
			return fmt.Errorf("failed to create cache: %w", err)
		}
	}

	if err := os.MkdirAll(*output, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	printConvertHeader(*input, pages, doc.PageCount(), *format, *dpi, *output, *cacheDir, *noCache)

	start := time.Now()
	outputFiles := []string{}

	for _, pageNum := range pages {
		pageStart := time.Now()

		page, err := doc.ExtractPage(pageNum)
		if err != nil {
			return fmt.Errorf("failed to extract page %d: %w", pageNum, err)
		}

		imageData, err := page.ToImage(renderer, c)
		if err != nil {
			return fmt.Errorf("failed to convert page %d: %w", pageNum, err)
		}

		baseName := strings.TrimSuffix(filepath.Base(*input), filepath.Ext(*input))
		imagePath := filepath.Join(*output, fmt.Sprintf("%s-page-%d.%s", baseName, pageNum, *format))

		if err := os.WriteFile(imagePath, imageData, 0644); err != nil {
			return fmt.Errorf("failed to write image file: %w", err)
		}

		outputFiles = append(outputFiles, imagePath)

		if *includeBase64 {
			var imgFormat document.ImageFormat
			if *format == "png" {
				imgFormat = document.PNG
			} else {
				imgFormat = document.JPEG
			}

			dataURI, err := encoding.EncodeImageDataURI(imageData, imgFormat)
			if err != nil {
				return fmt.Errorf("failed to encode data URI: %w", err)
			}

			txtPath := filepath.Join(*output, fmt.Sprintf("%s-page-%d.txt", baseName, pageNum))
			if err := os.WriteFile(txtPath, []byte(dataURI), 0644); err != nil {
				return fmt.Errorf("failed to write data URI file: %w", err)
			}

			outputFiles = append(outputFiles, txtPath)
		}

		elapsed := time.Since(pageStart)
		fmt.Printf("Converting page %d... done (%dms)\n", pageNum, elapsed.Milliseconds())
	}

	totalElapsed := time.Since(start)

	fmt.Printf("\nConverted %d page(s) in %s\n", len(pages), formatDuration(totalElapsed))
	fmt.Println("Output files:")

	for _, filePath := range outputFiles {
		info, _ := os.Stat(filePath)
		size := formatBytes(info.Size())
		ext := filepath.Ext(filePath)

		if ext == ".txt" {
			fmt.Printf("  - %s (%s, base64 data URI)\n", filePath, size)
		} else {
			fmt.Printf("  - %s (%s)\n", filePath, size)
		}
	}

	return nil
}

func runCache(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("cache command requires subcommand: clear, inspect, or stats")
	}

	subcommand := args[0]

	fs := flag.NewFlagSet("cache", flag.ExitOnError)
	cacheDir := fs.String("cache-dir", "/tmp/document-context-cache", "Cache directory")

	if err := fs.Parse(args[1:]); err != nil {
		return err
	}

	cacheCfg := &config.CacheConfig{
		Name: "filesystem",
		Logger: config.LoggerConfig{
			Level:  config.LogLevelDisabled,
			Output: config.LoggerOutputDiscard,
		},
		Options: map[string]any{
			"directory": *cacheDir,
		},
	}

	c, err := cache.Create(cacheCfg)
	if err != nil {
		return fmt.Errorf("failed to create cache: %w", err)
	}

	switch subcommand {
	case "clear":
		return runCacheClear(c, *cacheDir)
	case "inspect":
		return runCacheInspect(*cacheDir)
	case "stats":
		return runCacheStats(*cacheDir)
	default:
		return fmt.Errorf("unknown cache subcommand: %s", subcommand)
	}
}

func runCacheClear(c cache.Cache, cacheDir string) error {
	fmt.Printf("Clearing cache: %s\n", cacheDir)

	if err := c.Clear(); err != nil {
		return fmt.Errorf("failed to clear cache: %w", err)
	}

	fmt.Println("Cache cleared successfully")
	return nil
}

func runCacheInspect(cacheDir string) error {
	fmt.Printf("Cache directory: %s\n\n", cacheDir)

	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Cache directory does not exist (empty cache)")
			return nil
		}
		return fmt.Errorf("failed to read cache directory: %w", err)
	}

	if len(entries) == 0 {
		fmt.Println("Cache is empty")
		return nil
	}

	fmt.Printf("Cache entries (%d):\n\n", len(entries))

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		keyPath := filepath.Join(cacheDir, entry.Name())
		files, err := os.ReadDir(keyPath)
		if err != nil {
			continue
		}

		fmt.Printf("%s/\n", entry.Name())
		for _, file := range files {
			if file.IsDir() {
				continue
			}
			info, _ := file.Info()
			size := formatBytes(info.Size())
			fmt.Printf("  └─ %s (%s)\n", file.Name(), size)
		}
		fmt.Println()
	}

	return nil
}

func runCacheStats(cacheDir string) error {
	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("Cache directory: %s\n", cacheDir)
			fmt.Println("Status: Empty (directory does not exist)")
			return nil
		}
		return fmt.Errorf("failed to read cache directory: %w", err)
	}

	entryCount := 0
	fileCount := 0
	totalSize := int64(0)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		entryCount++
		keyPath := filepath.Join(cacheDir, entry.Name())
		files, err := os.ReadDir(keyPath)
		if err != nil {
			continue
		}

		for _, file := range files {
			if file.IsDir() {
				continue
			}
			fileCount++
			info, _ := file.Info()
			totalSize += info.Size()
		}
	}

	fmt.Printf("Cache directory: %s\n", cacheDir)
	fmt.Printf("Cache entries: %d\n", entryCount)
	fmt.Printf("Total files: %d\n", fileCount)
	fmt.Printf("Total size: %s\n", formatBytes(totalSize))

	if entryCount == 0 {
		fmt.Println("Status: Empty")
	} else {
		fmt.Println("Status: Active")
	}

	return nil
}

func parsePageSpec(spec string, pageCount int) ([]int, error) {
	if spec == "" {
		pages := make([]int, pageCount)
		for i := range pageCount {
			pages[i] = i + 1
		}
		return pages, nil
	}

	if strings.Contains(spec, ":") {
		parts := strings.Split(spec, ":")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid range syntax: %s", spec)
		}

		var start, end int
		var err error

		if parts[0] == "" {
			start = 1
		} else {
			start, err = strconv.Atoi(parts[0])
			if err != nil || start < 1 {
				return nil, fmt.Errorf("invalid start page: %s", parts[0])
			}
		}

		if parts[1] == "" {
			end = pageCount
		} else {
			end, err = strconv.Atoi(parts[1])
			if err != nil || end < 1 {
				return nil, fmt.Errorf("invalid end page: %s", parts[1])
			}
		}

		if start > end {
			return nil, fmt.Errorf("start page (%d) must be <= end page (%d)", start, end)
		}

		if end > pageCount {
			return nil, fmt.Errorf("end page (%d) exceeds document page count (%d)", end, pageCount)
		}

		pages := make([]int, end-start+1)
		for i := range pages {
			pages[i] = start + i
		}
		return pages, nil
	}

	if strings.Contains(spec, ",") {
		parts := strings.Split(spec, ",")
		pages := make([]int, len(parts))

		for i, part := range parts {
			pageNum, err := strconv.Atoi(strings.TrimSpace(part))
			if err != nil {
				return nil, fmt.Errorf("invalid page number: %s", part)
			}
			if pageNum < 1 || pageNum > pageCount {
				return nil, fmt.Errorf("page %d out of range (1-%d)", pageNum, pageCount)
			}
			pages[i] = pageNum
		}
		return pages, nil
	}

	pageNum, err := strconv.Atoi(spec)
	if err != nil {
		return nil, fmt.Errorf("invalid page number: %s", spec)
	}
	if pageNum < 1 || pageNum > pageCount {
		return nil, fmt.Errorf("page %d out of range (1-%d)", pageNum, pageCount)
	}

	return []int{pageNum}, nil
}

func printConvertHeader(input string, pages []int, totalPages int, format string, dpi int, output, cacheDir string, noCache bool) {
	fmt.Printf("Converting: %s\n", input)

	if len(pages) == totalPages {
		fmt.Printf("Pages: 1-%d (%d pages)\n", totalPages, totalPages)
	} else {
		fmt.Printf("Pages: %v (%d page(s))\n", pages, len(pages))
	}

	fmt.Printf("Format: %s @ %d DPI\n", strings.ToUpper(format), dpi)
	fmt.Printf("Output: %s\n", output)

	if noCache {
		fmt.Println("Cache: disabled")
	} else {
		fmt.Printf("Cache: %s\n", cacheDir)
	}

	fmt.Println()
}

func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%.2fs", d.Seconds())
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%d KB", bytes/div)
}
