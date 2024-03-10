package httptap

import (
	"bytes"
	"context"
	"net/http"
	"net/url"
)

type RequestResponse struct {
	Host       string
	URL        *url.URL
	ReqProto   string
	Method     string
	ReqHeader  http.Header
	ReqTrailer http.Header
	ReqBody    *bytes.Buffer

	StatusCode  int
	Status      string
	RespProto   string
	RespHeader  http.Header
	RespTrailer http.Header
	RespBody    *bytes.Buffer
}

type Tap interface {
	Serve(context.Context, *RequestResponse)
}

type TapFunc func(context.Context, *RequestResponse)

func (t TapFunc) Serve(ctx context.Context, r *RequestResponse) {
	t(ctx, r)
}
