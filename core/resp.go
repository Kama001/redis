package core

import (
	"errors"
	"fmt"
)

// 0 1  2  3 4
// 2 3 \r \n $

// 0 1 2  3  4 5
// * 2 3 \r \n $
func readLength(data []byte) (int, int) {
	index := 0
	argsCnt := 0
	for ; data[index] != '\r'; index++ {
		argsCnt = argsCnt*10 + int(data[index]-'0')
	}
	return argsCnt, index + 2
}

// to parse *2\r\n$3\r\nGET\r\n$3\r\foo\r\n
func readArray(data []byte) (interface{}, int, error) {
	// since we already know data[0] is '*'
	pos := 1

	// delta = index after *2\r\n is parsed
	// count = number of arguments to be expected
	count, delta := readLength(data[pos:])
	pos += delta
	var args []interface{} = make([]interface{}, count)
	for i := range args {
		arg, delta, err := ParseSymbol(data[pos:])
		if err != nil {
			return nil, 0, err
		}
		args[i] = arg
		pos += delta
	}
	return args, pos, nil
}

// 0 1  2  3 4 5 6  7  8
// $ 3 \r \n G E T \r \n
func readBulkString(data []byte) (string, int, error) {
	pos := 1
	len, delta := readLength(data[pos:])
	pos += delta
	return string(data[pos : pos+len]), pos + len + 2, nil
}

func ParseSymbol(data []byte) (interface{}, int, error) {
	switch data[0] {
	case '*':
		return readArray(data)
	case '$':
		return readBulkString(data)
	}
	return nil, 0, nil
}

func Parser(data []byte) ([]interface{}, error) {
	//	fmt.Println("no of bytes received when connection is closed: ", len(data))
	if len(data) == 0 {
		//		fmt.Println("error generated!!")
		return nil, errors.New("no data")
	}
	//	fmt.Println("error not generted")
	var args []interface{} = make([]interface{}, 0)
	var index int = 0
	for index < len(data) {
		arg, delta, err := ParseSymbol(data[index:])
		if err != nil {
			return nil, err
		}
		args = append(args, arg)
		index += delta
	}
	return args, nil
}

func encodeString(v string) []byte {
	return []byte(fmt.Sprintf("%d\r\n%s\r\n", len(v), v))
}

func Encode(value interface{}, isSimple bool) []byte {
	switch v := value.(type) {
	case string:
		if isSimple {
			return []byte(fmt.Sprintf("+%s\r\n", v))
		} else {
			return encodeString(v)
		}
	default:
		return RESP_NIL
	}
}
