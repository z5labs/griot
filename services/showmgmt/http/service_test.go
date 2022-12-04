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
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func TestShowMgmtService(t *testing.T) {
	t.Run("will return a 404", func(t *testing.T) {
		t.Run("if an unknown path is requested", func(t *testing.T) {
			ls, err := net.Listen("tcp", "127.0.0.1:0")
			if !assert.Nil(t, err) {
				return
			}
			addr := fmt.Sprintf("http://%s/", ls.Addr().String())

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			s, err := NewShowMgmtService(ServiceConfig{
				Logger: zap.NewNop(),
			})
			if !assert.Nil(t, err) {
				return
			}

			g, gctx := errgroup.WithContext(ctx)
			g.Go(func() error {
				err = s.Serve(gctx, ls)
				return err
			})

			doneCh := make(chan struct{}, 1)
			g.Go(func() error {
				select {
				case <-gctx.Done():
					return nil
				case <-doneCh:
				}
				cancel()
				return nil
			})
			defer g.Wait()
			defer close(doneCh)

			req, err := http.NewRequestWithContext(gctx, http.MethodGet, addr, nil)
			if !assert.Nil(t, err) {
				return
			}

			resp, err := http.DefaultClient.Do(req)
			if !assert.Nil(t, err) {
				return
			}

			if !assert.Equal(t, http.StatusNotFound, resp.StatusCode) {
				return
			}
		})
	})
}
