package utils

import (
	"bytes"
	"context"
	"github.com/peakedshout/go-pandorasbox/tool/uuid"
	"net/url"
	"path"
	"sync"
)

func GetLinkCache(ctx context.Context) *LinkCache {
	value := ctx.Value("link_cache")
	if value == nil {
		return gLinkCache
	}
	return value.(*LinkCache)
}

func SetLinkCache(ctx context.Context, lc *LinkCache) context.Context {
	return context.WithValue(ctx, "link_cache", lc)
}

func NewLinkCache() *LinkCache {
	return &LinkCache{
		idm:  make(map[string]*linkCacheUnit),
		srcm: make(map[string]*linkCacheUnit),
	}
}

var gLinkCache = &LinkCache{
	idm:  make(map[string]*linkCacheUnit),
	srcm: make(map[string]*linkCacheUnit),
}

type LinkCache struct {
	mux  sync.RWMutex
	idm  map[string]*linkCacheUnit
	srcm map[string]*linkCacheUnit
}

type linkCacheUnit struct {
	src  string
	id   string
	data []byte
}

func (lc *LinkCache) SetRaw(id string, data []byte) {
	if lc == nil {
		return
	}
	lc.mux.Lock()
	defer lc.mux.Unlock()
	lc.srcm[id] = &linkCacheUnit{
		id:   id,
		data: data,
	}
}

func (lc *LinkCache) SetX(id string, src string, data []byte) (string, error) {
	if lc == nil {
		return "", nil
	}
	lc.mux.Lock()
	defer lc.mux.Unlock()
	pu, err := url.Parse(src)
	if err != nil {
		return "", err
	}

	unit, ok := lc.srcm[src]
	if ok {
		if !bytes.Equal(unit.data, data) {
			unit.data = data
		}
		return unit.id, nil
	}
	u := &linkCacheUnit{
		src:  src,
		id:   "res_" + id + path.Ext(pu.Path),
		data: data,
	}
	lc.idm[u.id] = u
	lc.srcm[src] = u

	return u.id, nil
}

func (lc *LinkCache) Set(src string, data []byte) (string, error) {
	if lc == nil {
		return "", nil
	}
	return lc.SetX(uuid.NewId(1), src, data)
}

func (lc *LinkCache) Get(id string) []byte {
	if lc == nil {
		return nil
	}
	lc.mux.RLock()
	defer lc.mux.RUnlock()
	unit := lc.idm[id]
	if unit == nil {
		return nil
	}
	return unit.data
}

func (lc *LinkCache) Range(fn func(id string, data []byte) error) error {
	if lc == nil {
		return nil
	}
	lc.mux.RLock()
	defer lc.mux.RUnlock()
	for id, data := range lc.idm {
		err := fn(id, data.data)
		if err != nil {
			return err
		}
	}
	return nil
}

func (lc *LinkCache) Clear(ids [][]string) {
	if lc == nil {
		return
	}
	lc.mux.RLock()
	defer lc.mux.RUnlock()
	idm := make(map[string]bool)
	for _, l := range ids {
		for _, id := range l {
			idm[id] = true
		}
	}
	nidm := make(map[string]*linkCacheUnit, len(lc.idm))
	for id, unit := range lc.idm {
		if !idm[id] {
			continue
		}
		nidm[id] = unit
	}
	lc.idm = nidm
	lc.srcm = make(map[string]*linkCacheUnit, len(lc.idm))
	for _, u := range lc.idm {
		lc.srcm[u.src] = u
	}
}

type ExportCache struct {
	Src  string `json:"src"`
	Id   string `json:"id"`
	Data []byte `json:"data"`
}

func (lc *LinkCache) Export() map[string]*ExportCache {
	if lc == nil {
		return map[string]*ExportCache{}
	}
	lc.mux.Lock()
	defer lc.mux.Unlock()
	m := make(map[string]*ExportCache, len(lc.idm))
	for id, unit := range lc.idm {
		m[id] = &ExportCache{
			Src:  unit.src,
			Id:   unit.id,
			Data: unit.data,
		}
	}
	return m
}

func (lc *LinkCache) Import(ec map[string]*ExportCache) {
	if lc == nil {
		return
	}
	lc.mux.Lock()
	defer lc.mux.Unlock()
	lc.idm = make(map[string]*linkCacheUnit, len(ec))
	for id, unit := range ec {
		lc.idm[id] = &linkCacheUnit{
			src:  unit.Src,
			id:   unit.Id,
			data: unit.Data,
		}
	}
	lc.srcm = make(map[string]*linkCacheUnit, len(lc.idm))
	for _, u := range lc.idm {
		lc.srcm[u.src] = u
	}
}
