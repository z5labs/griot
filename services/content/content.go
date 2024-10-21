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

// Package content
package content

import (
	"context"
	"io"

	"github.com/z5labs/griot/services/content/contentpb"

	"go.opentelemetry.io/otel"
)

type Client struct{}

func NewClient() *Client {
	c := &Client{}
	return c
}

type UploadContentRequest struct {
	Name     string
	HashFunc contentpb.HashFunc
	Body     io.Reader
}

type UploadContentResponse struct {
	Id string `json:"id"`
}

func (c *Client) UploadContent(ctx context.Context, req *UploadContentRequest) (*UploadContentResponse, error) {
	_, span := otel.Tracer("content").Start(ctx, "Client.UploadContent")
	defer span.End()
	return nil, nil
}
