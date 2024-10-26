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

// Package content provides Content Service client and server implementations.
package content

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"

	"github.com/z5labs/griot/services/content/contentpb"

	"github.com/z5labs/humus/rest"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"
)

type HttpClient interface {
	Do(*http.Request) (*http.Response, error)
}

type Client struct {
	protoMarshal func(proto.Message) ([]byte, error)
	http         HttpClient
}

func NewClient(hc HttpClient) *Client {
	c := &Client{
		http: hc,
	}
	return c
}

type UploadContentRequest struct {
	Metadata *contentpb.Metadata
	Body     io.Reader
}

type UploadContentResponse struct {
	Id string `json:"id"`
}

func (c *Client) UploadContent(ctx context.Context, req *UploadContentRequest) (*UploadContentResponse, error) {
	spanCtx, span := otel.Tracer("content").Start(ctx, "Client.UploadContent")
	defer span.End()

	body, bodyWriter := io.Pipe()

	respCh := make(chan *http.Response)
	eg, egctx := errgroup.WithContext(spanCtx)
	eg.Go(func() error {
		defer bodyWriter.Close()

		return c.writeUploadRequest(egctx, bodyWriter, req)
	})
	eg.Go(func() error {
		defer close(respCh)
		defer body.Close()

		r, err := http.NewRequestWithContext(egctx, http.MethodPost, "", body)
		if err != nil {
			return err
		}
		r.Header.Set("Content-Type", "multipart/form-data")

		resp, err := c.http.Do(r)
		if err != nil {
			return err
		}
		select {
		case <-egctx.Done():
			return egctx.Err()
		case respCh <- resp:
		}
		return nil
	})

	err := eg.Wait()
	if err != nil {
		return nil, err
	}

	var resp *http.Response
	select {
	case <-spanCtx.Done():
		return nil, spanCtx.Err()
	case resp = <-respCh:
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected http status code: %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var uploadV1Resp contentpb.UploadContentV1Response
	err = proto.Unmarshal(b, &uploadV1Resp)
	if err != nil {
		return nil, err
	}

	uploadResp := UploadContentResponse{
		Id: uploadV1Resp.GetId().GetValue(),
	}
	return &uploadResp, nil
}

func (c *Client) writeUploadRequest(ctx context.Context, w io.Writer, req *UploadContentRequest) error {
	spanCtx, span := otel.Tracer("content").Start(ctx, "Client.writeUploadRequest")
	defer span.End()

	pw := multipart.NewWriter(w)
	err := c.writeMetadata(spanCtx, pw, req.Metadata)
	if err != nil {
		return err
	}

	err = c.writeContent(spanCtx, pw, req.Metadata.Checksum.Hash, req.Body)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) writeMetadata(ctx context.Context, pw *multipart.Writer, meta *contentpb.Metadata) error {
	_, span := otel.Tracer("content").Start(ctx, "Client.writeMetadata")
	defer span.End()

	b, err := c.protoMarshal(meta)
	if err != nil {
		return err
	}

	header := make(textproto.MIMEHeader)
	header.Set("Content-Disposition", `form-data; name="metadata"`)
	header.Set("Content-Type", rest.ProtobufContentType)

	part, err := pw.CreatePart(header)
	if err != nil {
		return err
	}

	n, err := io.Copy(part, bytes.NewReader(b))
	if err != nil {
		return err
	}
	if n != int64(len(b)) {
		// TODO
		return errors.New("did not write all metadata bytes")
	}
	return nil
}

type progressReader struct {
	ctx       context.Context
	r         io.Reader
	bytesRead metric.Int64Counter
}

func (r *progressReader) Read(b []byte) (int, error) {
	select {
	case <-r.ctx.Done():
		return 0, r.ctx.Err()
	default:
	}

	n, err := r.r.Read(b)
	r.bytesRead.Add(r.ctx, int64(n), metric.WithAttributes(
		attribute.String("griot.content.io.direction", "read"),
	))
	return n, err
}

func (c *Client) writeContent(ctx context.Context, pw *multipart.Writer, hash []byte, r io.Reader) error {
	spanCtx, span := otel.Tracer("content").Start(ctx, "Client.writeContent")
	defer span.End()

	bytesRead, err := otel.Meter("content").Int64Counter("griot.content.io", metric.WithUnit("By"))
	if err != nil {
		return err
	}

	pr := &progressReader{
		ctx:       spanCtx,
		r:         r,
		bytesRead: bytesRead,
	}

	filename := base64.StdEncoding.EncodeToString(hash)

	header := make(textproto.MIMEHeader)
	header.Set("Content-Disposition", fmt.Sprintf(`form-data; name="content"; filename=%q`, filename))
	header.Set("Content-Type", "application/octet-stream")

	part, err := pw.CreatePart(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(part, pr)
	return err
}
