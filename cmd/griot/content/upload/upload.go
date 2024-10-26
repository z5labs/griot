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
	"net/http"
	"os"
	"slices"
	"strings"

	"github.com/z5labs/griot/internal/command"
	"github.com/z5labs/griot/services/content"
	"github.com/z5labs/griot/services/content/contentpb"

	"github.com/spf13/pflag"
	"github.com/z5labs/humus"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
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

type hasher interface {
	hash.Hash

	HashFunc() contentpb.HashFunc
}

type sha256Hasher struct {
	hash.Hash
}

func (sha256Hasher) HashFunc() contentpb.HashFunc {
	return contentpb.HashFunc_SHA256
}

type handler struct {
	log *slog.Logger

	contentName string
	mediaType   string
	hasher      hasher
	src         io.ReadSeeker
	out         io.Writer

	content uploadClient
}

func initUploadHandler(ctx context.Context, cfg config) (command.Handler, error) {
	spanCtx, span := otel.Tracer("upload").Start(ctx, "initUploadHandler")
	defer span.End()

	log := humus.Logger("upload")

	hashFunc, exists := contentpb.HashFunc_value[cfg.HashFunc]
	if !exists {
		return nil, UnknownHashFuncError{
			Value: cfg.HashFunc,
		}
	}

	var contentHasher hasher
	switch contentpb.HashFunc(hashFunc) {
	case contentpb.HashFunc_SHA256:
		contentHasher = sha256Hasher{Hash: sha256.New()}
	default:
		return nil, fmt.Errorf("unsupported hash function: %s", cfg.HashFunc)
	}

	src, err := os.Open(cfg.SourceFile)
	if err != nil {
		log.ErrorContext(spanCtx, "failed to open source file", slog.String("error", err.Error()))
		return nil, err
	}

	hc := &http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}

	h := &handler{
		log:         log,
		contentName: cfg.Name,
		mediaType:   cfg.MediaType,
		hasher:      contentHasher,
		src:         src,
		out:         os.Stdout,
		content:     content.NewClient(hc),
	}
	return h, nil
}

type FailedToSeekReadBytesError struct {
	BytesRead   int64
	BytesSeeked int64
}

func (e FailedToSeekReadBytesError) Error() string {
	return fmt.Sprintf("bytes read do not match bytes seeked: %d:%d", e.BytesRead, e.BytesSeeked)
}

func (h *handler) Handle(ctx context.Context) error {
	spanCtx, span := otel.Tracer("upload").Start(ctx, "handler.Handle")
	defer span.End()

	bytesRead, err := io.Copy(h.hasher, h.src)
	if err != nil {
		span.RecordError(err)
		h.log.ErrorContext(spanCtx, "failed to compute hash", slog.String("error", err.Error()))
		return err
	}

	bytesSeeked, err := h.src.Seek(0, 0)
	if err != nil {
		span.RecordError(err)
		h.log.ErrorContext(spanCtx, "failed to perform seek on the source file", slog.String("error", err.Error()))
		return err
	}
	if bytesRead != bytesSeeked {
		err = FailedToSeekReadBytesError{
			BytesRead:   bytesRead,
			BytesSeeked: bytesSeeked,
		}

		span.RecordError(err)
		h.log.ErrorContext(spanCtx, "failed to seek to the start of the source file", slog.String("error", err.Error()))
		return err
	}

	mediaType, params, _ := mime.ParseMediaType(h.mediaType)

	req := &content.UploadContentRequest{
		Metadata: &contentpb.Metadata{
			Name: &h.contentName,
			MediaType: &contentpb.MediaType{
				Type:       &mediaType,
				Parameters: params,
			},
			Checksum: &contentpb.Checksum{
				HashFunc: h.hasher.HashFunc().Enum(),
				Hash:     h.hasher.Sum(nil),
			},
		},
		Body: h.src,
	}
	resp, err := h.content.UploadContent(spanCtx, req)
	if err != nil {
		h.log.ErrorContext(spanCtx, "failed to upload content", slog.String("error", err.Error()))
		return err
	}

	enc := json.NewEncoder(h.out)
	return enc.Encode(resp)
}
