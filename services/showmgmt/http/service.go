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
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/spf13/afero"
	"go.uber.org/zap"
)

type pathHandler struct {
	method  string
	path    string
	handler func(*fiber.Ctx) error
}

func post(path string, handler func(*fiber.Ctx) error) pathHandler {
	return pathHandler{
		method:  fiber.MethodPost,
		path:    path,
		handler: handler,
	}
}

// ShowMgmtService
type ShowMgmtService struct {
	app *fiber.App

	log       *zap.Logger
	validator *validator.Validate
	fs        afero.Fs
}

// ServiceConfig
type ServiceConfig struct {
	Logger *zap.Logger
	Dir    afero.Fs
}

// NewShowMgmtService
func NewShowMgmtService(cfg ServiceConfig) (*ShowMgmtService, error) {
	s := &ShowMgmtService{
		app:       fiber.New(),
		log:       cfg.Logger,
		validator: validator.New(),
		fs:        cfg.Dir,
	}
	s.app.Server().StreamRequestBody = true
	s.app.Use(cors.New())
	s.app.Use(pprof.New())

	phs := []pathHandler{
		addEpisodeHandler(s),
	}
	for _, ph := range phs {
		switch ph.method {
		case fiber.MethodPost:
			s.app.Post(ph.path, ph.handler)
		default:
			return nil, errors.New("unknown method")
		}
	}

	s.app.Use(func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNotFound)
	})

	return s, nil
}

// Serve
func (s *ShowMgmtService) Serve(ctx context.Context, ls net.Listener) error {
	errCh := make(chan error, 1)
	go func() {
		defer close(errCh)
		err := s.app.Listener(ls)
		if err == nil || errors.Is(err, context.Canceled) {
			return
		}
		errCh <- err
	}()

	select {
	case <-ctx.Done():
		err := s.app.Shutdown()
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

func (s *ShowMgmtService) test(req *http.Request, msTimeout ...int) (resp *http.Response, err error) {
	return s.app.Test(req, msTimeout...)
}
