package rodx

import (
	"context"
	"github.com/go-rod/rod"
	"sync"
)

type RodPool struct {
	rc *RodContext

	tasks  chan func(b *rod.Browser) error
	twg    sync.WaitGroup
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
}

func NewRodPool(rc *RodContext, num int) *RodPool {
	if num < 1 {
		num = 1
	}
	ctx, cancel := context.WithCancel(rc.Context())
	pool := &RodPool{
		rc:     rc,
		tasks:  make(chan func(b *rod.Browser) error),
		ctx:    ctx,
		cancel: cancel,
	}
	pool.wg.Add(num)
	for i := 0; i < num; i++ {
		go pool.worker()
	}
	return pool
}

func (p *RodPool) worker() {
	defer p.wg.Done()
	for {
		select {
		case <-p.ctx.Done():
			return
		case task := <-p.tasks:
			p.rDo(task)
			p.twg.Done()
		}
	}
}

func (p *RodPool) Do(task func(*rod.Browser) error) bool {
	select {
	case <-p.ctx.Done():
		p.rDo(task)
		return false
	case p.tasks <- task:
		p.twg.Add(1)
		return true
	}
}

func (p *RodPool) Stop() {
	p.cancel()
	p.wg.Wait()
}

func (p *RodPool) Wait() {
	p.twg.Wait()
}

func (p *RodPool) rDo(task func(*rod.Browser) error) {
	rs, err := p.rc.NewSession()
	if err != nil {
		p.rc.logger.Warn("rod session error", "err", err)
		return
	}
	err = task(rs.Browser())
	_ = rs.Close()
	if err != nil && p.ctx.Err() == nil {
		p.rDo(task)
	}
}
