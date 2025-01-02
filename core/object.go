package core

type Obj struct {
	TypeEncoding   uint8
	LastAccessedAt uint32
	Value          interface{}
}

var OBJ_TYPE_STRING uint8 = 0
var OBJ_ENCODING_RAW uint8 = 0
var OBJ_ENCODING_INT uint8 = 1
var OBJ_ENCODING_EMBSTR uint8 = 8