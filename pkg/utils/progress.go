package utils

import (
	"fmt"
	"sync"
)

type Progress struct {
	mux     sync.RWMutex
	total   int64
	current int64
	err     error
}

func (p *Progress) Error() error {
	p.mux.RLock()
	defer p.mux.RUnlock()
	return p.err
}

func (p *Progress) SetError(err error) {
	p.mux.Lock()
	defer p.mux.Unlock()
	p.err = err
}

func (p *Progress) Set(c, t int64) {
	p.mux.Lock()
	defer p.mux.Unlock()
	p.total = t
	p.current = c
}

func (p *Progress) Init(t int64) {
	p.mux.Lock()
	defer p.mux.Unlock()
	p.total = t
}

func (p *Progress) Add(c int64) {
	p.mux.Lock()
	defer p.mux.Unlock()
	p.current += c
}

func (p *Progress) Update(current int64) {
	p.mux.Lock()
	defer p.mux.Unlock()
	p.current = current
}

func (p *Progress) Total() int64 {
	p.mux.RLock()
	defer p.mux.RUnlock()
	return p.total
}

func (p *Progress) Current() int64 {
	p.mux.RLock()
	defer p.mux.RUnlock()
	return p.current
}

func (p *Progress) Percent() float64 {
	if p.total < 0 {
		return -1
	}
	if p.total == 0 {
		return 0
	}
	if p.current >= p.total {
		return 1
	}
	p.mux.RLock()
	defer p.mux.RUnlock()
	f := float64(p.current) / float64(p.total)
	if f > 1 {
		f = 1
	} else if f < 0 {
		f = 0
	}
	return f
}

func (p *Progress) String() string {
	p.mux.RLock()
	defer p.mux.RUnlock()
	if p.err != nil {
		return fmt.Errorf("err: %w", p.err).Error()
	}
	f := p.Percent()
	if f < 0 {
		return "null"
	}
	return fmt.Sprintf("%.2f%%", f*100)
}

func NewProgress(total int64) *Progress {
	return &Progress{
		total: total,
	}
}
