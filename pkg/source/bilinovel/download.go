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

func (p *Packager) download(sess *rodx.RodSession, id string, pcfg *model.PackageConfig) (err error) {
	if pcfg.OutputPath == "" {
		pcfg.OutputPath = "./"
	}
	stat, err := os.Stat(pcfg.OutputPath)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			p.logger.Warnf("Failed to check output path %s: %v", pcfg.OutputPath, err)
			return err
		}
		err = os.MkdirAll(pcfg.OutputPath, os.ModePerm)
		if err != nil {
			p.logger.Warnf("Failed to create output path %s: %v", pcfg.OutputPath, err)
			return err
		}
	}
	if !stat.IsDir() {
		p.logger.Warnf("Output path %s is not a directory", pcfg.OutputPath)
		return fmt.Errorf("output path %s is not a directory", pcfg.OutputPath)
	}

	rPath := path.Join(pcfg.OutputPath, fmt.Sprintf(CacheFile, id))
	record, err := utils.LoadRecord(rPath)
	if err != nil {
		p.logger.Warnf("Failed to load record for book %s: %v", id, err)
		record = &utils.Record{}
	}
	record.Config = pcfg
	if record.Config.Lang == "" {
		record.Config.Lang = "zh"
	}
	switch record.Config.PackageMode {
	case model.PackageModeNone, model.PackageModeDefault, model.PackageModeBook, model.PackageModeVolume, model.PackageModeChapter:
	default:
		return fmt.Errorf("invalid package mode %d", record.Config.PackageMode)
	}

	if record.Info == nil || !record.Config.DisSyncData {
		record.Info, err = p.getBookInfo(sess, id)
		if err != nil {
			p.logger.Warnf("Failed to get book info for book %s: %v", id, err)
			return err
		}
		err = p.getBookInfoFull(sess, record.Info)
		if err != nil {
			p.logger.Warnf("Failed to get book info for book %s: %v", id, err)
			return err
		}
		err = utils.SaveRecord(rPath, record, nil)
		if err != nil {
			p.logger.Warnf("Failed to get book info for book %s: %v", id, err)
			return err
		}
	}
	if record.Info == nil {
		err = fmt.Errorf("failed to get book info for book %s", id)
		p.logger.Warnf("Failed to get book info for book %s: %v", id, err)
		return err
	}
	err = p.downloadCheck(record, nil)
	if err != nil {
		p.logger.Warnf("Failed to download check book for book %s: %v", id, err)
		return err
	}
	err = p.downloadBook(sess, record)
	if err != nil {
		p.logger.Warnf("Failed to download book for book %s: %v", id, err)
		return err
	}
	if !record.Config.KeepRecord {
		_ = os.RemoveAll(rPath)
	}
	p.logger.Info("Download book %s success", id)
	return nil
}

func (p *Packager) downloadCheck(record *utils.Record, lc *utils.LinkCache) error {
	if record.Data == nil {
		record.Data = &model.BookData{}
	}
	record.Data.Loaded = true
	for i, volume := range record.Info.Volumes {
		if i >= len(record.Data.Volumes) {
			record.Data.Loaded = false
			record.Data.Volumes = append(record.Data.Volumes, &model.VolumeData{})
		}
		if record.Data.Volumes[i].Name != volume.Name || record.Data.Volumes[i].Id != volume.Id {
			record.Data.Loaded = false
			record.Data.Volumes[i].Loaded = false
		} else {
			record.Data.Volumes[i].Loaded = true
		}
		for k, chapter := range volume.Chapters {
			if k >= len(record.Data.Volumes[i].Chapters) {
				record.Data.Loaded = false
				record.Data.Volumes[i].Loaded = false
				record.Data.Volumes[i].Chapters = append(record.Data.Volumes[i].Chapters, &model.ChapterData{})
			}
			if chapter.Name != volume.Chapters[k].Name {
				record.Data.Loaded = false
				record.Data.Volumes[i].Loaded = false
				record.Data.Volumes[i].Chapters[k].Loaded = false
			} else {
				record.Data.Volumes[i].Chapters[k].Loaded = true
			}
		}
		record.Data.Volumes[i].Chapters = record.Data.Volumes[i].Chapters[:len(volume.Chapters)]
	}
	record.Data.Volumes = record.Data.Volumes[:len(record.Info.Volumes)]

	err := utils.SaveRecord(getCachePath(record), record, lc)
	if err != nil {
		p.logger.Warnf("Failed to save record for book %s: %v", record.Info.Name, err)
		return err
	}

	return nil
}

func (p *Packager) downloadBook(sess *rodx.RodSession, record *utils.Record) (err error) {
	lc := utils.NewLinkCache()
	lc.Import(record.Cache)
	record.Info.CoverId, _ = lc.SetX("cover", "cover"+path.Ext(record.Info.CoverId), record.Info.Cover)
	for i, info := range record.Info.Volumes {
		if len(record.Config.VolumeSelect) != 0 && !slices.Contains(record.Config.VolumeSelect, i+1) {
			continue
		}
		err = p.downloadVolume(sess, record, i, lc)
		if err != nil {
			p.logger.Warnf("Failed to download volume %d for book %s: %v", i+1, info.Id, err)
			return err
		}
	}
	if record.Config.PackageMode == model.PackageModeDefault || record.Config.PackageMode == model.PackageModeBook {
		vc := map[int]map[int]bool{}
		for _, index := range record.Config.VolumeSelect {
			vc[index-1] = make(map[int]bool)
		}
		err = epubx.Build(&epubx.Config{
			Info:        record.Info,
			Data:        record.Data,
			ImgCache:    lc,
			VC:          vc,
			Lang:        record.Config.Lang,
			Output:      record.Config.OutputPath,
			PackageMode: record.Config.PackageMode,
			Source:      Source,
		})
		if err != nil {
			return err
		}
	}
	err = p.downloadCheck(record, lc)
	if err != nil {
		p.logger.Warnf("Failed to save record for book %s: %v", record.Info.Name, err)
		return err
	}

	// clear
	cls := []string{record.Info.CoverId}
	for _, volume := range record.Info.Volumes {
		cls = append(cls, volume.CoverId)
	}
	rls := [][]string{cls}
	for _, volume := range record.Data.Volumes {
		for _, chapter := range volume.Chapters {
			rls = append(rls, chapter.Imgs)
		}
	}
	lc.Clear(rls)
	err = p.downloadCheck(record, lc)
	if err != nil {
		p.logger.Warnf("Failed to save record for book %s: %v", record.Info.Name, err)
		return err
	}
	return nil
}

func (p *Packager) downloadVolume(sess *rodx.RodSession, record *utils.Record, index int, lc *utils.LinkCache) error {
	if index >= len(record.Data.Volumes) {
		record.Data.Volumes = append(record.Data.Volumes, &model.VolumeData{})
	}
	vcid := fmt.Sprintf("cover_%d", index+1)
	volume := &record.Info.Volumes[index]
	volume.CoverId, _ = lc.SetX(vcid, vcid+path.Ext(volume.CoverId), volume.Cover)

	for i := range volume.Chapters {
		err := p.downloadChapter(sess, record, index, i, lc)
		if err != nil {
			p.logger.Warnf("Failed to download chapter %d for volume %d for book %s: %v", i+1, index+1, record.Info.Id, err)
			return err
		}
	}
	if record.Config.PackageMode == model.PackageModeVolume {
		err := epubx.Build(&epubx.Config{
			Info:     record.Info,
			Data:     record.Data,
			ImgCache: lc,
			VC: map[int]map[int]bool{
				index: {},
			},
			Lang:        record.Config.Lang,
			Output:      record.Config.OutputPath,
			PackageMode: record.Config.PackageMode,
			Source:      Source,
		})
		if err != nil {
			return err
		}
	}
	record.Data.Volumes[index].Name = record.Info.Volumes[index].Name
	record.Data.Volumes[index].Id = record.Info.Volumes[index].Id
	err := p.downloadCheck(record, lc)
	if err != nil {
		p.logger.Warnf("Failed to save record for book %s: %v", record.Info.Name, err)
		return err
	}
	return nil
}

func (p *Packager) downloadChapter(sess *rodx.RodSession, record *utils.Record, index, jndex int, lc *utils.LinkCache) error {
	if index >= len(record.Data.Volumes) {
		record.Data.Volumes[index].Chapters = append(record.Data.Volumes[index].Chapters, &model.ChapterData{})
	}
	cInfo := &record.Info.Volumes[index].Chapters[jndex]
	cData := record.Data.Volumes[index].Chapters[jndex]
	if !cData.Loaded || cData.Name != cInfo.Name {
		err := sess.PageLoop().DoWithNumX(func(page *rod.Page) (err error) {
			sess.Browser().MustSetCookies()
			time.Sleep(1 * time.Second)
			defer utils.ExpireClose(page, p.timeout)()
			err = p.refreshToken(page, &record.Info.Volumes[index])
			if err != nil {
				return err
			}
			time.Sleep(1 * time.Second)
			utils.UpdateExpireClose(page, p.timeout)
			p.logger.Info("Fetching chapter :", record.Info.Name, record.Info.Volumes[index].Name, cInfo.Name)
			err = p.checkoutChapter(page, cInfo, cData, lc)
			if err != nil {
				return err
			}
			return nil
		}, p.retryNum)

		if err != nil {
			return err
		}
		err = p.downloadCheck(record, lc)
		if err != nil {
			p.logger.Warnf("Failed to save record for book %s: %v", record.Info.Name, err)
			return err
		}
	}
	if record.Config.PackageMode == model.PackageModeChapter {
		err := epubx.Build(&epubx.Config{
			Info:     record.Info,
			Data:     record.Data,
			ImgCache: lc,
			VC: map[int]map[int]bool{
				index: {
					jndex: true,
				},
			},
			Lang:        record.Config.Lang,
			Output:      record.Config.OutputPath,
			PackageMode: record.Config.PackageMode,
			Source:      Source,
		})
		if err != nil {
			return err
		}
	}
	record.Data.Volumes[index].Chapters[jndex].Name = record.Info.Volumes[index].Chapters[jndex].Name
	err := p.downloadCheck(record, lc)
	if err != nil {
		p.logger.Warnf("Failed to save record for book %s: %v", record.Info.Name, err)
		return err
	}
	return nil
}
