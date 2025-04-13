package utils

import (
	"encoding/gob"
	"github.com/peakedshout/novelpackager/pkg/model"
	"os"
)

type Record struct {
	Config *model.PackageConfig    `json:"config"`
	Info   *model.BookInfo         `json:"info"`
	Data   *model.BookData         `json:"data"`
	Cache  map[string]*ExportCache `json:"cache"`
}

func SaveRecord(p string, r *Record, lc *LinkCache) error {
	if lc != nil {
		r.Cache = lc.Export()
	}
	file, err := os.Create(p)
	if err != nil {
		return err
	}
	defer file.Close()

	return gob.NewEncoder(file).Encode(r)
}

func LoadRecord(p string) (*Record, error) {
	file, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	r := new(Record)
	err = gob.NewDecoder(file).Decode(r)
	if err != nil {
		return nil, err
	}
	return r, nil
}
