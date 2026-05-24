package core

import (
	"errors"
	"fmt"
	"syscall"
)

func cmdPing(args []string) []byte {
	var res []byte
	if len(args) > 1 {
		return Encode(errors.New("ERR wrong number of argument for 'ping' command"), false)
	}

	if len(args) == 0 {
		res = Encode("PONG", true)
	} else {
		res = Encode(args[0], false)
	}
	return res
}

func ExecuteAndResponse(cmd *Command, connFd int) error {
	var res []byte
	switch cmd.Cmd {
	case "PING":
		res = cmdPing(cmd.Args)
	default:
		res = []byte(fmt.Sprintf("-CMD not found\r\n"))
	}

	_, err := syscall.Write(connFd, res)
	return err
}
