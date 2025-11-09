package watcher

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/pwhelan/robot10x/command"
	usbmon "github.com/rubiojr/go-usbmon"
)

// USBWatcher implements the Watcher interface for USB hotplug events.
type USBWatcher struct {
}

type Id uint64

type USBHotplugConfig struct {
	Product Id               `json:"product"`
	Vendor  Id               `json:"vendor"`
	CmdUp   command.Commands `json:"up"`
	CmdDown command.Commands `json:"down"`
}

func (id *Id) UnmarshalJSON(data []byte) error {
	var num string
	if err := json.Unmarshal(data, &num); err != nil {
		return err
	}
	val, err := strconv.ParseUint(num[2:], 16, 32)
	if err != nil {
		return err
	}
	*id = Id(val)
	return nil
}

func (id Id) Equals(sid string) bool {
	if sid[0:2] == "0x" {
		sid = sid[2:]
	}

	val, err := strconv.ParseUint(sid, 16, 32)
	if err != nil {
		return false
	}

	return id == Id(val)
}

// Init initializes the USB watcher.
func (w *USBWatcher) Init(ctx context.Context, cfg any) error {
	usbCfgs, ok := cfg.([]USBHotplugConfig)
	if !ok {
		return fmt.Errorf("invalid config type for USBWatcher")
	}

	go func() {
		cusbmon, err := usbmon.Listen(ctx)
		if err != nil {
			log.Fatal(err)
		}

		for {
			select {
			case <-ctx.Done():
				return
			case device := <-cusbmon:
				for _, cfg := range usbCfgs {
					if cfg.Vendor.Equals(device.VendorID()) && cfg.Product.Equals(device.ProductID()) {
						switch device.Action() {
						case "add":
							if errs := cfg.CmdUp.Exec(); len(errs) > 0 {
								for _, err := range errs {
									log.Printf("ERROR: %s", err)
								}
							}
						case "remove":
							if errs := cfg.CmdDown.Exec(); len(errs) > 0 {
								for _, err := range errs {
									log.Printf("ERROR: %s", err)
								}
							}
						}
					}
				}
			}
		}
	}()

	return nil
}
