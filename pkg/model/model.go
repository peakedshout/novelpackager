package model

type PackageMode = int8

const (
	PackageModeDefault PackageMode = iota
	PackageModeBook
	PackageModeVolume
	PackageModeChapter

	PackageModeNone = PackageModeDefault - 1
)

type PackageConfig struct {
	KeepRecord  bool   `json:"keepRecord" Barg:"kRecord,k" Harg:"Keep downloading cache files. If this parameter is not enabled, the files will be deleted directly after downloading is completed."`
	OutputPath  string `json:"output" Barg:"output,o" Harg:"The output folder path and packaged file name do not support customization."`
	DisSyncData bool   `json:"disSyncData" Barg:"disSyncData,s" Harg:"Disable updating the index and use cached data directly."`

	PackageMode PackageMode `json:"packageMode" Barg:"pMode,p" Harg:"Packaging Mode.（0,1. Package into one file; 2. Package by volume; 3. Package by chapter; -1. Do not package）"`

	VolumeSelect []int `json:"volumeSelect" Barg:"vSelect,l" Harg:"Select the volume you want to download, select according to the index."`

	Lang string `json:"lang" Barg:"lang" Harg:"Set the language attribute of the packaged epub. (The data of the download source will not be modified)"`
}

type BookInfo struct {
	Name        string   `json:"name"`
	Id          string   `json:"id"`
	Author      string   `json:"author"`
	Cover       []byte   `json:"cover"`
	CoverId     string   `json:"coverId"`
	Description string   `json:"description"`
	Metas       []string `json:"metas"`

	Volumes []VolumeInfo `json:"volumes"`
}

type VolumeInfo struct {
	Name        string `json:"name"`
	Id          string `json:"id"`
	Cover       []byte `json:"cover"`
	CoverId     string `json:"coverId"`
	Description string `json:"description"`
	Ahref       string `json:"ahref"`

	Chapters []ChapterInfo `json:"chapters"`
}

type ChapterInfo struct {
	Name  string `json:"name"`
	Ahref string `json:"ahref"`
}

type BookData struct {
	Loaded bool   `json:"loaded,omitempty"`
	Hash   string `json:"hash,omitempty"`

	Volumes []*VolumeData `json:"volumes,omitempty"`
}

type VolumeData struct {
	Loaded bool   `json:"loaded,omitempty"`
	Hash   string `json:"hash,omitempty"`

	Name     string         `json:"name"`
	Id       string         `json:"id"`
	Chapters []*ChapterData `json:"chapters,omitempty"`
}

type ChapterData struct {
	Loaded bool   `json:"loaded,omitempty"`
	Hash   string `json:"hash,omitempty"`

	Name string   `json:"name,omitempty"`
	Data []string `json:"data,omitempty"`
	Imgs []string `json:"imgs,omitempty"`
}

type SearchResult struct {
	Name        string   `json:"name"`
	Id          string   `json:"id"`
	Author      string   `json:"author"`
	Cover       []byte   `json:"cover"`
	Description string   `json:"description"`
	Ahref       string   `json:"ahref"`
	Metas       []string `json:"metas"`
}
