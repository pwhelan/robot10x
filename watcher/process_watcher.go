package watcher

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/elastic/gosigar/psnotify"
)

// ProcessWatcher implements the Watcher interface for process events.
type ProcessWatcher struct{}

type ExecConfig struct {
	Binary  string   `json:"bin"`
	CmdUp   Commands `json:"up"`
	CmdDown Commands `json:"down"`
}

func isNumeric(s string) bool {
	_, err := strconv.ParseInt(s, 10, 64)
	return err == nil
}

// Init initializes the process watcher.
func (w *ProcessWatcher) Init(ctx context.Context, cfg any) error {
	execCfgs, ok := cfg.([]ExecConfig)
	if !ok {
		return fmt.Errorf("invalid config type for ProcessWatcher")
	}

	execs := make(map[string]ExecConfig)
	execd := make(map[int]ExecConfig)

	pswatcher, err := psnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create process watcher: %w", err)
	}

	for _, exec := range execCfgs {
		execs[exec.Binary] = exec
	}

	go func() {
		for {
			select {
			case ev := <-pswatcher.Exec:
				bin, _ := os.Readlink(fmt.Sprintf("/proc/%d/exe", ev.Pid))
				if ex, ok := execs[bin]; ok {
					log.Printf("exec event: %d->%s", ev.Pid, bin)
					if errs := ex.CmdUp.Exec(); len(errs) > 0 {
						for _, err := range errs {
							fmt.Printf("ERROR: %s", err)
						}
					}
					execd[ev.Pid] = ex
				}
			case ev := <-pswatcher.Exit:
				if ex, ok := execd[ev.Pid]; ok {
					log.Printf("exit event: %d->%s (%+v)", ev.Pid, ex.Binary, ev)
					if errs := ex.CmdDown.Exec(); len(errs) > 0 {
						for _, err := range errs {
							fmt.Printf("ERROR: %s", err)
						}
					}
					delete(execd, ev.Pid)
				}
			case <-ctx.Done():
				pswatcher.Close()
				return
			}
		}
	}()

	files, err := os.ReadDir("/proc")
	if err != nil {
		return fmt.Errorf("failed to read /proc: %w", err)
	}

	for _, f := range files {
		if f.IsDir() && isNumeric(f.Name()) {
			pid, _ := strconv.ParseInt(f.Name(), 10, 64)
			err = pswatcher.Watch(int(pid), psnotify.PROC_EVENT_EXIT|psnotify.PROC_EVENT_EXEC)
			if err != nil {
				// Log and continue, as some processes might not be watchable
				log.Printf("failed to watch pid %d: %v", pid, err)
			}
		}
	}

	return nil
}
