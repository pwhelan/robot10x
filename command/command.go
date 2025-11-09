package command

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

type Command struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
}

func (command Command) Exec() error {
	cmd := exec.Command(command.Command, command.Args...)
	return cmd.Run()
}

type Commands []Command

func (commands Commands) Exec() []error {
	errs := make([]error, 0)
	for _, cmd := range commands {
		if err := cmd.Exec(); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func (commands *Commands) unmarshal(data interface{}) error {
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

func (commands *Commands) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var data interface{}
	if err := unmarshal(&data); err != nil {
		return err
	}
	return commands.unmarshal(data)
}

func (commands *Commands) UnmarshalJSON(buf []byte) error {
	var data interface{}
	if err := json.Unmarshal(buf, &data); err != nil {
		return err
	}
	return commands.unmarshal(data)
}
