// Copyright 2022 Z5Labs and Contributors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package http

import (
	"context"
	"errors"
	"net"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"go.uber.org/zap"
)

type ShowMgmtService struct {
	log      *zap.Logger
	handlers []func(*ShowMgmtService) func(*fiber.App)
}

type ServiceConfig struct {
	Logger *zap.Logger
}

// NewShowMgmtService
func NewShowMgmtService(cfg ServiceConfig) (*ShowMgmtService, error) {
	s := &ShowMgmtService{
		log: cfg.Logger,
	}
	return s, nil
}

// WithHandler
func (s *ShowMgmtService) WithHandler(h func(*ShowMgmtService) func(*fiber.App)) *ShowMgmtService {
	s.handlers = append(s.handlers, h)
	return s
}

// Serve
func (s *ShowMgmtService) Serve(ctx context.Context, ls net.Listener) error {
	app := fiber.New()

	app.Use(pprof.New())
	app.Use(func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNotFound)
	})

	for _, h := range s.handlers {
		h(s)(app)
	}

	errCh := make(chan error, 1)
	go func() {
		defer close(errCh)
		err := app.Listener(ls)
		if err == nil || errors.Is(err, context.Canceled) {
			return
		}
		errCh <- err
	}()

	select {
	case <-ctx.Done():
		err := app.Shutdown()
		if err == nil {
			return nil
		}
		zap.L().Error("failed to shutdown http server", zap.Error(err))
		return err
	case err := <-errCh:
		if err == nil {
			return nil
		}
		zap.L().Error("received unexpected error from http server", zap.Error(err))
		return err
	}
}
