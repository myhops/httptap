package tap

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"mime"
	"net/http"

	"github.com/myhops/httptap"
)

type ctError string

func (e ctError) Error() string {
	return fmt.Sprintf("expected application/json, got %s", string(e))
}

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
	if rr.RespBody != nil && t.isJson(rr.RespHeader) == nil{
		var obj any
		if t.unmarshalJSON(rr.RespBody, &obj) == nil {
			attrs = append(attrs, slog.Any("response_body_json", obj))
		}
	}
	if rr.ReqBody != nil && t.isJson(rr.ReqHeader) == nil{
		var obj any
		if t.unmarshalJSON(rr.ReqBody, &obj) == nil {
			attrs = append(attrs, slog.Any("request_body_json", obj))
		}
	}
	t.logger.LogAttrs(ctx, t.level, "upstream called", attrs...)
}

func (t *LogTap) unmarshalJSON(b *bytes.Buffer, obj *any) error {
	r := bytes.NewReader(b.Bytes())
	if err := json.NewDecoder(r).Decode(obj); err != nil {
		return err
	}
	return nil
}

func (t *LogTap) isJson(h http.Header) error {
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
