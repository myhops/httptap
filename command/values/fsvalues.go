package values

import (
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/url"
	"strings"

	"github.com/myhops/httptap/config"
	"gopkg.in/yaml.v3"
)

type FlagSet struct {
	*flag.FlagSet
}

func NewFlagSet(name string, errorHandling flag.ErrorHandling) *FlagSet {
	return &FlagSet{
		FlagSet: flag.NewFlagSet(name, errorHandling),
	}
}

type LogFormatValue string

func (l *LogFormatValue) Set(value string) error {
	switch strings.ToLower(value) {
	case "text", "json":
		*l = LogFormatValue(value)
	default:
		return fmt.Errorf("bad log format: %s", value)
	}
	return nil
}

func (l *LogFormatValue) String() string {
	if l == nil {
		return "LogFormat undefined"
	}
	return string(*l)
}

func newLogFormatValue(value string, p *string) *LogFormatValue {
	*p = value
	return (*LogFormatValue)(p)
}

type LogLevelValue slog.Level

func (l *LogLevelValue) Set(value string) error {
	switch strings.ToUpper(value) {
	case "DEBUG":
		*l = LogLevelValue(slog.LevelDebug)
	case "INFO":
		*l = LogLevelValue(slog.LevelInfo)
	case "WARN":
		*l = LogLevelValue(slog.LevelWarn)
	case "ERROR":
		*l = LogLevelValue(slog.LevelError)
	default:
		return fmt.Errorf("bad value for log level: %s", value)
	}
	return nil
}

func newLogLevelValue(value slog.Level, p *slog.Level) *LogLevelValue {
	*p = value
	return (*LogLevelValue)(p)
}

func (l *LogLevelValue) String() string {
	if l == nil {
		return "LogLevel undefined"
	}
	return slog.Level(*l).String()
}

type URLValue struct {
	URL *url.URL
}

func (v *URLValue) String() string {
	if v.URL == nil {
		return "URL undefined"
	}
	return v.URL.String()
}

func (v *URLValue) Set(s string) error {
	// if v.URL != nil {
	// 	return errors.New("url value already set")
	// }
	if u, err := url.Parse(s); err != nil {
		return err
	} else {
		*v.URL = *u
	}
	return nil
}

type TapHandlerValue struct {
	TapHandler *config.TapHandler
}

func newTapHandlerValue(value *config.TapHandler, p **config.TapHandler) *TapHandlerValue {
	if value != nil {
		*p = value
	}
	return &TapHandlerValue{
		TapHandler: *p,
	}
}

func (v *TapHandlerValue) Set(value string) error {
	if v == nil {
		return errors.New("TapHandlerValue is nil")
	}
	th, err := config.LoadTapHandler(value)
	if err != nil {
		return fmt.Errorf("error loading tap handler file: %w", err)
	}
	if th == nil {
		return errors.New("LoadTapHandler returned nil")
	}
	*v.TapHandler = *th
	return nil
}

func (v *TapHandlerValue) String() string {
	if v == nil || v.TapHandler == nil {
		return "error: unintialized"
	}
	b, err := yaml.Marshal(v.TapHandler)
	if err != nil {
		return fmt.Sprintf("error: %s", err.Error())
	}
	return string(b)
}

func newURLValue(value *url.URL, p **url.URL) *URLValue {
	*p = value
	return &URLValue{
		URL: *p,
	}
}

func (fs *FlagSet) URLVar(u **url.URL, name string, defaultValue *url.URL, usage string) {
	fs.FlagSet.Var(newURLValue(defaultValue, u), name, usage)
}

func (fs *FlagSet) TapHandlerVar(th **config.TapHandler, name string, defaultValue *config.TapHandler, usage string) {
	fs.FlagSet.Var(newTapHandlerValue(defaultValue, th), name, usage)
}

func (fs *FlagSet) LogLevelVar(l *slog.Level, name string, defaultValue slog.Level, usage string) {
	fs.FlagSet.Var(newLogLevelValue(defaultValue, l), name, usage)
}

func (fs *FlagSet) LogFormatVar(f *string, name string, defaultValue string, usage string) {
	fs.FlagSet.Var(newLogFormatValue(defaultValue, f), name, usage)
}
