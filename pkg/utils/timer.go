package utils

import (
	"context"
	"github.com/peakedshout/go-pandorasbox/tool/expired"
	"io"
	"time"
)

var expireManager = expired.Init(context.Background(), 5)

type _expireCloser struct {
	io.Closer
}

func (e *_expireCloser) Id() any {
	return e.Closer
}

func (e *_expireCloser) ExpiredFunc() {
	_ = e.Closer.Close()
}

func ExpireClose(c io.Closer, d time.Duration) func() {
	expireManager.SetWithDuration(&_expireCloser{Closer: c}, d)
	return func() {
		expireManager.Remove(c, true)
	}
}

func UpdateExpireClose(c io.Closer, d time.Duration) {
	expireManager.UpdateWithDuration(c, d)
}

func RecordTime(fn func(duration time.Duration)) func() {
	now := time.Now()
	return func() {
		fn(time.Since(now))
	}
}
