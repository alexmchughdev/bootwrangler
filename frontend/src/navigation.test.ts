import { describe, expect, it } from "vitest";
import { navigationItems } from "./navigation";

describe("navigationItems", () => {
  it("contains the primary product sections", () => {
    expect(navigationItems).toEqual([
      "Dashboard",
      "Profiles",
      "Render",
      "Images",
      "Media Builder",
      "Lab",
      "USB Devices",
      "Provisioning Server",
      "Settings",
    ]);
  });
});
