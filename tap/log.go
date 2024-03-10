package tap

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/myhops/httptap"
)

type LogTap struct {
	logger *slog.Logger
	level  slog.Level
}

func NewLogTap(logger *slog.Logger, level slog.Level) *LogTap {
	return &LogTap{
		logger: logger,
		level:  level,
	}
}

func (l *LogTap) Serve(ctx context.Context, rr *httptap.RequestResponse) {
	attrs := []slog.Attr{
		slog.String("host", rr.Host),
		slog.String("method", rr.Method),
		slog.String("path", rr.URL.Path),
		slog.String("url", rr.URL.String()),
		slog.String("status", rr.Status),
		slog.Any("request_header", slog.GroupValue(headerToAttrs(rr.ReqHeader)...)),
		slog.Any("request_trailer", slog.GroupValue(headerToAttrs(rr.ReqTrailer)...)),
		slog.Any("response_header", slog.GroupValue(headerToAttrs(rr.RespHeader)...)),
		slog.Any("response_trailer", slog.GroupValue(headerToAttrs(rr.RespTrailer)...)),
	}
	l.logger.LogAttrs(ctx, l.level, "upstream called", attrs...)
}

// Filter sensitive headers.
func headerToAttrs(h http.Header) []slog.Attr {
	var values []slog.Attr
	for k, v := range h {
		values = append(values, slog.String(k, v[0]))
	}

	slog.GroupValue()
	return values
}
