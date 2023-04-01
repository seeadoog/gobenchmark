package main

import (
	"context"
	"github.com/seeadoog/gobenchmark"
	"github.com/seeadoog/gobenchmark/app"
	"github.com/valyala/fasthttp"
)

func main() {
	a := app.New("benchfasthttp")
	cli := fasthttp.Client{}
	a.SetTask(func(t context.Context, b *gobenchmark.Benchmark) (err error) {
		req := fasthttp.AcquireRequest()
		resp := fasthttp.AcquireResponse()
		defer fasthttp.ReleaseRequest(req)
		defer fasthttp.ReleaseResponse(resp)

		req.Header.SetMethod("GET")
		req.SetHost("127.0.0.1:8972")

		err = cli.Do(req, resp)
		return err
	}, gobenchmark.Uniform(0, 100000, 100000))
	a.Start()
}
