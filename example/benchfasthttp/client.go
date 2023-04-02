package main

import (
	"context"
	"github.com/seeadoog/gobenchmark"
	"github.com/seeadoog/gobenchmark/app"
	"github.com/valyala/fasthttp"
	"net/http"
)

func main() {
	root := app.New("bench")
	a := root.NewApp("fasthttp")
	addr := root.Cmd().PersistentFlags().String("addr", "127.0.0.1:8972", "addr")
	cli := fasthttp.Client{}
	a.SetTask(func(t context.Context, b *gobenchmark.Benchmark) (err error) {
		req := fasthttp.AcquireRequest()
		resp := fasthttp.AcquireResponse()
		defer fasthttp.ReleaseRequest(req)
		defer fasthttp.ReleaseResponse(resp)

		req.Header.SetMethod("GET")
		req.SetHost(*addr)

		err = cli.Do(req, resp)
		return err
	}, gobenchmark.DefaultBuckets)

	b := root.NewApp("http")
	b.SetTask(func(t context.Context, b *gobenchmark.Benchmark) (err error) {
		resp, err := http.DefaultClient.Get("http://" + *addr)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		return err
	}, gobenchmark.Uniform(0, 100000, 100000))
	root.Start()
}
