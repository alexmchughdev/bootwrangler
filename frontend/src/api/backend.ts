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

export interface CustomSource {
  Type: string;
  Path: string;
  URL: string;
}

export interface Checksum {
  Type: string;
  Value: string;
}

export interface CustomImage {
  ID: string;
  Name: string;
  Source: CustomSource;
  Checksum?: Checksum;
  Compatibility: { WholeDrive: boolean; Partition: boolean; ISOFileBoot: boolean };
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
    Content: {
      Type: string;
      Image?: string;
      Version?: string;
      Profile?: string;
      Bundle?: string;
    };
    ContentResolution: {
      Status: string;
      Message: string;
      SourcePath: string;
      Warnings: string[];
      Errors: string[];
    };
    DevicePath: string;
  }>;
  DryRun: boolean;
  Ready: boolean;
  Warnings: string[];
  Errors: string[];
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

export interface QEMUAvailability {
  installed: boolean;
  source: string; // "bundled" | "system" | "installed" | "none"
  version: string;
  path: string;
  method: string;
  can_auto: boolean;
  hint: string;
}

export interface ImportResult {
  Profile: Profile;
  UnsupportedFields: string[];
  Warnings: string[];
}

export interface LibraryEntry {
  name: string;
  filename: string;
  os_family: string;
  os_version: string;
  updated_at: string;
}

export interface VersionEntry {
  hash: string;
  message: string;
  timestamp: string;
}

export interface PolicyViolation {
  Rule: string;
  Message: string;
}

export interface PolicyCheckResult {
  Passed: boolean;
  Violations: PolicyViolation[];
}

export interface PolicySnapshot {
  OSFamily: string;
  SSHPasswordAuth: boolean;
  SSHPermitRootLogin: boolean;
  DiskConfirmDestructive: boolean;
  Users: string[];
  Packages: string[];
  PackagePresets: string[];
}

export type EntryKind = "netboot" | "local-profile" | "local-image" | "shell" | "reboot";

export interface MenuEntry {
  Kind: EntryKind;
  Label: string;
  URL: string;
  Kernel: string;
  Initrd: string;
  Cmdline: string;
}

export interface NetbootImage {
  ID: string;
  Name: string;
  Family: string;
  Version: string;
  Arch: string;
  KernelURL: string;
  InitrdURL: string;
}

export interface CPUInfo {
  Model: string;
  Cores: number;
  Threads: number;
}

export interface MemoryInfo {
  TotalBytes: number;
  TotalHuman: string;
}

export interface NetworkInterface {
  Name: string;
  Addresses: string[];
  HWAddr: string;
}

export interface HostInfo {
  Hostname: string;
  CPU: CPUInfo;
  Memory: MemoryInfo;
  Interfaces: NetworkInterface[];
}

interface AppService {
  AvailableRenderers(): Promise<string[]>;
  Health(): Promise<HealthStatus>;
  ListImages(): Promise<CatalogueEntry[]>;
  ListCustomImages(): Promise<CustomImage[]>;
  CustomImagesPath(): Promise<string>;
  SaveCustomImage(img: CustomImage): Promise<void>;
  GetImage(id: string): Promise<CatalogueEntry>;
  ImageCacheStatus(id: string, version: string, arch: string): Promise<CacheStatus>;
  DownloadImage(id: string, version: string, arch: string): Promise<string>;
  ListDevices(): Promise<UsbDevice[]>;
  GetDevice(path: string): Promise<UsbDevice>;
  LoadProfile(path: string): Promise<Profile>;
  OpenInNeovim(path: string, readOnly: boolean): Promise<void>;
  PlanFlash(devicePath: string, imagePath: string): Promise<FlashPlan>;
  PlanCatalogueFlash(devicePath: string, id: string, version: string, arch: string): Promise<FlashPlan>;
  PlanCustomImageFlash(devicePath: string, id: string): Promise<FlashPlan>;
  PlanPartitionFlash(devicePath: string, partitionPath: string, imagePath: string): Promise<FlashPlan>;
  PlanCataloguePartitionFlash(
    devicePath: string,
    partitionPath: string,
    id: string,
    version: string,
    arch: string,
  ): Promise<FlashPlan>;
  PlanCustomImagePartitionFlash(
    devicePath: string,
    partitionPath: string,
    id: string,
  ): Promise<FlashPlan>;
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
  ExpandRolePreset(name: string, osFamily: string): Promise<[string[], string[], string[]]>;
  LabStart(profileName: string, memorymb: number, cpus: number): Promise<LabRun>;
  LabStop(runID: string): Promise<void>;
  LabStatus(runID: string): Promise<LabRun>;
  LabSerialLog(runID: string): Promise<string>;
  LabSSHCommand(runID: string, user: string): Promise<string>;
  LabListRuns(): Promise<LabRun[]>;
  LabQEMUStatus(): Promise<QEMUAvailability>;
  LabInstallQEMU(): Promise<string>;
  LabListSnapshots(runID: string): Promise<string[]>;
  LabCreateSnapshot(runID: string, name: string): Promise<void>;
  LabRevertSnapshot(runID: string, name: string): Promise<void>;
  LabDeleteSnapshot(runID: string, name: string): Promise<void>;
  LibraryGet(name: string): Promise<Profile>;
  LibraryList(): Promise<LibraryEntry[]>;
  LibraryExportBundle(name: string, path: string): Promise<void>;
  LibraryImportBundle(path: string): Promise<string>;
  LibraryCommit(name: string, message: string): Promise<void>;
  LibraryHistory(name: string): Promise<VersionEntry[]>;
  LibraryRestoreVersion(name: string, hash: string): Promise<void>;
  LibraryDiff(name: string, fromHash: string, toHash: string): Promise<string>;
  ImportConfig(content: string): Promise<ImportResult>;
  DetectConfigFormat(content: string): Promise<string>;
  GatherHostInfo(): Promise<HostInfo>;
  CheckPolicy(policyYAML: string, snap: PolicySnapshot): Promise<PolicyCheckResult>;
  WriteTextFile(path: string, content: string): Promise<void>;
  ReadRenderedFile(path: string): Promise<string>;
  RenderIPXEMenu(title: string, entries: MenuEntry[]): Promise<string>;
  RenderGRUBMenu(title: string, entries: MenuEntry[]): Promise<string>;
  ListNetbootImages(): Promise<NetbootImage[]>;
  FormatNetbootCmdline(osFamily: string, serverBaseURL: string): Promise<string>;
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

export async function listCustomImages(): Promise<CustomImage[]> {
  const service = getService();
  if (!service) return [];
  return service.ListCustomImages();
}

export async function customImagesPath(): Promise<string> {
  const service = getService();
  if (!service) return "";
  return service.CustomImagesPath();
}

export async function saveCustomImage(img: CustomImage): Promise<void> {
  return requireService().SaveCustomImage(img);
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

export async function downloadImage(
  id: string,
  version: string,
  arch: string,
): Promise<string> {
  return requireService().DownloadImage(id, version, arch);
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

export async function planCatalogueFlash(
  devicePath: string,
  id: string,
  version: string,
  arch: string,
): Promise<FlashPlan> {
  return requireService().PlanCatalogueFlash(devicePath, id, version, arch);
}

export async function planCustomImageFlash(
  devicePath: string,
  id: string,
): Promise<FlashPlan> {
  return requireService().PlanCustomImageFlash(devicePath, id);
}

export async function planPartitionFlash(
  devicePath: string,
  partitionPath: string,
  imagePath: string,
): Promise<FlashPlan> {
  return requireService().PlanPartitionFlash(devicePath, partitionPath, imagePath);
}

export async function planCataloguePartitionFlash(
  devicePath: string,
  partitionPath: string,
  id: string,
  version: string,
  arch: string,
): Promise<FlashPlan> {
  return requireService().PlanCataloguePartitionFlash(
    devicePath,
    partitionPath,
    id,
    version,
    arch,
  );
}

export async function planCustomImagePartitionFlash(
  devicePath: string,
  partitionPath: string,
  id: string,
): Promise<FlashPlan> {
  return requireService().PlanCustomImagePartitionFlash(devicePath, partitionPath, id);
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

export async function expandRolePreset(
  name: string,
  osFamily: string,
): Promise<[string[], string[], string[]]> {
  return requireService().ExpandRolePreset(name, osFamily);
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

export async function labQEMUStatus(): Promise<QEMUAvailability> {
  const service = getService();
  if (!service) {
    return {
      installed: false,
      source: "none",
      version: "",
      path: "",
      method: "manual",
      can_auto: false,
      hint: "Backend unavailable in browser preview.",
    };
  }
  return service.LabQEMUStatus();
}

export async function labInstallQEMU(): Promise<string> {
  return requireService().LabInstallQEMU();
}

export async function labListSnapshots(runID: string): Promise<string[]> {
  return requireService().LabListSnapshots(runID);
}

export async function labCreateSnapshot(runID: string, name: string): Promise<void> {
  return requireService().LabCreateSnapshot(runID, name);
}

export async function labRevertSnapshot(runID: string, name: string): Promise<void> {
  return requireService().LabRevertSnapshot(runID, name);
}

export async function labDeleteSnapshot(runID: string, name: string): Promise<void> {
  return requireService().LabDeleteSnapshot(runID, name);
}

export async function libraryGet(name: string): Promise<Profile> {
  return requireService().LibraryGet(name);
}

export async function libraryList(): Promise<LibraryEntry[]> {
  const service = getService();
  if (!service) return [];
  return service.LibraryList();
}

export async function libraryCommit(name: string, message: string): Promise<void> {
  return requireService().LibraryCommit(name, message);
}

export async function libraryHistory(name: string): Promise<VersionEntry[]> {
  return requireService().LibraryHistory(name);
}

export async function libraryRestoreVersion(name: string, hash: string): Promise<void> {
  return requireService().LibraryRestoreVersion(name, hash);
}

export async function libraryDiff(name: string, fromHash: string, toHash: string): Promise<string> {
  return requireService().LibraryDiff(name, fromHash, toHash);
}

export async function libraryExportBundle(name: string, path: string): Promise<void> {
  return requireService().LibraryExportBundle(name, path);
}

export async function libraryImportBundle(path: string): Promise<string> {
  return requireService().LibraryImportBundle(path);
}

export async function importConfig(content: string): Promise<ImportResult> {
  return requireService().ImportConfig(content);
}

export async function detectConfigFormat(content: string): Promise<string> {
  const service = getService();
  if (!service) return "unknown";
  return service.DetectConfigFormat(content);
}

export async function gatherHostInfo(): Promise<HostInfo> {
  const service = getService();
  if (!service) {
    return {
      Hostname: "preview-mode",
      CPU: { Model: "Unknown (backend not available)", Cores: 0, Threads: 0 },
      Memory: { TotalBytes: 0, TotalHuman: "—" },
      Interfaces: [],
    };
  }
  return service.GatherHostInfo();
}

export async function checkPolicy(policyYAML: string, snap: PolicySnapshot): Promise<PolicyCheckResult> {
  return requireService().CheckPolicy(policyYAML, snap);
}

export async function writeTextFile(path: string, content: string): Promise<void> {
  return requireService().WriteTextFile(path, content);
}

export async function readRenderedFile(path: string): Promise<string> {
  return requireService().ReadRenderedFile(path);
}

export async function renderIPXEMenu(title: string, entries: MenuEntry[]): Promise<string> {
  return requireService().RenderIPXEMenu(title, entries);
}

export async function renderGRUBMenu(title: string, entries: MenuEntry[]): Promise<string> {
  return requireService().RenderGRUBMenu(title, entries);
}

export async function listNetbootImages(): Promise<NetbootImage[]> {
  const service = getService();
  if (!service) return [];
  return service.ListNetbootImages();
}

export async function formatNetbootCmdline(osFamily: string, serverBaseURL: string): Promise<string> {
  return requireService().FormatNetbootCmdline(osFamily, serverBaseURL);
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
