package core

import (
	"bytes"
	"errors"
)

var RESP_NIL []byte = []byte("$-1\r\n")

func evalPing(args []string) []byte {
	var b []byte
	if len(args) > 1 {
		b = Encode(errors.New("ERR wrong number of arguments for 'ping' command"), false)
	}
	if len(args) == 0 {
		b = Encode("PONG", true)
	} else {
		b = Encode(args[0], false)
	}
	return b
}

func EvalAndRespond(cmds RedisCmds, f FDComm) {
	var response []byte
	buf := bytes.NewBuffer(response)
	for _, cmd := range cmds {
		switch cmd.Cmd {
		case "PING":
			buf.Write(evalPing(cmd.Args))
		default:
			buf.Write(evalPing(cmd.Args))
		}
	}
	f.Write(buf.Bytes())
}
