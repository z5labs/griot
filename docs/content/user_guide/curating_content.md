---
title: Curating Content
type: docs
description: Upload and catalog your content for easy searching.
---

## Key Concepts

### Content

Content is pretty much anything you can think of, some examples include, photos, videos, songs,
documents, etc. At the heart of griot, is the ability to arbitarily store and retrieve any type
of content you with.

### Collection

A collection is a ordered list of [Content](#content). The collection name does not need to be unique.

### Library

A library is uniquely named set of individual pieces of [Content](#content) and/or [Collection](#collection)s.
All collections and individual pieces of content in a library will be indexed for querying via [CEL](https://cel.dev/)
expressions.

## Adding your content to griot

### Step One: Upload a piece of content
```
$ griot upload --name "Naruto S01E01" --mime-type "video/av1" --source-file "Naruto S01E01.av1"
{"id": "content-1"}
```

### Step Two: (Optional) Create and add content to a collection
```
$ griot collection create --name "Season 1"
{"id":"collection-1"}

$ griot collection add item --collection-id "collection-1" --item-id "content-1" --order 1

$ griot collection create --name "Naruto"
{"id":"collection-2"}

$ griot collection add item --collection-id "collection-2" --item-id "collection-1" --order 1

$ griot collection list
[{"id":"collection-1","name":"Season 1"},{"id":"collection-2","name":"Naruto"}]

$ griot collection list --collection-id "collection-1"
[{"type":"content","id":"content-1","order":1}]

$ griot collection list --collection-id "collection-2"
[{"type":"collection","id":"collection-1","order":1}]
```

### Step Three: Create and add content/collection to a library
```
$ griot library create --name "Anime"
{"id":"library-1"}

$ griot library add item --library-name "Anime" --item-id "collection-2"
// or
$ griot library add item --library-id "library-1" --item-id "collection-2"

$ griot library list
[{"id":"library-1","name":"Anime"}]

$ griot library list --library-name "Anime"
// or
$ griot library list --library-id "library-1"
[{"type":"collection","id":"collection-2"}]

$ griot library search --library-name "Anime" --query "collection['name'] == 'Naruto'"
// or
$ griot library search --library-id "library-1" --query "collection['name'] == 'Naruto'"
```
