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
	"io"
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/z5labs/griot/cmd/internal/command"
)

func New() *cobra.Command {
	return command.New(
		"upload",
		command.Short("Upload content"),
		command.Flags(func(fs *pflag.FlagSet) {
			fs.String("name", "", "Provide an optional name to help identify this content later.")
			fs.String("media-type", "", "Specify the content Media Type.")
			fs.String("source-file", "", "Specify the content source file.")
		}),
		command.Handle(initUploadHandler),
	)
}

type config struct {
	command.LoggingConfig `flag:",squash"`

	Name       string `flag:"name"`
	MediaType  string `flag:"media-type"`
	SourceFile string `flag:"source-file"`
}

type handler struct {
	log *slog.Logger

	contentName string
	mediaType   string
	filename    string
	out         io.Writer
}

func initUploadHandler(ctx context.Context, cfg config) (command.Handler, error) {
	h := &handler{
		log: command.Logger(cfg.LoggingConfig),
	}
	return h, nil
}

func (h *handler) Handle(ctx context.Context) error {
	h.log.InfoContext(ctx, "hello")
	return nil
}
