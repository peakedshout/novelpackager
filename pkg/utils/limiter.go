package utils

import (
	"context"
	"fmt"
	"time"
)

func NewLimiter(ctx context.Context) *Limiter {
	return &Limiter{
		ctx: ctx,
		m:   make(map[string]chan struct{}),
	}
}

type Limiter struct {
	ctx context.Context
	m   map[string]chan struct{}
}

func (l *Limiter) Add(key string, limit int) *Limiter {
	if limit <= 0 {
		return l
	}
	l.m[key] = make(chan struct{}, limit)
	return l
}

func (l *Limiter) LimitWait(key string) func() {
	ch, ok := l.m[key]
	if !ok {
		return func() {}
	}
	select {
	case ch <- struct{}{}:
		return func() {
			<-ch
		}
	case <-l.ctx.Done():
		return func() {}
	}
}

func (l *Limiter) LimitReject(key string) (func(), error) {
	ch, ok := l.m[key]
	if !ok {
		return func() {}, nil
	}
	select {
	case ch <- struct{}{}:
		return func() {
			<-ch
		}, nil
	default:
		return nil, fmt.Errorf("limit reject: %s", key)
	}
}

func (l *Limiter) LimitTimeout(key string, td time.Duration) (func(), error) {
	ch, ok := l.m[key]
	if !ok {
		return func() {}, nil
	}
	timer := time.NewTimer(td)
	defer timer.Stop()
	select {
	case ch <- struct{}{}:
		return func() {
			<-ch
		}, nil
	case <-l.ctx.Done():
		return nil, l.ctx.Err()
	case <-timer.C:
		return nil, fmt.Errorf("limit timeout: %s", key)
	}
}
