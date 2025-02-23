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
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/z5labs/griot/internal/mimetype"
	"github.com/z5labs/griot/internal/ptr"
	"github.com/z5labs/griot/services/content/contentpb"

	"google.golang.org/protobuf/proto"
)

func ExampleClient_UploadContent() {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		_, err := io.Copy(io.Discard, r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		contentId := "example-id"
		b, err := proto.Marshal(&contentpb.UploadContentV1Response{
			Id: &contentpb.ContentId{
				Value: &contentId,
			},
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", mimetype.Protobuf)
		w.WriteHeader(http.StatusOK)
		io.Copy(w, bytes.NewReader(b))
	}))

	c := NewClient(http.DefaultClient, srv.URL)

	var content bytes.Buffer
	hasher := sha256.New()
	mw := io.MultiWriter(hasher, &content)
	_, err := io.Copy(mw, strings.NewReader("hello world"))
	if err != nil {
		fmt.Println(err)
		return
	}

	resp, err := c.UploadContent(context.Background(), &UploadContentRequest{
		Metadata: &contentpb.Metadata{
			Name: ptr.Ref("example-content"),
			MediaType: &contentpb.MediaType{
				Type:    ptr.Ref("text"),
				Subtype: ptr.Ref("plain"),
			},
			Checksum: &contentpb.Checksum{
				HashFunc: contentpb.HashFunc_SHA256.Enum(),
				Hash:     hasher.Sum(nil),
			},
		},
		Content: &content,
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(resp.Id)
	//Output: example-id
}
