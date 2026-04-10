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
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/ethanpham86/AutoClickAccepted/internal/engine"
	"github.com/ethanpham86/AutoClickAccepted/internal/learner"
	"github.com/ethanpham86/AutoClickAccepted/internal/logger"
	"github.com/ethanpham86/AutoClickAccepted/internal/selector"

	"gopkg.in/yaml.v3"
)

// appConfig maps to config.yaml
type appConfig struct {
	Keywords            []string `yaml:"keywords"`
	ScanIntervalMs      int      `yaml:"scan_interval_ms"`
	ConfidenceThreshold int      `yaml:"confidence_threshold"`
	LogLevel            string   `yaml:"log_level"`
	UseBackgroundClick  bool     `yaml:"use_background_click"`
}

func main() {
	configPath := flag.String("config", "config.yaml", "Path to config.yaml")
	debug := flag.Bool("debug", false, "Enable debug mode: save captures and show all detected words")
	imgDir := flag.String("imgdir", "img", "Path to sample images directory for learning")
	intervalOpt := flag.Int("interval", 0, "Scan interval in milliseconds (overrides config.yaml)")
	flag.Parse()

	// Load config initially to init logger
	cfg, err := loadConfig(*configPath)
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Init File Logger
	if err := logger.Init(cfg.LogLevel, "autoclick.log"); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Close()

	printBanner()

	// Override interval if flag is passed
	if *intervalOpt > 0 {
		cfg.ScanIntervalMs = *intervalOpt
		logger.Info("⏱ Overriding scan interval: %d ms", cfg.ScanIntervalMs)
	}

	// ===== LEARNER: Auto-learn keywords from img/ folder =====
	learnedWords, err := learner.LearnFromImages(*imgDir)
	if err != nil {
		logger.Error("⚠ Learner warning: %v", err)
	}
	if len(learnedWords) > 0 {
		logger.Info("  📚 Learned %d words from %s/:", len(learnedWords), *imgDir)
		for _, w := range learnedWords {
			logger.Info("     - %s", w)
		}
		// Merge learned keywords into config keywords
		cfg.Keywords = learner.MergeKeywords(cfg.Keywords, learnedWords)
	}

	logger.Info("🎯 Listening for keywords: %v", cfg.Keywords)
	if *debug {
		logger.Info("🛠  Debug Mode: ON (Captures will be saved to debug/)")
	}

	// Wait for user to select a region (unless configured statically later)
	logger.Info("\n[1] Màn hình sẽ mờ đi. Hãy KÉO CHUỘT KHOANH VÙNG khu vực hay xuất hiện button.")
	logger.Info("[2] Nhấn Ctrl+C bất cứ lúc nào để thoát.\n")

	region, err := selector.SelectRegion()
	if err != nil {
		logger.Fatal("Selector error: %v", err)
	}
	logger.Info("✅ Region selected: %v", region)

	// Context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		logger.Info("\n🛑 Shutting down gracefully...")
		cancel()
	}()

	// Debug: save one capture immediately to inspect
	if *debug {
		debugDir := filepath.Join(".", "debug")
		os.MkdirAll(debugDir, 0755)
		testCapture := filepath.Join(debugDir, "test_capture.png")
		if err := engine.SaveCaptureOnce(region, testCapture); err != nil {
			logger.Error("⚠ Debug capture failed: %v", err)
		} else {
			logger.Info("  🔍 Debug capture saved to: %s\n", testCapture)
		}
	}

	// Create engine
	debugDir := ""
	if *debug {
		debugDir = filepath.Join(".", "debug")
	}

	// Initialize and run the engine
	eng := engine.New(region, engine.Config{
		Keywords:            cfg.Keywords,
		ScanIntervalMs:      cfg.ScanIntervalMs,
		ConfidenceThreshold: cfg.ConfidenceThreshold,
		DebugSaveCaptures:   *debug,
		DebugDir:            debugDir,
		DebugMode:           *debug,
		UseBackgroundClick:  cfg.UseBackgroundClick,
	})

	if err := eng.Run(ctx); err != nil && err != context.Canceled {
		logger.Error("Engine stopped with error: %v", err)
	}

	stats := eng.Stats()
	logger.Info("\n📊 Final Statistics:")
	logger.Info("=====================")
	logger.Info("Total Scans : %d", stats.TotalScans)
	logger.Info("Total Clicks: %d", stats.TotalClicks)
	logger.Info("Total Errors: %d", stats.TotalErrors)
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
	// Background click defaults to true if omitted, for better UX
	if !cfg.UseBackgroundClick {
		// Just to log it properly, we allow false as explicit setting. But actually Go defaults bools to false.
		// So we can just leave it to whatever parsed.
	}

	return &cfg, nil
}

func printBanner() {
	fmt.Println()
	fmt.Println("  ╔══════════════════════════════════════════════════╗")
	fmt.Println("  ║     🖱️  AutoClickAccepted v2.6                  ║")
	fmt.Println("  ║     Background-Click Stealth Engine             ║")
	fmt.Println("  ╠══════════════════════════════════════════════════╣")
	fmt.Println("  ║  Features:                                      ║")
	fmt.Println("  ║  • Silent Background Click (No mouse hijack)    ║")
	fmt.Println("  ║  • Auto-learn keywords from img/ samples        ║")
	fmt.Println("  ║  • 3x Upscale for OCR accuracy                  ║")
	fmt.Println("  ║  Requires: tesseract.exe on PATH                ║")
	fmt.Println("  ╚══════════════════════════════════════════════════╝")
	fmt.Println()
}
