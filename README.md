# robot10x

`robot10x` is a lightweight, event-driven automation tool for Linux that listens for system events and executes predefined commands. It's designed to be a flexible "if-this-then-that" for your operating system, allowing you to automate tasks based on USB device connections, process execution, and more.

## What It Does

`robot10x` runs as a daemon, monitoring for events you've configured. When an event occurs, `robot10x` executes the corresponding command(s) you've specified.

Current supported watchers:
*   **USB Watcher:** Triggers commands when a specific USB device is plugged in or unplugged.
*   **Process Watcher:** Triggers commands when a specific application starts or stops.

## Configuration

`robot10x` can be configured using either a JSON or a YAML file. You must provide the path to the configuration file as a command-line argument.

The configuration file has two main sections: `usb` for USB device events and `commands` for process execution events.

### USB Watcher Configuration

The `usb` section is an array of objects, where each object defines a set of commands to run when a specific USB device is connected (`up`) or disconnected (`down`).

You can identify the `vendor` and `product` IDs for your device using the `lsusb` command (see section below).

**JSON Example (`config.json`):**
```json
{
  "usb": [
    {
      "vendor": "0x18d1",
      "product": "0x4ee9",
      "up": "touch /tmp/my-phone-connected.txt",
      "down": ["rm", "/tmp/my-phone-connected.txt"]
    },
    {
      "vendor": "0x14cd",
      "product": "0x1212",
      "up": [
        ["echo", "Device connected at $(date)"]
      ],
      "down": [
        ["echo", "Device disconnected at $(date)"]
      ]
    }
  ]
}
```

**YAML Example (`config.yaml`):**
```yaml
usb:
  - vendor: 0x18d1
    product: 0x4ee9
    up: touch /tmp/my-phone-connected.txt
    down:
      - rm
      - /tmp/my-phone-connected.txt
  - vendor: 0x14cd
    product: 0x1212
    up:
      - - echo
        - "Device connected at $(date)"
    down:
      - - echo
        - "Device disconnected at $(date)"
```

### Process Watcher Configuration

The `commands` section is an array of objects that define commands to run when a process starts (`up`) or stops (`down`). The process is identified by its binary path (`bin`).

**JSON Example (`config.json`):**
```json
{
  "commands": [
    {
      "bin": "/usr/bin/gedit",
      "up": "echo 'gedit started' > /tmp/gedit.log",
      "down": "echo 'gedit stopped' >> /tmp/gedit.log"
    }
  ]
}
```

**YAML Example (`config.yaml`):**
```yaml
commands:
  - bin: /usr/bin/gedit
    up: echo 'gedit started' > /tmp/gedit.log
    down: echo 'gedit stopped' >> /tmp/gedit.log
```

### Command Formats

Commands (`up` and `down`) can be specified in several ways:
- A single string: `"touch /tmp/file.txt"`
- An array of strings for command and arguments: `["touch", "/tmp/file.txt"]`
- An array of arrays for multiple commands: `[["touch", "/tmp/file1.txt"], ["touch", "/tmp/file2.txt"]]`

## Running robot10x

To run `robot10x`, execute the binary with the path to your configuration file:
```bash
./robot10x config.yaml
```

### Running as a systemd Service

For long-running automation, it's best to run `robot10x` as a systemd service. You can run it as a user service, which is recommended.

1.  Create a service file at `~/.config/systemd/user/robot10x.service`:

    ```ini
    [Unit]
    Description=robot10x event-driven automation

    [Service]
    ExecStart=/path/to/your/robot10x /path/to/your/config.json
    Restart=always

    [Install]
    WantedBy=default.target
    ```
    *Replace `/path/to/your/robot10x` and `/path/to/your/config.json` with the absolute paths to your binary and configuration file.*

2.  Enable and start the service:
    ```bash
    systemctl --user enable robot10x.service
    systemctl --user start robot10x.service
    ```

3.  Check the status and logs:
    ```bash
    systemctl --user status robot10x.service
    journalctl --user -u robot10x.service -f
    ```

## Identifying USB Devices with `lsusb`

To configure the USB watcher, you need the Vendor and Product IDs of your USB device. You can find these using the `lsusb` command.

1.  Run `lsusb` in your terminal. You will see a list of connected USB devices:
    ```
    $ lsusb
    Bus 001 Device 001: ID 1d6b:0002 Linux Foundation 2.0 root hub
    Bus 001 Device 002: ID 8087:0025 Intel Corp.
    Bus 001 Device 003: ID 04f2:b221 Chicony Electronics Co., Ltd
    Bus 002 Device 005: ID 18d1:4ee9 Google Inc.
    ```

2.  Find your device in the list. The ID is shown in the format `vendor:product`. For example, for the "Google Inc." device, the ID is `18d1:4ee9`.
    *   **Vendor ID:** `18d1`
    *   **Product ID:** `4ee9`

3.  Use these values (prefixed with `0x`) in your `config.json` or `config.yaml` file.

### Example: Calling a URL on USB connect/disconnect

You can use `curl` to call a URL when a USB device is plugged in or out. This is useful for triggering webhooks or other remote actions.

**`config.json`:**
```json
{
  "usb": [
    {
      "vendor": "0x18d1",
      "product": "0x4ee9",
      "up": [
        "curl",
        "-X",
        "POST",
        "https://your-webhook-url.com/usb-connected"
      ],
      "down": [
        "curl",
        "-X",
        "POST",
        "https://your-webhook-url.com/usb-disconnected"
      ]
    }
  ]
}
```
This configuration will send a POST request to the specified URLs when the Google device is connected or disconnected.
