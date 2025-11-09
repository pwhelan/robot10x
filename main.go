package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
)

type Command struct {
	Command string
	Args    []string
}

func (command Command) Exec() error {
	cmd := exec.Command(command.Command, command.Args...)
	return cmd.Run()
}

type Commands []Command

type config struct {
	USB  []usbhotplugconfig `json:"usb"`
	Exec []execconfig       `json:"commands"`
}

func (commands Commands) Exec() []error {
	errs := make([]error, 0)
	for _, cmd := range commands {
		if err := cmd.Exec(); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func (commands *Commands) UnmarshalJSON(buf []byte) error {
	var data interface{}

	if err := json.Unmarshal(buf, &data); err != nil {
		return err
	}

	switch cmds := data.(type) {
	case string:
		cmdparts := strings.Split(cmds, " ")
		if len(cmdparts) >= 2 {
			*commands = Commands{{
				Command: cmdparts[0],
				Args:    cmdparts[1:],
			}}
		} else if len(cmdparts) == 1 {
			*commands = Commands{{
				Command: cmdparts[0],
				Args:    make([]string, 0),
			}}
		} else {
			return fmt.Errorf("unable to support blank command")
		}
		return nil
	case []interface{}:
		if len(cmds) <= 0 {
			return fmt.Errorf("unable to support blank command")
		}
		if _, ok := cmds[0].(string); ok {
			args := make([]string, 0)
			for _, cmdparts := range cmds[1:] {
				args = append(args, cmdparts.(string))
			}
			*commands = Commands{{
				Command: cmds[0].(string),
				Args:    args,
			}}
			return nil
		}
		*commands = make(Commands, 0)
		for _, cmd := range cmds {
			cmdparts, ok := cmd.([]interface{})
			if !ok {
				return fmt.Errorf("unable to unmarshal command")
			}
			args := make([]string, 0)
			for _, cmdparts := range cmdparts[1:] {
				args = append(args, cmdparts.(string))
			}
			*commands = append(*commands, Command{
				Command: cmdparts[0].(string),
				Args:    args,
			})
		}
		return nil
	default:
		return fmt.Errorf("unable to unmarshal command format: %+v", data)
	}
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

	watchers := []Watcher{
		&USBWatcher{},
		&ProcessWatcher{},
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
