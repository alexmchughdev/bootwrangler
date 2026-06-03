// Package inspect provides lightweight host inspection by reading system
// information directly from /proc and the standard library where possible.
package inspect

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

// CPUInfo holds basic processor information.
type CPUInfo struct {
	Model   string
	Cores   int
	Threads int
}

// MemoryInfo holds physical memory information.
type MemoryInfo struct {
	TotalBytes int64
	TotalHuman string
}

// DiskInfo holds information about one block device or mount point.
type DiskInfo struct {
	Path      string
	SizeBytes int64
	SizeHuman string
	Fstype    string
}

// NetworkInterface holds information about one network interface.
type NetworkInterface struct {
	Name      string
	Addresses []string
	HWAddr    string
}

// HostInfo aggregates all gathered host information.
type HostInfo struct {
	Hostname   string
	CPU        CPUInfo
	Memory     MemoryInfo
	Disks      []DiskInfo
	Interfaces []NetworkInterface
}

// GatherHostInfo collects all available host information and returns it as a
// HostInfo. Partial results may be returned alongside a non-nil error if only
// some subsystems fail.
func GatherHostInfo() (HostInfo, error) {
	var info HostInfo
	var errs []string

	host, err := Hostname()
	if err != nil {
		errs = append(errs, "hostname: "+err.Error())
	}
	info.Hostname = host

	mem, err := MemoryInfo_()
	if err != nil {
		errs = append(errs, "memory: "+err.Error())
	}
	info.Memory = mem

	cpu, err := CPUInfo_()
	if err != nil {
		errs = append(errs, "cpu: "+err.Error())
	}
	info.CPU = cpu

	ifaces, err := NetworkInterfaces()
	if err != nil {
		errs = append(errs, "network: "+err.Error())
	}
	info.Interfaces = ifaces

	if len(errs) > 0 {
		return info, fmt.Errorf("inspect: %s", strings.Join(errs, "; "))
	}
	return info, nil
}

// Hostname returns the machine's hostname via os.Hostname.
func Hostname() (string, error) {
	return os.Hostname()
}

// MemoryInfo_ reads /proc/meminfo and returns total physical memory.
func MemoryInfo_() (MemoryInfo, error) {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return MemoryInfo{}, fmt.Errorf("open /proc/meminfo: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "MemTotal:") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			break
		}
		kb, err := strconv.ParseInt(fields[1], 10, 64)
		if err != nil {
			return MemoryInfo{}, fmt.Errorf("parse MemTotal: %w", err)
		}
		total := kb * 1024
		return MemoryInfo{
			TotalBytes: total,
			TotalHuman: FormatBytes(total),
		}, nil
	}
	if err := scanner.Err(); err != nil {
		return MemoryInfo{}, fmt.Errorf("scan /proc/meminfo: %w", err)
	}
	return MemoryInfo{}, fmt.Errorf("MemTotal not found in /proc/meminfo")
}

// CPUInfo_ reads /proc/cpuinfo and returns CPU model, physical core count, and
// logical thread count.
func CPUInfo_() (CPUInfo, error) {
	f, err := os.Open("/proc/cpuinfo")
	if err != nil {
		return CPUInfo{}, fmt.Errorf("open /proc/cpuinfo: %w", err)
	}
	defer f.Close()

	var info CPUInfo
	physicalIDs := make(map[string]bool)
	coreIDs := make(map[string]bool)
	threads := 0
	currentPhysical := ""

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "model name") {
			if info.Model == "" {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					info.Model = strings.TrimSpace(parts[1])
				}
			}
		} else if strings.HasPrefix(line, "physical id") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				currentPhysical = strings.TrimSpace(parts[1])
				physicalIDs[currentPhysical] = true
			}
		} else if strings.HasPrefix(line, "core id") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := currentPhysical + ":" + strings.TrimSpace(parts[1])
				coreIDs[key] = true
			}
		} else if strings.HasPrefix(line, "processor") {
			threads++
		}
	}
	if err := scanner.Err(); err != nil {
		return CPUInfo{}, fmt.Errorf("scan /proc/cpuinfo: %w", err)
	}

	info.Threads = threads
	if len(coreIDs) > 0 {
		info.Cores = len(coreIDs)
	} else {
		// Fallback: assume threads == cores (e.g. VMs without core id)
		info.Cores = threads
	}

	return info, nil
}

// NetworkInterfaces returns all network interfaces using the standard library.
func NetworkInterfaces() ([]NetworkInterface, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("net.Interfaces: %w", err)
	}

	result := make([]NetworkInterface, 0, len(ifaces))
	for _, iface := range ifaces {
		ni := NetworkInterface{
			Name:   iface.Name,
			HWAddr: iface.HardwareAddr.String(),
		}
		addrs, err := iface.Addrs()
		if err == nil {
			for _, a := range addrs {
				ni.Addresses = append(ni.Addresses, a.String())
			}
		}
		result = append(result, ni)
	}
	return result, nil
}

// FormatBytes converts a byte count into a human-readable string (KB/MB/GB).
func FormatBytes(n int64) string {
	const (
		kb = 1024
		mb = 1024 * kb
		gb = 1024 * mb
	)
	switch {
	case n >= gb:
		return fmt.Sprintf("%.1f GB", float64(n)/float64(gb))
	case n >= mb:
		return fmt.Sprintf("%.1f MB", float64(n)/float64(mb))
	case n >= kb:
		return fmt.Sprintf("%.1f KB", float64(n)/float64(kb))
	default:
		return fmt.Sprintf("%d B", n)
	}
}
