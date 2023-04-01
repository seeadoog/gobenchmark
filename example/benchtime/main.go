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
	myMetric := gobenchmark.NewHistogram("custom cost", gobenchmark.Uniform(0, 1000000000, 10000), "ns")
	a := app.New("benchtime")
	a.Start(func(t context.Context, b *gobenchmark.Benchmark) (err error) {
		st := time.Now()
		time.Sleep(time.Duration(rand.Int()%1) * time.Millisecond)
		myMetric.Add(float64(time.Since(st)))
		return nil
	}, gobenchmark.Uniform(0, 1000000, 10000), myMetric)
}
