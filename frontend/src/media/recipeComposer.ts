import type { CatalogueEntry, CustomImage, LibraryEntry } from "../api/backend";

export interface MediaPartitionDraft {
  label: string;
  size: string;
  filesystem: string;
  content: {
    type: string;
    image?: string;
    version?: string;
    profile?: string;
    bundle?: string;
  };
}

export interface MediaRecipeDraft {
  name: string;
  partitionTable: "gpt" | "mbr";
  bootMode: "uefi-bios" | "uefi" | "bios";
  bootMenu: "ipxe" | "grub" | "syslinux";
  partitions: MediaPartitionDraft[];
}

export function createDefaultMediaRecipeDraft(): MediaRecipeDraft {
  return {
    name: "lab-usb",
    partitionTable: "gpt",
    bootMode: "uefi-bios",
    bootMenu: "ipxe",
    partitions: [],
  };
}

export function addBootMenuPartition(draft: MediaRecipeDraft): MediaRecipeDraft {
  return addPartition(draft, {
    label: nextUniqueLabel(draft.partitions, "BOOTWRANGLER"),
    size: "2G",
    filesystem: "fat32",
    content: { type: "boot-menu" },
  });
}

export function addCatalogueImagePartition(
  draft: MediaRecipeDraft,
  entry: CatalogueEntry,
  version: string,
): MediaRecipeDraft {
  return addPartition(draft, {
    label: nextUniqueLabel(draft.partitions, sanitizePartitionLabel(`${entry.ID}_${version}`)),
    size: "6G",
    filesystem: "exfat",
    content: { type: "catalogue-image", image: entry.ID, version },
  });
}

export function addCustomImagePartition(draft: MediaRecipeDraft, image: CustomImage): MediaRecipeDraft {
  return addPartition(draft, {
    label: nextUniqueLabel(draft.partitions, sanitizePartitionLabel(image.ID)),
    size: "6G",
    filesystem: "exfat",
    content: { type: "custom-image", image: image.ID },
  });
}

export function addRenderedProfilePartition(
  draft: MediaRecipeDraft,
  entry: LibraryEntry,
): MediaRecipeDraft {
  return addPartition(draft, {
    label: nextUniqueLabel(draft.partitions, sanitizePartitionLabel(entry.name)),
    size: "2G",
    filesystem: "fat32",
    content: { type: "rendered-profile", profile: entry.name },
  });
}

export function addStoragePartition(draft: MediaRecipeDraft): MediaRecipeDraft {
  const fixedPartitions = draft.partitions.filter((partition) => partition.size !== "remaining");
  return {
    ...draft,
    partitions: [
      ...fixedPartitions,
      {
        label: nextUniqueLabel(fixedPartitions, "STORAGE"),
        size: "remaining",
        filesystem: "exfat",
        content: { type: "empty" },
      },
    ],
  };
}

export function serializeMediaRecipe(draft: MediaRecipeDraft): string {
  const lines = [
    `name: ${yamlString(draft.name)}`,
    "",
    "device:",
    `  partition_table: ${draft.partitionTable}`,
    "",
    "boot:",
    `  mode: ${draft.bootMode}`,
    `  menu: ${draft.bootMenu}`,
    "",
    "partitions:",
  ];
  for (const partition of draft.partitions) {
    lines.push(
      `  - label: ${partition.label}`,
      `    size: ${partition.size}`,
      `    filesystem: ${partition.filesystem}`,
      "    content:",
      `      type: ${partition.content.type}`,
    );
    if (partition.content.image) lines.push(`      image: ${yamlString(partition.content.image)}`);
    if (partition.content.version) lines.push(`      version: ${yamlString(partition.content.version)}`);
    if (partition.content.profile) lines.push(`      profile: ${yamlString(partition.content.profile)}`);
    if (partition.content.bundle) lines.push(`      bundle: ${yamlString(partition.content.bundle)}`);
  }
  return lines.join("\n");
}

export function sanitizePartitionLabel(value: string): string {
  return (
    value
      .trim()
      .toUpperCase()
      .replace(/[^A-Z0-9]+/g, "_")
      .replace(/^_+|_+$/g, "")
      .slice(0, 32) || "CONTENT"
  );
}

export function nextUniqueLabel(partitions: MediaPartitionDraft[], base: string): string {
  const normalizedBase = sanitizePartitionLabel(base);
  const existing = new Set(partitions.map((partition) => partition.label));
  if (!existing.has(normalizedBase)) return normalizedBase;

  for (let index = 2; index < 1000; index += 1) {
    const suffix = `_${index}`;
    const candidate = `${normalizedBase.slice(0, 32 - suffix.length)}${suffix}`;
    if (!existing.has(candidate)) return candidate;
  }
  return normalizedBase.slice(0, 28) + "_999";
}

function addPartition(draft: MediaRecipeDraft, partition: MediaPartitionDraft): MediaRecipeDraft {
  return {
    ...draft,
    partitions: [...draft.partitions, partition],
  };
}

function yamlString(value: string): string {
  return JSON.stringify(value);
}
