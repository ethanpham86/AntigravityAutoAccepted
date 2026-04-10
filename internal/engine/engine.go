// Package engine coordinates the capture → OCR → match → click loop.
package engine

import (
	"context"
	"fmt"
	"image"
	"log"
	"os"
	"strings"
	"time"

	"github.com/ethanpham86/AutoClickAccepted/internal/capture"
	"github.com/ethanpham86/AutoClickAccepted/internal/clicker"
	"github.com/ethanpham86/AutoClickAccepted/internal/ocr"
)

// Config holds engine configuration.
type Config struct {
	Keywords            []string
	ScanIntervalMs      int
	ConfidenceThreshold int
}

// Engine is the main auto-click coordinator.
type Engine struct {
	region image.Rectangle
	config Config
	stats  Stats
}

// Stats tracks runtime statistics.
type Stats struct {
	TotalScans   int
	TotalClicks  int
	TotalErrors  int
	StartTime    time.Time
}

// New creates a new Engine.
func New(region image.Rectangle, cfg Config) *Engine {
	if cfg.ScanIntervalMs <= 0 {
		cfg.ScanIntervalMs = 2000
	}
	if cfg.ConfidenceThreshold <= 0 {
		cfg.ConfidenceThreshold = 60
	}
	return &Engine{
		region: region,
		config: cfg,
		stats:  Stats{StartTime: time.Now()},
	}
}

// Run starts the main scan loop. Blocks until context is cancelled.
func (e *Engine) Run(ctx context.Context) error {
	interval := time.Duration(e.config.ScanIntervalMs) * time.Millisecond

	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════╗")
	fmt.Println("║        AUTO-CLICK ENGINE STARTED                ║")
	fmt.Println("╠══════════════════════════════════════════════════╣")
	fmt.Printf("║  Region : %dx%d at (%d,%d)%s║\n",
		e.region.Dx(), e.region.Dy(), e.region.Min.X, e.region.Min.Y,
		strings.Repeat(" ", max(1, 34-len(fmt.Sprintf("%dx%d at (%d,%d)", e.region.Dx(), e.region.Dy(), e.region.Min.X, e.region.Min.Y)))))
	fmt.Printf("║  Keywords: %-37s║\n", truncate(strings.Join(e.config.Keywords, ", "), 37))
	fmt.Printf("║  Interval: %-37s║\n", fmt.Sprintf("%dms", e.config.ScanIntervalMs))
	fmt.Println("║  Press Ctrl+C to stop                           ║")
	fmt.Println("╚══════════════════════════════════════════════════╝")
	fmt.Println()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Run first scan immediately
	e.scanAndClick()

	for {
		select {
		case <-ctx.Done():
			e.printStats()
			return ctx.Err()
		case <-ticker.C:
			e.scanAndClick()
		}
	}
}

// scanAndClick performs one capture → OCR → match → click cycle.
func (e *Engine) scanAndClick() {
	e.stats.TotalScans++

	// 1. Capture screen region
	imagePath, err := capture.CaptureToFile(e.region)
	if err != nil {
		e.stats.TotalErrors++
		log.Printf("[SCAN #%d] ✗ Capture failed: %v", e.stats.TotalScans, err)
		return
	}
	// Important: clean up temp file to avoid file lock issues / clutter
	defer os.Remove(imagePath)

	// 2. Run OCR
	matches, err := ocr.DetectText(imagePath)
	if err != nil {
		e.stats.TotalErrors++
		log.Printf("[SCAN #%d] ✗ OCR failed: %v", e.stats.TotalScans, err)
		return
	}

	// 3. Find keyword matches (including multi-word)
	found := ocr.FindMultiWordKeywords(matches, e.config.Keywords, e.config.ConfidenceThreshold)
	if len(found) == 0 {
		log.Printf("[SCAN #%d] — No keywords found (%d words detected)", e.stats.TotalScans, len(matches))
		return
	}

	// 4. Click on the first match
	hit := found[0]
	centerX := (hit.Bounds.Min.X + hit.Bounds.Max.X) / 2
	centerY := (hit.Bounds.Min.Y + hit.Bounds.Max.Y) / 2

	// Convert from image-relative to absolute screen coordinates
	absX, absY := clicker.RegionToScreen(e.region, centerX, centerY)

	log.Printf("[SCAN #%d] ✓ Found \"%s\" (conf=%d) → clicking at (%d, %d)",
		e.stats.TotalScans, hit.Text, hit.Confidence, absX, absY)

	if err := clicker.ClickAt(absX, absY); err != nil {
		e.stats.TotalErrors++
		log.Printf("[SCAN #%d] ✗ Click failed: %v", e.stats.TotalScans, err)
		return
	}

	e.stats.TotalClicks++

	// Small delay after click to let UI respond
	time.Sleep(500 * time.Millisecond)
}

func (e *Engine) printStats() {
	elapsed := time.Since(e.stats.StartTime)
	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════╗")
	fmt.Println("║        SESSION SUMMARY                          ║")
	fmt.Println("╠══════════════════════════════════════════════════╣")
	fmt.Printf("║  Duration : %-36s║\n", elapsed.Truncate(time.Second).String())
	fmt.Printf("║  Scans    : %-36d║\n", e.stats.TotalScans)
	fmt.Printf("║  Clicks   : %-36d║\n", e.stats.TotalClicks)
	fmt.Printf("║  Errors   : %-36d║\n", e.stats.TotalErrors)
	fmt.Println("╚══════════════════════════════════════════════════╝")
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
