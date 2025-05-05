package utils

import (
	"bytes"
	"encoding/gob"
	"errors"
	"github.com/peakedshout/go-pandorasbox/tool/xerror"
	"os"
	"path"
	"sort"
	"sync"
	"time"
)

type KVCache interface {
	Get(key string) ([]byte, error)
	Set(key string, data []byte, expired ...time.Duration) error
	Del(key string) error
	Range(fn func(key string, data []byte) error) error
	Flush() error
}

var (
	ErrNotFound = xerror.New("kv: not found %v")
	ErrFailed   = xerror.New("kv: failed %v")
)

func NewKVCache(flushFile string, bsMax int64) (KVCache, error) {
	kv := &kvCache{
		flushFile: flushFile,
		kvMap:     make(map[string]*ExpiredData[[]byte]),
		bsMax:     bsMax,
	}
	err := kv.init()
	return kv, err
}

type kvCache struct {
	flushFile string
	rw        sync.RWMutex
	kvMap     map[string]*ExpiredData[[]byte]
	bsMax     int64
}

func (kv *kvCache) Get(key string) ([]byte, error) {
	defer kv.lockR()()
	if v, ok := kv.kvMap[key]; ok {
		if !v.TD.IsZero() && time.Now().After(v.TD) {
			return nil, ErrNotFound.Errorf(key)
		}
		return v.Data, nil
	}
	return nil, ErrNotFound.Errorf(key)
}

func (kv *kvCache) Set(key string, data []byte, expired ...time.Duration) error {
	defer kv.lock()()
	ed := &ExpiredData[[]byte]{
		Key:  key,
		Data: data,
	}
	if len(expired) > 0 {
		ed.TD = time.Now().Add(expired[0])
	}
	kv.kvMap[key] = ed
	return kv.flush(false)
}

func (kv *kvCache) Del(key string) error {
	defer kv.lock()()
	delete(kv.kvMap, key)
	return kv.flush(false)
}

func (kv *kvCache) Range(fn func(key string, data []byte) error) error {
	defer kv.lockR()()
	t := time.Now()
	for k, v := range kv.kvMap {
		if !v.TD.IsZero() && t.After(v.TD) {
			continue
		}
		err := fn(k, v.Data)
		if err != nil {
			return err
		}
	}
	return nil
}

func (kv *kvCache) init() error {
	file, err := os.Open(kv.flushFile)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		return kv.Flush()
	}
	defer file.Close()
	return gob.NewDecoder(file).Decode(&kv.kvMap)
}

func (kv *kvCache) Flush() error {
	return kv.flush(true)
}

func (kv *kvCache) flush(lock bool) error {
	if lock {
		defer kv.lock()()
	}
	file, err := os.Create(kv.flushFile)
	if err != nil {
		return err
	}
	defer file.Close()

	for {
		err = gob.NewEncoder(file).Encode(kv.kvMap)
		if err != nil {
			return err
		}
		if len(kv.kvMap) != 0 && kv.bsMax > 0 {
			stat, err := file.Stat()
			if err != nil {
				return err
			}
			if stat.Size() > kv.bsMax {
				kv.clear()
				err = file.Truncate(0)
				if err != nil {
					return err
				}
				continue
			}
		}
		break
	}

	return nil
}

func (kv *kvCache) clear() {
	l := len(kv.kvMap)
	newMap := make(map[string]*ExpiredData[[]byte], l)
	t := time.Now()
	for k, v := range kv.kvMap {
		if !v.TD.IsZero() && t.After(v.TD) {
			continue
		}
		newMap[k] = v
	}
	if len(newMap) <= l/2 {
		kv.kvMap = newMap
		return
	}
	sl := make([]*ExpiredData[[]byte], 0, len(newMap))
	for _, v := range newMap {
		sl = append(sl, v)
	}
	sort.Slice(sl, func(i, j int) bool {
		a, b := sl[i], sl[j]
		aZero, bZero := a.TD.IsZero(), b.TD.IsZero()

		switch {
		case aZero && !bZero:
			return true
		case !aZero && bZero:
			return false
		case aZero && bZero:
			return false
		default:
			return a.TD.After(b.TD)
		}
	})
	kv.kvMap = make(map[string]*ExpiredData[[]byte], len(newMap)/2)
	for _, v := range sl[:len(newMap)/2] {
		kv.kvMap[v.Key] = v
	}
}

func (kv *kvCache) lockR() func() {
	kv.rw.RLock()
	return kv.rw.RUnlock
}

func (kv *kvCache) lock() func() {
	kv.rw.Lock()
	return kv.rw.Unlock
}

func KVCacheGetT[T any](kv KVCache, ks ...string) (t T, err error) {
	bs, err := kv.Get(path.Join(ks...))
	if err != nil {
		return t, err
	}
	err = gob.NewDecoder(bytes.NewReader(bs)).Decode(&t)
	return t, err
}

func KVCacheSetExpiredT[T any](kv KVCache, t T, expired time.Duration, ks ...string) error {
	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(t)
	if err != nil {
		return err
	}
	return kv.Set(path.Join(ks...), buf.Bytes(), expired)
}

func KVCacheSetT[T any](kv KVCache, t T, ks ...string) error {
	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(t)
	if err != nil {
		return err
	}
	return kv.Set(path.Join(ks...), buf.Bytes())
}

func KVCacheDelT[T any](kv KVCache, ks ...string) error {
	return kv.Del(path.Join(ks...))
}

type ExpiredData[T any] struct {
	Key  string
	Data T
	TD   time.Time
}
