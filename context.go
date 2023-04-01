package gobenchmark

import (
	"context"
	"os"
	"os/signal"
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
