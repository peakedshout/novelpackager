package bilinovel

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
	"github.com/go-rod/rod/lib/proto"
	"github.com/peakedshout/go-pandorasbox/logger"
	"github.com/peakedshout/novelpackager/pkg/model"
	"github.com/peakedshout/novelpackager/pkg/rodx"
	"github.com/peakedshout/novelpackager/pkg/utils"
	"html"
	"net/url"
	"path"
	"regexp"
	"strings"
	"time"
)

const Source = `bilinovel`

var (
	UrlRoot    = `https://www.bilinovel.com`
	UrlInfo    = `/novel/%s.html`
	UrlCatalog = `/novel/%s/catalog`
	UrlSearch  = `/search.html`

	UrlInfoPre = `/novel/`

	UrlSearchRegexp = regexp.MustCompile(`/novel/.*\.html`)

	CacheFile = `bn_%s.np`
)

type Config struct {
	Timeout  int `json:"timeout" Barg:"timeout,t" Harg:"Automation timeout.(s)"`
	RetryNum int `json:"retryNum" Barg:"retryNum,n" Harg:"Number of automated retries."`
}

type Packager struct {
	rc *rodx.RodContext

	timeout  time.Duration
	retryNum uint

	logger logger.Logger
}

func NewPackager(ctx *rodx.RodContext, cfg *Config) *Packager {
	l := logger.GetLogger(ctx.Context())
	p := &Packager{
		rc:       ctx,
		timeout:  30 * time.Second,
		retryNum: 3,
		logger:   l.Clone("bilinovel"),
	}
	if cfg.Timeout > 0 {
		p.timeout = time.Duration(cfg.Timeout) * time.Second
	}
	if cfg.RetryNum > 0 {
		p.retryNum = uint(cfg.RetryNum)
	} else if cfg.RetryNum < 0 {
		p.retryNum = 0
	}
	return p
}

func (p *Packager) GetInfo(id string, full bool) (*model.BookInfo, error) {
	sess, err := p.rc.NewSession()
	if err != nil {
		return nil, err
	}
	defer sess.Close()
	defer blockURLs(sess.Browser())()
	info, err := p.getBookInfo(sess, id)
	if err != nil {
		return nil, err
	}
	if !full {
		return info, nil
	}
	err = p.getBookInfoFull(sess, info)
	if err != nil {
		return nil, err
	}
	return info, nil
}

func (p *Packager) Search(name string, full bool, noImg bool) ([]model.SearchResult, error) {
	sess, err := p.rc.NewSession()
	if err != nil {
		return nil, err
	}
	defer sess.Close()
	defer blockURLs(sess.Browser())()
	return p.searchList(sess, name, full, noImg)
}

func (p *Packager) Download(id string, pcfg *model.PackageConfig) error {
	sess, err := p.rc.NewSession()
	if err != nil {
		return err
	}
	defer sess.Close()
	defer blockURLs(sess.Browser())()
	return p.download(sess, id, pcfg)
}

func (p *Packager) getBookInfo(sess *rodx.RodSession, id string) (*model.BookInfo, error) {
	turl := fmt.Sprintf(UrlRoot+UrlInfo, id)
	var info *model.BookInfo

	err := sess.PageLoop().DoWithNum(func(page *rod.Page) error {
		err := page.Navigate(turl)
		if err != nil {
			p.logger.Warnf("Failed to navigate to URL %s: %v", turl, err)
			return model.ErrPage.Errorf(turl, err)
		}
		p.printNav(page)

		defer utils.ExpireClose(page, p.timeout)()

		err = p.waitAndCheck404(page, turl)
		if err != nil {
			return err
		}

		bookInfo := &model.BookInfo{}

		bookInfo.Name, err = utils.Element("#bookDetailWrapper > div > div.book-layout > div.book-cell > h1").Text(page)
		if err != nil {
			p.logger.Warnf("Failed to get book name for URL %s: %v", turl, err)
			return err
		}

		bookInfo.Author, err = utils.Element("#bookDetailWrapper > div > div.book-layout > div.book-cell > div").Text(page)
		if err != nil {
			p.logger.Warnf("Failed to get book author for URL %s: %v", turl, err)
			return err
		}

		cE, err := utils.Element("#bookDetailWrapper > div > div.book-layout > div.module-book-cover > div > img").Element(page)
		if err != nil {
			p.logger.Warnf("Failed to get book cover for URL %s: %v", turl, err)
			return err
		}
		bs, src, err := waitImgDataSrc(cE)
		if err != nil {
			p.logger.Warnf("Failed to find cover element for URL %s: %v", turl, err)
			return err
		}

		bookInfo.Cover = bs
		bookInfo.CoverId = path.Ext(src)

		bookInfo.Description, err = utils.Element("#bookSummary > content").Text(page)
		if err != nil {
			p.logger.Warnf("Failed to get book description for URL %s: %v", turl, err)
			return err
		}

		mEs, err := page.Elements(`#bookDetailWrapper > div > div.book-layout > div.book-cell > p:nth-child(5) > span > *`)
		if err != nil {
			p.logger.Warnf("Failed to find meta element for URL %s: %v", turl, err)
			return err
		}
		for _, mE := range mEs {
			bookInfo.Metas = append(bookInfo.Metas, mE.MustText())
		}

		volumeInfos, err := p.getCatalog(page, id)
		if err != nil {
			p.logger.Warnf("Failed to get volume catalog for URL %s: %v", turl, err)
			return err
		}
		bookInfo.Volumes = volumeInfos

		bookInfo.Id = id
		info = bookInfo
		return nil
	}, p.retryNum)
	if err != nil {
		return nil, err
	}
	p.logger.Info("Successfully fetched book info for URL:", turl, info.Name)
	return info, nil
}

func (p *Packager) getBookInfoFull(sess *rodx.RodSession, info *model.BookInfo) error {
	for i := range info.Volumes {
		time.Sleep(1 * time.Second)
		err := p.getVolumeInfo(sess, info, i)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Packager) getCatalog(page *rod.Page, id string) (list []model.VolumeInfo, err error) {
	turl := fmt.Sprintf(UrlRoot+UrlCatalog, id)
	err = page.Navigate(turl)
	if err != nil {
		p.logger.Warnf("Failed to navigate to URL %s: %v", turl, err)
		return nil, model.ErrPage.Errorf(turl, err)
	}
	err = p.waitAndCheck404(page, turl)
	if err != nil {
		return nil, err
	}

	ve, err := page.Element(`#volumes`)
	if err != nil {
		p.logger.Warnf("Failed to find volume elements for URL %s: %v", turl, err)
		return nil, err
	}
	err = ve.WaitLoad()
	if err != nil {
		p.logger.Warnf("Failed to load volume elements for URL %s: %v", turl, err)
		return nil, err
	}

	vel, err := ve.Elements("*")
	if err != nil {
		p.logger.Warnf("Failed to find volume elements for URL %s: %v", turl, err)
		return nil, err
	}
	for _, div := range vel {
		class, _ := div.Attribute("class")
		if class == nil || *class != "catalog-volume" {
			continue
		}
		vnE, err := div.Element(`ul > li.chapter-bar.chapter-li`)
		if err != nil {
			p.logger.Warnf("Failed to find volume elements for URL %s: %v", turl, err)
			return nil, err
		}
		aE, err := div.Element(`ul > li.volume-cover.chapter-li > a`)
		if err != nil {
			p.logger.Warnf("Failed to find volume elements for URL %s: %v", turl, err)
			return nil, err
		}
		ahref, err := aE.Attribute("href")
		if err != nil {
			p.logger.Warnf("Failed to find volume elements for URL %s: %v", turl, err)
			return nil, err
		}
		if ahref == nil || *ahref == "" {
			p.logger.Warnf("Ahref is empty for volume")
			return nil, errors.New("ahref is empty")
		}
		vid, _ := strings.CutSuffix(path.Base(*ahref), ".html")
		vinfo := model.VolumeInfo{
			Name:  vnE.MustText(),
			Id:    vid,
			Ahref: *ahref,
		}
		list = append(list, vinfo)
	}
	return list, nil
}

func (p *Packager) getVolumeInfo(sess *rodx.RodSession, info *model.BookInfo, index int) error {
	volume := &info.Volumes[index]
	err := sess.PageLoop().DoWithNum(func(page *rod.Page) error {
		if volume.Ahref == "" {
			p.logger.Warnf("Ahref is empty for chapter: %s", volume.Name)
			return errors.New("ahref is empty")
		}
		turl, _ := url.JoinPath(UrlRoot, volume.Ahref)
		err := page.Navigate(turl)
		if err != nil {
			p.logger.Warnf("Failed to navigate to URL %s: %v", turl, err)
			return model.ErrPage.Errorf(turl, err)
		}
		defer utils.ExpireClose(page, p.timeout)()

		err = p.waitAndCheck404(page, turl)
		if err != nil {
			return err
		}

		cE, err := page.Element(`#bookDetailWrapper > div > div.book-layout > div.module-book-cover > div > img`)
		if err != nil {
			p.logger.Warnf("Failed to find cover element for URL %s: %v", turl, err)
			return err
		}
		bs, src, err := waitImgDataSrc(cE)
		if err != nil {
			p.logger.Warnf("Failed to find cover element for URL %s: %v", turl, err)
			return err
		}
		volume.Cover = bs
		volume.CoverId = path.Ext(src)

		els, err := page.Elements("#scroll > div.page.page-book-detail > div > div:nth-child(6) > div.catalog-volume > ul > *")
		if err != nil {
			p.logger.Warnf("Failed to find chapter elements for URL %s: %v", turl, err)
			return err
		}

		for _, cinfo := range els {
			ax, err := cinfo.ElementX("a")
			if err != nil {
				p.logger.Warnf("Failed to find chapter element for URL %s: %v", turl, err)
				return err
			}
			ahref, err := ax.Attribute("href")
			if err != nil {
				p.logger.Warnf("Failed to find chapter element for URL %s: %v", turl, err)
				return err
			}
			sx, err := ax.ElementX("span")
			if err != nil {
				p.logger.Warnf("Failed to find chapter element for URL %s: %v", turl, err)
				return err
			}
			volume.Chapters = append(volume.Chapters, model.ChapterInfo{
				Ahref: *ahref,
				Name:  sx.MustText(),
			})
		}

		volume.Description, err = utils.Element("#bookSummary > content").Text(page)
		if err != nil {
			p.logger.Warnf("Failed to get volume description for URL %s: %v", turl, err)
			return err
		}
		return nil
	}, p.retryNum)

	if err != nil {
		return err
	}
	p.logger.Info("Successfully fetched book info for URL:", volume.Ahref, info.Name, volume.Name, "Chapters", len(volume.Chapters))
	return nil
}

func (p *Packager) refreshToken(page *rod.Page, volume *model.VolumeInfo) error {
	if volume.Ahref == "" {
		p.logger.Warnf("Ahref is empty for chapter: %s", volume.Name)
		return errors.New("ahref is empty")
	}
	turl, _ := url.JoinPath(UrlRoot, volume.Ahref)
	err := page.Navigate(turl)
	if err != nil {
		p.logger.Warnf("Failed to navigate to URL %s: %v", turl, err)
		return model.ErrPage.Errorf(turl, err)
	}

	err = p.waitAndCheck404(page, turl)
	if err != nil {
		return err
	}
	return nil
}

func (p *Packager) checkoutChapter(page *rod.Page, info *model.ChapterInfo, data *model.ChapterData, lc *utils.LinkCache) error {
	turl := UrlRoot + info.Ahref
	err := page.Navigate(turl)
	if err != nil {
		p.logger.Warnf("Failed to create page for URL %s: %v", turl, err)
		return model.ErrPage.Errorf(turl, err)
	}

	*data = model.ChapterData{}

	hash := sha256.New()

	count := 0
	for {
		err = p.waitAndCheck404(page, turl)
		if err != nil {
			return err
		}

		//p.logger.Info("Fetching checkout info from URL:", turl, "wait load over")

		result, err := page.Eval(`() => {
        for (let sheet of document.styleSheets) {
            try {
                for (let rule of sheet.cssRules) {
                   	if (rule.cssText.includes('p:last-of-type')) {
                        return true;
                    }
                }
            } catch (e) {
                console.log('Cannot read stylesheet', sheet.href, e);
            }
        }
        return false;
    }`)
		if err != nil {
			return err
		}

		pFont := result.Value.Bool()

		ael, err := page.Element("#acontent")
		if err != nil {
			p.logger.Warnf("Failed to find acontent element for URL %s: %v", turl, err)
			return err
		}

		_, _ = ael.Eval(`() => this.removeAttribute('style')`)

		el, err := page.Elements("#acontent > *")
		if err != nil {
			p.logger.Warnf("Failed to find elements for URL %s: %v", turl, err)
			return err
		}

		var lastP *rod.Element
		var countP int

		for _, element := range el {
			switch utils.ElementType(element) {
			case "p":
				count++
				countP = count
				lastP = element

				data.Data = append(data.Data, fmt.Sprintf("<p>%s</p>", html.EscapeString(element.MustText())))

			case "br":
				count++
				data.Data = append(data.Data, "<br/>")
			case "img":
				count++

				utils.UpdateExpireClose(page, p.timeout)

				bs, src, err := waitImgDataSrc(element)
				if err != nil {
					p.logger.Warnf("Failed to find img element for URL %s: %v", turl, err)
					return err
				}

				id, err := lc.Set(src, bs)
				if err != nil {
					p.logger.Warnf("Failed to set resource in cache for URL %s: %v", turl, err)
					return err
				}
				data.Data = append(data.Data, fmt.Sprintf(`<img src="../images/%s" alt="%s"/>`, id, id))
				data.Imgs = append(data.Imgs, id)
			default:
				p.logger.Debug("Unknown element type:", utils.ElementType(element), "for URL:", turl)
				continue
			}
		}

		if lastP != nil && pFont {
			data.Data[countP-1] = fmt.Sprintf(`<p>%s</p>`, decryptionFont(html.EscapeString(lastP.MustText())))
		}

		nextE, err := page.Element("#footlink > a:nth-child(4)")
		if err != nil {
			p.logger.Warnf("Failed to find next element for URL %s: %v", turl, err)
			return err
		}
		if nextE.MustText() == "下一頁" {
			utils.UpdateExpireClose(page, p.timeout)
			time.Sleep(1 * time.Second)
			nextE.MustClick()
		} else {
			break
		}

	}
	p.logger.Info("Successfully fetched chapter for URL:", turl, info.Name)

	for _, datum := range data.Data {
		hash.Write([]byte(datum))
	}

	data.Hash = hex.EncodeToString(hash.Sum(nil))
	data.Loaded = true

	return nil
}

func (p *Packager) searchList(sess *rodx.RodSession, name string, full bool, noImg bool) ([]model.SearchResult, error) {
	turl := ""
	var list []model.SearchResult
	err := sess.PageLoop().DoWithNum(func(page *rod.Page) error {
		turl = UrlRoot
		// home page
		err := page.Navigate(turl)
		if err != nil {
			p.logger.Warnf("Failed to navigate to URL %s: %v", turl, err)
			return model.ErrPage.Errorf(turl, err)
		}

		defer utils.ExpireClose(page, p.timeout)()

		err = page.WaitLoad()
		if err != nil {
			p.logger.Warnf("Failed to load page for URL %s: %v", turl, err)
			return model.ErrPage.Errorf(turl, err)
		}
		searchE, err := page.Element(`body > div.page.page-home > div > div.white-content > a`)
		if err != nil {
			p.logger.Warnf("Failed to find search element for URL %s: %v", turl, err)
			return model.ErrPage.Errorf(turl, err)
		}
		searchE.MustClick()

		// search page
		turl = UrlRoot + UrlSearch
		err = page.WaitLoad()
		if err != nil {
			p.logger.Warnf("Failed to load page for URL %s: %v", turl, err)
			return model.ErrPage.Errorf(turl, err)
		}
		inputE, err := page.Element(`#searchkey`)
		if err != nil {
			p.logger.Warnf("Failed to find input element for URL %s: %v", turl, err)
			return err
		}
		err = inputE.Input(name)
		if err != nil {
			p.logger.Warnf("Failed to input element for URL %s: %v", turl, err)
			return err
		}
		inputE.MustKeyActions().Press(input.Enter).MustDo()

		utils.UpdateExpireClose(page, p.timeout)

		for {
			err = p.waitAndCheck404(page, turl)
			if err != nil {
				return err
			}

			// wait one page
			err = page.WaitStable(3 * time.Second)
			if err != nil {
				p.logger.Warnf("Failed to wait stable for URL %s: %v", turl, err)
				return err
			}

			info, err := page.Info()
			if err != nil {
				p.logger.Warnf("Failed to get info for URL %s: %v", turl, err)
				return err
			}
			if strings.HasPrefix(info.URL, UrlRoot+UrlInfoPre) {
				result, err := p.searchOne(page, info.URL)
				if err != nil {
					p.logger.Warnf("Failed to search one for URL %s: %v", turl, err)
					return err
				}
				list = append(list, *result)
				return nil
			}

			el, err := page.Element(`body > div.page.page-finish > div > div.module > ol`)
			if err != nil {
				p.logger.Warnf("Failed to find element for URL %s: %v", turl, err)
				return err
			}
			err = el.WaitLoad()
			if err != nil {
				p.logger.Warnf("Failed to load element for URL %s: %v", turl, err)
				return err
			}

			els, err := el.Elements("*")
			if err != nil {
				p.logger.Warnf("Failed to find elements for URL %s: %v", turl, err)
				return err
			}

			for _, element := range els {
				if utils.ElementType(element) != "li" {
					continue
				}
				dataE, err := element.ElementX("a")
				if err != nil {
					p.logger.Warnf("Failed to find data element for URL %s: %v", turl, err)
					return err
				}
				ahref, err := dataE.Attribute("href")
				if err != nil {
					p.logger.Warnf("Failed to get href attribute for URL %s: %v", turl, err)
					return err
				}
				if ahref == nil {
					p.logger.Warn("Ahref is nil for URL:", turl)
					continue
				}
				var result model.SearchResult
				result.Ahref = *ahref
				after, _ := strings.CutPrefix(result.Ahref, `/novel/`)
				before, _ := strings.CutSuffix(after, `.html`)
				result.Id = before

				nameE, err := dataE.Element(`div.book-cell > div.book-title-x > h4`)
				if err != nil {
					p.logger.Warnf("Failed to find aEl element for URL %s: %v", turl, err)
					return err
				}
				result.Name = nameE.MustText()

				aE, err := dataE.Element(`div.book-cell > div.book-meta > div.book-meta-l > span`)
				if err != nil {
					p.logger.Warnf("Failed to find aEl element for URL %s: %v", turl, err)
					return err
				}
				result.Author = aE.MustText()

				dE, err := dataE.Element(`div.book-cell > p`)
				if err != nil {
					p.logger.Warnf("Failed to find aEl element for URL %s: %v", turl, err)
					return err
				}
				result.Description = dE.MustText()

				metaEs, err := dataE.Elements(`div.book-cell > div.book-meta > div.book-meta-r > span > *`)
				if err != nil {
					p.logger.Warnf("Failed to find aEl element for URL %s: %v", turl, err)
					return err
				}
				var str string
				for i, metaE := range metaEs {
					str += strings.TrimSpace(metaE.MustText())
					if i < len(metaEs)-1 {
						str += " "
					}
				}
				result.Metas = strings.Split(str, " ")

				if !noImg {
					cE, err := dataE.Element(`div.book-cover > img`)
					if err != nil {
						p.logger.Warnf("Failed to find aEl element for URL %s: %v", turl, err)
						return err
					}
					bs, _, err := waitImgDataSrc(cE)
					if err != nil {
						p.logger.Warnf("Failed to find volume element for URL %s: %v", turl, err)
						return err
					}
					result.Cover = bs
				}

				utils.UpdateExpireClose(page, p.timeout)

				list = append(list, result)
			}

			if !full {
				break
			}

			// next
			nextE, err := page.Element("#pagelink > a.next")
			if err != nil {
				p.logger.Warnf("Failed to find next element for URL %s: %v", turl, err)
				return err
			}
			attr, err := nextE.Attribute("href")
			if err != nil {
				return err
			}

			if attr != nil && *attr != "#" {
				p.logger.Info("search next ...", *attr)
				utils.UpdateExpireClose(page, p.timeout)
				nextE.MustClick()
			} else {
				break
			}
		}
		return nil
	}, p.retryNum)
	if err != nil {
		return nil, err
	}
	p.logger.Info("Successfully fetched search list for URL:", turl)
	return list, nil
}

func (p *Packager) searchOne(page *rod.Page, turl string) (*model.SearchResult, error) {
	var sr model.SearchResult
	var err error
	sr.Name, err = utils.Element("#bookDetailWrapper > div > div.book-layout > div.book-cell > h1").Text(page)
	if err != nil {
		p.logger.Warnf("Failed to get book name for URL %s: %v", turl, err)
		return nil, err
	}

	sr.Author, err = utils.Element("#bookDetailWrapper > div > div.book-layout > div.book-cell > div").Text(page)
	if err != nil {
		p.logger.Warnf("Failed to get book author for URL %s: %v", turl, err)
		return nil, err
	}

	cE := utils.Element("#bookDetailWrapper > div > div.book-layout > div.module-book-cover > div > img")

	ct, err := cE.Attribute(page, "src")
	if err != nil {
		p.logger.Warnf("Failed to get cover attribute for URL %s: %v", turl, err)
		return nil, err
	}
	if ct == nil {
		p.logger.Warnf("Cover type is empty for URL: %s", turl)
		return nil, errors.New("cover type is empty")
	}

	sr.Cover, err = cE.Resource(page)
	if err != nil {
		p.logger.Warnf("Failed to get cover resource for URL %s: %v", turl, err)
		return nil, err
	}

	sr.Description, err = utils.Element("#bookSummary > content").Text(page)
	if err != nil {
		p.logger.Warnf("Failed to get book description for URL %s: %v", turl, err)
		return nil, err
	}

	sr.Ahref, _ = strings.CutPrefix(turl, UrlRoot)
	after, _ := strings.CutPrefix(sr.Ahref, `/novel/`)
	before, _ := strings.CutSuffix(after, `.html`)
	sr.Id = before

	mEs, err := page.Elements(`#bookDetailWrapper > div > div.book-layout > div.book-cell > p:nth-child(5) > span > *`)
	if err != nil {
		p.logger.Warnf("Failed to find meta element for URL %s: %v", turl, err)
		return nil, err
	}
	for _, mE := range mEs {
		sr.Metas = append(sr.Metas, mE.MustText())
	}
	return &sr, nil
}

func (p *Packager) printNav(page *rod.Page) {
	info := page.MustEval(`() => ({
	   userAgent: navigator.userAgent,
	   platform: navigator.platform,
	   language: navigator.language,
	   languages: navigator.languages,
	   webdriver: navigator.webdriver,
	   deviceMemory: navigator.deviceMemory,
	   hardwareConcurrency: navigator.hardwareConcurrency,
	   vendor: navigator.vendor,
	})`)
	p.logger.Info("Navigator:", info)
}

func (p *Packager) waitAndCheck404(page *rod.Page, turl string) error {
	err := page.WaitLoad()
	if err != nil {
		p.logger.Warnf("Failed to load page for URL %s: %v", turl, err)
		return model.ErrPage.Errorf(turl, err)
	}

	err = p.reloadBlock(page)
	if err != nil {
		p.logger.Warnf("Failed to reload page for URL %s: %v", turl, err)
		return model.ErrPage.Errorf(turl, err)
	}

	hasX, hasEl, err := page.Has(`body > div > div > div.c1 > a > img`)
	if err != nil {
		p.logger.Warnf("Failed to check hasX for URL %s: %v", turl, err)
		return model.ErrPage.Errorf(turl, err)
	}
	if hasX {
		src, err := hasEl.Attribute("src")
		if err != nil {
			p.logger.Warnf("Failed to get src attribute for URL %s: %v", turl, err)
			return err
		}
		if src != nil && *src == "/404.png" {
			p.logger.Warnf("Failed to load page for URL %s: %v", turl, 404)
			return model.ErrPage.Errorf(turl, err)
		}
	}

	_, err = page.Eval(`() => {
		var imgs = document.querySelectorAll('img[data-src]');
		imgs.forEach(function(img) {
	    if (img.dataset.src) {
	        img.src = img.dataset.src;
	        img.removeAttribute('data-src');
	    }
	});
	}`)

	if err != nil {
		p.logger.Warnf("Failed to checkout data-src for URL %s: %v", turl, err)
		return err
	}

	return nil
}

func (p *Packager) reloadBlock(page *rod.Page) error {
	for {
		has, _, err := page.Has(`#cookie-alert`)
		if err != nil {
			return err
		}
		if has {
			time.Sleep(3 * time.Second)
			err = page.Reload()
			if err != nil {
				return err
			}
			err = page.WaitLoad()
			if err != nil {
				return err
			}
		} else {
			return nil
		}
	}
}

func blockURLs(b *rod.Browser) func() {
	ls := []string{
		`https://hm.baidu.com/`,
		`https://www.googletagmanager.com`,
		`https://pagead2.googlesyndication.com/`,
		`https://www.bilinovel.com/images/`,
		`https://www.bilinovel.com/public/`,
		`https://www.bilinovel.com/favicon.ico`,
		`https://www.bilinovel.com/cdn-cgi/`,
		`https://www.bilinovel.com/files/article/image/banner/`,
	}
	router := b.HijackRequests()
	router.MustAdd("*", func(hijack *rod.Hijack) {
		u := hijack.Request.URL()
		for _, str := range ls {
			if strings.HasPrefix(u.String(), str) {
				hijack.Response.Fail(proto.NetworkErrorReasonAborted)
				//fmt.Println("Blocked URL:", u.String())
				return
			}
		}
		//fmt.Println("Continue URL:", u.String())
		hijack.ContinueRequest(&proto.FetchContinueRequest{})
	})
	go router.Run()
	return func() {
		_ = router.Stop()
	}
}

func waitImgDataSrc(element *rod.Element) ([]byte, string, error) {
	err := element.Focus()
	if err != nil {
		return nil, "", err
	}
	src, err := element.Attribute("src")
	if err != nil {
		return nil, "", err
	}
	ds, err := element.Attribute("data-src")
	if err != nil {
		return nil, "", err
	}
	if ds != nil {
		for *src != *ds {
			time.Sleep(1 * time.Second)
			src, err = element.Attribute("src")
			if err != nil {
				return nil, "", err
			}
		}
	}
	bs, err := element.Resource()
	if err != nil {
		return nil, "", err
	}
	return bs, *src, nil
}

func getCachePath(r *utils.Record) string {
	return path.Join(r.Config.OutputPath, fmt.Sprintf(CacheFile, r.Info.Id))
}
