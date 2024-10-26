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
	"github.com/z5labs/griot/cmd/griot/content/upload"
	"github.com/z5labs/griot/cmd/internal/command"

	"github.com/spf13/cobra"
)

func New() *cobra.Command {
	return command.New(
		"content",
		command.Short("Manage content"),
		command.Sub(upload.New()),
	)
}
