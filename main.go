package main

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/pwhelan/robot10x/watcher"

	"github.com/charmbracelet/log"
	"github.com/goccy/go-yaml"
)

type config struct {
	USB  []watcher.USBHotplugConfig `json:"usb" yaml:"usb"`
	Exec []watcher.ExecConfig       `json:"commands" yaml:"commands"`
}

func main() {
	logger := log.New(os.Stderr)

	if len(os.Args) < 2 {
		logger.Fatal("Usage: ", os.Args[0], " <config.json|config.yaml>")
	}

	cfgFile := os.Args[1]
	cfgdata, err := os.ReadFile(cfgFile)
	if err != nil {
		logger.Fatal("Failed to read config file: ", err)
	}

	var cfg config
	ext := filepath.Ext(cfgFile)
	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(cfgdata, &cfg); err != nil {
			logger.Fatal("Failed to parse YAML config: ", err)
		}
	default:
		if err := json.Unmarshal(cfgdata, &cfg); err != nil {
			logger.Fatal("Failed to parse JSON config: ", err)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	watchers := []watcher.Watcher{
		&watcher.USBWatcher{},
		&watcher.ProcessWatcher{},
	}

	wctx := context.WithValue(ctx, "logger", logger)

	if len(cfg.USB) > 0 {
		if err := watchers[0].Init(wctx, cfg.USB); err != nil {
			logger.Fatalf("Failed to initialize USB watcher: %v", err)
		}
	}

	if len(cfg.Exec) > 0 {
		if err := watchers[1].Init(wctx, cfg.Exec); err != nil {
			logger.Fatalf("Failed to initialize Process watcher: %v", err)
		}
	}

	logger.Info("Watchers initialized. Waiting for events...")

	// Wait for termination signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Info("Shutting down...")
}
