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
	requestContexttKey struct{}
	requestResponseKey struct{}
	handlerKey         struct{}
	proxyLoggerKey     struct{}
)

type ProxyTap struct {
	Pattern  string
	TypeName string
}

type ProxyConfig struct {
	Listen   string
	Upstream string
	Taps     []ProxyTap
	LogLevel slog.Level
}

type RequestContext struct {
	Handler         *Handler
	RequestResponse *RequestResponse
	Logger          *slog.Logger
}

func withRequestContext(ctx context.Context, rc *RequestContext) context.Context {
	return context.WithValue(ctx, requestContexttKey{}, rc)
}

func requestContextValue(ctx context.Context) *RequestContext {
	res, ok := ctx.Value(requestContexttKey{}).(*RequestContext)
	if !ok {
		return nil
	}
	return res
}

// withRequestResponseValue add the request response to the context.
func withRequestResponseValue(ctx context.Context, rr *RequestResponse) context.Context {
	return context.WithValue(ctx, requestResponseKey{}, rr)
}

// requestResponseValue returns the request response if present.
func requestResponseValue(ctx context.Context) *RequestResponse {
	rr, ok := ctx.Value(requestResponseKey{}).(*RequestResponse)
	if !ok {
		return nil
	}
	return rr
}

// withRequestResponseValue add the request response to the context.
func withHandlerValue(ctx context.Context, h *Handler) context.Context {
	return context.WithValue(ctx, handlerKey{}, h)
}

// HandlerValue returns the request response if present.
func handlerValue(ctx context.Context) *Handler {
	res, ok := ctx.Value(handlerKey{}).(*Handler)
	if !ok {
		return nil
	}
	return res
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
	// Add the request context to the request.
	rc := &RequestContext{
		Logger: p.logger,
	}
	r = r.WithContext(withRequestContext(r.Context(), rc))
	p.ServeMux.ServeHTTP(w, r)

	// Call the handler.
	rc.Handler.Serve(r.Context(), rc.RequestResponse)
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
		h.withResponseBody = len(yes) != 1 || yes[0]
	})
}

func WithRequestJSON(yes ...bool) tapOption {
	return tapOption(func(h *Handler) {
		h.withRequestJSON = len(yes) != 1 || yes[0]
	})
}

func WithResponseJSON(yes ...bool) tapOption {
	return tapOption(func(h *Handler) {
		h.withResponseJSON = len(yes) != 1 || yes[0]
	})
}
