package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/pwhelan/gousb"
)

// USBWatcher implements the Watcher interface for USB hotplug events.
type USBWatcher struct {
	usb *gousb.Context
}

type ID gousb.ID

type usbhotplugconfig struct {
	Product ID       `json:"product"`
	Vendor  ID       `json:"vendor"`
	CmdUp   Commands `json:"up"`
	CmdDown Commands `json:"down"`
}

func (id *ID) UnmarshalJSON(data []byte) error {
	var num string
	if err := json.Unmarshal(data, &num); err != nil {
		return err
	}
	val, err := strconv.ParseUint(num[2:], 16, 32)
	if err != nil {
		return err
	}
	*id = ID(val)
	return nil
}

// Init initializes the USB watcher.
func (w *USBWatcher) Init(ctx context.Context, cfg any) error {
	usbCfgs, ok := cfg.([]usbhotplugconfig)
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

		fmt.Printf("%+v\n", desc)
		for _, cfg := range usbCfgs {
			if desc.Vendor == gousb.ID(cfg.Vendor) && desc.Product == gousb.ID(cfg.Product) {
				if ev.Type() == gousb.HotplugEventDeviceArrived {
					fmt.Printf("UP=%+v\n", cfg.CmdUp)
					if errs := cfg.CmdUp.Exec(); len(errs) > 0 {
						for _, err := range errs {
							fmt.Printf("ERROR: %s", err)
						}
					}
				} else if ev.Type() == gousb.HotplugEventDeviceLeft {
					fmt.Printf("DOWN=%+v\n", cfg.CmdDown)
					if errs := cfg.CmdDown.Exec(); len(errs) > 0 {
						for _, err := range errs {
							fmt.Printf("ERROR: %s", err)
						}
					}
				}
			} else {
				fmt.Printf("%v != %v\n", desc.Vendor, cfg.Vendor)
				fmt.Printf("%v != %v\n", desc.Product, cfg.Product)
			}
		}
	})

	go func() {
		<-ctx.Done()
		w.usb.Close()
	}()

	return nil
}
