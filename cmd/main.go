// AutoClickAccepted — Auto-click program that detects text on screen via OCR
// and clicks buttons matching configured keywords.
//
// Usage:
//
//	go run ./cmd/main.go
//	go run ./cmd/main.go -config path/to/config.yaml
//	go run ./cmd/main.go -debug    (saves captures to debug/ folder)
package main

import (
	"context"
	"flag"
	"fmt"
	"image"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/ethanpham86/AutoClickAccepted/internal/engine"
	"github.com/ethanpham86/AutoClickAccepted/internal/hotkey"
	"github.com/ethanpham86/AutoClickAccepted/internal/learner"
	"github.com/ethanpham86/AutoClickAccepted/internal/logger"
	"github.com/ethanpham86/AutoClickAccepted/internal/ocr"
	"github.com/ethanpham86/AutoClickAccepted/internal/selector"

	"golang.org/x/sys/windows"
	"gopkg.in/yaml.v3"
)

// appConfig maps to config.yaml
type appConfig struct {
	Keywords            []string `yaml:"keywords"`
	ScanIntervalMs      int      `yaml:"scan_interval_ms"`
	ConfidenceThreshold int      `yaml:"confidence_threshold"`
	LogLevel            string   `yaml:"log_level"`
	UseBackgroundClick  bool     `yaml:"use_background_click"`
	ScanRegion          string   `yaml:"scan_region"` // Optional: "x,y,width,height" to skip interactive selector
	MaxClicksPerScan    int      `yaml:"max_clicks_per_scan"`
	DedupRadiusPx       int      `yaml:"dedup_radius_px"`
	TemplateThreshold   float64  `yaml:"template_threshold"`
	ClickDelayMs        int      `yaml:"click_delay_ms"`
}

func main() {
	// Ensure we have a visible console window even if built with -H windowsgui
	ensureConsole()

	// Set console to UTF-8 so log text renders correctly
	setConsoleUTF8()

	// Set DPI awareness so screen coordinates are accurate on high-DPI displays
	setDPIAware()

	configPath := flag.String("config", "config.yaml", "Path to config.yaml")
	debug := flag.Bool("debug", false, "Enable debug mode: save captures and show all detected words")
	imgDir := flag.String("imgdir", "img", "Path to sample images directory for learning")
	intervalOpt := flag.Int("interval", 0, "Scan interval in milliseconds (overrides config.yaml)")
	flag.Parse()

	// Resolve paths relative to the exe location (not CWD) for portability
	exeDir := getExeDir()
	resolvedConfig := resolvePath(exeDir, *configPath)
	resolvedImgDir := resolvePath(exeDir, *imgDir)

	// Load config initially to init logger
	cfg, err := loadConfig(resolvedConfig)
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		fmt.Println("Press Enter to exit...")
		fmt.Scanln()
		os.Exit(1)
	}

	// Init File Logger (resolve log file relative to exe dir too)
	logPath := filepath.Join(exeDir, "autoclick.log")
	if err := logger.Init(cfg.LogLevel, logPath); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		fmt.Println("Press Enter to exit...")
		fmt.Scanln()
		os.Exit(1)
	}
	defer logger.Close()

	printBanner()

	// Override interval if flag is passed
	if *intervalOpt > 0 {
		cfg.ScanIntervalMs = *intervalOpt
		logger.Info("[CONFIG] Overriding scan interval: %d ms", cfg.ScanIntervalMs)
	}

	// ===== CHECK TESSERACT AVAILABILITY =====
	ocrAvailable := ocr.IsAvailable()
	if ocrAvailable {
		logger.Info("[OK] Tesseract OCR: Installed -- OCR fallback enabled")
	} else {
		logger.Info("[WARN] Tesseract OCR: NOT FOUND -- Running template-matching only")
		logger.Info("[TIP] Install Tesseract from dependencies/ folder to enable OCR fallback")
	}

	// ===== LEARNER: Auto-learn keywords from img/ folder =====
	learnedWords, templates, err := learner.LearnFromImages(resolvedImgDir)
	if err != nil {
		logger.Error("[LEARN] Warning: %v", err)
	}
	if len(templates) > 0 {
		logger.Info("[LEARN] Loaded %d image templates from %s/", len(templates), *imgDir)
	}
	if len(learnedWords) > 0 {
		logger.Info("[LEARN] Learned %d words from %s/:", len(learnedWords), *imgDir)
		for _, w := range learnedWords {
			logger.Info("     - %s", w)
		}
		// Merge learned keywords into config keywords
		cfg.Keywords = learner.MergeKeywords(cfg.Keywords, learnedWords)
	}

	logger.Info("[KEYWORDS] Listening for: %v", cfg.Keywords)
	if *debug {
		logger.Info("[DEBUG] Debug Mode: ON (Captures saved to debug/)")
	}

	// ===== REGION SELECTION =====
	var region image.Rectangle

	if cfg.ScanRegion != "" {
		// Parse fixed region from config: "x,y,width,height"
		region, err = parseScanRegion(cfg.ScanRegion)
		if err != nil {
			logger.Fatal("Invalid scan_region in config: %v", err)
		}
		logger.Info("[REGION] Using fixed scan region from config: %v", region)
	} else {
		// Interactive selector
		logger.Info("\n[1] Screen will dim. DRAG MOUSE to select the scan region.")
		logger.Info("[2] Press Ctrl+C to exit at any time.\n")

		region, err = selector.SelectRegion()
		if err != nil {
			logger.Fatal("Selector error: %v", err)
		}
		logger.Info("[REGION] Selected: %v", region)
		logger.Info("[TIP] Add to config.yaml to skip selection next time:")
		logger.Info("   scan_region: \"%d,%d,%d,%d\"", region.Min.X, region.Min.Y, region.Dx(), region.Dy())
	}

	// Context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		logger.Info("\n[STOP] Shutting down gracefully...")
		cancel()
	}()

	// Debug: save one capture immediately to inspect
	if *debug {
		debugDir := filepath.Join(exeDir, "debug")
		os.MkdirAll(debugDir, 0755)
		testCapture := filepath.Join(debugDir, "test_capture.png")
		if err := engine.SaveCaptureOnce(region, testCapture); err != nil {
			logger.Error("[DEBUG] Capture failed: %v", err)
		} else {
			logger.Info("[DEBUG] Capture saved to: %s\n", testCapture)
		}
	}

	// Create engine
	debugDir := ""
	if *debug {
		debugDir = filepath.Join(exeDir, "debug")
	}

	// Initialize and run the engine
	eng := engine.New(region, engine.Config{
		Keywords:            cfg.Keywords,
		Templates:           templates,
		ScanIntervalMs:      cfg.ScanIntervalMs,
		ConfidenceThreshold: cfg.ConfidenceThreshold,
		DebugSaveCaptures:   *debug,
		DebugDir:            debugDir,
		DebugMode:           *debug,
		UseBackgroundClick:  cfg.UseBackgroundClick,
		OCRAvailable:        ocrAvailable,
		MaxClicksPerScan:    cfg.MaxClicksPerScan,
		DedupRadiusPx:       cfg.DedupRadiusPx,
		TemplateThreshold:   cfg.TemplateThreshold,
		ClickDelayMs:        cfg.ClickDelayMs,
	})

	// ===== GLOBAL HOTKEY LISTENER =====
	hotkeyChan := make(chan hotkey.Action, 4)
	go func() {
		if err := hotkey.Listen(hotkeyChan); err != nil {
			logger.Error("[HOTKEY] Registration failed: %v", err)
			logger.Info("[HOTKEY] F6/F7 unavailable (another instance may be running)")
		}
	}()

	go func() {
		for action := range hotkeyChan {
			switch action {
			case hotkey.ActionTogglePause:
				eng.TogglePause()
			case hotkey.ActionStop:
				logger.Info("[STOP] F7 pressed -- Shutting down...")
				cancel()
			}
		}
	}()

	logger.Info("[HOTKEY] F6 = Pause/Resume | F7 = Stop")

	if err := eng.Run(ctx); err != nil && err != context.Canceled {
		logger.Error("Engine stopped with error: %v", err)
	}

	stats := eng.Stats()
	logger.Info("\n=== Final Statistics ===")
	logger.Info("=====================")
	logger.Info("Total Scans : %d", stats.TotalScans)
	logger.Info("Total Clicks: %d", stats.TotalClicks)
	logger.Info("Total Errors: %d", stats.TotalErrors)

	// Per-keyword click breakdown
	if len(stats.ClicksByLabel) > 0 {
		logger.Info("--- Clicks by Keyword ---")
		for label, count := range stats.ClicksByLabel {
			logger.Info("  %-20s : %d clicks", label, count)
		}
	}
	logger.Info("=====================")
}

// loadConfig reads keywords and settings from a YAML file
func loadConfig(path string) (*appConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg appConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// Defaults
	if len(cfg.Keywords) == 0 {
		cfg.Keywords = []string{"Allow Once", "ACCEPTED", "RUN", "Retry", "Allow", "Accept"}
	}
	if cfg.ScanIntervalMs <= 0 {
		cfg.ScanIntervalMs = 2000
	}
	if cfg.ConfidenceThreshold <= 0 {
		cfg.ConfidenceThreshold = 30
	}
	if cfg.MaxClicksPerScan <= 0 {
		cfg.MaxClicksPerScan = 10
	}
	if cfg.DedupRadiusPx <= 0 {
		cfg.DedupRadiusPx = 80
	}
	if cfg.TemplateThreshold <= 0.0 {
		cfg.TemplateThreshold = 0.92
	}
	if cfg.ClickDelayMs <= 0 {
		cfg.ClickDelayMs = 300
	}

	return &cfg, nil
}

// parseScanRegion parses "x,y,width,height" into an image.Rectangle
func parseScanRegion(s string) (image.Rectangle, error) {
	parts := strings.Split(strings.TrimSpace(s), ",")
	if len(parts) != 4 {
		return image.Rectangle{}, fmt.Errorf("expected 4 values (x,y,w,h), got %d", len(parts))
	}

	x, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
	y, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
	w, err3 := strconv.Atoi(strings.TrimSpace(parts[2]))
	h, err4 := strconv.Atoi(strings.TrimSpace(parts[3]))

	if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
		return image.Rectangle{}, fmt.Errorf("invalid numbers in scan_region: %q", s)
	}

	if w < 10 || h < 10 {
		return image.Rectangle{}, fmt.Errorf("scan_region too small: %dx%d (minimum 10x10)", w, h)
	}

	return image.Rect(x, y, x+w, y+h), nil
}

// ensureConsole allocates a console window if the app was built with -H windowsgui.
// This allows the exe to show output even when built as a GUI subsystem app.
func ensureConsole() {
	kernel32 := windows.NewLazySystemDLL("kernel32.dll")
	procGetConsoleWindow := kernel32.NewProc("GetConsoleWindow")
	procAllocConsole := kernel32.NewProc("AllocConsole")

	hwnd, _, _ := procGetConsoleWindow.Call()
	if hwnd == 0 {
		// No console attached — allocate one
		procAllocConsole.Call()
		// Reopen stdout/stderr to the new console
		conOut, err := os.OpenFile("CONOUT$", os.O_WRONLY, 0)
		if err == nil {
			os.Stdout = conOut
			os.Stderr = conOut
		}
	}
}

// setDPIAware tells Windows not to scale coordinates, so screen capture
// and click positions are pixel-accurate on high-DPI displays.
func setDPIAware() {
	user32 := windows.NewLazySystemDLL("user32.dll")
	proc := user32.NewProc("SetProcessDPIAware")
	proc.Call()
}

// getExeDir returns the directory where the running executable is located.
func getExeDir() string {
	exe, err := os.Executable()
	if err != nil {
		return "."
	}
	return filepath.Dir(exe)
}

// resolvePath resolves a path relative to baseDir if it's not already absolute.
func resolvePath(baseDir, path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(baseDir, path)
}

// setConsoleUTF8 sets the Windows console output codepage to UTF-8
// so that log text renders correctly instead of garbled characters.
func setConsoleUTF8() {
	kernel32 := windows.NewLazySystemDLL("kernel32.dll")
	proc := kernel32.NewProc("SetConsoleOutputCP")
	proc.Call(65001) // CP_UTF8
}

func printBanner() {
	banner := `
  +----------------------------------------------------+
  |     AutoClickAccepted v3.2.0 Hybrid                |
  |     Background-Click Stealth Engine                |
  +----------------------------------------------------+
  |  Features:                                         |
  |  * Silent Background Click (PostMessage)           |
  |  * Exact Pixel Template Match + OCR Fallback       |
  |  * Auto-learn templates from img/ folder           |
  |  * Pause/Resume (F6) & Stop (F7) hotkeys           |
  +----------------------------------------------------+
`
	fmt.Print(banner)
	logger.Info("AutoClickAccepted v3.2.0 Hybrid -- Engine started")
}
