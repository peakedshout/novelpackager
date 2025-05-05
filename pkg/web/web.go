package web

import (
	"context"
	"crypto/tls"
	"embed"
	"github.com/peakedshout/go-pandorasbox/pcrypto"
	"github.com/peakedshout/go-pandorasbox/xnet/xtool/xhttp"
	"github.com/peakedshout/novelpackager/pkg/rodx"
	"github.com/peakedshout/novelpackager/pkg/utils"
	"github.com/spf13/cobra"
	"net"
	"path"
)

//go:embed frontend/dist
var embedFs embed.FS

func Serve(ctx context.Context, cfg *webConfig) error {
	var tcfg *tls.Config
	if cfg.Tls {
		tc, err := pcrypto.MakeTlsConfigFromFile(cfg.CertFile, cfg.KeyFile)
		if err != nil {
			return err
		}
		tcfg = tc
		tcfg.InsecureSkipVerify = cfg.Insecure
	}
	if cfg.Network == "" {
		cfg.Network = "tcp"
	}
	if cfg.Address == "" {
		cfg.Address = ":8080"
	}
	var auth func(u string, p string) bool
	if cfg.Username != "" && cfg.Password != "" {
		auth = func(u, p string) bool {
			if u != cfg.Username || p != cfg.Password {
				return false
			}
			return true
		}
	}

	sr := newServer(&xhttp.Config{
		Ctx:    ctx,
		Type:   xhttp.TypeNone,
		TlsCfg: tcfg,
		Auth:   auth,
		Prefix: "",
	})

	ln, err := net.Listen(cfg.Network, cfg.Address)
	if err != nil {
		return err
	}
	defer ln.Close()
	return sr.Serve(ln)
}

type webConfig struct {
	CacheDir   string `json:"cacheDir" Barg:"cacheDir" Harg:"cache dir"`
	MaxCacheBs int64  `json:"maxCacheBs" Barg:"maxCacheBs" Harg:"kv cache max size"`

	Network  string `Barg:"web.nk" Harg:"cmd network"`
	Address  string `Barg:"web.addr" Harg:"cmd address"`
	Username string `Barg:"web.u" Harg:"cmd username" Garg:"up"`
	Password string `Barg:"web.p" Harg:"cmd password" Garg:"up"`
	Tls      bool   `Barg:"web.tls" Harg:"cmd tls enable"`
	Insecure bool   `Barg:"web.i" Harg:"cmd tls insecure"`

	CertFile string `Barg:"web.cert" Harg:"cmd tls cert file" Garg:"ck"`
	KeyFile  string `Barg:"web.key" Harg:"cmd tls key file" Garg:"ck"`
}

var rootCmd = &cobra.Command{
	Use:   "web",
	Short: "web",
	RunE: func(cmd *cobra.Command, args []string) error {
		rcfg := utils.GetKeyT[rodx.RodConfig](cmd, "rodx")
		rcfg.Ctx = cmd.Context()
		rc, err := rodx.NewRodContext(rcfg)
		if err != nil {
			return err
		}
		defer rc.Close()

		cfg := utils.GetKeyT[webConfig](cmd, "cfg")
		kvCache, err := utils.NewKVCache(path.Join(cfg.CacheDir, ".web.KVCache"), cfg.MaxCacheBs)
		if err != nil {
			return err
		}

		buildSource(&BuildContext{
			Ctx:        cmd.Context(),
			RodContext: rc,
			Cache:      kvCache,
			CacheDir:   cfg.CacheDir,
		})

		return Serve(cmd.Context(), cfg)
	},
}

func Init(c *cobra.Command) {
	c.AddCommand(rootCmd)
	utils.BindKey(rootCmd, "rodx", new(rodx.RodConfig))
	utils.BindKey(rootCmd, "cfg", &webConfig{CacheDir: "./.np_cache", MaxCacheBs: 10 * 1024 * 1024})
}
