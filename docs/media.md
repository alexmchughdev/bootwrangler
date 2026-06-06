# Media Builder

The media builder creates bootable USB drives with multiple partitions, each containing different OS images or installer assets.

## Recipe format

```yaml
label: my-lab-usb
partition_table: gpt        # gpt | mbr
boot_menu: grub             # grub | ipxe | none
boot_mode: [uefi, bios]

partitions:
  - label: EFI
    size: 512MiB
    filesystem: fat32
    content_type: efi

  - label: ubuntu-server
    size: 4GiB
    filesystem: ext4
    content_type: installer-assets
    source: /rendered/ubuntu-server/

  - label: data
    size: remaining
    filesystem: ext4
    content_type: data
```

## Validate a recipe

```sh
bootwrangler recipe validate my-usb.yaml
```

## Preview the build plan

```sh
bootwrangler recipe plan my-usb.yaml --device /dev/sdb
```

This prints the partition layout and copy actions without writing anything.

## Image flash compatibility

Catalogue images declare the modes they support:

- `whole_drive`: write the image to an entire USB/disk.
- `partition`: write the image directly to a selected partition.
- `iso_file_boot`: store the ISO as a file and boot it from a menu.

The desktop flash workflows resolve catalogue metadata before planning. If an
image is whole-drive only, partition flashing is rejected; use whole-drive flash
or ISO-file boot mode instead.

## Image download verification

Catalogue downloads use checksum metadata when the image entry provides it. The
download flow stores a verified cache marker only after the image bytes match
the expected SHA256 checksum.

Cache states shown in the desktop image catalogue:

- `Not downloaded`: no cached image is present.
- `Downloaded, not verified`: a cached image exists without a verified checksum
  marker.
- `Verified`: the image was downloaded and matched its SHA256 checksum.

Checksum failures remove the newly downloaded image and any stale verification
marker so flashing workflows do not treat an unverified image as safe.

## Custom images

User-defined images are loaded from:

```text
~/.bootwrangler/catalogue/custom-images.yaml
```

Example:

```yaml
images:
  - id: company-os
    name: Company OS
    source:
      type: local-file
      path: /path/to/company-os.iso
    checksum:
      type: sha256
      value: 0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef
    compatibility:
      whole_drive: true
      partition: false
      iso_file_boot: true
```

Supported custom image sources:

- `local-file`: references an existing image file by path.
- `url`: references an absolute `http` or `https` image URL.

Custom images must declare at least one compatibility mode. Checksums are
optional, but when present they must be valid `sha256` or `md5` hex values.

The desktop Images view can add or replace custom image definitions by ID. The
CLI can inspect the same file:

```sh
bootwrangler images custom path
bootwrangler images custom list
bootwrangler images custom validate
```

Local-file custom images appear in the whole-drive and partition flash
workflows when their compatibility flags allow that mode. URL custom images can
be catalogued now; download and cache support for those entries is handled
separately from local-file flash planning.

## Contained provisioning mode

Rendered installer assets can be written directly to a mounted USB partition so the machine boots and installs without a network provisioning server:

```sh
# 1. Render a profile
bootwrangler render my-server.yaml --out /tmp/my-server/

# 2. Mount the target USB partition
mount /dev/sdb1 /mnt/usb

# 3. Copy assets (BootWrangler plans and copies the files)
bootwrangler flash partition /tmp/my-server/ --target /mnt/usb
```

## Boot menu

A GRUB or iPXE menu is generated automatically from the recipe and written to the EFI partition. The menu lists each installer partition and chains to the appropriate bootloader.
