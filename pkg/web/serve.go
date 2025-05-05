package web

import (
	"fmt"
	"github.com/peakedshout/go-pandorasbox/tool/hjson"
	"github.com/peakedshout/go-pandorasbox/xnet/xtool/xhttp"
	"io/fs"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
)

type server struct {
	*xhttp.Server
}

func newServer(cfg *xhttp.Config) *server {
	sr := xhttp.NewServer(cfg)
	sub, err := fs.Sub(embedFs, "frontend/dist")
	if err != nil {
		panic(err)
	}
	s := &server{Server: sr}

	sh := http.FileServerFS(sub)
	s.Set("", func(context *xhttp.Context) error {
		sh.ServeHTTP(context.Raw())
		return nil
	})
	s.Set("/api/source_list", func(context *xhttp.Context) error {
		sl := make([]string, 0, len(sourceMap))
		for k := range sourceMap {
			sl = append(sl, k)
		}
		slices.Sort(sl)
		_, err = context.Write(hjson.MustMarshal(sl))
		return err
	})
	sr.Set("/api/get_info", s.getInfo)
	sr.Set("/api/search", s.search)
	sr.Set("/api/progress", s.progress)
	sr.Set("/api/caching", s.caching)
	sr.Set("/api/enable_download", s.enableDownload)
	sr.Set("/api/download", s.download)
	return s
}

func (sr *server) getInfo(context *xhttp.Context) error {
	source := context.Query().Get("source")
	s, err := getSource(source)
	if err != nil {
		return context.WriteAny(NewError(err))
	}

	id := context.Query().Get("id")
	full, err := strconv.ParseBool(context.Query().Get("full"))
	if err != nil {
		return context.WriteAny(NewError(err))
	}
	return context.WriteAny(NewMsg(s.GetInfo(context, id, full)))
}

func (sr *server) search(context *xhttp.Context) error {
	source := context.Query().Get("source")
	s, err := getSource(source)
	if err != nil {
		return context.WriteAny(NewError(err))
	}

	name := context.Query().Get("name")
	full, err := strconv.ParseBool(context.Query().Get("full"))
	if err != nil {
		return context.WriteAny(NewError(err))
	}

	return context.WriteAny(NewMsg(s.Search(context, name, full, false)))
}

func (sr *server) progress(context *xhttp.Context) error {
	source := context.Query().Get("source")
	s, err := getSource(source)
	if err != nil {
		return context.WriteAny(NewError(err))
	}

	return context.WriteAny(NewMsg(s.Progress(context)))
}

func (sr *server) caching(context *xhttp.Context) error {
	source := context.Query().Get("source")
	s, err := getSource(source)
	if err != nil {
		return context.WriteAny(NewError(err))
	}

	id := context.Query().Get("id")
	return context.WriteAny(NewError(s.Caching(context, id)))
}

func (sr *server) enableDownload(context *xhttp.Context) error {
	source := context.Query().Get("source")
	s, err := getSource(source)
	if err != nil {
		return context.WriteAny(NewError(err))
	}

	id := context.Query().Get("id")
	return context.WriteAny(NewMsg(s.EnableDownload(context, id)))
}

func (sr *server) download(context *xhttp.Context) error {
	source := context.Query().Get("source")
	s, err := getSource(source)
	if err != nil {
		return err
	}

	id := context.Query().Get("id")

	volsStr := context.Query().Get("vols")
	vols, err := parseVols(volsStr)
	if err != nil {
		return err
	}

	fd, err := s.Download(context, id, vols...)
	if err != nil {
		return err
	}

	var filename string
	if len(vols) <= 1 {
		volsStr = ""
		filename = fd.Name
	} else {
		volsStr = fmt.Sprintf("[%s]", volsStr)
		fd.Name, _ = strings.CutSuffix(fd.Name, ".epub")
		filename = fmt.Sprintf("%s%s.epub", fd.Name, volsStr)
	}
	encodedFilename := url.PathEscape(filename)
	context.WHeader().Set("Content-Type", "application/epub+zip")
	context.WHeader().Set("Content-Disposition",
		fmt.Sprintf(`attachment; filename="%s"; filename*=UTF-8''%s`, filename, encodedFilename))
	context.WHeader().Set("Content-Length", strconv.Itoa(len(fd.Data)))
	_, err = context.Write(fd.Data)
	return err
}

func parseVols(volsStr string) ([]int, error) {
	if volsStr == "" {
		return []int{}, nil
	}
	parts := strings.Split(volsStr, ",")
	vols := make([]int, 0, len(parts))

	for _, part := range parts {
		vol, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("invalid number: %s", part)
		}
		vols = append(vols, vol)
	}

	return vols, nil
}
