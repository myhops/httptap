package httptap

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type (
	requestResponseKey struct{}
	proxyLoggerKey struct{}
)

func WithRequestResponseValue(ctx context.Context, rr *RequestResponse) context.Context {
	return context.WithValue(ctx, requestResponseKey{}, rr)
}

func RequestResponseValue(ctx context.Context) *RequestResponse {
	rr, ok := ctx.Value(requestResponseKey{}).(*RequestResponse)
	if !ok {
		return nil
	}
	return rr
}

func ProxyLoggerValue(ctx context.Context) *slog.Logger {
	l, ok := ctx.Value(proxyLoggerKey{}).(*slog.Logger)
	if !ok {
		return nil
	}
	return l
}

type Proxy struct {
	http.ServeMux
	upstream   *url.URL
	logger     *slog.Logger
	hasDefault bool
}

func New(upstream string, options ...proxyOption) (*Proxy, error) {
	p := &Proxy{
		ServeMux: *http.NewServeMux(),
	}

	// Process options
	for _, o := range options {
		o(p)
	}
	// Set logger
	if p.logger == nil {
		p.logger = slog.Default()
	}

	var err error
	// Parse upstream.
	if p.upstream, err = url.Parse(upstream); err != nil {
		return nil, fmt.Errorf("error parsing upstream: %w", err)
	}
	return p, nil
}

type proxyOption = func(p *Proxy)

func WithLogger(logger *slog.Logger) proxyOption {
	return proxyOption(func(p *Proxy) {
		p.logger = logger
	})
}

// Tap options can modify the handler.
// The passed handler has proxy, logger and upstream set.
type tapOption func(p *Handler)

func (p *Proxy) Tap(pattern string, tap Tap, options ...tapOption) {
	logger := p.logger
	th := &Handler{
		p:        p,
		upstream: p.upstream,
		tap:      tap,
		logger:   p.logger,
	}

	for _, o := range options {
		o(th)
	}

	rp := &httputil.ReverseProxy{
		Rewrite:        th.rewrite,
		ModifyResponse: th.modifyResponse,
		ErrorLog:       slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}
	th.logger = logger
	p.ServeMux.Handle(pattern, rp)
	p.hasDefault = p.hasDefault || pattern == "/"
}

func nopTap(logger *slog.Logger) TapFunc {
	log := logger.With("tap", "noTap")
	return TapFunc(func(_ context.Context, _ *RequestResponse) {
		log.Info("called")
	})
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !p.hasDefault {
		p.Tap("/", nopTap(p.logger), WithRequestBody(false), WithResponseBody(false))
		p.hasDefault = true
	}
	p.ServeMux.ServeHTTP(w, r)
}

func WithLogAttrs(attrs ...slog.Attr) tapOption {
	return tapOption(func(h *Handler) {
		h.logger = slog.New(h.logger.Handler().WithAttrs(attrs))
	})
}

func WithRequestBody(yes ...bool) tapOption {
	return tapOption(func(h *Handler) {
		h.withRequestBody = len(yes) != 1 || yes[0]
	})
}

func WithResponseBody(yes ...bool) tapOption {
	return tapOption(func(h *Handler) {
		h.WithResponseBody = len(yes) != 1 || yes[0]
	})
}
