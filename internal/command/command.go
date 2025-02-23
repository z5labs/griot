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

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/z5labs/bedrock"
	"github.com/z5labs/bedrock/appbuilder"
	"github.com/z5labs/humus"
	"go.opentelemetry.io/otel"
)

type Option func(*App)

func Args(args ...string) Option {
	return func(a *App) {
		a.cmd.SetArgs(args)
	}
}

func Short(desription string) Option {
	return func(a *App) {
		a.cmd.Short = desription
	}
}

func Flags(f func(*pflag.FlagSet)) Option {
	return func(a *App) {
		f(a.cmd.Flags())
	}
}

func PersistentFlags(f func(*pflag.FlagSet)) Option {
	return func(a *App) {
		f(a.cmd.PersistentFlags())
	}
}

func Sub(sub *App) Option {
	return func(a *App) {
		a.cmd.AddCommand(sub.cmd)
	}
}

type Validator interface {
	Validate(context.Context) error
}

type ValidatorFunc func(context.Context) error

func (f ValidatorFunc) Validate(ctx context.Context) error {
	return f(ctx)
}

func ValidateAll(ctx context.Context, vs ...Validator) error {
	errs := make([]error, 0, len(vs))
	for _, v := range vs {
		err := v.Validate(ctx)
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

type Handler interface {
	Handle(context.Context) error
}

func Handle[T Validator](f func(context.Context, T) (Handler, error)) Option {
	return func(a *App) {
		a.cmd.RunE = func(cmd *cobra.Command, args []string) error {
			spanCtx, span := otel.Tracer("command").Start(cmd.Context(), "cobra.Command.RunE")
			defer span.End()

			var cfg T
			err := decodeFlags(cmd.Flags(), &cfg)
			if err != nil {
				return err
			}

			err = cfg.Validate(spanCtx)
			if err != nil {
				return err
			}

			h, err := f(spanCtx, cfg)
			if err != nil {
				return err
			}
			return h.Handle(spanCtx)
		}
	}
}

type App struct {
	cmd *cobra.Command
}

func NewApp(name string, opts ...Option) *App {
	a := &App{
		cmd: &cobra.Command{
			Use:           name,
			SilenceErrors: true,
			SilenceUsage:  true,
		},
	}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

func Run[T any](r io.Reader, f func(context.Context, T) (*App, error)) {
	builder := bedrock.AppBuilderFunc[T](func(ctx context.Context, cfg T) (bedrock.App, error) {
		return nil, nil
	})
	runner := humus.NewRunner(appbuilder.FromConfig(builder))
	runner.Run(context.Background(), humus.DefaultConfig())
}

func (a *App) Run(ctx context.Context) error {
	return a.cmd.ExecuteContext(ctx)
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
