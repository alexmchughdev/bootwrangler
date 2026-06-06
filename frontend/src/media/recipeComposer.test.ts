import { describe, expect, it } from "vitest";
import type { CatalogueEntry, CustomImage, LibraryEntry } from "../api/backend";
import {
  addBootMenuPartition,
  addCatalogueImagePartition,
  addCustomImagePartition,
  addRenderedProfilePartition,
  addStoragePartition,
  createDefaultMediaRecipeDraft,
  nextUniqueLabel,
  sanitizePartitionLabel,
  serializeMediaRecipe,
} from "./recipeComposer";

describe("recipeComposer", () => {
  it("serializes a stable schema-correct recipe", () => {
    const draft = addStoragePartition(
      addCatalogueImagePartition(addBootMenuPartition(createDefaultMediaRecipeDraft()), catalogueEntry(), "24.04"),
    );

    expect(serializeMediaRecipe(draft)).toBe(`name: "lab-usb"

device:
  partition_table: gpt

boot:
  mode: uefi-bios
  menu: ipxe

partitions:
  - label: BOOTWRANGLER
    size: 2G
    filesystem: fat32
    content:
      type: boot-menu
  - label: UBUNTU_SERVER_24_04
    size: 6G
    filesystem: exfat
    content:
      type: catalogue-image
      image: "ubuntu-server"
      version: "24.04"
  - label: STORAGE
    size: remaining
    filesystem: exfat
    content:
      type: empty`);
  });

  it("keeps generated labels valid and unique", () => {
    expect(sanitizePartitionLabel("Ubuntu Server 24.04")).toBe("UBUNTU_SERVER_24_04");
    expect(sanitizePartitionLabel("!!!")).toBe("CONTENT");

    const draft = addBootMenuPartition(addBootMenuPartition(createDefaultMediaRecipeDraft()));
    expect(draft.partitions.map((partition) => partition.label)).toEqual([
      "BOOTWRANGLER",
      "BOOTWRANGLER_2",
    ]);
    expect(nextUniqueLabel(draft.partitions, "BOOTWRANGLER")).toBe("BOOTWRANGLER_3");
  });

  it("emits custom image and rendered profile content types", () => {
    const draft = addRenderedProfilePartition(
      addCustomImagePartition(createDefaultMediaRecipeDraft(), customImage()),
      libraryEntry(),
    );
    const yaml = serializeMediaRecipe(draft);

    expect(yaml).toContain("      type: custom-image\n      image: \"company-os\"");
    expect(yaml).toContain("      type: rendered-profile\n      profile: \"edge-node\"");
  });

  it("keeps only one remaining storage partition", () => {
    const draft = addStoragePartition(addStoragePartition(createDefaultMediaRecipeDraft()));

    expect(draft.partitions.filter((partition) => partition.size === "remaining")).toHaveLength(1);
    expect(serializeMediaRecipe(draft)).toContain("    size: remaining");
  });
});

function catalogueEntry(): CatalogueEntry {
  return {
    ID: "ubuntu-server",
    Name: "Ubuntu Server",
    Family: "ubuntu",
    Versions: [],
  };
}

function customImage(): CustomImage {
  return {
    ID: "company-os",
    Name: "Company OS",
    Source: { Type: "local-file", Path: "/tmp/company-os.iso", URL: "" },
    Compatibility: { WholeDrive: true, Partition: false, ISOFileBoot: true },
  };
}

function libraryEntry(): LibraryEntry {
  return {
    name: "edge-node",
    filename: "edge-node.yaml",
    os_family: "ubuntu",
    os_version: "24.04",
    updated_at: "2026-06-06T23:35:00Z",
  };
}
