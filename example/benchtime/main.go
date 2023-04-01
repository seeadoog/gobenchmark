package main

import (
	"context"
	"github.com/seeadoog/gobenchmark"
	"github.com/seeadoog/gobenchmark/app"
	"math/rand"
	"time"
)

// main
func main() {
	myMetric := gobenchmark.NewHistogram("custom cost", gobenchmark.Uniform(0, 1000000, 10000), "us")
	a := app.New("benchmark")

	bentime := a.NewApp("time")
	bentime.SetTask(func(t context.Context, b *gobenchmark.Benchmark) (err error) {
		st := time.Now()
		time.Sleep(time.Duration(rand.Int()%1) * time.Millisecond)
		myMetric.Add(float64(time.Since(st).Nanoseconds() / 1000))
		return nil
	}, gobenchmark.Uniform(0, 1000000, 10000), myMetric)

	benchCost := a.NewApp("cost")

	benchCost.SetTask(func(t context.Context, b *gobenchmark.Benchmark) (err error) {
		st := time.Now()
		time.Sleep(time.Duration(rand.Int()%1) * time.Millisecond)
		myMetric.Add(float64(time.Since(st).Nanoseconds() / 1000))
		return nil
	}, gobenchmark.Uniform(0, 1000000, 10000), myMetric)

	a.Start()
}
