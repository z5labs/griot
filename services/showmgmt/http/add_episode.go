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
	"fmt"
	"io"
	"mime/multipart"
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type addEpisodeParams struct {
	Title   string `schema:"title,required" validate:"required"`
	Season  string `schema:"season,required" validate:"required"`
	Episode string `schema:"episode,required" validate:"required"`
}

func addEpisodeHandler(s *ShowMgmtService) pathHandler {
	return post("/show/:title/season/:season/episode/:episode", addEpisode(s))
}

func addEpisode(s *ShowMgmtService) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		s.log.Info("parsing request parameters")
		params, err := parseParams(s, c)
		if err != nil {
			s.log.Error("failed to get request parameters", zap.Error(err))
			return fiber.ErrBadRequest
		}
		s.log.Info("parsed request parameters")

		// TODO: validate file doesn't already exist
		// TODO: validate request body is multipart
		// TODO: get episode video content type
		videoFormat := "mkv"

		s.log.Info(
			"opening file for episode content",
			zap.String("title", params.Title),
			zap.String("season", params.Season),
			zap.String("episode", params.Episode),
			zap.String("video_format", videoFormat),
		)
		epFilePath := buildEpisodeFilePath(params, videoFormat)
		episode, err := s.fs.Open(epFilePath)
		if err != nil {
			s.log.Error(
				"failed to create episode file",
				zap.String("title", params.Title),
				zap.String("season", params.Season),
				zap.String("episode", params.Episode),
				zap.String("video_format", videoFormat),
				zap.Error(err),
			)
			return fiber.ErrInternalServerError
		}
		s.log.Info(
			"opened file for episode content",
			zap.String("title", params.Title),
			zap.String("season", params.Season),
			zap.String("episode", params.Episode),
			zap.String("video_format", videoFormat),
		)

		s.log.Info(
			"saving episode content",
			zap.String("title", params.Title),
			zap.String("season", params.Season),
			zap.String("episode", params.Episode),
			zap.String("video_format", videoFormat),
		)
		body := c.Context().RequestBodyStream()
		pr := multipart.NewReader(body, "")
		bytesWritten, err := s.copyEpisodeContent(episode, pr, params)
		if err != nil {
			s.log.Error(
				"failed to save episode content",
				zap.String("title", params.Title),
				zap.String("season", params.Season),
				zap.String("episode", params.Episode),
				zap.Int64("bytes_written", bytesWritten),
				zap.String("video_format", videoFormat),
				zap.Error(err),
			)
			return fiber.ErrInternalServerError
		}
		s.log.Info(
			"saved episode content",
			zap.String("title", params.Title),
			zap.String("season", params.Season),
			zap.String("episode", params.Episode),
			zap.Int64("bytes_written", bytesWritten),
			zap.String("video_format", videoFormat),
		)

		return nil
	}
}

func parseParams(s *ShowMgmtService, c *fiber.Ctx) (addEpisodeParams, error) {
	var params addEpisodeParams
	err := c.ParamsParser(&params)
	if err != nil {
		s.log.Error("failed to parse request parameters", zap.Error(err))
		return params, err
	}

	sanitizers := []func(*addEpisodeParams) error{
		sanitizeTitle,
		sanitizeSeason,
		sanitizeEpisode,
	}
	for _, sanitize := range sanitizers {
		err = sanitize(&params)
		if err != nil {
			s.log.Error(
				"failed to sanitize request parameters",
				zap.String("title", params.Title),
				zap.String("season", params.Season),
				zap.String("episode", params.Episode),
				zap.Error(err),
			)
			return params, err
		}
	}

	err = s.validator.Struct(params)
	if err != nil {
		s.log.Error(
			"invalid path variables",
			zap.String("title", params.Title),
			zap.String("season", params.Season),
			zap.String("episode", params.Episode),
			zap.Error(err),
		)
		return params, err
	}
	return params, nil
}

func sanitizeTitle(params *addEpisodeParams) (err error) {
	params.Title, err = url.QueryUnescape(params.Title)
	if err != nil {
		return
	}
	params.Title = strings.TrimSpace(params.Title)
	return
}

func sanitizeSeason(params *addEpisodeParams) (err error) {
	params.Season, err = url.QueryUnescape(params.Season)
	if err != nil {
		return
	}
	params.Season = strings.TrimSpace(params.Season)
	return
}

func sanitizeEpisode(params *addEpisodeParams) (err error) {
	params.Episode, err = url.QueryUnescape(params.Episode)
	if err != nil {
		return
	}
	params.Episode = strings.TrimSpace(params.Episode)
	return
}

func buildEpisodeFilePath(params addEpisodeParams, videoFormat string) string {
	return fmt.Sprintf(
		"/%s/%s/%s - %s.%s",
		params.Title,
		params.Season,
		params.Title,
		params.Episode,
		videoFormat,
	)
}

func (s *ShowMgmtService) copyEpisodeContent(w io.Writer, r *multipart.Reader, params addEpisodeParams) (int64, error) {
	n := int64(0)
	for {
		p, err := r.NextPart()
		if err == io.EOF {
			return n, nil
		}
		if err != nil {
			s.log.Error(
				"failed to read multipart part",
				zap.String("title", params.Title),
				zap.String("season", params.Season),
				zap.String("episode", params.Episode),
				zap.Error(err),
			)
			return n, err
		}

		m, err := io.Copy(w, p)
		n += m
		if err != nil {
			s.log.Error(
				"failed to save episode",
				zap.String("title", params.Title),
				zap.String("season", params.Season),
				zap.String("episode", params.Episode),
				zap.Error(err),
			)
			return n, err
		}
	}
}
