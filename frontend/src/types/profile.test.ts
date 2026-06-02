import { describe, expect, it } from "vitest";
import { createDefaultProfile } from "./profile";

describe("createDefaultProfile", () => {
  it("starts with secure SSH daemon defaults", () => {
    const profile = createDefaultProfile();

    expect(profile.ssh).toEqual({
      enabled: true,
      permit_root_login: false,
      password_authentication: false,
    });
    expect(profile.disk.confirm_destructive).toBe(true);
  });
});
