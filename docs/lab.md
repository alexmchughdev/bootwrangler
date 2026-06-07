---
title: QEMU Lab
description: Test BootWrangler profiles in disposable QEMU virtual machines.
sidebar_position: 5
---

# QEMU Lab

The Lab feature launches ephemeral QEMU VMs from a rendered BootWrangler profile, letting you validate installer configs before flashing to real hardware.

## Requirements

- `qemu-system-x86_64` in PATH
- `qemu-img` in PATH (for disk image creation and snapshots)

Install on Ubuntu/Debian:
```sh
sudo apt install qemu-system-x86 qemu-utils
```

Install on macOS:
```sh
brew install qemu
```

## Start a lab VM

```sh
bootwrangler lab start my-server --memory 2048 --cpus 2
```

This creates a 20 GiB qcow2 disk image, boots QEMU, and returns a run ID.

## Check status

```sh
bootwrangler lab status <run-id>
```

## View serial console log

```sh
bootwrangler lab console <run-id>
```

## Connect via SSH

```sh
$(bootwrangler lab ssh <run-id>)
```

## Stop a VM

```sh
bootwrangler lab stop <run-id>
```

## Snapshots

```sh
# Create
bootwrangler lab snapshot create <run-id> post-install

# List
bootwrangler lab snapshot list <run-id>

# Revert
bootwrangler lab snapshot revert <run-id> post-install
```

## Dry-run / plan

Print the QEMU command that would be run without starting anything:

```sh
bootwrangler lab plan my-server
```
