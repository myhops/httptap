package httptap

import (
	"log/slog"

	jsonpatch "github.com/evanphx/json-patch"
)

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

func WithRequestBodyPatch(patch []byte) tapOption {
	return tapOption(func(h *Handler) {
		p, err := jsonpatch.DecodePatch(patch)
		if err != nil {
			h.logger.Error("ReqBody DecodePatch error", slog.String("err", err.Error()))
			return
		}
		h.reqBodyPatch = p
	})
}

func WithResponseBodyPatch(patch []byte) tapOption {
	return tapOption(func(h *Handler) {
		logger := h.logger.With(slog.String("step", "WithResponseBodyPatch"))
		p, err := jsonpatch.DecodePatch(patch)
		if err != nil {
			logger.Error("RespBody DecodePatch error",
				slog.String("err", err.Error()),
				slog.String("patch", string(patch)),
			)
			return
		}
		logger.Info("adding patch")
		h.respBodyPatch = p
	})
}

func WithIncludeHeaders(header []string) tapOption {
	return tapOption(func(h *Handler) {
		h.includeHeaders = canonicalHeaders(header)
	})
}

func WithExcludeHeaders(header []string) tapOption {
	return tapOption(func(h *Handler) {
		h.excludeHeaders = canonicalHeaders(header)
	})
}
