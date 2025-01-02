package core

import (
	"container/list"
	"redis/config"
	"time"
)

type Pair struct {
	key string
	obj *Obj
}

var store map[string]*list.Element
var expires map[*Obj]uint64
var storeList *list.List

func init() {
	store = make(map[string]*list.Element)
	storeList = list.New()
	expires = make(map[*Obj]uint64)
}

func setExpiry(obj *Obj, exDurationMs int64) {
	expires[obj] = uint64(time.Now().UnixMilli()) + uint64(exDurationMs)
}

func NewObj(value interface{}, exDurationMs int64, oType uint8, oEnc uint8) *Obj {

	obj := &Obj{
		Value:          value,
		TypeEncoding:   oType | oEnc,
		LastAccessedAt: getCurrentClock(),
	}
	if exDurationMs > 0 {
		setExpiry(obj, exDurationMs)
	}
	return obj
}

func Put(key string, obj *Obj) {
	if _, exists := store[key]; exists {
		keyPtr := store[key]
		storeList.Remove(keyPtr)
		storeList.PushFront(Pair{key, obj})
	} else if storeList.Len() >= config.KeysLimit {
		leastAccessedPtr := storeList.Back()
		leastAccessedPair := leastAccessedPtr.Value.(Pair) // type assert it is a pair, then only .key or .obj will work
		delete(store, leastAccessedPair.key)
		storeList.Remove(leastAccessedPtr)
		storeList.PushFront(Pair{key, obj})
	} else {
		storeList.PushFront(Pair{key, obj})
	}
	store[key] = storeList.Front()
}

func Get(key string) *Obj {
	if _, exists := store[key]; exists {
		keyPtr := store[key]
		obj := keyPtr.Value.(Pair).obj
		storeList.Remove(keyPtr)
		storeList.PushFront(Pair{key, obj})
		store[key] = storeList.Front()
		return obj
	}
	return nil
}
