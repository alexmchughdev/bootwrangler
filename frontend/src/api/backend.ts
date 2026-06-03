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

export interface ArchImage {
  Type: string;
  URL: string;
  ChecksumURL: string;
  Compatibility: { WholeDrive: boolean; Partition: boolean; ISOFileBoot: boolean };
  BootMode: string[];
}

export interface ArchEntry {
  Arch: string;
  Images: ArchImage[];
}

export interface VersionEntry {
  Version: string;
  Architectures: ArchEntry[];
}

export interface CatalogueEntry {
  ID: string;
  Name: string;
  Family: string;
  Versions: VersionEntry[];
}

export interface CacheStatus {
  Cached: boolean;
  Path: string;
  Verified: boolean;
}

export interface Partition {
  Path: string;
  Name: string;
  Size: number;
  SizeHuman: string;
  Filesystem: string;
  MountPoint: string;
  Label: string;
}

export interface UsbDevice {
  Path: string;
  Name: string;
  Model: string;
  Size: number;
  SizeHuman: string;
  Serial: string;
  Transport: string;
  Removable: boolean;
  MountPoint: string;
  Partitions: Partition[];
  Safe: boolean;
  SafetyNote: string;
}

export interface FlashPlan {
  ImagePath: string;
  DevicePath: string;
  ImageSize: number;
  ImageSizeHuman: string;
  DeviceSize: number;
  DeviceSizeHuman: string;
  Command: string;
  DryRun: boolean;
}

interface AppService {
  AvailableRenderers(): Promise<string[]>;
  Health(): Promise<HealthStatus>;
  ListImages(): Promise<CatalogueEntry[]>;
  GetImage(id: string): Promise<CatalogueEntry>;
  ImageCacheStatus(id: string, version: string, arch: string): Promise<CacheStatus>;
  ListDevices(): Promise<UsbDevice[]>;
  GetDevice(path: string): Promise<UsbDevice>;
  LoadProfile(path: string): Promise<Profile>;
  OpenInNeovim(path: string, readOnly: boolean): Promise<void>;
  PlanFlash(devicePath: string, imagePath: string): Promise<FlashPlan>;
  ExecuteFlash(plan: FlashPlan): Promise<void>;
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

export async function listImages(): Promise<CatalogueEntry[]> {
  const service = getService();
  if (!service) return [];
  return service.ListImages();
}

export async function getImage(id: string): Promise<CatalogueEntry> {
  return requireService().GetImage(id);
}

export async function imageCacheStatus(
  id: string,
  version: string,
  arch: string,
): Promise<CacheStatus> {
  return requireService().ImageCacheStatus(id, version, arch);
}

export async function listDevices(): Promise<UsbDevice[]> {
  const service = getService();
  if (!service) return [];
  return service.ListDevices();
}

export async function getDevice(path: string): Promise<UsbDevice> {
  return requireService().GetDevice(path);
}

export async function planFlash(
  devicePath: string,
  imagePath: string,
): Promise<FlashPlan> {
  return requireService().PlanFlash(devicePath, imagePath);
}

export async function executeFlash(plan: FlashPlan): Promise<void> {
  return requireService().ExecuteFlash(plan);
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
