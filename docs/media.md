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
