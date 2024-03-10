package httptap

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/myhops/httptap/bufpool"
)

type Handler struct {
	p        *Proxy
	upstream *url.URL
	tap      Tap
	logger   *slog.Logger

	withRequestBody  bool
	WithResponseBody bool
}

func (h *Handler) copyRequest(rr *RequestResponse, pr *httputil.ProxyRequest) {
	// Allocate a buffer for the outgoing request.
	if h.withRequestBody && pr.Out.Body != nil {
		b := bufpool.Get()
		rr.ReqBody = bufpool.Get()
		// Read the body twice.
		body := pr.Out.Body
		b.ReadFrom(io.TeeReader(body, rr.ReqBody))
		// Add the copy to the outgoing request.
		pr.Out.Body = io.NopCloser(b)
	}

	rr.Host = pr.Out.Host
	rr.URL = pr.Out.URL
	rr.ReqProto = pr.Out.Proto
	// Save the headers.
	rr.ReqHeader = pr.Out.Header.Clone()
	rr.ReqTrailer = pr.Out.Trailer.Clone()
	rr.Method = pr.Out.Method
}

func (h *Handler) copyResponse(rr *RequestResponse, r *http.Response) {
	// Save the response body.
	if h.WithResponseBody && r.Body != nil {
		rr.RespBody = bufpool.Get()
		b := bufpool.Get()
		b.ReadFrom(io.TeeReader(r.Body, rr.RespBody))
		r.Body = io.NopCloser(b)
		r.Body.Close()
	}

	// Here we can collect the data
	rr.StatusCode = r.StatusCode
	rr.Status = r.Status
	rr.RespHeader = r.Header.Clone()
	rr.ReqTrailer = r.Trailer.Clone()
	rr.RespProto = r.Proto
}

func (h *Handler) rewrite(pr *httputil.ProxyRequest) {

	// Create the request response and add it to the context of the outgoing request.
	rr := &RequestResponse{
		Start: time.Now(),
	}
	pr.Out = pr.Out.WithContext(WithRequestResponseValue(pr.Out.Context(), rr))

	// set upstream.
	pr.SetURL(h.p.upstream)

	pr.Out.Host = pr.In.Host

	pr.SetXForwarded()

	// Record the data.
	h.copyRequest(rr, pr)
}

func (h *Handler) modifyResponse(r *http.Response) error {
	// Get the request respose.
	ctx := r.Request.Context()
	rr := RequestResponseValue(ctx)
	if rr == nil {
		h.logger.ErrorContext(ctx, "request response not in context")
		return nil
	}
	rr.End = time.Now()
	rr.Duration = rr.End.Sub(rr.Start)  

	// Record the data.
	h.copyResponse(rr, r)

	// Call the tap.
	ctx = context.WithValue(ctx, proxyLoggerKey{}, h.logger)
	h.tap.Serve(ctx, rr)
	// Return the buffers.
	bufpool.Put(rr.ReqBody)
	bufpool.Put(rr.RespBody)
	return nil
}
