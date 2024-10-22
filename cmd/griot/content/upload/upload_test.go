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
	"mime"
	"os"
	"testing"

	"github.com/z5labs/griot/cmd/internal/command"

	"github.com/stretchr/testify/assert"
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
