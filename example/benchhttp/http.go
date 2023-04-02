package main

import (
	"context"
	"github.com/seeadoog/gobenchmark"
	"github.com/seeadoog/gobenchmark/app"
	"net/http"
)

func main() {
	a := app.New("benchfasthttp")
	a.SetTask(func(t context.Context, b *gobenchmark.Benchmark) (err error) {
		resp, err := http.DefaultClient.Get("http://127.0.0.1:8972")
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		return err
	}, gobenchmark.Uniform(0, 100000, 100000))
	a.Start()
}
