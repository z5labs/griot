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
	"encoding/json"
	"fmt"
	"hash"
	"io"
	"log/slog"
	"maps"
	"mime"
	"os"
	"slices"
	"strings"

	"github.com/z5labs/griot/cmd/internal/command"
	"github.com/z5labs/griot/services/content"
	"github.com/z5labs/griot/services/content/contentpb"

	"github.com/spf13/pflag"
	"github.com/z5labs/humus"
	"go.opentelemetry.io/otel"
)

func New(args ...string) *command.App {
	return command.NewApp(
		"upload",
		command.Args(args...),
		command.Short("Upload content"),
		command.Flags(func(fs *pflag.FlagSet) {
			fs.String("name", "", "Provide an optional name to help identify this content later.")
			fs.String("media-type", "", "Specify the content Media Type.")
			fs.String("source-file", "", "Specify the content source file.")
			fs.String(
				"hash-func",
				contentpb.HashFunc_SHA256.String(),
				fmt.Sprintf(
					"Specify hash function used for calculating content checksum. (values %s)",
					strings.Join(slices.Collect(maps.Values(contentpb.HashFunc_name)), ","),
				),
			)
		}),
		command.Handle(initUploadHandler),
	)
}

type config struct {
	Name       string `flag:"name"`
	MediaType  string `flag:"media-type"`
	SourceFile string `flag:"source-file"`
	HashFunc   string `flag:"hash-func"`
}

func (c config) Validate(ctx context.Context) error {
	validators := []command.Validator{
		validateMediaType(c.MediaType),
		validateSourceFile(c.SourceFile),
		validateHashFunc(c.HashFunc),
	}

	return command.ValidateAll(ctx, validators...)
}

func validateMediaType(mediaType string) command.ValidatorFunc {
	return func(ctx context.Context) error {
		if len(mediaType) == 0 {
			return command.InvalidFlagError{
				Name:  "media-type",
				Cause: command.ErrFlagRequired,
			}
		}
		_, _, err := mime.ParseMediaType(mediaType)
		if err != nil {
			return command.InvalidFlagError{
				Name:  "media-type",
				Cause: err,
			}
		}
		return nil
	}
}

func validateSourceFile(filename string) command.ValidatorFunc {
	return func(ctx context.Context) error {
		if len(filename) == 0 {
			return command.InvalidFlagError{
				Name:  "source-file",
				Cause: command.ErrFlagRequired,
			}
		}

		info, err := os.Stat(filename)
		if err != nil {
			return command.InvalidFlagError{
				Name:  "source-file",
				Cause: err,
			}
		}
		if info.IsDir() {
			return command.InvalidFlagError{
				Name:  "source-file",
				Cause: command.ErrMustBeAFile,
			}
		}
		return nil
	}
}

type UnknownHashFuncError struct {
	Value string
}

func (e UnknownHashFuncError) Error() string {
	return fmt.Sprintf("unknown hash func value: %s", e.Value)
}

func validateHashFunc(name string) command.ValidatorFunc {
	return func(ctx context.Context) error {
		if len(name) == 0 {
			return command.InvalidFlagError{
				Name:  "hash-func",
				Cause: command.ErrFlagRequired,
			}
		}

		_, found := contentpb.HashFunc_value[name]
		if !found {
			return command.InvalidFlagError{
				Name: "hash-func",
				Cause: UnknownHashFuncError{
					Value: name,
				},
			}
		}
		return nil
	}
}

type uploadClient interface {
	UploadContent(context.Context, *content.UploadContentRequest) (*content.UploadContentResponse, error)
}

type handler struct {
	log *slog.Logger

	contentName string
	mediaType   string
	hasher      hash.Hash
	src         io.ReadSeeker
	out         io.Writer

	content uploadClient
}

func initUploadHandler(ctx context.Context, cfg config) (command.Handler, error) {
	spanCtx, span := otel.Tracer("upload").Start(ctx, "initUploadHandler")
	defer span.End()

	log := humus.Logger("upload")

	var hasher hash.Hash
	hashFunc, exists := contentpb.HashFunc_value[cfg.HashFunc]
	if !exists {
		return nil, fmt.Errorf("unknown hash function: %s", cfg.HashFunc)
	}

	switch contentpb.HashFunc(hashFunc) {
	case contentpb.HashFunc_SHA256:
		hasher = sha256.New()
	default:
		return nil, fmt.Errorf("unsupported hash function: %s", cfg.HashFunc)
	}

	src, err := os.Open(cfg.SourceFile)
	if err != nil {
		log.ErrorContext(spanCtx, "failed to open source file", slog.String("error", err.Error()))
		return nil, err
	}

	h := &handler{
		log:         log,
		contentName: cfg.Name,
		mediaType:   cfg.MediaType,
		hasher:      hasher,
		src:         src,
		out:         os.Stdout,
		content:     content.NewClient(),
	}
	return h, nil
}

func (h *handler) Handle(ctx context.Context) error {
	spanCtx, span := otel.Tracer("upload").Start(ctx, "handler.Handle")
	defer span.End()

	_, err := io.Copy(h.hasher, h.src)
	if err != nil {
		h.log.ErrorContext(spanCtx, "failed to compute hash", slog.String("error", err.Error()))
		return err
	}

	_, err = h.src.Seek(0, 0)
	if err != nil {
		h.log.ErrorContext(spanCtx, "failed to seek to start of content source", slog.String("error", err.Error()))
		return err
	}

	req := &content.UploadContentRequest{
		Name:     h.contentName,
		HashFunc: contentpb.HashFunc_SHA256,
		Body:     h.src,
	}
	resp, err := h.content.UploadContent(spanCtx, req)
	if err != nil {
		h.log.ErrorContext(spanCtx, "failed to upload content", slog.String("error", err.Error()))
		return err
	}

	enc := json.NewEncoder(h.out)
	return enc.Encode(resp)
}
