package core

import (
	"bytes"
	"errors"
	"log"
	"strconv"
)

var RESP_NIL []byte = []byte("$-1\r\n")
var RESP_OK []byte = []byte("+OK\r\n")

func evalPing(args []string) []byte {
	var b []byte
	if len(args) > 1 {
		return Encode(errors.New("ERR wrong number of arguments for 'ping' command"), false)
	}
	if len(args) == 0 {
		b = Encode("PONG", true)
	} else {
		b = Encode(args[0], false)
	}
	return b
}

// SET key value EX 60
// SET key value
// SET mykey myvalue NX. Setting a Value Only If the Key Does Not Exist
// If mykey already exists, the command will do nothing and return (nil).
// SET mykey mynewvalue XX. Setting a Value Only If the Key Exists
// If mykey doesn't exist, Redis will return an error.
func evalSet(args []string) []byte {
	if len(args) > 4 {
		return Encode(errors.New("ERR wrong number of arguments for 'SET' command"), false)
	}
	var key, value string
	var exDurationMs int64 = -1
	key, value = args[0], args[1]
	oType, oEnc := deduceTypeEncoding(value)
	for i := 2; i < len(args); i++ {
		switch args[i] {
		case "EX", "ex":
			i++
			if i == len(args) {
				return Encode(errors.New("ERR syntax error"), false)
			}
			exDurationSec, err := strconv.ParseInt(args[i], 10, 64)
			if err != nil {
				return Encode(errors.New("ERR expiration is not a number or out of range"), false)
			}
			exDurationMs = exDurationSec * 1000
		default:
			return Encode(errors.New("ERR syntax error"), false)
		}
	}
	Put(key, NewObj(value, exDurationMs, oType, oEnc))
	return RESP_OK
}

func evalGet(args []string) []byte {
	if len(args) != 1 {
		return Encode(errors.New("ERR wrong number of arguments for 'GET' command"), false)
	}
	obj := Get(args[0])

	if obj == nil {
		return RESP_NIL
	}

	if hasExpired(obj) {
		return RESP_NIL
	}

	return Encode(obj.Value, true)
}

func EvalAndRespond(cmds RedisCmds, f FDComm) {
	var response []byte
	buf := bytes.NewBuffer(response)
	for _, cmd := range cmds {
		switch cmd.Cmd {
		case "PING":
			buf.Write(evalPing(cmd.Args))
		case "SET":
			buf.Write(evalSet(cmd.Args))
		case "GET":
			buf.Write(evalGet(cmd.Args))
		default:
			buf.Write(evalPing(cmd.Args))
		}
	}
	log.Printf("sent data is: %s", buf.String())
	f.Write(buf.Bytes())
}
