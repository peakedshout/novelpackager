package bilinovel

import (
	"errors"
	"fmt"
	"github.com/go-rod/rod"
	"github.com/peakedshout/novelpackager/pkg/epubx"
	"github.com/peakedshout/novelpackager/pkg/model"
	"github.com/peakedshout/novelpackager/pkg/rodx"
	"github.com/peakedshout/novelpackager/pkg/utils"
	"os"
	"path"
	"slices"
	"time"
)

func (p *Packager) downloadOver(out, id string) bool {
	rPath := path.Join(out, fmt.Sprintf(CacheFile, id))
	record, err := utils.LoadRecord(rPath)
	if err != nil {
		p.logger.Warnf("Failed to load record for book %s: %v", id, err)
		return false
	}
	if record.Data == nil {
		return false
	}
	return record.Data.Loaded
}

type downloadContext struct {
	id     string
	pcfg   *model.PackageConfig
	pr     *utils.Progress
	record *utils.Record
	lc     *utils.LinkCache
}

func (p *Packager) download(sess *rodx.RodSession, ctx *downloadContext) (err error) {
	if ctx.pcfg.OutputPath == "" {
		ctx.pcfg.OutputPath = "./"
	}
	stat, err := os.Stat(ctx.pcfg.OutputPath)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			p.logger.Warnf("Failed to check output path %s: %v", ctx.pcfg.OutputPath, err)
			return err
		}
		err = os.MkdirAll(ctx.pcfg.OutputPath, os.ModePerm)
		if err != nil {
			p.logger.Warnf("Failed to create output path %s: %v", ctx.pcfg.OutputPath, err)
			return err
		}
	}
	if !stat.IsDir() {
		p.logger.Warnf("Output path %s is not a directory", ctx.pcfg.OutputPath)
		return fmt.Errorf("output path %s is not a directory", ctx.pcfg.OutputPath)
	}

	rPath := path.Join(ctx.pcfg.OutputPath, fmt.Sprintf(CacheFile, ctx.id))
	record, err := utils.LoadRecord(rPath)
	if err != nil {
		p.logger.Warnf("Failed to load record for book %s: %v", ctx.id, err)
		record = &utils.Record{}
	}
	ctx.record = record
	if ctx.pcfg.Lang == "" {
		ctx.pcfg.Lang = "zh"
	}
	switch ctx.pcfg.PackageMode {
	case model.PackageModeNone, model.PackageModeDefault, model.PackageModeBook, model.PackageModeVolume, model.PackageModeChapter:
	default:
		return fmt.Errorf("invalid package mode %d", ctx.pcfg.PackageMode)
	}

	if record.Info == nil || !ctx.pcfg.DisSyncData {
		record.Info, err = p.getBookInfo(sess, ctx.id)
		if err != nil {
			p.logger.Warnf("Failed to get book info for book %s: %v", ctx.id, err)
			return err
		}
		err = p.getBookInfoFull(sess, record.Info)
		if err != nil {
			p.logger.Warnf("Failed to get book info for book %s: %v", ctx.id, err)
			return err
		}
		err = utils.SaveRecord(rPath, record, nil)
		if err != nil {
			p.logger.Warnf("Failed to get book info for book %s: %v", ctx.id, err)
			return err
		}
	}
	if record.Info == nil {
		err = fmt.Errorf("failed to get book info for book %s", ctx.id)
		p.logger.Warnf("Failed to get book info for book %s: %v", ctx.id, err)
		return err
	}
	err = p.downloadCheck(ctx)
	if err != nil {
		p.logger.Warnf("Failed to download check book for book %s: %v", ctx.id, err)
		return err
	}
	err = p.downloadBook(sess, ctx)
	if err != nil {
		p.logger.Warnf("Failed to download book for book %s: %v", ctx.id, err)
		return err
	}
	if !ctx.pcfg.KeepRecord {
		_ = os.RemoveAll(rPath)
	}
	p.logger.Info("Download book %s success", ctx.id)
	return nil
}

func (p *Packager) downloadCheck(ctx *downloadContext) error {
	if ctx.record.Data == nil {
		ctx.record.Data = &model.BookData{}
	}
	ctx.record.Data.Loaded = true

	var t int64
	var load bool

	for i, volume := range ctx.record.Info.Volumes {
		if len(ctx.pcfg.VolumeSelect) == 0 {
			load = true
		} else if slices.Contains(ctx.pcfg.VolumeSelect, i) {
			load = true
		} else {
			load = false
		}

		if i >= len(ctx.record.Data.Volumes) {
			ctx.record.Data.Loaded = false
			ctx.record.Data.Volumes = append(ctx.record.Data.Volumes, &model.VolumeData{})
		}
		if ctx.record.Data.Volumes[i].Name != volume.Name || ctx.record.Data.Volumes[i].Id != volume.Id {
			ctx.record.Data.Loaded = false
			ctx.record.Data.Volumes[i].Loaded = false
		} else {
			ctx.record.Data.Volumes[i].Loaded = true
		}
		for k, chapter := range volume.Chapters {
			if k >= len(ctx.record.Data.Volumes[i].Chapters) {
				ctx.record.Data.Loaded = false
				ctx.record.Data.Volumes[i].Loaded = false
				ctx.record.Data.Volumes[i].Chapters = append(ctx.record.Data.Volumes[i].Chapters, &model.ChapterData{})
			}

			if chapter.Name != volume.Chapters[k].Name {
				ctx.record.Data.Loaded = false
				ctx.record.Data.Volumes[i].Loaded = false
				ctx.record.Data.Volumes[i].Chapters[k].Loaded = false
			} else {
				ctx.record.Data.Volumes[i].Chapters[k].Loaded = true
			}
			if load {
				t++
			}
		}
		ctx.record.Data.Volumes[i].Chapters = ctx.record.Data.Volumes[i].Chapters[:len(volume.Chapters)]
	}
	ctx.record.Data.Volumes = ctx.record.Data.Volumes[:len(ctx.record.Info.Volumes)]

	err := utils.SaveRecord(getCachePath(ctx), ctx.record, ctx.lc)
	if err != nil {
		p.logger.Warnf("Failed to save record for book %s: %v", ctx.record.Info.Name, err)
		return err
	}

	ctx.pr.Init(t)

	return nil
}

func (p *Packager) downloadBook(sess *rodx.RodSession, ctx *downloadContext) (err error) {
	lc := utils.NewLinkCache()
	lc.Import(ctx.record.Cache)
	ctx.lc = lc
	ctx.record.Info.CoverId, _ = lc.SetX("cover", "cover"+path.Ext(ctx.record.Info.CoverId), ctx.record.Info.Cover)
	for i, info := range ctx.record.Info.Volumes {
		if len(ctx.pcfg.VolumeSelect) != 0 && !slices.Contains(ctx.pcfg.VolumeSelect, i+1) {
			continue
		}
		err = p.downloadVolume(sess, i, ctx)
		if err != nil {
			p.logger.Warnf("Failed to download volume %d for book %s: %v", i+1, info.Id, err)
			return err
		}
	}
	if ctx.pcfg.PackageMode == model.PackageModeDefault || ctx.pcfg.PackageMode == model.PackageModeBook {
		vc := map[int]map[int]bool{}
		for _, index := range ctx.pcfg.VolumeSelect {
			vc[index-1] = make(map[int]bool)
		}
		err = epubx.Build(&epubx.Config{
			Info:        ctx.record.Info,
			Data:        ctx.record.Data,
			ImgCache:    lc,
			VC:          vc,
			Lang:        ctx.pcfg.Lang,
			Output:      ctx.pcfg.OutputPath,
			PackageMode: ctx.pcfg.PackageMode,
			Source:      Source,
		})
		if err != nil {
			return err
		}
	}
	err = p.downloadCheck(ctx)
	if err != nil {
		p.logger.Warnf("Failed to save record for book %s: %v", ctx.record.Info.Name, err)
		return err
	}

	// clear
	cls := []string{ctx.record.Info.CoverId}
	for _, volume := range ctx.record.Info.Volumes {
		cls = append(cls, volume.CoverId)
	}
	rls := [][]string{cls}
	for _, volume := range ctx.record.Data.Volumes {
		for _, chapter := range volume.Chapters {
			rls = append(rls, chapter.Imgs)
		}
	}
	lc.Clear(rls)
	err = p.downloadCheck(ctx)
	if err != nil {
		p.logger.Warnf("Failed to save record for book %s: %v", ctx.record.Info.Name, err)
		return err
	}
	return nil
}

func (p *Packager) downloadVolume(sess *rodx.RodSession, index int, ctx *downloadContext) error {
	if index >= len(ctx.record.Data.Volumes) {
		ctx.record.Data.Volumes = append(ctx.record.Data.Volumes, &model.VolumeData{})
	}
	vcid := fmt.Sprintf("cover_%d", index+1)
	volume := &ctx.record.Info.Volumes[index]
	volume.CoverId, _ = ctx.lc.SetX(vcid, vcid+path.Ext(volume.CoverId), volume.Cover)

	for i := range volume.Chapters {
		err := p.downloadChapter(sess, index, i, ctx)
		if err != nil {
			p.logger.Warnf("Failed to download chapter %d for volume %d for book %s: %v", i+1, index+1, ctx.record.Info.Id, err)
			return err
		}
		ctx.pr.Add(1)
		p.logger.Infof("[%s] Successfully downloaded chapter %d for volume %d for book %s", ctx.pr.String(), i+1, index+1, ctx.record.Info.Id)
	}
	if ctx.pcfg.PackageMode == model.PackageModeVolume {
		err := epubx.Build(&epubx.Config{
			Info:     ctx.record.Info,
			Data:     ctx.record.Data,
			ImgCache: ctx.lc,
			VC: map[int]map[int]bool{
				index: {},
			},
			Lang:        ctx.pcfg.Lang,
			Output:      ctx.pcfg.OutputPath,
			PackageMode: ctx.pcfg.PackageMode,
			Source:      Source,
		})
		if err != nil {
			return err
		}
	}
	ctx.record.Data.Volumes[index].Name = ctx.record.Info.Volumes[index].Name
	ctx.record.Data.Volumes[index].Id = ctx.record.Info.Volumes[index].Id
	err := p.downloadCheck(ctx)
	if err != nil {
		p.logger.Warnf("Failed to save record for book %s: %v", ctx.record.Info.Name, err)
		return err
	}
	return nil
}

func (p *Packager) downloadChapter(sess *rodx.RodSession, index, jndex int, ctx *downloadContext) error {
	if index >= len(ctx.record.Data.Volumes) {
		ctx.record.Data.Volumes[index].Chapters = append(ctx.record.Data.Volumes[index].Chapters, &model.ChapterData{})
	}
	cInfo := &ctx.record.Info.Volumes[index].Chapters[jndex]
	cData := ctx.record.Data.Volumes[index].Chapters[jndex]
	if !cData.Loaded || cData.Name != cInfo.Name {
		err := sess.PageLoop().DoWithNumX(func(page *rod.Page) (err error) {
			sess.Browser().MustSetCookies()
			time.Sleep(1 * time.Second)
			defer utils.ExpireClose(page, p.timeout)()
			err = p.refreshToken(page, &ctx.record.Info.Volumes[index])
			if err != nil {
				return err
			}
			time.Sleep(1 * time.Second)
			utils.UpdateExpireClose(page, p.timeout)
			p.logger.Info("Fetching chapter :", ctx.record.Info.Name, ctx.record.Info.Volumes[index].Name, cInfo.Name)
			err = p.checkoutChapter(page, cInfo, cData, ctx)
			if err != nil {
				return err
			}
			return nil
		}, p.retryNum)

		if err != nil {
			return err
		}
		err = p.downloadCheck(ctx)
		if err != nil {
			p.logger.Warnf("Failed to save record for book %s: %v", ctx.record.Info.Name, err)
			return err
		}
	}
	if ctx.pcfg.PackageMode == model.PackageModeChapter {
		err := epubx.Build(&epubx.Config{
			Info:     ctx.record.Info,
			Data:     ctx.record.Data,
			ImgCache: ctx.lc,
			VC: map[int]map[int]bool{
				index: {
					jndex: true,
				},
			},
			Lang:        ctx.pcfg.Lang,
			Output:      ctx.pcfg.OutputPath,
			PackageMode: ctx.pcfg.PackageMode,
			Source:      Source,
		})
		if err != nil {
			return err
		}
	}
	ctx.record.Data.Volumes[index].Chapters[jndex].Name = ctx.record.Info.Volumes[index].Chapters[jndex].Name
	err := p.downloadCheck(ctx)
	if err != nil {
		p.logger.Warnf("Failed to save record for book %s: %v", ctx.record.Info.Name, err)
		return err
	}
	return nil
}
