package core

import "time"

func hasExpired(obj *Obj) bool {
	if exp, exists := expires[obj]; exists {
		return uint64(time.Now().UnixMilli()) > exp
	} else {
		return false
	}
}
