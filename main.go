package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/pwhelan/robot10x/watcher"
)

type config struct {
	USB  []watcher.USBHotplugConfig `json:"usb"`
	Exec []watcher.ExecConfig       `json:"commands"`
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: ", os.Args[0], " <config.json>")
	}

	cfgdata, err := os.ReadFile(os.Args[1])
	if err != nil {
		log.Fatal("Failed to read config file: ", err)
	}

	var cfg config
	if err := json.Unmarshal(cfgdata, &cfg); err != nil {
		log.Fatal("Failed to parse config: ", err)
	}

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
