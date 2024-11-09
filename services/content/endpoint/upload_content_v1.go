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

package endpoint

import (
	"context"
	"mime/multipart"
	"net/http"

	"github.com/z5labs/griot/internal/ptr"
	"github.com/z5labs/griot/services/content/contentpb"

	"github.com/swaggest/openapi-go/openapi3"
	"github.com/z5labs/humus/rest"
)

type uploadContentV1Handler struct{}

func UploadContentV1() rest.Endpoint {
	h := &uploadContentV1Handler{}

	return rest.NewEndpoint(
		http.MethodPost,
		"/content/upload",
		rest.ConsumesMultipartFormData[UploadContentV1Schema](
			rest.ProducesProto(h),
		),
	)
}

type UploadContentV1Schema struct{}

func (UploadContentV1Schema) OpenApiV3Schema() (*openapi3.Schema, error) {
	var req rest.ProtoRequest[contentpb.Metadata, *contentpb.Metadata]
	metadataSchema, err := req.OpenApiV3Schema()
	if err != nil {
		return nil, err
	}

	var schema openapi3.Schema
	schema.WithType(openapi3.SchemaTypeObject)
	schema.WithProperties(map[string]openapi3.SchemaOrRef{
		"metadata": {
			Schema: metadataSchema,
		},
		"content": {
			Schema: &openapi3.Schema{
				Type:   ptr.Ref(openapi3.SchemaTypeString),
				Format: ptr.Ref("binary"),
			},
		},
	})
	return &schema, nil
}

func (h *uploadContentV1Handler) Handle(ctx context.Context, req *multipart.Reader) (*contentpb.UploadContentV1Response, error) {
	return nil, nil
}
