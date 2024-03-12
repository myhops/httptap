package httptap

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/myhops/httptap/bufpool"
)

type ctError string

func (e ctError) Error() string {
	return fmt.Sprintf("expected application/json, got %s", string(e))
}

type Handler struct {
	p        *Proxy
	upstream *url.URL
	tap      Tap
	logger   *slog.Logger

	withRequestBody  bool
	withResponseBody bool
	withRequestJSON  bool
	withResponseJSON bool
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
	if h.withResponseBody && r.Body != nil {
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
	// Add the request context to the outgoing request.
	rc := requestContextValue(pr.In.Context())
	pr.Out = pr.Out.WithContext(withRequestContext(pr.Out.Context(), rc))

	// Add myself to the request context.
	rc.Handler = h

	// Create the request response and add it to the request context
	rr := &RequestResponse{
		Start: time.Now(),
	}
	rc.RequestResponse = rr

	// set upstream.
	pr.SetURL(h.p.upstream)

	pr.Out.Host = pr.In.Host

	pr.SetXForwarded()

	// Record the data.
	h.copyRequest(rr, pr)
}

func (h *Handler) modifyResponse(r *http.Response) error {
	// Get the request context
	rc := requestContextValue(r.Request.Context())

	rr := rc.RequestResponse
	rr.End = time.Now()
	rr.Duration = rr.End.Sub(rr.Start)

	// Record the data.
	h.copyResponse(rr, r)

	// Unmarshal json bodies.
	h.unmarshalBodies(rr)

	return nil
}

func (h *Handler) Serve(ctx context.Context, rr *RequestResponse) error {
	// Call the tap.
	h.tap.Serve(ctx, rr)

	// Return the buffers.
	bufpool.Put(rr.ReqBody)
	bufpool.Put(rr.RespBody)
	return nil
}

func (h *Handler) unmarshalBodies(rr *RequestResponse) {
	if h.withRequestJSON && h.isJson(rr.ReqHeader) == nil {
		h.unmarshalJSON(rr.ReqBody, &rr.ReqBodyJSON)
	}
	if h.withResponseJSON && h.isJson(rr.RespHeader) == nil {
		h.unmarshalJSON(rr.RespBody, &rr.RespBodyJSON)
	}
}

func (t *Handler) unmarshalJSON(b *bytes.Buffer, obj *any) error {
	r := bytes.NewReader(b.Bytes())
	if err := json.NewDecoder(r).Decode(obj); err != nil {
		return err
	}
	return nil
}

func (t *Handler) isJson(h http.Header) error {
	cth := h.Get("Content-Type")
	if cth == "" {
		return errors.New("no Content-Type header")
	}
	ct, _, err := mime.ParseMediaType(cth)
	if err != nil {
		return fmt.Errorf("error parsing Content-Type: %w", err)
	}
	if ct != "application/json" {
		return ctError(ct)
	}
	return nil
}
