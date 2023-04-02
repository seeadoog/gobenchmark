package gobenchmark

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"
)

func NewContext(ctx context.Context, duration time.Duration) context.Context {
	ctx, cf := context.WithTimeout(ctx, duration)
	go func() {
		sigc := make(chan os.Signal, 1)
		signal.Notify(sigc, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGHUP)
		<-sigc
		cf()
	}()
	return ctx
}

type Counter struct {
	n int64

	ct     atomic.Int64
	ctx    *counterContext
	closed atomic.Bool
}

type counterContext struct {
	counter *Counter
	parent  context.Context
	done    chan struct{}
	err     error
}

func (c *counterContext) Deadline() (deadline time.Time, ok bool) {
	return c.parent.Deadline()
}

func (c *counterContext) Err() error {
	if c.err == nil {
		return c.parent.Err()
	}
	return c.err
}

func (c *counterContext) Value(key any) any {
	return c.parent.Value(key)
}

func (c *counterContext) Done() <-chan struct{} {
	return c.done
}

func (c *Counter) Add(n int64) {
	res := c.ct.Add(n)
	if res >= c.n {
		c.Cancel()
	}
}

func (c *Counter) Cancel() {
	if c.closed.CompareAndSwap(false, true) {
		c.ctx.err = errors.New("counter done")
		close(c.ctx.done)
	}
}

func NewCounterContext(ctx context.Context, n int64) (context.Context, *Counter) {
	c := new(Counter)
	c.n = n
	cc := &counterContext{
		counter: c,
		parent:  ctx,
		done:    make(chan struct{}),
	}

	c.ctx = cc

	go func() {
		select {
		case <-ctx.Done():
			c.Cancel()
		}
	}()
	return cc, c
}
