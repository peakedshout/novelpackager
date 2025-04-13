package rodx

import (
	"context"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/defaults"
	"github.com/go-rod/rod/lib/devices"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/utils"
	"github.com/google/uuid"
	"github.com/peakedshout/go-pandorasbox/logger"
	"github.com/peakedshout/go-pandorasbox/tool/xerror"
	"os"
	"path"
	"sync"
	"time"
)

const (
	cachePath    = `.np_cache`
	binPath      = `bin`
	userDataPath = `userData`
)

type RodConfig struct {
	Ctx     context.Context `json:"-"`
	BinDir  string          `json:"binDir" Barg:"binDir,b" Harg:"Rod bin dir (First use the browser scanned locally, followed by the browser in the directory, if there is no browser in the directory to download the browser)"`
	UserDir string          `json:"userDir" Barg:"userDir,u" Harg:"The user-data-dir used by the browser will be automatically deleted when the program exits normally."`
	Delay   uint            `json:"delay" Barg:"delay,d" Harg:"Set the delay for each control action, such as the simulation of the human inputs.（ms）"` //ms
	View    bool            `json:"view" Barg:"view" Harg:"The debug view is used to show what the automated process does."`
}

type RodContext struct {
	ctx    context.Context
	cl     context.CancelCauseFunc
	closer sync.Once

	binPath     string
	userDataDir string
	delay       time.Duration
	view        bool

	sessionMap map[string]*RodSession
	sessionMux sync.Mutex
	sessionWg  sync.WaitGroup

	logger logger.Logger
}

func NewRodContext(cfg *RodConfig) (*RodContext, error) {
	rc := RodContext{
		ctx:        context.Background(),
		logger:     logger.Init("rodContext"),
		sessionMap: make(map[string]*RodSession),
		view:       cfg.View,
	}
	if cfg.Ctx != nil {
		rc.ctx = cfg.Ctx
	}
	if cfg.UserDir == "" {
		rc.userDataDir = path.Join("./", cachePath, userDataPath)
	} else {
		rc.userDataDir = cfg.UserDir
	}
	if cfg.Delay > 0 {
		rc.delay = time.Duration(cfg.Delay) * time.Millisecond
	}
	bin, err := rc.lookup(cfg.BinDir)
	if err != nil {
		return nil, err
	}
	rc.binPath = bin

	rc.ctx, rc.cl = context.WithCancelCause(logger.SetLogger(rc.ctx, rc.logger))

	return &rc, nil
}

func (rc *RodContext) Context() context.Context {
	return rc.ctx
}

func (rc *RodContext) Close() error {
	err := rc.ctx.Err()
	rc.closer.Do(func() {
		rc.sessionMux.Lock()
		rc.cl(ErrRodContextClosed)
		rc.sessionMux.Unlock()
		rc.sessionWg.Wait()
		_ = os.RemoveAll(rc.userDataDir)
	})
	return err
}

func (rc *RodContext) NewSession() (*RodSession, error) {
	rc.sessionMux.Lock()
	defer rc.sessionMux.Unlock()
	if rc.ctx.Err() != nil {
		return nil, rc.ctx.Err()
	}
	id := uuid.New().String()
	rs, err := rc.launch(id)
	if err != nil {
		return nil, err
	}
	rc.sessionMap[id] = rs
	rc.sessionWg.Add(1)
	return rs, nil
}

func (rc *RodContext) Loop() *RodLoop {
	return NewRodLoop(rc)
}

func (rc *RodContext) Pool(num int) *RodPool {
	return NewRodPool(rc, num)
}

func (rc *RodContext) launch(id string) (*RodSession, error) {
	ctx, cl := context.WithCancelCause(rc.ctx)
	rs := &RodSession{
		rc:       rc,
		id:       id,
		ctx:      ctx,
		cl:       cl,
		userdata: path.Join(rc.userDataDir, id),
	}

	var err error
	rs.launcher, rs.browser, err = rc.launchRod(rs.ctx, rs.userdata)
	if err != nil {
		_ = rs.Close()
		return nil, err
	}
	return rs, nil
}

func (rc *RodContext) launchRod(ctx context.Context, userdata string) (*launcher.Launcher, *rod.Browser, error) {
	l := launcher.New().Bin(rc.binPath).
		UserDataDir(userdata).
		Set("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36").
		Set("disable-blink-features", "AutomationControlled").
		//Set("--enable-parallel-downloading").
		//Set("--disable-gpu").
		Context(ctx).
		//HeadlessNew(!rc.view).
		Headless(!rc.view).
		Devtools(rc.view).
		NoSandbox(true)

	b := rod.New().Context(rc.ctx).SlowMotion(rc.delay).NoDefaultDevice()

	launch, err := l.Launch()
	if err != nil {
		return nil, nil, err
	}

	err = b.ControlURL(launch).Logger(utils.Log(rc.logger.Debug)).Connect()
	if err != nil {
		return nil, nil, err
	}
	return l, b, nil
}

func (rc *RodContext) lookup(bp string) (string, error) {
	lookPath, has := launcher.LookPath()
	if has {
		return lookPath, nil
	}
	if bp == "" {
		bp = path.Join("./", cachePath, binPath)
	}
	browser := &launcher.Browser{
		Context:  context.Background(),
		Revision: launcher.RevisionDefault,
		Hosts:    []launcher.Host{launcher.HostGoogle, launcher.HostNPM, launcher.HostPlaywright},
		RootDir:  bp,
		Logger:   utils.Log(rc.logger.Info),
		LockPort: defaults.LockPort,
	}
	return browser.Get()
}

type RodSession struct {
	rc *RodContext

	id       string
	userdata string
	launcher *launcher.Launcher
	browser  *rod.Browser

	ctx    context.Context
	cl     context.CancelCauseFunc
	closer sync.Once
}

func (rs *RodSession) Launcher() *launcher.Launcher {
	return rs.launcher
}

func (rs *RodSession) Browser() *rod.Browser {
	return rs.browser
}

func (rs *RodSession) Close() error {
	err := rs.ctx.Err()
	rs.closer.Do(func() {
		if rs.browser != nil {
			err = rs.browser.Close()
			if err != nil {
				rs.cl(err)
			}
		}
		if rs.launcher != nil {
			rs.launcher.Kill()
		}
		_ = os.RemoveAll(rs.userdata)
		rs.cl(ErrRodSessionClosed.Errorf(rs.id))
		rs.rc.sessionMux.Lock()
		defer rs.rc.sessionMux.Unlock()
		delete(rs.rc.sessionMap, rs.id)
		rs.rc.sessionWg.Done()
	})
	return err
}

func (rs *RodSession) Reload() (err error) {
	if rs.browser != nil {
		err = rs.browser.Close()
		if err != nil {
			_ = rs.browser.Close()
			return err
		}
	}
	if rs.launcher != nil {
		rs.launcher.Kill()
	}
	_ = os.RemoveAll(rs.userdata)

	rs.launcher, rs.browser, err = rs.rc.launchRod(rs.ctx, rs.userdata)
	if err != nil {
		_ = rs.Close()
		return err
	}
	return nil
}

func (rs *RodSession) PageLoop() *RodPageLoop {
	return NewPageLoop(rs)
}

var (
	ErrRodSessionClosed = xerror.New("rod session[%s] closed")
	ErrRodContextClosed = xerror.New("rod context closed")
)

var (
	defaultDevice = devices.Device{
		Capabilities:   []string{},
		UserAgent:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36",
		AcceptLanguage: "zh",
		Screen: devices.Screen{
			DevicePixelRatio: 1,
			Horizontal: devices.ScreenSize{
				Width:  1280,
				Height: 800,
			},
			Vertical: devices.ScreenSize{
				Width:  800,
				Height: 1280,
			},
		},
		Title: "np",
	}
)
