---
title: Website integration
description: Public contract for the documentation site and private marketing site.
sidebar_position: 6
---

# Website integration

BootWrangler has two public-facing web surfaces:

| Surface | Source | Purpose |
| --- | --- | --- |
| Documentation | This open-source repository | Product docs, guides, examples, and contributor-editable reference material |
| Marketing site | Separate private repository | Brand storytelling, launch pages, commercial copy, analytics, and private website assets |

The two surfaces should feel like one product, but they should not share source
ownership. Keep implementation code for the marketing site out of this
repository.

## Recommended hosting model

Use separate deployments with shared navigation:

```text
https://bootwrangler.dev          private marketing site
https://docs.bootwrangler.dev     public Docusaurus docs site
```

If the marketing site needs docs under a path instead of a subdomain, host the
Docusaurus build behind a reverse proxy:

```text
https://bootwrangler.dev/docs
```

Set `BOOTWRANGLER_DOCS_BASE_URL=/docs/` when building the docs for that path.

## Shared navigation contract

The marketing site should link to:

- `/` for product landing content.
- `/download` for desktop downloads.
- `https://docs.bootwrangler.dev/quickstart` for the getting-started path.
- `https://docs.bootwrangler.dev/profiles` for profile documentation.
- `https://docs.bootwrangler.dev/media` for media builder documentation.

The docs site links back to the marketing site through:

- `BOOTWRANGLER_MARKETING_URL`
- `BOOTWRANGLER_DOCS_URL`
- `BOOTWRANGLER_DOCS_BASE_URL`

## Shared visual contract

The docs site should use the same product posture as the desktop app:

- Flat surfaces.
- Strong borders and rules.
- Small-radius controls.
- Muted steel-blue accent.
- Monospace for device paths, commands, hashes, and identifiers.
- Light and dark themes.

The marketing site can be more editorial, but should avoid breaking the product
language with glossy effects, decorative gradients, or western costume themes.

## Source boundary

Do not add private website assets, proprietary analytics configuration, launch
plans, or unreleased marketing copy to this repository. Public docs should remain
useful to users and contributors without depending on private website code.
