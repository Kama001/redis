package core

import "time"

func getCurrentClock() uint32 {
	return uint32(time.Now().Unix()) & 0x00FFFFFF
}
