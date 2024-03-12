package httptap

import (
	"bytes"
	"context"
	"net/http"
	"net/url"
	"time"
)

type RequestResponse struct {
	Start    time.Time
	End      time.Time
	Duration time.Duration

	Host       string
	URL        *url.URL
	ReqProto   string
	Method     string
	ReqHeader  http.Header
	ReqTrailer http.Header
	// This buffer is valid until the end of Serve.
	// Do not read from this buffer, but create a reader from it.
	ReqBody     *bytes.Buffer
	ReqBodyJSON any

	StatusCode  int
	Status      string
	RespProto   string
	RespHeader  http.Header
	RespTrailer http.Header
	// RespBody contains the body of the response, can be nil.
	//
	// This buffer is valid until the end of Serve.
	//
	// Do not read from this buffer, but create a reader from it.
	// 	r := bytes.NewReader(RespBody.Bytes())
	//
	// If you need access to this buffer outside of Serve, then create a copy.
	// 	savedBody := make([]byte, RespBody.Len())
	//  copy(savedBody, RespBody.Bytes())
	//  r := bytes.NewReader(savedBody)
	RespBody     *bytes.Buffer
	RespBodyJSON any
}

type Tap interface {
	// Serve handles the tap.
	Serve(context.Context, *RequestResponse)
}

type TapFunc func(context.Context, *RequestResponse)

func (t TapFunc) Serve(ctx context.Context, r *RequestResponse) {
	t(ctx, r)
}

