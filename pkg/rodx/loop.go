package rodx

import (
	"context"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"time"
)

type RodLoop struct {
	rc *RodContext
}

func NewRodLoop(rc *RodContext) *RodLoop {
	return &RodLoop{rc: rc}
}

func (rl *RodLoop) Do(fn func(b *rod.Browser) error) error {
	rs, err := rl.rc.NewSession()
	if err != nil {
		return err
	}
	err = fn(rs.Browser())
	_ = rs.Close()
	if err != nil && rl.rc.ctx.Err() == nil {
		return rl.Do(fn)
	}
	return rl.rc.ctx.Err()
}

func (rl *RodLoop) DoWithNum(fn func(b *rod.Browser) error, num uint) error {
	if num <= 0 {
		return rl.Do(fn)
	}
	rs, err := rl.rc.NewSession()
	if err != nil {
		return err
	}
	err = fn(rs.Browser())
	_ = rs.Close()
	if err != nil && rl.rc.ctx.Err() == nil {
		num--
		if num == 0 {
			return err
		}
		return rl.DoWithNum(fn, num)
	}
	return rl.rc.ctx.Err()
}

func (rl *RodLoop) DoWithTimeout(fn func(b *rod.Browser) error, timeout time.Duration) error {
	if timeout <= 0 {
		return rl.Do(fn)
	}
	ctx, cl := context.WithTimeout(rl.rc.ctx, timeout)
	defer cl()
	for ctx.Err() == nil {
		rs, err := rl.rc.NewSession()
		if err != nil {
			return err
		}
		err = fn(rs.Browser().Context(ctx))
		_ = rs.Close()
		if err == nil {
			break
		}
	}
	return ctx.Err()
}

type RodPageLoop struct {
	sess *RodSession
}

func NewPageLoop(s *RodSession) *RodPageLoop {
	return &RodPageLoop{sess: s}
}

func (rpl *RodPageLoop) Do(fn func(page *rod.Page) error) error {
	page, err := rpl.sess.Browser().Page(proto.TargetCreateTarget{})
	if err != nil {
		return err
	}
	err = fn(page)
	_ = page.Close()
	if err != nil && rpl.sess.Browser().GetContext().Err() == nil {
		return rpl.Do(fn)
	}
	return rpl.sess.Browser().GetContext().Err()
}

func (rpl *RodPageLoop) DoWithNum(fn func(page *rod.Page) error, num uint) error {
	if num <= 0 {
		return rpl.Do(fn)
	}
	page, err := rpl.sess.Browser().Page(proto.TargetCreateTarget{})
	if err != nil {
		return err
	}
	err = fn(page)
	_ = page.Close()
	if err != nil && rpl.sess.Browser().GetContext().Err() == nil {
		num--
		if num == 0 {
			return err
		}
		return rpl.DoWithNum(fn, num)
	}
	return rpl.sess.Browser().GetContext().Err()
}

func (rpl *RodPageLoop) DoWithTimeout(fn func(b *rod.Page) error, timeout time.Duration) error {
	if timeout <= 0 {
		return rpl.Do(fn)
	}
	ctx, cl := context.WithTimeout(rpl.sess.Browser().GetContext(), timeout)
	defer cl()
	for ctx.Err() == nil {
		page, err := rpl.sess.Browser().Page(proto.TargetCreateTarget{})
		if err != nil {
			return err
		}
		err = fn(page.Context(ctx))
		_ = page.Close()
		if err == nil {
			break
		}
	}
	return rpl.sess.Browser().GetContext().Err()
}

func (rpl *RodPageLoop) DoX(fn func(page *rod.Page) error) error {
	page, err := rpl.sess.Browser().Page(proto.TargetCreateTarget{})
	if err != nil {
		return err
	}
	err = fn(page)
	_ = page.Close()
	if err != nil && rpl.sess.Browser().GetContext().Err() == nil {
		err = rpl.sess.Reload()
		if err != nil {
			return err
		}
		return rpl.Do(fn)
	}
	return rpl.sess.Browser().GetContext().Err()
}

func (rpl *RodPageLoop) DoWithNumX(fn func(page *rod.Page) error, num uint) error {
	if num <= 0 {
		return rpl.DoX(fn)
	}
	page, err := rpl.sess.Browser().Page(proto.TargetCreateTarget{})
	if err != nil {
		return err
	}
	err = fn(page)
	_ = page.Close()
	if err != nil && rpl.sess.Browser().GetContext().Err() == nil {
		num--
		if num == 0 {
			return err
		}
		err = rpl.sess.Reload()
		if err != nil {
			return err
		}
		return rpl.DoWithNum(fn, num)
	}
	return rpl.sess.Browser().GetContext().Err()
}

func (rpl *RodPageLoop) DoWithTimeoutX(fn func(b *rod.Page) error, timeout time.Duration) error {
	if timeout <= 0 {
		return rpl.Do(fn)
	}
	ctx, cl := context.WithTimeout(rpl.sess.ctx, timeout)
	defer cl()
	for ctx.Err() == nil {
		page, err := rpl.sess.Browser().Page(proto.TargetCreateTarget{})
		if err != nil {
			return err
		}
		err = fn(page.Context(ctx))
		_ = page.Close()
		if err == nil {
			break
		} else {
			err = rpl.sess.Reload()
			if err != nil {
				return err
			}
		}
	}
	return rpl.sess.Browser().GetContext().Err()
}
