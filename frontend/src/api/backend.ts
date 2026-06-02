import type { Profile } from "../types/profile";

export interface HealthStatus {
  status: string;
  version: string;
}

export interface ValidationResult {
  valid: boolean;
  problems: string[];
}

export interface ManifestFileEntry {
  path: string;
  purpose: string;
  sha256: string;
}

export interface RenderManifest {
  profile_name: string;
  os_family: string;
  os_version: string;
  renderer: string;
  files: ManifestFileEntry[];
  warnings: string[];
}

interface AppService {
  AvailableRenderers(): Promise<string[]>;
  Health(): Promise<HealthStatus>;
  LoadProfile(path: string): Promise<Profile>;
  OpenInNeovim(path: string, readOnly: boolean): Promise<void>;
  RenderProfile(value: Profile, outDir: string): Promise<RenderManifest>;
  SaveProfile(path: string, value: Profile): Promise<void>;
  ValidateProfile(value: Profile): Promise<ValidationResult>;
  ValidateSSHPublicKey(value: string): Promise<ValidationResult>;
  Version(): Promise<string>;
}

declare global {
  interface Window {
    go?: {
      app?: {
        Service?: AppService;
      };
    };
  }
}

export async function getHealth(): Promise<HealthStatus> {
  const service = getService();
  if (!service) {
    return {
      status: "preview",
      version: "frontend-only",
    };
  }

  return service.Health();
}

export async function loadProfile(path: string): Promise<Profile> {
  return requireService().LoadProfile(path);
}

export async function openInNeovim(path: string): Promise<void> {
  return requireService().OpenInNeovim(path, false);
}

export async function saveProfile(path: string, value: Profile): Promise<void> {
  return requireService().SaveProfile(path, value);
}

export async function validateProfile(value: Profile): Promise<ValidationResult> {
  const service = getService();
  if (!service) {
    return {
      valid: false,
      problems: ["Backend validation is unavailable in browser preview mode."],
    };
  }
  return service.ValidateProfile(value);
}

export async function validateSSHPublicKey(value: string): Promise<ValidationResult> {
  const service = getService();
  if (!service) {
    return {
      valid: false,
      problems: ["SSH key validation is unavailable in browser preview mode."],
    };
  }
  return service.ValidateSSHPublicKey(value);
}

export async function availableRenderers(): Promise<string[]> {
  const service = getService();
  if (!service) return [];
  return service.AvailableRenderers();
}

export async function renderProfile(
  value: Profile,
  outDir: string,
): Promise<RenderManifest> {
  return requireService().RenderProfile(value, outDir);
}

function getService(): AppService | undefined {
  return window.go?.app?.Service;
}

function requireService(): AppService {
  const service = getService();
  if (!service) {
    throw new Error("BootWrangler backend is unavailable.");
  }
  return service;
}
