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
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestAddEpisodeHandler(t *testing.T) {
	t.Run("will return a 404", func(t *testing.T) {
		t.Run("if the show title is empty", func(t *testing.T) {
			s, err := NewShowMgmtService(ServiceConfig{
				Logger: zap.L(),
			})
			if !assert.Nil(t, err) {
				return
			}

			req := httptest.NewRequest(http.MethodPost, "/show//season/1/episode/1", strings.NewReader("hello"))

			resp, err := s.test(req, 5)
			if !assert.Nil(t, err) {
				return
			}

			if !assert.Equal(t, fiber.StatusNotFound, resp.StatusCode) {
				return
			}
		})

		t.Run("if the season is empty", func(t *testing.T) {
			s, err := NewShowMgmtService(ServiceConfig{
				Logger: zap.L(),
			})
			if !assert.Nil(t, err) {
				return
			}

			req := httptest.NewRequest(http.MethodPost, "/show/1/season//episode/1", strings.NewReader("hello"))

			resp, err := s.test(req, 5)
			if !assert.Nil(t, err) {
				return
			}

			if !assert.Equal(t, fiber.StatusNotFound, resp.StatusCode) {
				return
			}
		})

		t.Run("if the episode is empty", func(t *testing.T) {
			s, err := NewShowMgmtService(ServiceConfig{
				Logger: zap.L(),
			})
			if !assert.Nil(t, err) {
				return
			}

			req := httptest.NewRequest(http.MethodPost, "/show/1/season/1/episode/", strings.NewReader("hello"))

			resp, err := s.test(req, 5)
			if !assert.Nil(t, err) {
				return
			}

			if !assert.Equal(t, fiber.StatusNotFound, resp.StatusCode) {
				return
			}
		})
	})

	t.Run("will return a 400", func(t *testing.T) {
		t.Run("if the show title is all spaces", func(t *testing.T) {
			s, err := NewShowMgmtService(ServiceConfig{
				Logger: zap.L(),
			})
			if !assert.Nil(t, err) {
				return
			}

			req := httptest.NewRequest(http.MethodPost, "/show/+++/season/1/episode/1", strings.NewReader("hello"))

			resp, err := s.test(req, 5)
			if !assert.Nil(t, err) {
				return
			}

			if !assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode) {
				return
			}
		})

		t.Run("if the season is all spaces", func(t *testing.T) {
			s, err := NewShowMgmtService(ServiceConfig{
				Logger: zap.L(),
			})
			if !assert.Nil(t, err) {
				return
			}

			req := httptest.NewRequest(http.MethodPost, "/show/1/season/+++/episode/1", strings.NewReader("hello"))

			resp, err := s.test(req, 5)
			if !assert.Nil(t, err) {
				return
			}

			if !assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode) {
				return
			}
		})

		t.Run("if the episode is all spaces", func(t *testing.T) {
			s, err := NewShowMgmtService(ServiceConfig{
				Logger: zap.L(),
			})
			if !assert.Nil(t, err) {
				return
			}

			req := httptest.NewRequest(http.MethodPost, "/show/1/season/1/episode/+++", strings.NewReader("hello"))

			resp, err := s.test(req, 5)
			if !assert.Nil(t, err) {
				return
			}

			if !assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode) {
				return
			}
		})
	})

	t.Run("will return a 200", func(t *testing.T) {
		t.Run("if provided request is completely valid", func(t *testing.T) {})
	})
}
