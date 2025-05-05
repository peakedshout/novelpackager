package bilinovel

import (
	"context"
	"fmt"
	"github.com/peakedshout/novelpackager/pkg/epubx"
	"github.com/peakedshout/novelpackager/pkg/model"
	"github.com/peakedshout/novelpackager/pkg/utils"
	"github.com/peakedshout/novelpackager/pkg/web"
	"path"
	"sync"
	"time"
)

func init() {
	web.Register(func(ctx *web.BuildContext) web.Source {
		lr := utils.NewLimiter(ctx.Ctx)
		lr.Add("GetInfo", 1)
		lr.Add("Search", 1)
		lr.Add("Cache", 1)
		lr.Add("Download", 3)
		p := NewPackager(ctx.RodContext, &Config{})
		return &WebSource{
			ctx: ctx.Ctx,
			p:   p,
			pcfg: &model.PackageConfig{
				KeepRecord:   true,
				OutputPath:   ctx.CacheDir,
				DisSyncData:  false,
				PackageMode:  model.PackageModeNone,
				VolumeSelect: nil,
				Lang:         "",
			},
			kvCache: ctx.Cache,
			prMap:   make(map[string]*utils.Progress),
			limiter: lr,
		}
	})
}

type WebSource struct {
	ctx context.Context

	p    *Packager
	pcfg *model.PackageConfig

	kvCache utils.KVCache

	prMux sync.Mutex
	prMap map[string]*utils.Progress

	limiter *utils.Limiter
}

func (w *WebSource) Name() string {
	return Source
}

func (w *WebSource) GetInfo(ctx context.Context, id string, full bool) (*model.BookInfo, error) {
	info, err := utils.KVCacheGetT[*model.BookInfo](w.kvCache, Source, "BookInfo", id)
	if err == nil {
		return info, nil
	}
	fn, err := w.limiter.LimitTimeout("Info", 3*time.Second)
	if err != nil {
		return nil, err
	}
	defer fn()
	info, err = w.p.GetInfo(ctx, id, full)
	if err != nil {
		return nil, err
	}
	_ = utils.KVCacheSetExpiredT(w.kvCache, info, 24*time.Hour, Source, "BookInfo", id)
	return info, nil
}

func (w *WebSource) Search(ctx context.Context, name string, full bool, noImg bool) ([]model.SearchResult, error) {
	sl, err := utils.KVCacheGetT[[]model.SearchResult](w.kvCache, Source, "SearchResult", name)
	if err == nil {
		return sl, nil
	}
	fn, err := w.limiter.LimitTimeout("Search", 3*time.Second)
	if err != nil {
		return nil, err
	}
	defer fn()
	sl, err = w.p.Search(ctx, name, full, noImg)
	if err != nil {
		return nil, err
	}
	_ = utils.KVCacheSetExpiredT(w.kvCache, sl, 24*time.Hour, Source, "SearchResult", name)
	return sl, nil
}

func (w *WebSource) Progress(ctx context.Context) map[string]string {
	w.prMux.Lock()
	defer w.prMux.Unlock()
	m := make(map[string]string, len(w.prMap))
	for k, v := range w.prMap {
		m[k] = v.String()
	}
	return m
}

func (w *WebSource) Caching(ctx context.Context, id string) error {
	fn, err := w.limiter.LimitTimeout("Cache", 3*time.Second)
	if err != nil {
		return err
	}
	go func() {
		defer fn()
		_ = w.caching(w.ctx, id)
	}()
	return nil
}

func (w *WebSource) caching(ctx context.Context, id string) (err error) {
	sess, err := w.p.rc.NewSession(ctx)
	if err != nil {
		return err
	}
	defer sess.Close()
	defer blockURLs(sess.Browser())()

	w.prMux.Lock()
	pr, ok := w.prMap[id]
	if ok && (pr.Error() == nil && pr.Percent() != 1) {
		w.prMux.Unlock()
		return fmt.Errorf("already caching %s", id)
	}
	pr = utils.NewProgress(-1)
	w.prMap[id] = pr
	w.prMux.Unlock()

	defer func() {
		if err != nil {
			pr.SetError(err)
		}
	}()

	return w.p.download(sess, &downloadContext{
		id:     id,
		pcfg:   w.pcfg,
		pr:     pr,
		record: nil,
		lc:     nil,
	})
}

func (w *WebSource) EnableDownload(ctx context.Context, id string) ([]string, error) {
	rPath := path.Join(w.pcfg.OutputPath, fmt.Sprintf(CacheFile, id))
	record, err := utils.LoadRecord(rPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load record for book %s: %v", id, err)
	}
	if record.Data == nil || !record.Data.Loaded {
		return nil, fmt.Errorf("record for book %s is not loaded", id)
	}
	var sl []string
	for _, volume := range record.Data.Volumes {
		sl = append(sl, volume.Name)
	}
	return sl, nil
}

func (w *WebSource) Download(ctx context.Context, id string, vols ...int) (*epubx.FBytesData, error) {
	fn, err := w.limiter.LimitTimeout("Download", 3*time.Second)
	if err != nil {
		return nil, err
	}
	defer fn()
	return w.p.RecordExtract(w.pcfg.OutputPath, id, vols...)
}
