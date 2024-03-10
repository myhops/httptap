package tap

import (
	"context"
	"io"
	"log/slog"
	"text/template"
	"time"

	"github.com/myhops/httptap"
)

const (
	DefaultTemplate = `Time={{.Time -}}, Host={{.Data.Host}}, URL={{.URL.Scheme}}://{.URL.Host}}{{.URL.Path}}`
)

type TemplateTap struct {
	w   io.Writer
	tpl *template.Template
}

type TemplateObject struct {
	Time time.Time
	Data *httptap.RequestResponse
}



func NewTemplateTap(w io.Writer, text string) (*TemplateTap, error) {
	tt, err := template.New("tap").Parse(text)
	if err != nil {
		return nil, err
	}
	return &TemplateTap{
		w: w,
		tpl: tt,
	}, nil
}

func (t *TemplateTap) Serve(ctx context.Context, rr *httptap.RequestResponse) {
	logger := slog.Default()
	if ll := httptap.ProxyLoggerValue(ctx); ll != nil {
		logger = ll
	}
	to := TemplateObject{
		Time: time.Now(),
		Data: rr,
	}
	if err := t.tpl.Execute(t.w, to); err != nil {
		logger.ErrorContext(ctx, "cannot execute template", slog.String("err", err.Error()))	
	}
}

