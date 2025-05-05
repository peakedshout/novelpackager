package main

import (
	"github.com/peakedshout/novelpackager/pkg/boot"
	"github.com/peakedshout/novelpackager/pkg/web"
	"github.com/spf13/cobra"
)

func main() {
	err := root.Execute()
	if err != nil {
		panic(err)
	}
}

var root = &cobra.Command{
	Use:     "novelpackager",
	Short:   "novel packager cli",
	Version: boot.Version,
}

func init() {
	boot.Init(root)
	web.Init(root)
}
