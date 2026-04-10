package engine

import (
	"context"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/ethanpham86/AutoClickAccepted/internal/capture"
	"github.com/ethanpham86/AutoClickAccepted/internal/clicker"
	"github.com/ethanpham86/AutoClickAccepted/internal/logger"
	"github.com/ethanpham86/AutoClickAccepted/internal/ocr"
)

const ScaleFactor = 3

type Config struct {
	Keywords            []string
	ScanIntervalMs      int
	ConfidenceThreshold int
	DebugSaveCaptures   bool
	DebugDir            string
	DebugMode           bool
	UseBackgroundClick  bool
}

type Engine struct {
	region image.Rectangle
	config Config
	stats  Stats
}

type Stats struct {
	TotalScans  int
	TotalClicks int
	TotalErrors int
	StartTime   time.Time
}

func New(region image.Rectangle, cfg Config) *Engine {
	if cfg.ScanIntervalMs <= 0 {
		cfg.ScanIntervalMs = 2000
	}
	if cfg.ConfidenceThreshold <= 0 {
		cfg.ConfidenceThreshold = 30
	}
	return &Engine{
		region: region,
		config: cfg,
		stats:  Stats{StartTime: time.Now()},
	}
}

func (e *Engine) Stats() Stats {
	return e.stats
}

func (e *Engine) Run(ctx context.Context) error {
	interval := time.Duration(e.config.ScanIntervalMs) * time.Millisecond

	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════╗")
	fmt.Println("║        AUTO-CLICK ENGINE STARTED                 ║")
	fmt.Println("╠══════════════════════════════════════════════════╣")
	fmt.Printf("║  Region : %dx%d at (%d,%d)%s║\n",
		e.region.Dx(), e.region.Dy(), e.region.Min.X, e.region.Min.Y,
		strings.Repeat(" ", max(1, 33-len(fmt.Sprintf("%dx%d at (%d,%d)", e.region.Dx(), e.region.Dy(), e.region.Min.X, e.region.Min.Y)))))
	fmt.Printf("║  Keywords: %-36s║\n", truncate(strings.Join(e.config.Keywords, ", "), 36))
	fmt.Printf("║  Interval: %-36s║\n", fmt.Sprintf("%dms", e.config.ScanIntervalMs))
	fmt.Println("║  Press Ctrl+C to stop                            ║")
	fmt.Println("╚══════════════════════════════════════════════════╝")
	fmt.Println()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	e.scanAndClick()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			e.scanAndClick()
		}
	}
}

func (e *Engine) scanAndClick() {
	e.stats.TotalScans++

	capturePath, err := capture.CaptureToFile(e.region)
	if err != nil {
		e.stats.TotalErrors++
		logger.Error("[SCAN #%d] ✗ Capture failed: %v", e.stats.TotalScans, err)
		return
	}
	defer os.Remove(capturePath)

	if e.config.DebugSaveCaptures {
		e.saveDebugCapture(capturePath)
	}

	matches, err := ocr.DetectText(capturePath)
	if err != nil {
		if strings.Contains(err.Error(), "0xc000013a") || err == context.Canceled {
			return
		}
		e.stats.TotalErrors++
		logger.Error("[SCAN #%d] ✗ OCR failed: %v", e.stats.TotalScans, err)
		return
	}

	if e.config.DebugMode {
		var words []string
		for _, m := range matches {
			words = append(words, fmt.Sprintf("%s(%d%%)", m.Text, m.Confidence))
		}
		logger.Debug("[SCAN #%d] 🔍 Words: %s", e.stats.TotalScans, strings.Join(words, ", "))
	}

	found := ocr.FindMultiWordKeywords(matches, e.config.Keywords, e.config.ConfidenceThreshold)
	if len(found) == 0 {
		logger.Debug("[SCAN #%d] — No keywords found (%d words detected)", e.stats.TotalScans, len(matches))
		return
	}

	getPriority := func(kw string) int {
		for i, k := range e.config.Keywords {
			if strings.EqualFold(k, kw) {
				return i
			}
		}
		return 999
	}

	sort.Slice(found, func(i, j int) bool {
		p1, p2 := getPriority(found[i].Keyword), getPriority(found[j].Keyword)
		if p1 != p2 {
			return p1 < p2
		}
		return found[i].Confidence > found[j].Confidence
	})

	type coord struct{ x, y int }
	var clickedSpots []coord

	for _, hit := range found {
		centerX := ((hit.Bounds.Min.X + hit.Bounds.Max.X) / 2) / ScaleFactor
		centerY := ((hit.Bounds.Min.Y + hit.Bounds.Max.Y) / 2) / ScaleFactor

		absX, absY := clicker.RegionToScreen(e.region, centerX, centerY)

		duplicate := false
		for _, c := range clickedSpots {
			dx, dy := absX-c.x, absY-c.y
			if dx*dx+dy*dy < 400 {
				duplicate = true
				break
			}
		}
		if duplicate {
			continue
		}
		clickedSpots = append(clickedSpots, coord{x: absX, y: absY})

		logger.Info("[SCAN #%d] ✓ Found \"%s\" (conf=%d) → clicking at (%d, %d) [Background=%v]",
			e.stats.TotalScans, hit.Keyword, hit.Confidence, absX, absY, e.config.UseBackgroundClick)

		if err := clicker.ClickAt(absX, absY, e.config.UseBackgroundClick); err != nil {
			e.stats.TotalErrors++
			logger.Error("[SCAN #%d] ✗ Click failed: %v", e.stats.TotalScans, err)
			continue
		}

		e.stats.TotalClicks++
		time.Sleep(500 * time.Millisecond)
	}
}

func (e *Engine) saveDebugCapture(srcPath string) {
	destPath := filepath.Join(e.config.DebugDir, fmt.Sprintf("scan_%d.png", e.stats.TotalScans))
	data, err := os.ReadFile(srcPath)
	if err == nil {
		os.WriteFile(destPath, data, 0644)
	}
}

func SaveCaptureOnce(region image.Rectangle, destPath string) error {
	srcPath, err := capture.CaptureToFile(region)
	if err != nil {
		return err
	}
	defer os.Remove(srcPath)
	data, err := os.ReadFile(srcPath)
	if err != nil {
		return err
	}
	return os.WriteFile(destPath, data, 0644)
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
