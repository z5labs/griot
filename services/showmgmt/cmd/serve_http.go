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

package cmd

import (
	"net"

	"github.com/z5labs/griot/services/showmgmt/http"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func withServeHttpCmd() func(*viper.Viper) *cobra.Command {
	return func(v *viper.Viper) *cobra.Command {
		cmd := &cobra.Command{
			Use:               "http",
			PersistentPreRunE: withPersistentPreRun()(v),
			RunE: func(cmd *cobra.Command, args []string) error {
				addr := v.GetString("addr")
				ls, err := net.Listen("tcp", addr)
				if err != nil {
					zap.L().Error("failed to initialize network listener", zap.String("addr", addr), zap.Error(err))
					return Error{
						Cmd:   cmd,
						Cause: err,
					}
				}

				fs := afero.NewBasePathFs(afero.NewOsFs(), v.GetString("content-dir"))
				cfg := http.ServiceConfig{
					Logger: zap.L(),
					Dir:    fs,
				}
				s, err := http.NewShowMgmtService(cfg)
				if err != nil {
					zap.L().Error("failed to initialize show mgmt service", zap.String("addr", addr), zap.Error(err))
					return Error{
						Cmd:   cmd,
						Cause: err,
					}
				}

				err = s.Serve(cmd.Context(), ls)
				if err == nil {
					return nil
				}
				zap.L().Error("unexpected error while serving requests", zap.String("addr", addr), zap.Error(err))
				return Error{
					Cmd:   cmd,
					Cause: err,
				}
			},
		}

		cmd.Flags().String("addr", ":0", "")
		cmd.Flags().String("content-dir", ".", "Specify root directory for shows to be stored in.")

		v.BindPFlag("addr", cmd.Flags().Lookup("addr"))
		v.BindPFlag("content-dir", cmd.Flags().Lookup("content-dir"))

		return cmd
	}
}
