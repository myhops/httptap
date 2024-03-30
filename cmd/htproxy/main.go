package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"

	"github.com/myhops/httptap/command"
	"github.com/myhops/httptap/command/serve"
	"github.com/myhops/httptap/command/values"
)

type cmd struct {
	fs *values.FlagSet
	gc *command.GlobalCmd
	sc *serve.ServeCmd
}

func newCmd() *cmd {
	gc := &command.GlobalCmd{}
	res := &cmd{
		fs: values.NewFlagSet("htproxy", flag.ExitOnError),
		gc: gc,
		sc:serve.NewServeCmd(gc),
	}
	res.gc.Flags(res.fs)
	res.sc.Flags(res.fs)
	return res
}

func (c *cmd) init() error {
	if err := c.gc.Init(); err != nil {
		return err
	}
	return nil
}

func (c *cmd) run(ctx context.Context) error {
	return c.sc.Run(ctx)
}

func run(args []string) error {
	cmd := newCmd()
	if err := cmd.fs.Parse(args[1:]); err != nil {
		return err
	}

	// Init the commands.
	if err := cmd.init(); err != nil {
		return fmt.Errorf("error calling cmd.init: %w", err)
	}
	// We have a logger.
	slog.SetDefault(cmd.gc.Logger)
	slog.SetLogLoggerLevel(cmd.gc.LogLevel.Level())
	cmd.sc.GlobalCmd = cmd.gc

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	return cmd.run(ctx)
}

func main() {
	err := run(os.Args)
	if err != nil {
		slog.Error("run returned error", slog.String("err", err.Error()))
	}
}
