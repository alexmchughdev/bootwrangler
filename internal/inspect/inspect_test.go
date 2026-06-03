package inspect

import (
	"runtime"
	"strings"
	"testing"
)

func TestGatherHostInfo_NoError(t *testing.T) {
	skipUnsupportedHostInspection(t)

	info, err := GatherHostInfo()
	if err != nil {
		t.Fatalf("GatherHostInfo() unexpected error: %v", err)
	}
	if info.Hostname == "" {
		t.Error("GatherHostInfo() returned empty hostname")
	}
}

func TestHostname(t *testing.T) {
	h, err := Hostname()
	if err != nil {
		t.Fatalf("Hostname() error = %v", err)
	}
	if h == "" {
		t.Error("Hostname() returned empty string")
	}
}

func TestMemoryInfo(t *testing.T) {
	skipUnsupportedHostInspection(t)

	m, err := MemoryInfo_()
	if err != nil {
		t.Fatalf("MemoryInfo_() error = %v", err)
	}
	if m.TotalBytes <= 0 {
		t.Errorf("MemoryInfo_() TotalBytes = %d, want > 0", m.TotalBytes)
	}
	if m.TotalHuman == "" {
		t.Error("MemoryInfo_() TotalHuman is empty")
	}
}

func TestCPUInfo(t *testing.T) {
	skipUnsupportedHostInspection(t)

	c, err := CPUInfo_()
	if err != nil {
		t.Fatalf("CPUInfo_() error = %v", err)
	}
	if c.Threads <= 0 {
		t.Errorf("CPUInfo_() Threads = %d, want > 0", c.Threads)
	}
	if c.Cores <= 0 {
		t.Errorf("CPUInfo_() Cores = %d, want > 0", c.Cores)
	}
}

func TestMemoryInfoFromReader(t *testing.T) {
	m, err := memoryInfoFromReader(strings.NewReader("MemTotal:        2048 kB\n"), "test meminfo")
	if err != nil {
		t.Fatalf("memoryInfoFromReader() error = %v", err)
	}
	if m.TotalBytes != 2*1024*1024 {
		t.Errorf("memoryInfoFromReader() TotalBytes = %d, want %d", m.TotalBytes, 2*1024*1024)
	}
	if m.TotalHuman != "2.0 MB" {
		t.Errorf("memoryInfoFromReader() TotalHuman = %q, want %q", m.TotalHuman, "2.0 MB")
	}
}

func TestMemoryInfoFromHostinfo(t *testing.T) {
	m, err := memoryInfoFromHostinfo("Primary memory available: 16.00 gigabytes\n")
	if err != nil {
		t.Fatalf("memoryInfoFromHostinfo() error = %v", err)
	}
	if m.TotalBytes != 16*1024*1024*1024 {
		t.Errorf("memoryInfoFromHostinfo() TotalBytes = %d, want %d", m.TotalBytes, 16*1024*1024*1024)
	}
	if m.TotalHuman != "16.0 GB" {
		t.Errorf("memoryInfoFromHostinfo() TotalHuman = %q, want %q", m.TotalHuman, "16.0 GB")
	}
}

func TestCPUInfoFromReader(t *testing.T) {
	data := strings.Join([]string{
		"processor\t: 0",
		"model name\t: Test CPU",
		"physical id\t: 0",
		"core id\t\t: 0",
		"",
		"processor\t: 1",
		"model name\t: Test CPU",
		"physical id\t: 0",
		"core id\t\t: 1",
		"",
	}, "\n")

	c, err := cpuInfoFromReader(strings.NewReader(data), "test cpuinfo")
	if err != nil {
		t.Fatalf("cpuInfoFromReader() error = %v", err)
	}
	if c.Model != "Test CPU" {
		t.Errorf("cpuInfoFromReader() Model = %q, want %q", c.Model, "Test CPU")
	}
	if c.Threads != 2 {
		t.Errorf("cpuInfoFromReader() Threads = %d, want 2", c.Threads)
	}
	if c.Cores != 2 {
		t.Errorf("cpuInfoFromReader() Cores = %d, want 2", c.Cores)
	}
}

func TestNetworkInterfaces(t *testing.T) {
	ifaces, err := NetworkInterfaces()
	if err != nil {
		t.Fatalf("NetworkInterfaces() error = %v", err)
	}
	if len(ifaces) == 0 {
		t.Error("NetworkInterfaces() returned no interfaces")
	}
}

func skipUnsupportedHostInspection(t *testing.T) {
	t.Helper()
	switch runtime.GOOS {
	case "darwin", "linux":
	default:
		t.Skipf("host inspection is not implemented on %s", runtime.GOOS)
	}
}

func TestFormatBytes(t *testing.T) {
	cases := []struct {
		input int64
		want  string
	}{
		{500, "500 B"},
		{2048, "2.0 KB"},
		{3 * 1024 * 1024, "3.0 MB"},
		{4 * 1024 * 1024 * 1024, "4.0 GB"},
	}
	for _, tc := range cases {
		got := FormatBytes(tc.input)
		if got != tc.want {
			t.Errorf("FormatBytes(%d) = %q, want %q", tc.input, got, tc.want)
		}
	}
}
