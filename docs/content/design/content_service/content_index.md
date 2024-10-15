---
title: Content Index
type: docs
description: Index content which has been successfully stored.
---

The Content Index is a quick and simple indexing of the individual pieces of content.
It's primary purpose is to help avoid the need to scan [Content Storage]({{% ref "/design/content_service/content_storage" %}})
when the user wants to list out all of their content. This is typically going to provide
lower level functionality for higher-level user experiences like curating content into
[Collections]({{% ref "/user_guide/curating_content#collection" %}}) and/or [Libraries]({{% ref "/user_guide/curating_content#library" %}}).

## Index Record

The data which needs to be stored in the Content Index is described by the protobuf message: [Record](https://github.com/z5labs/griot/blob/main/services/content/indexpb/index_record.proto)

## Primary Key

The Content Index primary key is the [Content ID]({{% ref "/design/content_service/_index.md#content-id" %}}).

## Supported Queries

The following queries are supported:
- Get by [Content ID]({{% ref "/design/content_service/_index.md#content-id" %}})
- Get by [Checksum](https://github.com/z5labs/griot/blob/main/services/content/contentpb/checksum.proto)
- Query by [Media Type](https://en.wikipedia.org/wiki/Media_type) type and optional sub type filter
- Query by name
