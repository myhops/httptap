package serve

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/myhops/httptap"
	"github.com/myhops/httptap/command"
	"github.com/myhops/httptap/command/values"
	"github.com/myhops/httptap/config"
	"github.com/myhops/httptap/tap"
)

var (
	ErrNoTapDefined    = errors.New("no tap defined")
	ErrNoPatterns      = errors.New("no patterns defined")
	ErrShutdownTimeout = errors.New("server shutdown timed out")
	ErrTooFewArguments = errors.New("too few arguments")
)

type ServeCmd struct {
	GlobalCmd *command.GlobalCmd

	TapHandlerConfig *config.TapHandler

	Address  string
	Upstream *url.URL
}

func NewServeCmd(global *command.GlobalCmd) *ServeCmd {
	return &ServeCmd{
		GlobalCmd:        global,
		TapHandlerConfig: &config.TapHandler{},
	}
}

func mustURL(u string) *url.URL {
	uu, err := url.Parse(u)
	if err != nil {
		panic(fmt.Sprintf("mustURL error: %s", err.Error()))
	}
	return uu
}

func (c *ServeCmd) Flags(fs *values.FlagSet) {

	fs.URLVar(&c.Upstream, "upstream", mustURL("http://localhost:18080"), "upstream service")
	fs.TapHandlerVar(&c.TapHandlerConfig, "tap-config-file", nil, "Tap handlers config file")
	fs.StringVar(&c.Address, "address", ":8080", "listen address")
}

func (c *ServeCmd) createLogTap() (httptap.Tap, error) {
	tt := tap.NewLogTap(c.GlobalCmd.Logger, slog.LevelInfo)
	return tt, nil
}

func (c *ServeCmd) createTap(tcfg *config.Tap) (httptap.Tap, error) {
	var tt httptap.Tap
	if tcfg.LogTap != nil {
		d, err := c.createLogTap()
		if err != nil {
			return nil, err
		}
		tt = d
	}
	return tt, nil
}

func (c *ServeCmd) addTaps(p *httptap.Proxy) error {
	logger := c.GlobalCmd.Logger.With(slog.String("step", "addTaps"))
	if c.TapHandlerConfig == nil {
		logger.Info("no tap handler configured")
		return ErrNoTapDefined
	}
	if len(c.TapHandlerConfig.Taps) == 0 {
		logger.Info("no taps found")
	}
	// Add the taps
	for _, tcfg := range c.TapHandlerConfig.Taps {
		logger.Info("adding tap", slog.String("name", tcfg.Name))
		t, err := c.createTap(tcfg)
		if err != nil || t == nil {
			return err
		}
		logger.Info("adding tap to pattern", slog.Any("pattern", tcfg.Patterns))
		p.Tap(tcfg.Patterns, t, c.getTapOptions(tcfg)...)
	}
	return nil
}

func (c *ServeCmd) getTapOptions(tcfg *config.Tap) httptap.TapOptions {
	logger := c.GlobalCmd.Logger.With(slog.String("step", "getTapOptions"))
	var opts httptap.TapOptions
	if o := tcfg.Header.Exclude; len(o) > 0 {
		logger.Info("exclude headers", slog.Any("headers", o))
		opts = append(opts, httptap.WithExcludeHeaders(o))
	}
	if o := tcfg.Header.Include; len(o) > 0 {
		logger.Info("include headers", slog.Any("headers", o))
		opts = append(opts, httptap.WithIncludeHeaders(o))
	}
	mustMarshal := func(obj any) []byte {
		var res []byte
		res, err := json.Marshal(obj)
		if err != nil {
			panic("json patch marshal failed")
		}
		return res
	}
	if tcfg.RequestIn != nil {
		logger.Info("setting request body in")
		opts = append(opts, httptap.WithRequestBody(tcfg.RequestIn.Body))
		opts = append(opts, httptap.WithRequestJSON(tcfg.RequestIn.BodyJSON))
		if tcfg.RequestIn.BodyPatch != nil {
			logger.Info("addding response body patch")
			opts = append(opts, httptap.WithRequestBodyPatch(mustMarshal(tcfg.RequestIn.BodyPatch)))
		}
		
	}
	if tcfg.Response != nil {
		logger.Info("setting request body out")
		opts = append(opts, httptap.WithResponseBody(tcfg.Response.Body))
		opts = append(opts, httptap.WithResponseJSON(tcfg.Response.BodyJSON))
		if tcfg.Response.BodyPatch != nil {
			logger.Info("addding response body patch")
			opts = append(opts, httptap.WithResponseBodyPatch(mustMarshal(tcfg.Response.BodyPatch)))
		}
	}
	return opts
}

func (c *ServeCmd) Run(ctx context.Context) error {
	logger := c.GlobalCmd.Logger
	logger.Debug("debug enabled")
	// Create the proxy.
	p, err := httptap.New(c.Upstream.String(), httptap.WithLogger(c.GlobalCmd.Logger))
	if err != nil {
		return err
	}

	// Add the taps.
	c.addTaps(p)

	// Create the server.
	srv := &http.Server{
		Handler:           p,
		Addr:              c.Address,
		BaseContext:       func(_ net.Listener) context.Context { return ctx },
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
	}
	logger.Info("starting server", slog.String("address", c.Address))
	go srv.ListenAndServe()
	<-ctx.Done()
	logger.Info("Ctx Done")

	logger.Info("shutting down server")
	shutdownCtx, shudownCancel := context.WithTimeoutCause(context.Background(), 10*time.Second, ErrShutdownTimeout)
	defer shudownCancel()
	err = srv.Shutdown(shutdownCtx)
	if err != nil {
		logger.Error("shutdown return with error", slog.String("err", err.Error()))
		return err
	}
	logger.Info("server shut down")
	return nil
}
