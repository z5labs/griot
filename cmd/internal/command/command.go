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
	"log/slog"
	"os"
	"os/signal"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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
		c.RunE = func(cmd *cobra.Command, args []string) error {
			m := make(map[string]any)
			cmd.Flags().VisitAll(func(f *pflag.Flag) {
				m[f.Name] = f.Value
			})

			var cfg T
			dec, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
				Result:  &cfg,
				TagName: "flag",
			})
			if err != nil {
				return err
			}

			err = dec.Decode(m)
			if err != nil {
				return err
			}

			h, err := f(cmd.Context(), cfg)
			if err != nil {
				return err
			}
			return h.Handle(cmd.Context())
		}
	}
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
}
