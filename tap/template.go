package tap

import (
	"context"
	"io"
	"log/slog"
	"sync"
	"text/template"

	"github.com/myhops/httptap"
)

const (
	DefaultTemplate = `Time={{.Data.Start.String -}}, Method={{.Data.Method}}, Host={{.Data.Host}}, URL={{.Data.URL.Scheme}}://{{.Data.URL.Host}}{{.Data.URL.Path}}{{"\n"}}`
)

type TemplateTap struct {
	w   io.Writer
	tpl *template.Template
	m   sync.Mutex
}

type TemplateObject struct {
	Data *httptap.RequestResponse
}

func NewTemplateTap(w io.Writer, text string) (*TemplateTap, error) {
	tt, err := template.New("tap").Parse(text)
	if err != nil {
		return nil, err
	}
	return &TemplateTap{
		w:   w,
		tpl: tt,
	}, nil
}

func (t *TemplateTap) Serve(ctx context.Context, rr *httptap.RequestResponse) {
	logger := slog.Default()
	// Get the context
	rc := httptap.RequestContextValue(ctx)
	if rc.Logger != nil {
		logger = rc.Logger
	}
	to := TemplateObject{
		Data: rr,
	}
	// Lock writer.
	t.m.Lock()
	defer t.m.Unlock()
	if err := t.tpl.Execute(t.w, to); err != nil {
		logger.ErrorContext(ctx, "cannot execute template", slog.String("err", err.Error()))
	}
}
