package watcher

import (
	"context"
)

type Watcher interface {
	// Init initializes the watcher by passing it the main context
	// as well as the relevant subsection of configuration for the watcher.
	Init(context.Context, any) error
}
