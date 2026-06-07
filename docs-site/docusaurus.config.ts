import { themes as prismThemes } from "prism-react-renderer";
import type { Config } from "@docusaurus/types";
import type * as Preset from "@docusaurus/preset-classic";

const marketingUrl = process.env.BOOTWRANGLER_MARKETING_URL ?? "https://bootwrangler.dev";
const docsUrl = process.env.BOOTWRANGLER_DOCS_URL ?? "https://docs.bootwrangler.dev";
const docsBaseUrl = process.env.BOOTWRANGLER_DOCS_BASE_URL ?? "/";

const config: Config = {
  title: "BootWrangler Docs",
  tagline: "Linux provisioning and boot media studio",
  favicon: "img/mark.svg",
  url: docsUrl,
  baseUrl: docsBaseUrl,
  organizationName: "alexmchughdev",
  projectName: "bootwrangler",
  onBrokenLinks: "throw",
  markdown: {
    hooks: {
      onBrokenMarkdownLinks: "warn",
    },
  },
  i18n: {
    defaultLocale: "en",
    locales: ["en"],
  },
  presets: [
    [
      "classic",
      {
        docs: {
          path: "../docs",
          routeBasePath: "/",
          sidebarPath: "./sidebars.ts",
          editUrl: "https://github.com/alexmchughdev/bootwrangler/edit/main/docs/",
          showLastUpdateAuthor: false,
          showLastUpdateTime: true,
        },
        blog: false,
        theme: {
          customCss: "./src/css/custom.css",
        },
      } satisfies Preset.Options,
    ],
  ],
  themeConfig: {
    metadata: [
      {
        name: "description",
        content: "User-facing documentation for BootWrangler, a GUI-first Linux provisioning and boot media studio.",
      },
    ],
    navbar: {
      title: "BootWrangler",
      logo: {
        alt: "BootWrangler",
        src: "img/mark.svg",
      },
      items: [
        { type: "docSidebar", sidebarId: "docs", position: "left", label: "Docs" },
        { to: "/quickstart", label: "Quickstart", position: "left" },
        { to: "/profiles", label: "Profiles", position: "left" },
        { to: "/media", label: "Media", position: "left" },
        { href: marketingUrl, label: "Product", position: "right" },
        { href: `${marketingUrl}/download`, label: "Download", position: "right" },
        { href: "https://github.com/alexmchughdev/bootwrangler", label: "GitHub", position: "right" },
      ],
    },
    footer: {
      style: "light",
      links: [
        {
          title: "Use BootWrangler",
          items: [
            { label: "Quickstart", to: "/quickstart" },
            { label: "Profiles", to: "/profiles" },
            { label: "Media Builder", to: "/media" },
            { label: "QEMU Lab", to: "/lab" },
          ],
        },
        {
          title: "Project",
          items: [
            { label: "Source", href: "https://github.com/alexmchughdev/bootwrangler" },
            { label: "Issues", href: "https://github.com/alexmchughdev/bootwrangler/issues" },
            { label: "Website integration", to: "/website-integration" },
          ],
        },
        {
          title: "Product",
          items: [
            { label: "Product site", href: marketingUrl },
            { label: "Download", href: `${marketingUrl}/download` },
          ],
        },
      ],
      copyright: `Copyright © ${new Date().getFullYear()} BootWrangler contributors.`,
    },
    prism: {
      theme: prismThemes.github,
      darkTheme: prismThemes.oneDark,
    },
  } satisfies Preset.ThemeConfig,
};

export default config;
