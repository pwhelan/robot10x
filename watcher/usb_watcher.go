package watcher

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/pwhelan/gousb"
	"github.com/pwhelan/robot10x/command"
)

// USBWatcher implements the Watcher interface for USB hotplug events.
type USBWatcher struct {
	usb *gousb.Context
}

type Id gousb.ID

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

// Init initializes the USB watcher.
func (w *USBWatcher) Init(ctx context.Context, cfg any) error {
	usbCfgs, ok := cfg.([]USBHotplugConfig)
	if !ok {
		return fmt.Errorf("invalid config type for USBWatcher")
	}

	w.usb = gousb.NewContext()
	w.usb.RegisterHotplug(func(ev gousb.HotplugEvent) {
		desc, err := ev.DeviceDesc()
		if err != nil {
			log.Printf("could not get device description: %v", err)
			return
		}

		for _, cfg := range usbCfgs {
			if desc.Vendor == gousb.ID(cfg.Vendor) && desc.Product == gousb.ID(cfg.Product) {
				if ev.Type() == gousb.HotplugEventDeviceArrived {
					if errs := cfg.CmdUp.Exec(); len(errs) > 0 {
						for _, err := range errs {
							log.Printf("ERROR: %s", err)
						}
					}
				} else if ev.Type() == gousb.HotplugEventDeviceLeft {
					if errs := cfg.CmdDown.Exec(); len(errs) > 0 {
						for _, err := range errs {
							log.Printf("ERROR: %s", err)
						}
					}
				}
			}
		}
	})

	go func() {
		<-ctx.Done()
		w.usb.Close()
	}()

	return nil
}
