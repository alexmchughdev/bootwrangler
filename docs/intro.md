---
title: BootWrangler documentation
description: User-facing documentation for BootWrangler.
slug: /
sidebar_position: 1
---

# BootWrangler documentation

BootWrangler is a GUI-first Linux provisioning and boot media studio. It helps
you prepare bootable USB media, manage OS images, build unattended installer
profiles, compose multi-boot media, and test installs before touching hardware.

The desktop application is the primary product surface. The `bootwrangler` CLI
is the companion for automation, repeatable tests, and scripted workflows.

## Start here

- [Quickstart](./quickstart.md): build and run BootWrangler locally.
- [Profiles](./profiles.md): write unattended installer profiles.
- [Media Builder](./media.md): compose multi-partition USB media.
- [QEMU Lab](./lab.md): test profile renders in disposable virtual machines.

## Safety model

BootWrangler is designed around explicit planning before destructive work:

- Flash workflows produce a plan before writing.
- Disk operations require typed confirmation.
- Device paths are never treated as safe by default.
- SSH public keys are preferred; private keys are not required or stored.

## Project surfaces

The public documentation lives in this repository and is built with Docusaurus.
The product marketing website is maintained separately so brand, launch, and
commercial material can evolve without mixing private website code into the
open-source project.
