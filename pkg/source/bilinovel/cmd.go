package bilinovel

import (
	"fmt"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/peakedshout/novelpackager/pkg/model"
	"github.com/peakedshout/novelpackager/pkg/rodx"
	"github.com/peakedshout/novelpackager/pkg/utils"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

const Version = `v0.1.0`

func init() {
	rootCmd.AddCommand(searchCmd)
	utils.BindKey(searchCmd, "rodx", new(rodx.RodConfig))
	utils.BindKey(searchCmd, "bcfg", new(Config))
	utils.BindKey(searchCmd, "args", new(searchArgs))

	rootCmd.AddCommand(infoCmd)
	utils.BindKey(infoCmd, "rodx", new(rodx.RodConfig))
	utils.BindKey(infoCmd, "bcfg", new(Config))
	utils.BindKey(infoCmd, "args", new(infoArgs))

	rootCmd.AddCommand(downloadCmd)
	utils.BindKey(downloadCmd, "rodx", new(rodx.RodConfig))
	utils.BindKey(downloadCmd, "bcfg", new(Config))
	utils.BindKey(downloadCmd, "args", new(model.PackageConfig))

	utils.RegisterCommand(rootCmd)
}

var rootCmd = &cobra.Command{
	Use:     "bilinovel",
	Short:   "bilinovel packager",
	Version: Version,
}

type searchArgs struct {
	Short bool `json:"short,omitempty" Barg:"short" Harg:"List short information"`
	Full  bool `json:"full,omitempty" Barg:"full" Harg:"List all search results (may be long)"`
	NoImg bool `json:"noImg,omitempty" Barg:"noImg" Harg:"Do not obtain images, which is more friendly to some retrieval environments with poor resources."`
}

var searchCmd = &cobra.Command{
	Use:   "search key",
	Short: "search by key",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		rcfg := utils.GetKeyT[rodx.RodConfig](cmd, "rodx")
		rcfg.Ctx = cmd.Context()

		pcfg := utils.GetKeyT[Config](cmd, "bcfg")
		sas := utils.GetKeyT[searchArgs](cmd, "args")

		rc, err := rodx.NewRodContext(rcfg)
		if err != nil {
			return err
		}
		defer rc.Close()
		pr := NewPackager(rc, pcfg)
		results, err := pr.Search(args[0], sas.Full, sas.NoImg)
		if err != nil {
			return err
		}

		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.SetStyle(table.StyleColoredBright)
		tableSetColColor(t, []text.Colors{
			{text.FgHiCyan},
			{text.FgHiYellow},
			{text.FgHiBlue},
			{text.FgHiMagenta},
			{text.FgHiGreen},
			{text.FgHiWhite},
		})

		if sas.Short {
			wmax := []int{5, 10, 20, 20, 20}
			t.AppendHeader(table.Row{"Index", "Id", "Name", "Author", "Metas"})
			for i, result := range results {
				tableAppendRow(t, wmax, table.Row{i + 1, result.Id, result.Name, result.Author, result.Metas})
			}
		} else {
			wmax := []int{5, 10, 20, 20, 20, 45}
			t.AppendHeader(table.Row{"Index", "Id", "Name", "Author", "Metas", "Description"})
			for i, result := range results {
				tableAppendRow(t, wmax, table.Row{i + 1, result.Id, result.Name, result.Author, result.Metas, result.Description})
			}
		}
		t.AppendFooter(table.Row{"TOTAL", len(results)}, table.RowConfig{AutoMerge: true})
		t.Render()
		return nil
	},
}

type infoArgs struct {
	Full bool `json:"full,omitempty" Barg:"full" Harg:"A complete acquisition will list some additional information."`
}

var infoCmd = &cobra.Command{
	Use:   "info id",
	Short: "get info by id",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		rcfg := utils.GetKeyT[rodx.RodConfig](cmd, "rodx")
		rcfg.Ctx = cmd.Context()

		pcfg := utils.GetKeyT[Config](cmd, "bcfg")
		rc, err := rodx.NewRodContext(rcfg)
		if err != nil {
			return err
		}
		defer rc.Close()
		pr := NewPackager(rc, pcfg)

		ias := utils.GetKeyT[infoArgs](cmd, "args")

		info, err := pr.GetInfo(args[0], ias.Full)
		if err != nil {
			return err
		}
		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.SetStyle(table.StyleColoredBright)
		tableSetColColor(t, []text.Colors{
			{text.FgHiCyan},
			{text.FgHiWhite},
		})

		t.AppendHeader(table.Row{"", ""})
		t.AppendFooter(table.Row{"", ""})

		wmax := []int{15, 105}

		tableAppendRow(t, wmax, table.Row{"Id", info.Id})
		tableAppendRow(t, wmax, table.Row{"Name", info.Name})
		tableAppendRow(t, wmax, table.Row{"Author", info.Author})
		tableAppendRow(t, wmax, table.Row{"Metas", info.Metas})
		tableAppendRow(t, wmax, table.Row{"Description", info.Description})
		tableAppendRow(t, wmax, table.Row{"Volumes", fmt.Sprintf("total: %d", len(info.Volumes))})

		for i, volume := range info.Volumes {
			cCount := ""
			if ias.Full {
				cCount = fmt.Sprintf("chapters: %d", len(volume.Chapters))
			}
			tableAppendRow(t, wmax, table.Row{cCount, fmt.Sprintf("%d.  %s", i+1, volume.Name)})

		}

		t.Render()
		return nil
	},
}

var downloadCmd = &cobra.Command{
	Use:   "download id",
	Short: "download by id",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		rcfg := utils.GetKeyT[rodx.RodConfig](cmd, "rodx")
		rcfg.Ctx = cmd.Context()

		pcfg := utils.GetKeyT[Config](cmd, "bcfg")
		rc, err := rodx.NewRodContext(rcfg)
		if err != nil {
			return err
		}
		defer rc.Close()
		pr := NewPackager(rc, pcfg)

		pas := utils.GetKeyT[model.PackageConfig](cmd, "args")

		err = pr.Download(args[0], pas)

		return err
	},
}

func tableSetColColor(t table.Writer, wcs []text.Colors) {
	cc := make([]table.ColumnConfig, 0, len(wcs))
	for i, wc := range wcs {
		cc = append(cc, table.ColumnConfig{Number: i + 1, Colors: wc})
	}
	t.SetColumnConfigs(cc)
}

func tableAppendRow(t table.Writer, wmaxs []int, row table.Row, configs ...table.RowConfig) {
	for i, data := range row {
		row[i] = textWrap(fmt.Sprint(data), wmaxs[i])
	}
	t.AppendRow(row, configs...)
}

func textWrap(str string, num int) string {
	sb := strings.Builder{}
	currentWidth := 0
	word := ""
	wordWidth := 0
	nNum := 0

	for _, r := range str {
		if r == '\n' {
			nNum++
			if nNum > 2 {
				continue
			}
			sb.WriteString(word)
			sb.WriteRune(r)
			word = ""
			wordWidth = 0
			currentWidth = 0
			continue
		} else {
			nNum = 0
		}

		charWidth := text.StringWidthWithoutEscSequences(string(r))
		if currentWidth+charWidth > num {
			sb.WriteString(word)
			sb.WriteString("\n")
			word = string(r)
			wordWidth = charWidth
			currentWidth = wordWidth
		} else {
			word += string(r)
			wordWidth += charWidth
			currentWidth += charWidth
		}
	}

	sb.WriteString(word)
	return sb.String()
}
