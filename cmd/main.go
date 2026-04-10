// AutoClickAccepted — Auto-click program that detects text on screen via OCR
// and clicks buttons matching configured keywords.
//
// Usage:
//
//	go run ./cmd/main.go
//	go run ./cmd/main.go -config path/to/config.yaml
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/ethanpham86/AutoClickAccepted/internal/engine"
	"github.com/ethanpham86/AutoClickAccepted/internal/selector"

	"gopkg.in/yaml.v3"
)

// appConfig maps to config.yaml
type appConfig struct {
	Keywords            []string `yaml:"keywords"`
	ScanIntervalMs      int      `yaml:"scan_interval_ms"`
	ConfidenceThreshold int      `yaml:"confidence_threshold"`
}

func main() {
	configPath := flag.String("config", "config.yaml", "Path to config.yaml")
	flag.Parse()

	printBanner()

	// Load config
	cfg, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Printf("  Keywords : %s\n", strings.Join(cfg.Keywords, ", "))
	fmt.Printf("  Interval : %dms\n", cfg.ScanIntervalMs)
	fmt.Printf("  Min Conf : %d%%\n\n", cfg.ConfidenceThreshold)

	// Select region
	region, err := selector.SelectRegion()
	if err != nil {
		log.Fatalf("Region selection failed: %v", err)
	}

	// Create engine
	eng := engine.New(region, engine.Config{
		Keywords:            cfg.Keywords,
		ScanIntervalMs:      cfg.ScanIntervalMs,
		ConfidenceThreshold: cfg.ConfidenceThreshold,
	})

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\n  ⚠ Shutting down...")
		cancel()
	}()

	// Run
	if err := eng.Run(ctx); err != nil && err != context.Canceled {
		log.Fatalf("Engine error: %v", err)
	}

	fmt.Println("  Goodbye!")
}

func loadConfig(path string) (*appConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg appConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	// Defaults
	if len(cfg.Keywords) == 0 {
		cfg.Keywords = []string{"Allow Once", "ACCEPTED", "RUN", "Retry", "Allow", "Accept"}
	}
	if cfg.ScanIntervalMs <= 0 {
		cfg.ScanIntervalMs = 2000
	}
	if cfg.ConfidenceThreshold <= 0 {
		cfg.ConfidenceThreshold = 60
	}

	return &cfg, nil
}

func printBanner() {
	fmt.Println()
	fmt.Println("  ╔══════════════════════════════════════════════════╗")
	fmt.Println("  ║     🖱️  AutoClickAccepted v1.0                  ║")
	fmt.Println("  ║     OCR-based Auto-Clicker for Windows          ║")
	fmt.Println("  ╠══════════════════════════════════════════════════╣")
	fmt.Println("  ║  Detects text on screen and clicks matching     ║")
	fmt.Println("  ║  buttons automatically.                         ║")
	fmt.Println("  ║  Requires: tesseract.exe on PATH                ║")
	fmt.Println("  ╚══════════════════════════════════════════════════╝")
	fmt.Println()
}
