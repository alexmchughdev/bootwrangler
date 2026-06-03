package inspect

import (
	"testing"
)

func TestGatherHostInfo_NoError(t *testing.T) {
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

func TestNetworkInterfaces(t *testing.T) {
	ifaces, err := NetworkInterfaces()
	if err != nil {
		t.Fatalf("NetworkInterfaces() error = %v", err)
	}
	if len(ifaces) == 0 {
		t.Error("NetworkInterfaces() returned no interfaces")
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
