package media

import (
	"fmt"
	"strconv"
	"strings"
)

// Validate checks a Recipe for structural correctness.
func Validate(r Recipe) error {
	if r.Name == "" {
		return fmt.Errorf("media recipe: name is required")
	}
	if err := validateDevice(r.Device); err != nil {
		return err
	}
	if err := validateBoot(r.Boot); err != nil {
		return err
	}
	if len(r.Partitions) == 0 {
		return fmt.Errorf("media recipe: at least one partition is required")
	}
	remainingCount := 0
	labels := map[string]bool{}
	for i, p := range r.Partitions {
		if p.Label == "" {
			return fmt.Errorf("media recipe: partitions[%d]: label is required", i)
		}
		if labels[p.Label] {
			return fmt.Errorf("media recipe: duplicate partition label %q", p.Label)
		}
		labels[p.Label] = true
		if p.Size == "" {
			return fmt.Errorf("media recipe: partition %q: size is required", p.Label)
		}
		if strings.EqualFold(p.Size, "remaining") {
			remainingCount++
		} else {
			if _, err := parseSize(p.Size); err != nil {
				return fmt.Errorf("media recipe: partition %q: invalid size %q: %w", p.Label, p.Size, err)
			}
		}
		if p.Filesystem == "" {
			return fmt.Errorf("media recipe: partition %q: filesystem is required", p.Label)
		}
		if err := validateContent(p.Label, p.Content); err != nil {
			return err
		}
	}
	if remainingCount > 1 {
		return fmt.Errorf("media recipe: only one partition may use size \"remaining\"")
	}
	if remainingCount > 0 {
		last := r.Partitions[len(r.Partitions)-1]
		if !strings.EqualFold(last.Size, "remaining") {
			return fmt.Errorf("media recipe: only the last partition may use size \"remaining\"")
		}
	}
	return nil
}

func validateDevice(d DeviceSpec) error {
	switch d.PartitionTable {
	case PartitionTableGPT, PartitionTableMBR:
		return nil
	case "":
		return fmt.Errorf("media recipe: device.partition_table is required")
	default:
		return fmt.Errorf("media recipe: unsupported partition_table %q (use gpt or mbr)", d.PartitionTable)
	}
}

func validateBoot(b BootSpec) error {
	switch b.Mode {
	case BootModeUEFIBIOS, BootModeUEFI, BootModeBIOS, "":
		// empty is allowed; defaults applied at plan time
	default:
		return fmt.Errorf("media recipe: unsupported boot mode %q", b.Mode)
	}
	switch b.Menu {
	case BootMenuiPXE, BootMenuGRUB, BootMenuSyslinux, "":
		// empty is allowed
	default:
		return fmt.Errorf("media recipe: unsupported boot menu type %q", b.Menu)
	}
	return nil
}

func validateContent(label string, c PartitionContent) error {
	switch c.Type {
	case ContentBootMenu:
		return nil
	case ContentCatalogueImage:
		if c.Image == "" {
			return fmt.Errorf("media recipe: partition %q: catalogue-image content requires image id", label)
		}
		return nil
	case ContentCustomImage:
		if c.Image == "" {
			return fmt.Errorf("media recipe: partition %q: custom-image content requires image id", label)
		}
		return nil
	case ContentImageFile:
		if c.Image == "" {
			return fmt.Errorf("media recipe: partition %q: image-file content requires image id", label)
		}
		return nil
	case ContentRenderedProfile:
		if c.Profile == "" {
			return fmt.Errorf("media recipe: partition %q: rendered-profile content requires profile name", label)
		}
		return nil
	case ContentProfileBundle:
		if c.Bundle == "" {
			return fmt.Errorf("media recipe: partition %q: profile-bundle content requires bundle path", label)
		}
		return nil
	case ContentEmpty, ContentCustomFiles:
		return nil
	case "":
		return fmt.Errorf("media recipe: partition %q: content type is required", label)
	default:
		return fmt.Errorf("media recipe: partition %q: unsupported content type %q", label, c.Type)
	}
}

// parseSize parses a human-readable size string like "2G", "500M", "1.5G".
// Returns bytes.
func parseSize(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty size string")
	}
	upper := strings.ToUpper(s)
	multipliers := []struct {
		suffix string
		factor int64
	}{
		{"TIB", 1 << 40},
		{"GIB", 1 << 30},
		{"MIB", 1 << 20},
		{"KIB", 1 << 10},
		{"TB", 1_000_000_000_000},
		{"GB", 1_000_000_000},
		{"MB", 1_000_000},
		{"KB", 1_000},
		{"T", 1 << 40},
		{"G", 1 << 30},
		{"M", 1 << 20},
		{"K", 1 << 10},
	}
	for _, m := range multipliers {
		if strings.HasSuffix(upper, m.suffix) {
			numStr := s[:len(s)-len(m.suffix)]
			f, err := strconv.ParseFloat(strings.TrimSpace(numStr), 64)
			if err != nil {
				return 0, fmt.Errorf("invalid size %q", s)
			}
			if f <= 0 {
				return 0, fmt.Errorf("size must be positive, got %q", s)
			}
			return int64(f * float64(m.factor)), nil
		}
	}
	// plain bytes
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid size %q", s)
	}
	return n, nil
}

// ParseSize is the exported form of parseSize for use by the planner.
func ParseSize(s string) (int64, error) {
	if strings.EqualFold(s, "remaining") {
		return -1, nil
	}
	return parseSize(s)
}
