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
	"os"
	"strings"

	"github.com/z5labs/griot/cmd/internal/command"
	"github.com/z5labs/griot/services/content"
	"github.com/z5labs/griot/services/content/contentpb"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.opentelemetry.io/otel"
)

func New() *cobra.Command {
	return command.New(
		"upload",
		command.Short("Upload content"),
		command.Flags(func(fs *pflag.FlagSet) {
			fs.String("name", "", "Provide an optional name to help identify this content later.")
			fs.String("media-type", "", "Specify the content Media Type.")
			fs.String("source-file", "", "Specify the content source file.")
			fs.String("hash-func", "sha256", "Specify hash function used for calculating content checksum.")
		}),
		command.Handle(initUploadHandler),
	)
}

type config struct {
	command.LoggingConfig `flag:",squash"`

	Name       string `flag:"name"`
	MediaType  string `flag:"media-type"`
	SourceFile string `flag:"source-file"`
	HashFunc   string `flag:"hash-func"`
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
	var hasher hash.Hash
	switch cfg.HashFunc {
	case "sha256":
		hasher = sha256.New()
	default:
		return nil, fmt.Errorf("unsupported hash function: %s", cfg.HashFunc)
	}

	src, err := openSourceFile(cfg.SourceFile)
	if err != nil {
		return nil, err
	}

	h := &handler{
		log:         command.Logger(cfg.LoggingConfig),
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
		Metadata: &contentpb.Metadata{
			Name: &h.contentName,
			Checksum: &contentpb.Checksum{
				HashFunc: contentpb.HashFunc_SHA256.Enum(),
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

func openSourceFile(filename string) (*os.File, error) {
	filename = strings.TrimSpace(filename)
	if len(filename) == 0 {
		return os.Stdin, nil
	}
	return os.Open(filename)
}
