package httptap

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
)

type (
	requestContexttKey struct{}

// requestResponseKey struct{}
// handlerKey         struct{}
// proxyLoggerKey     struct{}
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

	closers []io.Closer
}

func withRequestContext(ctx context.Context, rc *RequestContext) context.Context {
	return context.WithValue(ctx, requestContexttKey{}, rc)
}

func RequestContextValue(ctx context.Context) *RequestContext {
	res, ok := ctx.Value(requestContexttKey{}).(*RequestContext)
	if !ok {
		return nil
	}
	return res
}

var _ httputil.BufferPool = (*bytesPool)(nil)

type bytesPool struct {
	pool *sync.Pool
	size int
}

func newBytesPool(size int) *bytesPool {
	if size <= 0 {
		size = 32 * 1024
	}
	return &bytesPool{
		pool: &sync.Pool{
			New: func() any {
				return make([]byte, size)
			},
		},
		size: size,
	}
}

func (p *bytesPool) Get() []byte {
	return p.pool.Get().([]byte)
}

func (p *bytesPool) Put(b []byte) {
	p.pool.Put(b)
}

type Proxy struct {
	http.ServeMux
	upstream   *url.URL
	logger     *slog.Logger
	hasDefault bool

	bytespool *bytesPool
}

func New(upstream string, options ...proxyOption) (*Proxy, error) {
	p := &Proxy{
		ServeMux:  *http.NewServeMux(),
		bytespool: newBytesPool(0),
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

func (p *Proxy) Tap(patterns []string, tap Tap, options ...tapOption) {
	logger := p.logger
	h := NewHandler(p.upstream, p, tap, logger, options...)

	rp := &httputil.ReverseProxy{
		Rewrite:        h.rewrite,
		ModifyResponse: h.modifyResponse,
		ErrorLog:       slog.NewLogLogger(logger.Handler(), slog.LevelError),
		BufferPool:     p.bytespool,
	}
	for _, pattern := range patterns {
		p.ServeMux.Handle(pattern, rp)
		p.hasDefault = p.hasDefault || pattern == "/"
	}
}

func nopTap(logger *slog.Logger) TapFunc {
	log := logger.With("tap", "noTap")
	return TapFunc(func(_ context.Context, _ *RequestResponse) {
		log.Info("called")
	})
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !p.hasDefault {
		p.Tap([]string{"/"}, nopTap(p.logger), WithRequestBody(false), WithResponseBody(false))
		p.hasDefault = true
	}
	// Add the request context to the request.
	rc := &RequestContext{
		Logger: p.logger,
	}
	r = r.WithContext(withRequestContext(r.Context(), rc))
	p.ServeMux.ServeHTTP(w, r)
	rc.closers = append(rc.closers, r.Body)

	// Call the handler.
	rc.Handler.Serve(r.Context(), rc.RequestResponse)

	// Close all bodies.
	for _, c := range rc.closers {
		if c != nil {
			c.Close()
		}
	}
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
