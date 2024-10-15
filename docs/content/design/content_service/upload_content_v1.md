---
title: Upload Content v1
type: docs
description: Upload and index content from the user.
---

## Context Diagrams

### Happy Path

```mermaid
sequenceDiagram
    User ->> Content Service: Upload Content v1

    Content Service ->> Content Service: Compute checksum while uploading

    Content Service ->> Object Storage: Store content with base64 checksum as object key
    Object Storage ->> Content Service: Success

    Content Service ->> Content Service: Compare uploaded and computed checksums
    Content Service ->> Content Service: Checksums match!

    Content Service ->> Object Index: Record object in index
    Object Index -->> Content Service: Success

    Content Service -->> User: HTTP 200
```

## API Description

| Descriptor | Value |
|------------|-------|
| API Type | RESTful |
| HTTP Method | POST |
| Path | /v1/content |

## Request Headers

| Name | Type | Constraint |
|------|------|------------|
| Content-Type | string | must be multipart/form-data |

## Request Body

### Form Field: metadata

| Content-Type |
|--------------|
| application/x-protobuf |

For proto message type which will be returned, please see: [Metadata](https://github.com/z5labs/griot/blob/main/services/content/contentpb/metadata.proto)

### Form Field: content

| Content-Type |
|--------------|
| Any valid [Media Type](https://en.wikipedia.org/wiki/Media_type)

## Response Headers

| Name | Value |
|------|-------|
| Content-Type | application/x-protobuf |

## Response Body

### HTTP 200

For proto message type which will be returned, please see: [UploadContentV1Response](https://github.com/z5labs/griot/blob/main/services/content/contentpb/upload_content_v1_response.proto)

### HTTP 400

For proto message type which will be returned, please see: [Status](https://github.com/z5labs/humus/blob/main/humus.proto#L14)

### HTTP 500

For proto message type which will be returned, please see: [Status](https://github.com/z5labs/humus/blob/main/humus.proto#L14)