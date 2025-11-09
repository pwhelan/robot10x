package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/goccy/go-yaml"
	"github.com/pwhelan/robot10x/watcher"
)

type config struct {
	USB  []watcher.USBHotplugConfig `json:"usb" yaml:"usb"`
	Exec []watcher.ExecConfig       `json:"commands" yaml:"commands"`
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: ", os.Args[0], " <config.json|config.yaml>")
	}

	cfgFile := os.Args[1]
	cfgdata, err := os.ReadFile(cfgFile)
	if err != nil {
		log.Fatal("Failed to read config file: ", err)
	}

	var cfg config
	ext := filepath.Ext(cfgFile)
	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(cfgdata, &cfg); err != nil {
			log.Fatal("Failed to parse YAML config: ", err)
		}
	default:
		if err := json.Unmarshal(cfgdata, &cfg); err != nil {
			log.Fatal("Failed to parse JSON config: ", err)
		}
	}
	fmt.Printf("cfg=%+v\n", cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	watchers := []watcher.Watcher{
		&watcher.USBWatcher{},
		&watcher.ProcessWatcher{},
	}

	if len(cfg.USB) > 0 {
		if err := watchers[0].Init(ctx, cfg.USB); err != nil {
			log.Fatalf("Failed to initialize USB watcher: %v", err)
		}
	}

	if len(cfg.Exec) > 0 {
		if err := watchers[1].Init(ctx, cfg.Exec); err != nil {
			log.Fatalf("Failed to initialize Process watcher: %v", err)
		}
	}

	log.Println("Watchers initialized. Waiting for events...")

	// Wait for termination signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down...")
}
