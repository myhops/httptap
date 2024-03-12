package tap

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/myhops/httptap"
)

// TODO: Add body filter with json patch.
type LogTap struct {
	logger  *slog.Logger
	level   slog.Level
	blocked []string
}

func NewLogTap(logger *slog.Logger, level slog.Level) *LogTap {
	return &LogTap{
		logger:  logger,
		level:   level,
		blocked: []string{http.CanonicalHeaderKey("authorization")},
	}
}

func (t *LogTap) Serve(ctx context.Context, rr *httptap.RequestResponse) {
	attrs := []slog.Attr{
		slog.String("host", rr.Host),
		slog.String("method", rr.Method),
		slog.String("path", rr.URL.Path),
		slog.String("url", rr.URL.String()),
		slog.String("status", rr.Status),
		slog.Any("request_header", slog.GroupValue(t.headerToAttrs(rr.ReqHeader)...)),
		slog.Any("request_trailer", slog.GroupValue(t.headerToAttrs(rr.ReqTrailer)...)),
		slog.Any("response_header", slog.GroupValue(t.headerToAttrs(rr.RespHeader)...)),
		slog.Any("response_trailer", slog.GroupValue(t.headerToAttrs(rr.RespTrailer)...)),
	}
	if rr.ReqBodyJSON != nil {
		attrs = append(attrs, slog.Any("request_body_json", rr.ReqBodyJSON))
	}
	if rr.RespBodyJSON != nil {
		attrs = append(attrs, slog.Any("response_body_json", rr.RespBodyJSON))
	}
	t.logger.LogAttrs(ctx, t.level, "upstream called", attrs...)
}

func (t *LogTap) isBlocked(key string) bool {
	for _, k := range t.blocked {
		if k == key {
			return true
		}
	}
	return false
}

// Filter sensitive headers.
func (t *LogTap) headerToAttrs(h http.Header) []slog.Attr {
	var values []slog.Attr
	for k, v := range h {
		if !t.isBlocked(k) {
			values = append(values, slog.String(k, v[0]))
		}
	}

	slog.GroupValue()
	return values
}
