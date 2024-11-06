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

package content

import (
	"bytes"
	"context"
	_ "embed"

	"github.com/z5labs/griot/services/content/endpoint"

	"github.com/z5labs/humus/rest"
)

//go:embed config.yaml
var ConfigBytes []byte

func Run(f func(context.Context, AppConfig) (*App, error)) {
	rest.Run(bytes.NewReader(ConfigBytes), func(ctx context.Context, cfg AppConfig) (*rest.App, error) {
		app, err := f(ctx, cfg)
		if err != nil {
			return nil, err
		}
		return app.rest, nil
	})
}

type AppConfig struct {
	rest.Config `config:",squash"`
}

type App struct {
	rest *rest.App
}

func InitApp(ctx context.Context, cfg AppConfig) (*App, error) {
	ra := rest.New(
		rest.ListenOn(cfg.Http.Port),
		rest.Title(cfg.OpenApi.Title),
		rest.Version(cfg.OpenApi.Version),
		rest.Readiness(nil),
		rest.Liveness(nil),
		rest.RegisterEndpoint(endpoint.UploadContentV1()),
	)

	a := &App{
		rest: ra,
	}
	return a, nil
}

func (a *App) Run(ctx context.Context) error {
	return a.rest.Run(ctx)
}
