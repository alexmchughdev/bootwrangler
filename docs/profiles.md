---
title: Profiles
description: Profile schema, rendering, importing, and policy checks.
sidebar_position: 3
---

# Profiles

A profile is a YAML document that describes one target machine's operating system, network, disk, SSH, and user configuration. BootWrangler renders a profile into the unattended installer assets for the chosen OS family.

## Supported OS families

| Family | Renderer | Installer format |
|--------|----------|-----------------|
| `ubuntu` | ubuntu-autoinstall | `user-data` + `meta-data` |
| `debian` | debian-preseed | `preseed.cfg` |
| `alpine` | alpine-answerfile | `answerfile` |
| `rocky` | kickstart | `ks.cfg` |
| `fedora` | kickstart | `ks.cfg` |
| `arch` | archinstall | `archinstall.json` |
| `opensuse` | autoyast | `autoinst.xml` |

## Profile fields

```yaml
name: my-server
os:
  family: ubuntu        # required
  version: "24.04"      # required
  architecture: x86_64  # default

system:
  hostname: my-server
  timezone: Europe/London
  keyboard: uk
  locale: en_GB.UTF-8

network:
  mode: dhcp            # dhcp | static
  address: ""           # required if mode=static
  gateway: ""
  dns: []

disk:
  mode: wipe            # wipe | partition
  target: auto          # device path or "auto"
  filesystem: ext4
  install_mode: server  # server | desktop
  confirm_destructive: true   # must be true to render

ssh:
  enabled: true
  permit_root_login: false
  password_authentication: false
  authorized_keys: []

users: []
packages: []
```

## Rendering

```sh
bootwrangler render my-profile.yaml --out /tmp/my-server/
```

Output files are written to the specified directory and a `manifest.json` is created listing each file's purpose and SHA-256 digest.

## Importing from existing configs

```sh
bootwrangler import /path/to/existing.cfg
```

BootWrangler auto-detects Alpine answerfile, Ubuntu autoinstall, Debian preseed, Kickstart, and AutoYaST formats and converts them to a BootWrangler profile on a best-effort basis.

## Policy checks

```sh
bootwrangler policy check my-profile.yaml --policy company.yaml
```

A policy YAML can enforce rules such as requiring SSH key-only auth, forbidding root login, mandating disk confirmation, and restricting allowed distros.
