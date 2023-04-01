package main

import (
	"github.com/valyala/fasthttp"
	"testing"
)

func TestServer(t *testing.T) {

	fasthttp.ListenAndServe(":8972", func(ctx *fasthttp.RequestCtx) {
		
	})
}
