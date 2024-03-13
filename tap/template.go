package tap

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"text/template"

	"github.com/myhops/httptap"
	"github.com/myhops/httptap/bufpool"
)

const (
	DefaultTemplate = `Time={{.Data.Start.String -}}, Method={{.Data.Method}}, Host={{.Data.Host}}, URL={{.Data.URL.Scheme}}://{{.Data.URL.Host}}{{.Data.URL.Path}}{{"\n"}}`
)

type TemplateTapConfig struct {
	Text   string `yaml:"text"`
	Group  string `yaml:"group,omitempty"`
	Logger string `yaml:"logger,omitempty"`
	Format string `yaml:"format,omitempty"`

	logger *slog.Logger
}

type TemplateTap struct {
	logger     *slog.Logger
	tpl        *template.Template
	group      string
	formatJSON bool
}

type TemplateObject struct {
	Data *httptap.RequestResponse
}

func NewTemplateTapCfg(cfg *TemplateTapConfig) (*TemplateTap, error) {
	return newTemplateTap(cfg.logger, cfg.Text, cfg.Group, cfg.Format)
}

func NewTemplateTap(logger *slog.Logger, text string, group string, format string) (*TemplateTap, error) {
	return newTemplateTap(logger, text, group, format)
}

func newTemplateTap(logger *slog.Logger, text string, group string, format string) (*TemplateTap, error) {
	tt, err := template.New("tap").Parse(text)
	if err != nil {
		return nil, err
	}
	if group != "" {
		logger = logger.WithGroup(group)
	}
	return &TemplateTap{
		logger: logger,
		tpl:    tt,
		group:  group,
		formatJSON: strings.ToLower(format) == "json",
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
	// Execute the template.
	b := bufpool.Get()
	defer bufpool.Put(b)
	if err := t.tpl.Execute(b, to); err != nil {
		logger.ErrorContext(ctx, "cannot execute template", slog.String("err", err.Error()))
	}
	if !t.formatJSON {
		t.logger.InfoContext(ctx, "audit log", slog.String("data", b.String()))
		return
	}
	var obj any
	if err := json.Unmarshal(b.Bytes(), &obj); err != nil {
		logger.ErrorContext(ctx, "unmarshal failed", slog.String("err", err.Error()))
	}
	t.logger.InfoContext(ctx, "audit log", slog.Any("data", obj))
}
