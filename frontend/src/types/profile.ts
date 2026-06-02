export interface Profile {
  name: string;
  os: OS;
  system: System;
  network: Network;
  disk: Disk;
  users: User[];
  ssh: SSH;
  packages: Packages;
  services: Services;
  validation: Validation;
  post_install: PostInstall;
  os_specific: Record<string, unknown>;
}

export interface OS {
  family: string;
  version: string;
  architecture: string;
}

export interface System {
  hostname: string;
  timezone: string;
  keyboard: string;
  locale: string;
}

export interface Network {
  mode: string;
  interface: string;
  address: string;
  gateway: string;
  dns: string[];
}

export interface Disk {
  mode: string;
  target: string;
  install_mode: string;
  filesystem: string;
  confirm_destructive: boolean;
}

export interface User {
  name: string;
  shell: string;
  groups: string[];
  sudo: boolean;
  ssh_keys: string[];
}

export interface SSH {
  enabled: boolean;
  permit_root_login: boolean;
  password_authentication: boolean;
}

export interface Packages {
  presets: string[];
  names: string[];
}

export interface Services {
  enable: string[];
}

export interface Validation {
  ssh: SSHValidation;
  commands: ValidationCommand[];
}

export interface SSHValidation {
  enabled: boolean;
  user: string;
}

export interface ValidationCommand {
  name: string;
  command: string;
  expect_exit_code: number;
}

export interface PostInstall {
  scripts: Script[];
}

export interface Script {
  name: string;
  content: string;
}

export function createDefaultProfile(): Profile {
  return {
    name: "new-profile",
    os: {
      family: "ubuntu",
      version: "24.04",
      architecture: "x86_64",
    },
    system: {
      hostname: "new-profile",
      timezone: "Europe/London",
      keyboard: "gb",
      locale: "en_GB.UTF-8",
    },
    network: {
      mode: "dhcp",
      interface: "auto",
      address: "",
      gateway: "",
      dns: [],
    },
    disk: {
      mode: "wipe",
      target: "auto",
      install_mode: "server",
      filesystem: "ext4",
      confirm_destructive: true,
    },
    users: [createDefaultUser()],
    ssh: {
      enabled: true,
      permit_root_login: false,
      password_authentication: false,
    },
    packages: {
      presets: ["minimal"],
      names: [],
    },
    services: {
      enable: ["ssh"],
    },
    validation: {
      ssh: {
        enabled: true,
        user: "deploy",
      },
      commands: [],
    },
    post_install: {
      scripts: [],
    },
    os_specific: {},
  };
}

export function createDefaultUser(): User {
  return {
    name: "deploy",
    shell: "/bin/bash",
    groups: ["sudo"],
    sudo: true,
    ssh_keys: [],
  };
}
