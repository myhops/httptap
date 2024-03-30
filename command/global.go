package command

import (
	"log/slog"
	"os"

	"github.com/myhops/httptap/command/values"
)

type GlobalCmd struct {
	LogFormat string
	logLevel  slog.Level

	Logger   *slog.Logger
	LogLevel *slog.LevelVar
}

func (c *GlobalCmd) Flags(fs *values.FlagSet) {
	fs.LogFormatVar(&c.LogFormat, "logformat", "text", "set the log format, text or json")
	fs.LogLevelVar(&c.logLevel, "loglevel", slog.LevelInfo, "set the log level to debug, info, warn or error")
}

func (c *GlobalCmd) Init() error {
	// Create a variable log level.
	c.LogLevel = &slog.LevelVar{}
	c.LogLevel.Set(c.logLevel)

	// Create the handler options
	ho := &slog.HandlerOptions{
		Level: c.LogLevel,
	}

	// Create the handler.
	var h slog.Handler
	switch c.LogFormat {
	case "json":
		h = slog.NewJSONHandler(os.Stdout, ho)
	default:
		h = slog.NewTextHandler(os.Stdout, ho)
	}

	// Create the logger.
	c.Logger = slog.New(h)
	return nil
}
