// Copyright 2024 Z5Labs and Contributors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package command

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

var DefaultMinLogLevel LogLevel

func init() {
	DefaultMinLogLevel.Set(slog.LevelWarn.String())
}

type LogLevel slog.LevelVar

func (l *LogLevel) Set(v string) error {
	return (*slog.LevelVar)(l).UnmarshalText([]byte(v))
}

func (l *LogLevel) String() string {
	return (*slog.LevelVar)(l).Level().String()
}

func (l *LogLevel) Type() string {
	return "slog.Level"
}

type LoggingConfig struct {
	Level *LogLevel `flag:"log-level"`
}

func Logger(cfg LoggingConfig) *slog.Logger {
	return slog.New(slog.NewJSONHandler(
		os.Stderr,
		&slog.HandlerOptions{
			AddSource: true,
			Level:     (*slog.LevelVar)(cfg.Level).Level(),
		},
	))
}

type OTelConfig struct {
	Enabled          bool   `flag:"enable-otel"`
	TraceDestination string `flag:"trace-out"`
}

type Option func(*cobra.Command)

func Short(desription string) Option {
	return func(c *cobra.Command) {
		c.Short = desription
	}
}

func Flags(f func(*pflag.FlagSet)) Option {
	return func(c *cobra.Command) {
		f(c.Flags())
	}
}

func PersistentFlags(f func(*pflag.FlagSet)) Option {
	return func(c *cobra.Command) {
		f(c.PersistentFlags())
	}
}

func Sub(sub *cobra.Command) Option {
	return func(c *cobra.Command) {
		c.AddCommand(sub)
	}
}

type Handler interface {
	Handle(context.Context) error
}

func Handle[T any](f func(context.Context, T) (Handler, error)) Option {
	return func(c *cobra.Command) {
		var postRunHooks []func(context.Context) error

		c.PreRunE = func(cmd *cobra.Command, args []string) error {
			var cfg OTelConfig
			err := decodeFlags(cmd.Flags(), &cfg)
			if err != nil {
				return err
			}

			if !cfg.Enabled {
				return nil
			}

			otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
				propagation.Baggage{},
				propagation.TraceContext{},
			))

			out, err := openDestination(cfg.TraceDestination)
			if err != nil {
				return err
			}

			tp, err := initTracerProvider(cmd.Context(), out)
			if err != nil {
				return err
			}

			otel.SetTracerProvider(tp)
			postRunHooks = append(postRunHooks, tp.Shutdown)
			return nil
		}

		c.RunE = func(cmd *cobra.Command, args []string) error {
			spanCtx, span := otel.Tracer("command").Start(cmd.Context(), "cobra.Command.RunE")
			defer span.End()

			var cfg T
			err := decodeFlags(cmd.Flags(), &cfg)
			if err != nil {
				return err
			}

			h, err := f(spanCtx, cfg)
			if err != nil {
				return err
			}
			return h.Handle(spanCtx)
		}

		c.PostRunE = func(cmd *cobra.Command, args []string) error {
			var errs []error
			for _, hook := range postRunHooks {
				err := hook(cmd.Context())
				if err == nil {
					continue
				}
				errs = append(errs, err)
			}
			if len(errs) == 0 {
				return nil
			}
			if len(errs) == 1 {
				return errs[0]
			}
			return errors.Join(errs...)
		}
	}
}

func decodeFlags(fs *pflag.FlagSet, v interface{}) error {
	m := make(map[string]any)
	fs.VisitAll(func(f *pflag.Flag) {
		m[f.Name] = f.Value
	})

	dec, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:  v,
		TagName: "flag",
	})
	if err != nil {
		return err
	}
	return dec.Decode(m)
}

func openDestination(filename string) (*os.File, error) {
	filename = strings.TrimSpace(filename)
	if len(filename) > 0 {
		return os.Create(filename)
	}
	return os.CreateTemp("", "griot_traces_*")
}

func initTracerProvider(ctx context.Context, out io.Writer) (*sdktrace.TracerProvider, error) {
	rsc, err := resource.Detect(ctx)
	if err != nil {
		return nil, err
	}

	exp, err := stdouttrace.New(
		stdouttrace.WithWriter(out),
	)
	if err != nil {
		return nil, err
	}

	sp := sdktrace.NewSimpleSpanProcessor(
		exp,
	)

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithResource(rsc),
		sdktrace.WithSpanProcessor(sp),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)
	return tp, nil
}

func New(use string, opts ...Option) *cobra.Command {
	cmd := &cobra.Command{
		Use: use,
	}
	for _, opt := range opts {
		opt(cmd)
	}
	return cmd
}

func Run(c *cobra.Command) {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	err := c.ExecuteContext(ctx)
	if err == nil {
		return
	}

	log := Logger(LoggingConfig{
		Level: &DefaultMinLogLevel,
	})
	log.Error("encountered unexpected error", slog.String("error", err.Error()))
}
