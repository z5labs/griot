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
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/z5labs/griot/internal/mimetype"
	"github.com/z5labs/griot/services/content/contentpb"

	"github.com/stretchr/testify/assert"
	"github.com/z5labs/humus/humuspb"
	"google.golang.org/protobuf/proto"
)

type httpClientFunc func(*http.Request) (*http.Response, error)

func (f httpClientFunc) Do(r *http.Request) (*http.Response, error) {
	return f(r)
}

type readFunc func([]byte) (int, error)

func (f readFunc) Read(b []byte) (int, error) {
	return f(b)
}

func TestClient_UploadContent(t *testing.T) {
	t.Run("will return an error", func(t *testing.T) {
		t.Run("if it fails to marshal the content metadata", func(t *testing.T) {
			hc := httpClientFunc(func(r *http.Request) (*http.Response, error) {
				defer r.Body.Close()
				_, err := io.Copy(io.Discard, r.Body)
				return nil, err
			})

			c := NewClient(hc, "")

			marshalErr := errors.New("failed to marshal proto")
			c.protoMarshal = func(m proto.Message) ([]byte, error) {
				return nil, marshalErr
			}

			_, err := c.UploadContent(context.Background(), &UploadContentRequest{})
			if !assert.Equal(t, marshalErr, err) {
				return
			}
		})

		t.Run("if it fails to read the content", func(t *testing.T) {
			hc := httpClientFunc(func(r *http.Request) (*http.Response, error) {
				defer r.Body.Close()
				_, err := io.Copy(io.Discard, r.Body)
				return nil, err
			})

			c := NewClient(hc, "")

			readErr := errors.New("failed to read")
			content := readFunc(func(b []byte) (int, error) {
				return 0, readErr
			})

			_, err := c.UploadContent(context.Background(), &UploadContentRequest{
				Metadata: &contentpb.Metadata{
					Checksum: &contentpb.Checksum{},
				},
				Content: content,
			})
			if !assert.Equal(t, readErr, err) {
				return
			}
		})

		t.Run("if it fails to do the http request", func(t *testing.T) {
			httpErr := errors.New("failed to do http request")
			hc := httpClientFunc(func(r *http.Request) (*http.Response, error) {
				return nil, httpErr
			})

			c := NewClient(hc, "")

			content := strings.NewReader(`hello world`)

			_, err := c.UploadContent(context.Background(), &UploadContentRequest{
				Metadata: &contentpb.Metadata{
					Checksum: &contentpb.Checksum{},
				},
				Content: content,
			})
			if !assert.Equal(t, httpErr, err) {
				return
			}
		})

		t.Run("if the context is cancelled while doing http request", func(t *testing.T) {
			hc := httpClientFunc(func(r *http.Request) (*http.Response, error) {
				defer r.Body.Close()
				_, err := io.Copy(io.Discard, r.Body)
				return nil, err
			})

			c := NewClient(hc, "")

			ctx, cancel := context.WithCancel(context.Background())
			content := readFunc(func(b []byte) (int, error) {
				cancel()
				n := copy(b, make([]byte, len(b)))
				return n, nil
			})

			_, err := c.UploadContent(ctx, &UploadContentRequest{
				Metadata: &contentpb.Metadata{
					Checksum: &contentpb.Checksum{},
				},
				Content: content,
			})
			if !assert.Equal(t, context.Canceled, err) {
				return
			}
		})

		t.Run(fmt.Sprintf("if the response content type is not %s", mimetype.Protobuf), func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer r.Body.Close()
				_, err := io.Copy(io.Discard, r.Body)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				w.Header().Set("Content-Type", "application/json")
			}))

			hc := http.DefaultClient

			c := NewClient(hc, srv.URL)

			content := strings.NewReader("hello world")

			_, err := c.UploadContent(context.Background(), &UploadContentRequest{
				Metadata: &contentpb.Metadata{
					Checksum: &contentpb.Checksum{},
				},
				Content: content,
			})

			var uerr UnsupportedResponseContentTypeError
			if !assert.ErrorAs(t, err, &uerr) {
				return
			}
			if !assert.NotEmpty(t, uerr.Error()) {
				return
			}
			if !assert.Equal(t, "application/json", uerr.ContentType) {
				return
			}
		})

		t.Run("if it fails to read the http response body", func(t *testing.T) {
			readErr := errors.New("failed to read")
			respBody := readFunc(func(b []byte) (int, error) {
				return 0, readErr
			})
			hc := httpClientFunc(func(r *http.Request) (*http.Response, error) {
				defer r.Body.Close()
				_, err := io.Copy(io.Discard, r.Body)
				if err != nil {
					return nil, err
				}

				resp := &http.Response{
					Header: make(http.Header),
					Body:   io.NopCloser(respBody),
				}
				resp.Header.Set("Content-Type", mimetype.Protobuf)

				return resp, nil
			})

			c := NewClient(hc, "")

			content := strings.NewReader("hello world")

			_, err := c.UploadContent(context.Background(), &UploadContentRequest{
				Metadata: &contentpb.Metadata{
					Checksum: &contentpb.Checksum{},
				},
				Content: content,
			})
			if !assert.Equal(t, readErr, err) {
				return
			}
		})

		t.Run("if it fails to unmarshal the humuspb.Status response", func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer r.Body.Close()
				_, err := io.Copy(io.Discard, r.Body)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				b, err := proto.Marshal(&humuspb.Status{
					Code: humuspb.Code_INTERNAL.Enum(),
				})
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				w.Header().Set("Content-Type", mimetype.Protobuf)
				w.WriteHeader(http.StatusInternalServerError)
				io.Copy(w, bytes.NewReader(b))
			}))

			hc := http.DefaultClient

			c := NewClient(hc, srv.URL)

			unmarshalErr := errors.New("failed to unmarshal")
			c.protoUnmarshal = func(b []byte, m proto.Message) error {
				return unmarshalErr
			}

			content := strings.NewReader("hello world")

			_, err := c.UploadContent(context.Background(), &UploadContentRequest{
				Metadata: &contentpb.Metadata{
					Checksum: &contentpb.Checksum{},
				},
				Content: content,
			})
			if !assert.Equal(t, unmarshalErr, err) {
				return
			}
		})

		t.Run("if the response code is not HTTP 200", func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer r.Body.Close()
				_, err := io.Copy(io.Discard, r.Body)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				b, err := proto.Marshal(&humuspb.Status{
					Code: humuspb.Code_INTERNAL.Enum(),
				})
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				w.Header().Set("Content-Type", mimetype.Protobuf)
				w.WriteHeader(http.StatusInternalServerError)
				io.Copy(w, bytes.NewReader(b))
			}))

			hc := http.DefaultClient

			c := NewClient(hc, srv.URL)

			content := strings.NewReader("hello world")

			_, err := c.UploadContent(context.Background(), &UploadContentRequest{
				Metadata: &contentpb.Metadata{
					Checksum: &contentpb.Checksum{},
				},
				Content: content,
			})

			var status *humuspb.Status
			if !assert.ErrorAs(t, err, &status) {
				return
			}
			if !assert.Equal(t, humuspb.Code_INTERNAL, status.GetCode()) {
				return
			}
		})

		t.Run("if it fails to unmarshal the contentpb.UploadContentV1Response response", func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer r.Body.Close()
				_, err := io.Copy(io.Discard, r.Body)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				b, err := proto.Marshal(&contentpb.UploadContentV1Response{})
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				w.Header().Set("Content-Type", mimetype.Protobuf)
				w.WriteHeader(http.StatusOK)
				io.Copy(w, bytes.NewReader(b))
			}))

			hc := http.DefaultClient

			c := NewClient(hc, srv.URL)

			unmarshalErr := errors.New("failed to unmarshal")
			c.protoUnmarshal = func(b []byte, m proto.Message) error {
				return unmarshalErr
			}

			content := strings.NewReader("hello world")

			_, err := c.UploadContent(context.Background(), &UploadContentRequest{
				Metadata: &contentpb.Metadata{
					Checksum: &contentpb.Checksum{},
				},
				Content: content,
			})
			if !assert.Equal(t, unmarshalErr, err) {
				return
			}
		})
	})
}
