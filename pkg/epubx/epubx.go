package epubx

import (
	"fmt"
	"github.com/go-shiori/go-epub"
	"github.com/google/uuid"
	"github.com/peakedshout/novelpackager/pkg/model"
	"github.com/peakedshout/novelpackager/pkg/utils"
	"html"
	"os"
	"path"
	"strings"
)

type Config struct {
	Info     *model.BookInfo
	Data     *model.BookData
	ImgCache *utils.LinkCache
	VC       map[int]map[int]bool

	Lang        string
	Output      string
	PackageMode model.PackageMode
	Source      string
}

func Build(cfg *Config) error {
	ec := &epubContext{
		id:     uuid.New().String(),
		info:   cfg.Info,
		data:   cfg.Data,
		lc:     cfg.ImgCache,
		lang:   cfg.Lang,
		output: cfg.Output,
		mode:   cfg.PackageMode,
		tmpDir: "",
		source: cfg.Source,
	}
	err := ec.build()
	if err != nil {
		return err
	}
	return nil
}

type epubContext struct {
	id string

	info *model.BookInfo
	data *model.BookData
	lc   *utils.LinkCache

	vcm map[int]map[int]bool

	lang   string
	output string
	mode   model.PackageMode

	tmpDir string
	source string
}

func (ec *epubContext) build() (err error) {
	ec.tmpDir, err = os.MkdirTemp("", "np")
	if err != nil {
		return err
	}
	defer os.RemoveAll(ec.tmpDir)

	switch ec.mode {
	case model.PackageModeBook, model.PackageModeDefault:
		return ec.buildBookContent()
	case model.PackageModeVolume:
		return ec.buildBookContentVolume()
	case model.PackageModeChapter:
		return ec.buildBookContentChapter()
	default:
		return fmt.Errorf("unknown package mode")
	}
}

func (ec *epubContext) buildResById(ep *epub.Epub, lc *utils.LinkCache, id string, tmpMap map[string]bool) (err error) {
	if tmpMap[id] {
		return nil
	}
	data := lc.Get(id)

	p := path.Join(ec.tmpDir, id)
	err = os.WriteFile(p, data, 0666)
	if err != nil {
		return err
	}
	_, err = ep.AddImage(p, id)
	if err != nil {
		return err
	}
	tmpMap[id] = true
	return nil
}

func (ec *epubContext) buildBookContentChapter() error {
	for i, volume := range ec.info.Volumes {
		if !verifyVolume(ec.vcm, i) {
			continue
		}
		if !ec.data.Volumes[i].Loaded {
			continue
		}
		for k, chapter := range volume.Chapters {
			if !verifyChapter(ec.vcm, i, k) {
				continue
			}

			fp := path.Join(ec.output, fmt.Sprintf("%s_%d_%s_%d_%s.epub", verifyFileName(ec.info.Name), i+1, verifyFileName(volume.Name), k+1, verifyFileName(chapter.Name)))
			if ec.data.Volumes[i].Chapters[k].Loaded {
				fh, _ := utils.FileHashSha256(fp)
				if len(fh) != 0 && fh == ec.data.Volumes[i].Chapters[k].Hash {
					continue
				}
			} else {
				continue
			}

			ep, err := epub.NewEpub(fmt.Sprintf("%s %s %s", ec.info.Name, volume.Name, chapter.Name))
			if err != nil {
				return err
			}
			tmpMap := make(map[string]bool)

			ep.SetAuthor(ec.info.Author)

			err = ep.SetCover(fmt.Sprintf("../images/%s", volume.CoverId), "")
			if err != nil {
				return err
			}
			err = ec.buildResById(ep, ec.lc, volume.CoverId, tmpMap)
			if err != nil {
				return err
			}

			ep.SetDescription(volume.Description)
			ep.SetLang(ec.lang)
			ep.SetIdentifier(fmt.Sprintf("%s_%s_%s_%d_%s", ec.source, ec.id, volume.Id, k, chapter.Name))

			cbody := fmt.Sprintf(`<h1>%s</h1>
<h2>%s</h2>
<h3>%s</h3>
%s`, html.EscapeString(ec.info.Name), html.EscapeString(volume.Name), html.EscapeString(chapter.Name), strings.Join(ec.data.Volumes[i].Chapters[k].Data, "\n"))
			_, err = ep.AddSection(cbody, chapter.Name, fmt.Sprintf("chapter%d_%d.xhtml", i+1, k+1), "")
			if err != nil {
				return err
			}
			for _, cid := range ec.data.Volumes[i].Chapters[k].Imgs {
				err = ec.buildResById(ep, ec.lc, cid, tmpMap)
				if err != nil {
					return err
				}
			}
			err = ep.Write(fp)
			if err != nil {
				return err
			}
			fh, err := utils.FileHashSha256(fp)
			if err != nil {
				return err
			}
			ec.data.Volumes[i].Chapters[k].Hash = fh
		}
	}
	return nil
}

func (ec *epubContext) buildBookContentVolume() error {
	for i, volume := range ec.info.Volumes {
		if !verifyVolume(ec.vcm, i) {
			continue
		}

		fp := path.Join(ec.output, fmt.Sprintf("%s_%d_%s.epub", verifyFileName(ec.info.Name), i+1, verifyFileName(volume.Name)))
		if ec.data.Volumes[i].Loaded {
			fh, _ := utils.FileHashSha256(fp)
			if len(fh) != 0 && fh == ec.data.Volumes[i].Hash {
				continue
			}
		} else {
			continue
		}

		ep, err := epub.NewEpub(fmt.Sprintf("%s %s", ec.info.Name, volume.Name))
		if err != nil {
			return err
		}
		tmpMap := make(map[string]bool)

		ep.SetAuthor(ec.info.Author)

		err = ep.SetCover(fmt.Sprintf("../images/%s", volume.CoverId), "")
		if err != nil {
			return err
		}
		err = ec.buildResById(ep, ec.lc, volume.CoverId, tmpMap)
		if err != nil {
			return err
		}

		ep.SetDescription(volume.Description)
		ep.SetLang(ec.lang)
		ep.SetIdentifier(fmt.Sprintf("%s_%s_%s", ec.source, ec.id, volume.Id))

		vbody := fmt.Sprintf(`<h1>%s</h1>
<h2>%s</h2>`, html.EscapeString(ec.info.Name), html.EscapeString(volume.Name))
		if volume.Description != "" {
			vbody += fmt.Sprintf("\n"+`<h3>%s</h3>`, html.EscapeString(volume.Description))
		}
		if volume.CoverId != "" {
			vbody += fmt.Sprintf("\n"+`<img src="../images/%s" alt="%s"/>`, volume.CoverId, volume.CoverId)
			err = ec.buildResById(ep, ec.lc, volume.CoverId, tmpMap)
			if err != nil {
				return err
			}
		}
		vs, err := ep.AddSection(vbody, volume.Name, fmt.Sprintf("volumes_%d.xhtml", i+1), "")
		if err != nil {
			return err
		}
		for k, chapter := range volume.Chapters {
			if !verifyChapter(ec.vcm, i, k) {
				continue
			}
			if !ec.data.Volumes[i].Chapters[k].Loaded {
				continue
			}
			cbody := fmt.Sprintf(`<h1>%s</h1>
<h2>%s</h2>
<h3>%s</h3>
%s`, html.EscapeString(ec.info.Name), html.EscapeString(volume.Name), html.EscapeString(chapter.Name), strings.Join(ec.data.Volumes[i].Chapters[k].Data, "\n"))
			_, err = ep.AddSubSection(vs, cbody, chapter.Name, fmt.Sprintf("chapter%d_%d.xhtml", i+1, k+1), "")
			if err != nil {
				return err
			}
			for _, cid := range ec.data.Volumes[i].Chapters[k].Imgs {
				err = ec.buildResById(ep, ec.lc, cid, tmpMap)
				if err != nil {
					return err
				}
			}
		}
		err = ep.Write(fp)
		if err != nil {
			return err
		}
		fh, err := utils.FileHashSha256(fp)
		if err != nil {
			return err
		}
		ec.data.Volumes[i].Hash = fh
	}
	return nil
}

func (ec *epubContext) buildBookContent() error {
	fp := path.Join(ec.output, fmt.Sprintf("%s.epub", verifyFileName(ec.info.Name)))
	if ec.data.Loaded {
		fh, _ := utils.FileHashSha256(fp)
		if len(fh) != 0 && fh == ec.data.Hash {
			return nil
		}
	}

	ep, err := epub.NewEpub(ec.info.Name)
	if err != nil {
		return err
	}

	tmpMap := make(map[string]bool)

	ep.SetAuthor(ec.info.Author)

	err = ec.buildResById(ep, ec.lc, ec.info.CoverId, tmpMap)
	if err != nil {
		return err
	}
	err = ep.SetCover(fmt.Sprintf("../images/%s", ec.info.CoverId), "")
	if err != nil {
		return err
	}

	ep.SetDescription(ec.info.Description)
	ep.SetLang(ec.lang)
	ep.SetIdentifier(fmt.Sprintf("%s_%s", ec.source, ec.id))

	for i, volume := range ec.info.Volumes {
		if !verifyVolume(ec.vcm, i) {
			continue
		}
		if !ec.data.Volumes[i].Loaded {
			continue
		}
		vbody := fmt.Sprintf(`<h1>%s</h1>
<h2>%s</h2>`, html.EscapeString(ec.info.Name), html.EscapeString(volume.Name))
		if volume.Description != "" {
			vbody += fmt.Sprintf("\n"+`<h3>%s</h3>`, html.EscapeString(volume.Description))
		}
		if volume.CoverId != "" {
			vbody += fmt.Sprintf("\n"+`<img src="../images/%s" alt="%s"/>`, volume.CoverId, volume.CoverId)
			err = ec.buildResById(ep, ec.lc, volume.CoverId, tmpMap)
			if err != nil {
				return err
			}
		}
		vs, err := ep.AddSection(vbody, volume.Name, fmt.Sprintf("volumes_%d.xhtml", i+1), "")
		if err != nil {
			return err
		}
		for k, chapter := range volume.Chapters {
			if !verifyChapter(ec.vcm, i, k) {
				continue
			}
			if !ec.data.Volumes[i].Chapters[k].Loaded {
				continue
			}
			cbody := fmt.Sprintf(`<h1>%s</h1>
<h2>%s</h2>
<h3>%s</h3>
%s`, html.EscapeString(ec.info.Name), html.EscapeString(volume.Name), html.EscapeString(chapter.Name), strings.Join(ec.data.Volumes[i].Chapters[k].Data, "\n"))
			_, err = ep.AddSubSection(vs, cbody, chapter.Name, fmt.Sprintf("chapter%d_%d.xhtml", i+1, k+1), "")
			if err != nil {
				return err
			}
			for _, cid := range ec.data.Volumes[i].Chapters[k].Imgs {
				err = ec.buildResById(ep, ec.lc, cid, tmpMap)
				if err != nil {
					return err
				}
			}
		}
	}
	err = ep.Write(fp)
	if err != nil {
		return err
	}
	fh, err := utils.FileHashSha256(fp)
	if err != nil {
		return err
	}
	ec.data.Hash = fh
	return nil
}

func verifyVolume(VC map[int]map[int]bool, v int) bool {
	if len(VC) == 0 {
		return true
	}
	if _, ok := VC[v]; ok {
		return true
	}
	return false
}

func verifyChapter(VC map[int]map[int]bool, v, c int) bool {
	if len(VC) == 0 {
		return true
	}
	if cm, ok := VC[v]; ok {
		return verifyChapter2(cm, c)
	}
	return false
}

func verifyChapter2(cm map[int]bool, c int) bool {
	if len(cm) == 0 {
		return true
	}
	if _, ok := cm[c]; ok {
		return true
	}
	return false
}

func verifyFileName(fp string) string {
	specialChars := map[rune]string{
		'\\': "_",
		'/':  "_",
		':':  "_",
		'*':  "_",
		'?':  "_",
		'"':  "_",
		'<':  "_",
		'>':  "_",
		'|':  "_",
		' ':  "_",
		'\t': "_",
		'\n': "_",
		'\r': "_",
		' ':  "_",
	}
	result := strings.Builder{}
	for _, r := range fp {
		if replacement, ok := specialChars[r]; ok {
			result.WriteString(replacement)
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}
