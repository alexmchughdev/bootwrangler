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

export interface BuildPlan {
  RecipeName: string;
  DevicePath: string;
  DeviceSize: number;
  TotalBytes: number;
  Actions: Array<{
    Label: string;
    SizeBytes: number;
    Filesystem: string;
    DevicePath: string;
  }>;
  DryRun: boolean;
}

export interface PackagePreset {
  Name: string;
  Description: string;
  Packages: Record<string, string[]>;
}

export interface RolePreset {
  Name: string;
  Description: string;
  Packages: string[];
  Services: string[];
  Groups: string[];
}

export interface LabRun {
  ID: string;
  ProfileName: string;
  State: string; // "pending" | "running" | "stopped" | "failed"
  SSHPort: number;
  VNCPort: number;
  SerialLog: string;
  DiskPath: string;
  Error: string;
}

export interface ImportResult {
  Profile: Profile;
  UnsupportedFields: string[];
  Warnings: string[];
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
  PlanMediaBuild(recipeYAML: string, devicePath: string): Promise<BuildPlan>;
  FormatMediaBuildPlan(plan: BuildPlan): Promise<string>;
  RenderProfile(value: Profile, outDir: string, serverBaseURL: string): Promise<RenderManifest>;
  SaveProfile(path: string, value: Profile): Promise<void>;
  ValidateProfile(value: Profile): Promise<ValidationResult>;
  ValidateSSHPublicKey(value: string): Promise<ValidationResult>;
  Version(): Promise<string>;
  ListPackagePresets(): Promise<PackagePreset[]>;
  ListRolePresets(): Promise<RolePreset[]>;
  ExpandPackagePresets(names: string[], osFamily: string): Promise<string[]>;
  LabStart(profileName: string, memorymb: number, cpus: number): Promise<LabRun>;
  LabStop(runID: string): Promise<void>;
  LabStatus(runID: string): Promise<LabRun>;
  LabSerialLog(runID: string): Promise<string>;
  LabSSHCommand(runID: string, user: string): Promise<string>;
  LabListRuns(): Promise<LabRun[]>;
  LabListSnapshots(runID: string): Promise<string[]>;
  LabCreateSnapshot(runID: string, name: string): Promise<void>;
  ImportConfig(content: string): Promise<ImportResult>;
  DetectConfigFormat(content: string): Promise<string>;
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
  serverBaseURL = "",
): Promise<RenderManifest> {
  return requireService().RenderProfile(value, outDir, serverBaseURL);
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

export async function planMediaBuild(
  recipeYAML: string,
  devicePath: string,
): Promise<BuildPlan> {
  return requireService().PlanMediaBuild(recipeYAML, devicePath);
}

export async function formatMediaBuildPlan(plan: BuildPlan): Promise<string> {
  return requireService().FormatMediaBuildPlan(plan);
}

export async function listPackagePresets(): Promise<PackagePreset[]> {
  const service = getService();
  if (!service) return [];
  return service.ListPackagePresets();
}

export async function listRolePresets(): Promise<RolePreset[]> {
  const service = getService();
  if (!service) return [];
  return service.ListRolePresets();
}

export async function expandPackagePresets(
  names: string[],
  osFamily: string,
): Promise<string[]> {
  const service = getService();
  if (!service) return [];
  return service.ExpandPackagePresets(names, osFamily);
}

export async function labStart(
  profileName: string,
  memorymb: number,
  cpus: number,
): Promise<LabRun> {
  return requireService().LabStart(profileName, memorymb, cpus);
}

export async function labStop(runID: string): Promise<void> {
  return requireService().LabStop(runID);
}

export async function labStatus(runID: string): Promise<LabRun> {
  return requireService().LabStatus(runID);
}

export async function labSerialLog(runID: string): Promise<string> {
  return requireService().LabSerialLog(runID);
}

export async function labSSHCommand(runID: string, user: string): Promise<string> {
  return requireService().LabSSHCommand(runID, user);
}

export async function labListRuns(): Promise<LabRun[]> {
  const service = getService();
  if (!service) return [];
  return service.LabListRuns();
}

export async function labListSnapshots(runID: string): Promise<string[]> {
  return requireService().LabListSnapshots(runID);
}

export async function labCreateSnapshot(runID: string, name: string): Promise<void> {
  return requireService().LabCreateSnapshot(runID, name);
}

export async function importConfig(content: string): Promise<ImportResult> {
  return requireService().ImportConfig(content);
}

export async function detectConfigFormat(content: string): Promise<string> {
  const service = getService();
  if (!service) return "unknown";
  return service.DetectConfigFormat(content);
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
