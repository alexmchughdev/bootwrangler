package lab

import "time"

// timeNow returns the current time. Exposed as a variable for testability.
var timeNow = func() time.Time { return time.Now() }

// MemoryMB returns the RAM allocation in MiB (default 2048).
func (r *Run) MemoryMB() int {
	if r.memoryMB == 0 {
		return 2048
	}
	return r.memoryMB
}

// CPUs returns the virtual CPU count (default 2).
func (r *Run) CPUs() int {
	if r.cpus == 0 {
		return 2
	}
	return r.cpus
}

// DiskSizeGB returns the disk size in GiB (default 20).
func (r *Run) DiskSizeGB() int {
	if r.diskSizeGB == 0 {
		return 20
	}
	return r.diskSizeGB
}
