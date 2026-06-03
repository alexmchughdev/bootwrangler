import { describe, expect, it } from "vitest";
import { navigationItems } from "./navigation";

describe("navigationItems", () => {
  it("contains the primary product sections", () => {
    expect(navigationItems).toEqual([
      "Dashboard",
      "Profiles",
      "Library",
      "Render",
      "Policy",
      "Images",
      "Flash Image",
      "Flash Partition",
      "Media Builder",
      "Boot Menu",
      "Lab",
      "Import",
      "USB Devices",
      "Host Info",
      "Provisioning Server",
      "Settings",
    ]);
  });
});
