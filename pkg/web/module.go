package web

import (
	"context"
	"fmt"
	"github.com/peakedshout/novelpackager/pkg/epubx"
	"github.com/peakedshout/novelpackager/pkg/model"
	"github.com/peakedshout/novelpackager/pkg/rodx"
	"github.com/peakedshout/novelpackager/pkg/utils"
)

func getSource(name string) (Source, error) {
	source, ok := sourceMap[name]
	if !ok {
		return nil, fmt.Errorf("unknown source: %s", name)
	}
	return source, nil
}

func buildSource(ctx *BuildContext) {
	for _, s := range sourceList {
		source := s(ctx)
		sourceMap[source.Name()] = source
	}
}

func Register(s BuildSource) {
	sourceList = append(sourceList, s)
}

var sourceMap = make(map[string]Source)
var sourceList []BuildSource

type BuildContext struct {
	Ctx        context.Context
	RodContext *rodx.RodContext
	Cache      utils.KVCache
	CacheDir   string
}

type BuildSource func(ctx *BuildContext) Source

type Source interface {
	Name() string
	GetInfo(ctx context.Context, id string, full bool) (*model.BookInfo, error)
	Search(ctx context.Context, name string, full bool, noImg bool) ([]model.SearchResult, error)
	Progress(ctx context.Context) map[string]string
	Caching(ctx context.Context, id string) error
	EnableDownload(ctx context.Context, id string) ([]string, error)
	Download(ctx context.Context, id string, vols ...int) (*epubx.FBytesData, error)
}
