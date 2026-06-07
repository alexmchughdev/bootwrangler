# BootWrangler Docs Site

This is the public Docusaurus documentation site for BootWrangler.

The source content lives in the repository-level `docs/` directory so open-source
contributors can edit Markdown without working inside the site shell.

## Local development

```sh
npm install
npm run start
```

## Build

```sh
npm run build
```

## Marketing site boundary

The product marketing website is intentionally not stored in this repository.
Keep this site focused on user-facing product documentation and use
`docs/website-integration.md` as the public contract between the docs site and
the private marketing site.
