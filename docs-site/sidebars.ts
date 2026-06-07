import type { SidebarsConfig } from "@docusaurus/plugin-content-docs";

const sidebars: SidebarsConfig = {
  docs: [
    "intro",
    "quickstart",
    {
      type: "category",
      label: "Core workflows",
      items: ["profiles", "media", "lab"],
    },
    {
      type: "category",
      label: "Project surfaces",
      items: ["website-integration"],
    },
  ],
};

export default sidebars;
