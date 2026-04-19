package engine

import (
	"context"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/ethanpham86/AutoClickAccepted/internal/capture"
	"github.com/ethanpham86/AutoClickAccepted/internal/clicker"
	"github.com/ethanpham86/AutoClickAccepted/internal/logger"
	"github.com/ethanpham86/AutoClickAccepted/internal/matcher"
	"github.com/ethanpham86/AutoClickAccepted/internal/ocr"
)

const ScaleFactor = 3
const maxClicksPerScan = 3  // Strict limit: max 3 clicks per scan to prevent spam
const dedupRadiusSq = 90000 // 300px radius squared — prevent clicking same button region

type Config struct {
	Keywords            []string
	Templates           []matcher.Template
	ScanIntervalMs      int
	ConfidenceThreshold int
	DebugSaveCaptures   bool
	DebugDir            string
	DebugMode           bool
	UseBackgroundClick  bool
	OCRAvailable        bool // Whether Tesseract is installed and usable
}

type Engine struct {
	region image.Rectangle
	config Config
	stats  Stats
	mu     sync.RWMutex
	paused bool
}

type Stats struct {
	TotalScans    int
	TotalClicks   int
	TotalErrors   int
	StartTime     time.Time
	ClicksByLabel map[string]int // Per-keyword/template click counts
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
		stats: Stats{
			StartTime:     time.Now(),
			ClicksByLabel: make(map[string]int),
		},
	}
}

func (e *Engine) Stats() Stats {
	e.mu.RLock()
	defer e.mu.RUnlock()
	// Return a copy of the stats with a copied map
	s := e.stats
	s.ClicksByLabel = make(map[string]int, len(e.stats.ClicksByLabel))
	for k, v := range e.stats.ClicksByLabel {
		s.ClicksByLabel[k] = v
	}
	return s
}

// Pause pauses the scan loop.
func (e *Engine) Pause() {
	e.mu.Lock()
	defer e.mu.Unlock()
	if !e.paused {
		e.paused = true
		logger.Info(">>> PAUSED -- Press F6 to resume")
	}
}

// Resume resumes the scan loop.
func (e *Engine) Resume() {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.paused {
		e.paused = false
		logger.Info(">>> RESUMED -- Scanning")
	}
}

// TogglePause toggles between paused and running states.
func (e *Engine) TogglePause() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.paused = !e.paused
	if e.paused {
		logger.Info(">>> PAUSED -- Press F6 to resume")
	} else {
		logger.Info(">>> RESUMED -- Scanning")
	}
}

// IsPaused returns whether the engine is currently paused.
func (e *Engine) IsPaused() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.paused
}

func (e *Engine) Run(ctx context.Context) error {
	interval := time.Duration(e.config.ScanIntervalMs) * time.Millisecond

	bgLabel := "Physical"
	if e.config.UseBackgroundClick {
		bgLabel = "Background"
	}

	ocrLabel := "OFF"
	if e.config.OCRAvailable {
		ocrLabel = "ON"
	}

	fmt.Println()
	fmt.Println("+----------------------------------------------------+")
	fmt.Println("|        AUTO-CLICK ENGINE STARTED                   |")
	fmt.Println("+----------------------------------------------------+")
	fmt.Printf("|  Region : %dx%d at (%d,%d)%s|\n",
		e.region.Dx(), e.region.Dy(), e.region.Min.X, e.region.Min.Y,
		strings.Repeat(" ", max(1, 35-len(fmt.Sprintf("%dx%d at (%d,%d)", e.region.Dx(), e.region.Dy(), e.region.Min.X, e.region.Min.Y)))))
	fmt.Printf("|  Keywords: %-38s|\n", truncate(strings.Join(e.config.Keywords, ", "), 38))
	fmt.Printf("|  Interval: %-38s|\n", fmt.Sprintf("%dms", e.config.ScanIntervalMs))
	fmt.Printf("|  Click   : %-38s|\n", bgLabel)
	fmt.Printf("|  OCR     : %-38s|\n", ocrLabel)
	fmt.Println("+----------------------------------------------------+")
	fmt.Println("|  F6 = Pause/Resume  |  F7 = Stop  |  Ctrl+C        |")
	fmt.Println("+----------------------------------------------------+")
	fmt.Println()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	e.scanAndClick()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if e.IsPaused() {
				continue
			}
			e.scanAndClick()
		}
	}
}

func (e *Engine) recordClick(label string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.stats.TotalClicks++
	e.stats.ClicksByLabel[label]++
}

func (e *Engine) scanAndClick() {
	e.stats.TotalScans++

	// Capture 1:1 image for exact matching
	capturePathRaw, err := capture.CaptureToFile(e.region, 1)
	if err != nil {
		e.stats.TotalErrors++
		logger.Error("[SCAN #%d] Capture failed: %v", e.stats.TotalScans, err)
		return
	}
	defer os.Remove(capturePathRaw)

	if e.config.DebugSaveCaptures {
		e.saveDebugCapture(capturePathRaw, "_1x")
	}

	bgTag := "Physical"
	if e.config.UseBackgroundClick {
		bgTag = "BG"
	}

	// 1. FAST IMAGE TEMPLATE MATCHING
	if len(e.config.Templates) > 0 {
		f, err := os.Open(capturePathRaw)
		if err == nil {
			capImg, _, err := image.Decode(f)
			f.Close()
			if err == nil {
				// Require 85% exact pixel similarity for click execution
				fastMatches, bestMissName, highestConf := matcher.MatchSingle(capImg, e.config.Templates, 0.85)

				// Hard cap: only process top N matches to prevent spam clicking
				if len(fastMatches) > maxClicksPerScan {
					logger.Info("[SCAN #%d] Capped %d template matches to %d", e.stats.TotalScans, len(fastMatches), maxClicksPerScan)
					fastMatches = fastMatches[:maxClicksPerScan]
				}

				if len(fastMatches) == 0 && highestConf >= 0.60 {
					logger.Info("[SCAN #%d] ~~ Almost Matched \"%s\" (%.1f%%, Needs: 85%%)", e.stats.TotalScans, bestMissName, highestConf*100)
				}

				if len(fastMatches) > 0 {
					var clickedSpots []struct{ x, y int }

					for _, tm := range fastMatches {
						centerX := (tm.Bounds.Min.X + tm.Bounds.Max.X) / 2
						centerY := (tm.Bounds.Min.Y + tm.Bounds.Max.Y) / 2

						absX, absY := clicker.RegionToScreen(e.region, centerX, centerY)

						duplicate := false
						for _, c := range clickedSpots {
							dx, dy := absX-c.x, absY-c.y
							if dx*dx+dy*dy < dedupRadiusSq { // 300 pixels radius
								duplicate = true
								break
							}
						}
						if duplicate {
							continue
						}
						clickedSpots = append(clickedSpots, struct{ x, y int }{absX, absY})

						logger.Click("\"%s\" @ (%d,%d) | Template %.0f%% | %s",
							tm.TemplateName, absX, absY, tm.Confidence*100, bgTag)

						if err := clicker.ClickAt(absX, absY, e.config.UseBackgroundClick); err != nil {
							e.stats.TotalErrors++
							logger.Error("[SCAN #%d] Click failed: %v", e.stats.TotalScans, err)
							continue
						}
						e.recordClick(tm.TemplateName)
						time.Sleep(300 * time.Millisecond)

						if len(clickedSpots) >= maxClicksPerScan {
							break
						}
					}
					// If visual match worked, skip expensive OCR completely.
					if len(clickedSpots) > 0 {
						return
					}
				}
			}
		}
	}

	// 2. OCR FALLBACK MATCHING (Requires 3x upscaled image)
	if !e.config.OCRAvailable {
		logger.Debug("[SCAN #%d] — OCR skipped (Tesseract not installed)", e.stats.TotalScans)
		return
	}

	ocrCapturePath, err := capture.CaptureToFile(e.region, 3)
	if err != nil {
		e.stats.TotalErrors++
		logger.Error("[SCAN #%d] OCR Capture failed: %v", e.stats.TotalScans, err)
		return
	}
	defer os.Remove(ocrCapturePath)

	if e.config.DebugSaveCaptures {
		e.saveDebugCapture(ocrCapturePath, "_3x")
	}

	matches, err := ocr.DetectText(ocrCapturePath)
	if err != nil {
		if strings.Contains(err.Error(), "0xc000013a") || err == context.Canceled {
			return
		}
		e.stats.TotalErrors++
		logger.Error("[SCAN #%d] OCR failed: %v", e.stats.TotalScans, err)
		return
	}

	if e.config.DebugMode {
		var words []string
		for _, m := range matches {
			words = append(words, fmt.Sprintf("%s(%d%%)", m.Text, m.Confidence))
		}
		logger.Debug("[SCAN #%d] Words: %s", e.stats.TotalScans, strings.Join(words, ", "))
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
			if dx*dx+dy*dy < dedupRadiusSq { // 300 pixels radius
				duplicate = true
				break
			}
		}
		if duplicate {
			continue
		}
		clickedSpots = append(clickedSpots, coord{x: absX, y: absY})

		logger.Click("\"%s\" @ (%d,%d) | OCR conf=%d%% | %s",
			hit.Keyword, absX, absY, hit.Confidence, bgTag)

		if err := clicker.ClickAt(absX, absY, e.config.UseBackgroundClick); err != nil {
			e.stats.TotalErrors++
			logger.Error("[SCAN #%d] Click failed: %v", e.stats.TotalScans, err)
			continue
		}

		e.recordClick(hit.Keyword)
		time.Sleep(300 * time.Millisecond)

		if len(clickedSpots) >= maxClicksPerScan {
			break
		}
	}
}

func (e *Engine) saveDebugCapture(srcPath string, suffix string) {
	destPath := filepath.Join(e.config.DebugDir, fmt.Sprintf("scan_%d%s.png", e.stats.TotalScans, suffix))
	data, err := os.ReadFile(srcPath)
	if err == nil {
		os.WriteFile(destPath, data, 0644)
	}
}

func SaveCaptureOnce(region image.Rectangle, destPath string) error {
	srcPath, err := capture.CaptureToFile(region, 1)
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
