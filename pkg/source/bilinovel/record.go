package bilinovel

import (
	"fmt"
	"github.com/peakedshout/novelpackager/pkg/epubx"
	"github.com/peakedshout/novelpackager/pkg/model"
	"github.com/peakedshout/novelpackager/pkg/utils"
	"path"
)

func (p *Packager) RecordExtract(out, id string, vols ...int) (*epubx.FBytesData, error) {
	rPath := path.Join(out, fmt.Sprintf(CacheFile, id))
	record, err := utils.LoadRecord(rPath)
	if err != nil {
		p.logger.Warnf("Failed to load record for book %s: %v", id, err)
		return nil, err
	}
	if record.Info == nil {
		return nil, fmt.Errorf("book %s not loaded", id)
	}

	vc := map[int]map[int]bool{}
	for _, index := range vols {
		if index > len(record.Info.Volumes) || index <= 0 {
			return nil, fmt.Errorf("book %s volume %d not found", id, index)
		}
		vc[index-1] = make(map[int]bool)
	}
	pm := model.PackageModeDefault
	if len(vols) == 1 {
		pm = model.PackageModeVolume
	}

	if record.Data == nil || record.Data.Loaded == false {
		return nil, fmt.Errorf("book %s not loaded", id)
	}
	lc := utils.NewLinkCache()
	lc.Import(record.Cache)
	ch := make(chan *epubx.FBytesData, 1)
	err = epubx.Build(&epubx.Config{
		Info:        record.Info,
		Data:        record.Data,
		ImgCache:    lc,
		VC:          vc,
		Lang:        "zh",
		OutputChan:  ch,
		PackageMode: pm,
		Source:      Source,
	})
	if err != nil {
		return nil, err
	}
	return <-ch, nil
}
