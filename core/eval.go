package core

import (
	"bytes"
	"errors"
	"fmt"
)

var RESP_NIL []byte = []byte("$-1\r\n")

func evalPing(args []string) []byte {
	var b []byte
	if len(args) > 1 {
		fmt.Printf("Recieved %d args", len(args))
		return Encode(errors.New("ERR wrong number of arguments for 'ping' command"), false)
	}
	if len(args) == 0 {
		fmt.Printf("Recieved %d args", len(args))
		b = Encode("PONG", true)
	} else {
		fmt.Printf("Recieved %d args", len(args))
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
			fmt.Println("in ping case.....")
			buf.Write(evalPing(cmd.Args))
		default:
			fmt.Println("in default case.....")
			buf.Write(evalPing(cmd.Args))
		}
	}
	fmt.Println("sent data is ", buf.String())
	f.Write(buf.Bytes())
}
