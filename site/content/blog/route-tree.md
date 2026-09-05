---
title: One tree for docs and publishing
description: Why the blog archive and RSS feed use the same Router as documentation.
date: 2026-08-28
authors: [fastr-docs]
tags: [architecture, routing]
---

# One tree for docs and publishing

The blog is a Markdown collection owned by the same Router that owns the documentation tree. That keeps navigation, search, metadata, static export, and RSS on the same set of routes.

## One source, many surfaces

The archive, taxonomy pages, RSS feed, and post template all read the same
published route metadata. There is no second list to keep in sync.

## A better reading surface

Posts add article landmarks, share and copy actions, related reading, and a
responsive table of contents while keeping the global site shell familiar.
