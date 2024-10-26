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

package upload

import (
	"context"
	"crypto/sha256"
	"errors"
	"io"
	"log/slog"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/z5labs/griot/internal/command"
	"github.com/z5labs/griot/services/content"
	"github.com/z5labs/griot/services/content/contentpb"

	"github.com/stretchr/testify/assert"
	"github.com/z5labs/bedrock/pkg/noop"
)

func TestApp(t *testing.T) {
	t.Run("will return an error", func(t *testing.T) {
		t.Run("if the media type is not set", func(t *testing.T) {
			f, err := os.CreateTemp(t.TempDir(), "*")
			if !assert.Nil(t, err) {
				return
			}
			err = f.Close()
			if !assert.Nil(t, err) {
				return
			}

			app := New("--source-file", f.Name())
			err = app.Run(context.Background())

			var iferr command.InvalidFlagError
			if !assert.ErrorAs(t, err, &iferr) {
				return
			}
			if !assert.Equal(t, "media-type", iferr.Name) {
				return
			}
			if !assert.ErrorIs(t, iferr, command.ErrFlagRequired) {
				return
			}
		})

		t.Run("if the media type is invalid", func(t *testing.T) {
			f, err := os.CreateTemp(t.TempDir(), "*")
			if !assert.Nil(t, err) {
				return
			}
			err = f.Close()
			if !assert.Nil(t, err) {
				return
			}

			app := New("--media-type", "text/plain; hello=", "--source-file", f.Name())
			err = app.Run(context.Background())

			var iferr command.InvalidFlagError
			if !assert.ErrorAs(t, err, &iferr) {
				return
			}
			if !assert.Equal(t, "media-type", iferr.Name) {
				return
			}
			if !assert.ErrorIs(t, iferr, mime.ErrInvalidMediaParameter) {
				return
			}
		})

		t.Run("if the source file is not set", func(t *testing.T) {
			app := New("--media-type", "text/plain")
			err := app.Run(context.Background())

			var iferr command.InvalidFlagError
			if !assert.ErrorAs(t, err, &iferr) {
				return
			}
			if !assert.Equal(t, "source-file", iferr.Name) {
				return
			}
			if !assert.ErrorIs(t, iferr, command.ErrFlagRequired) {
				return
			}
		})

		t.Run("if the source file does not exist", func(t *testing.T) {
			app := New("--media-type", "text/plain", "--source-file", "test.txt")
			err := app.Run(context.Background())

			var iferr command.InvalidFlagError
			if !assert.ErrorAs(t, err, &iferr) {
				return
			}
			if !assert.Equal(t, "source-file", iferr.Name) {
				return
			}

			var perr *os.PathError
			if !assert.ErrorAs(t, err, &perr) {
				return
			}
		})

		t.Run("if the source file name is a directory instead of a file", func(t *testing.T) {
			dir := t.TempDir()

			app := New("--media-type", "text/plain", "--source-file", dir)
			err := app.Run(context.Background())

			var iferr command.InvalidFlagError
			if !assert.ErrorAs(t, err, &iferr) {
				return
			}
			if !assert.Equal(t, "source-file", iferr.Name) {
				return
			}
			if !assert.ErrorIs(t, iferr, command.ErrMustBeAFile) {
				return
			}
		})

		t.Run("if the hash func is set to empty value", func(t *testing.T) {
			f, err := os.CreateTemp(t.TempDir(), "*")
			if !assert.Nil(t, err) {
				return
			}
			err = f.Close()
			if !assert.Nil(t, err) {
				return
			}

			app := New("--media-type", "text/plain", "--source-file", f.Name(), "--hash-func", "")
			err = app.Run(context.Background())

			var iferr command.InvalidFlagError
			if !assert.ErrorAs(t, err, &iferr) {
				return
			}
			if !assert.Equal(t, "hash-func", iferr.Name) {
				return
			}
			if !assert.ErrorIs(t, iferr, command.ErrFlagRequired) {
				return
			}
		})

		t.Run("if the hash func is set to an unknown value", func(t *testing.T) {
			f, err := os.CreateTemp(t.TempDir(), "*")
			if !assert.Nil(t, err) {
				return
			}
			err = f.Close()
			if !assert.Nil(t, err) {
				return
			}

			app := New("--media-type", "text/plain", "--source-file", f.Name(), "--hash-func", "SHA")
			err = app.Run(context.Background())

			var iferr command.InvalidFlagError
			if !assert.ErrorAs(t, err, &iferr) {
				return
			}
			if !assert.Equal(t, "hash-func", iferr.Name) {
				return
			}

			var uerr UnknownHashFuncError
			if !assert.ErrorAs(t, iferr, &uerr) {
				return
			}
			if !assert.NotEmpty(t, uerr.Error()) {
				return
			}
			if !assert.Equal(t, "SHA", uerr.Value) {
				return
			}
		})
	})
}

func TestInitUploadHandler(t *testing.T) {
	t.Run("will return an error", func(t *testing.T) {
		t.Run("if a unknown hash function is provided", func(t *testing.T) {
			cfg := config{
				HashFunc: "SHA",
			}

			_, err := initUploadHandler(context.Background(), cfg)

			var uerr UnknownHashFuncError
			if !assert.ErrorAs(t, err, &uerr) {
				return
			}
			if !assert.NotEmpty(t, uerr.Error()) {
				return
			}
			if !assert.Equal(t, "SHA", uerr.Value) {
				return
			}
		})

		t.Run("if it fails to open the source file", func(t *testing.T) {
			cfg := config{
				HashFunc:   contentpb.HashFunc_SHA256.String(),
				SourceFile: filepath.Join(t.TempDir(), "test.txt"),
			}

			_, err := initUploadHandler(context.Background(), cfg)

			var perr *os.PathError
			if !assert.ErrorAs(t, err, &perr) {
				return
			}
			if !assert.Equal(t, "open", perr.Op) {
				return
			}
		})
	})
}

type readNoopSeeker func([]byte) (int, error)

func (f readNoopSeeker) Read(b []byte) (int, error) {
	return f(b)
}

func (f readNoopSeeker) Seek(offset int64, whence int) (int64, error) {
	return 0, nil
}

type seekNoopReader func(int64, int) (int64, error)

func (f seekNoopReader) Read(b []byte) (int, error) {
	return 0, io.EOF
}

func (f seekNoopReader) Seek(offset int64, whence int) (int64, error) {
	return f(offset, whence)
}

type uploadClientFunc func(context.Context, *content.UploadContentRequest) (*content.UploadContentResponse, error)

func (f uploadClientFunc) UploadContent(ctx context.Context, req *content.UploadContentRequest) (*content.UploadContentResponse, error) {
	return f(ctx, req)
}

func TestHandler_Handle(t *testing.T) {
	t.Run("will return an error", func(t *testing.T) {
		t.Run("if it fails to compute the content hash", func(t *testing.T) {
			readErr := errors.New("read failed")
			src := readNoopSeeker(func(b []byte) (int, error) {
				return 0, readErr
			})

			h := &handler{
				log:    slog.New(noop.LogHandler{}),
				hasher: sha256Hasher{Hash: sha256.New()},
				src:    src,
			}

			err := h.Handle(context.Background())
			if !assert.Equal(t, readErr, err) {
				return
			}
		})

		t.Run("if it fails to perform seek on the source", func(t *testing.T) {
			seekErr := errors.New("failed to seek")
			src := seekNoopReader(func(i1 int64, i2 int) (int64, error) {
				return 0, seekErr
			})

			h := &handler{
				log:    slog.New(noop.LogHandler{}),
				hasher: sha256Hasher{Hash: sha256.New()},
				src:    src,
			}

			err := h.Handle(context.Background())
			if !assert.Equal(t, seekErr, err) {
				return
			}
		})

		t.Run("if it fails to seek to the start of the source file", func(t *testing.T) {
			bytesSeeked := int64(10)
			src := seekNoopReader(func(i1 int64, i2 int) (int64, error) {
				return bytesSeeked, nil
			})

			h := &handler{
				log:    slog.New(noop.LogHandler{}),
				hasher: sha256Hasher{Hash: sha256.New()},
				src:    src,
			}

			err := h.Handle(context.Background())

			var ferr FailedToSeekReadBytesError
			if !assert.ErrorAs(t, err, &ferr) {
				return
			}
			if !assert.NotEmpty(t, ferr.Error()) {
				return
			}
			if !assert.Equal(t, int64(0), ferr.BytesRead) {
				return
			}
			if !assert.Equal(t, bytesSeeked, ferr.BytesSeeked) {
				return
			}
		})

		t.Run("if it fails to upload the content", func(t *testing.T) {
			src := strings.NewReader(``)

			uploadErr := errors.New("failed to upload")
			client := uploadClientFunc(func(ctx context.Context, ucr *content.UploadContentRequest) (*content.UploadContentResponse, error) {
				return nil, uploadErr
			})

			h := &handler{
				log:     slog.New(noop.LogHandler{}),
				hasher:  sha256Hasher{Hash: sha256.New()},
				src:     src,
				content: client,
			}

			err := h.Handle(context.Background())
			if !assert.Equal(t, uploadErr, err) {
				return
			}
		})
	})
}
