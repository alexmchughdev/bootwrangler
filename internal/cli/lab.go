package cli

import (
	"fmt"
	"io"
	"strconv"

	"github.com/alexmchughdev/bootwrangler/internal/lab"
)

func runLab(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		printLabUsage(stderr)
		return 2
	}

	switch args[0] {
	case "start":
		return runLabStart(args[1:], stdout, stderr)
	case "stop":
		return runLabStop(args[1:], stdout, stderr)
	case "status":
		return runLabStatus(args[1:], stdout, stderr)
	case "console":
		return runLabConsole(args[1:], stdout, stderr)
	case "ssh":
		return runLabSSH(args[1:], stdout, stderr)
	case "plan":
		return runLabPlan(args[1:], stdout, stderr)
	case "snapshot":
		return runLabSnapshot(args[1:], stdout, stderr)
	default:
		fmt.Fprintf(stderr, "unknown lab command %q\n", args[0])
		return 2
	}
}

func runLabStart(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "usage: bootwrangler lab start <profile-name> [--memory 2048] [--cpus 2]")
		return 2
	}

	profileName := args[0]
	opts := lab.StartOptions{
		ProfileName: profileName,
		MemoryMB:    2048,
		CPUs:        2,
	}

	i := 1
	for i < len(args) {
		switch args[i] {
		case "--memory":
			if i+1 >= len(args) {
				fmt.Fprintln(stderr, "lab start: --memory requires a value")
				return 2
			}
			v, err := strconv.Atoi(args[i+1])
			if err != nil {
				fmt.Fprintf(stderr, "lab start: invalid --memory value %q\n", args[i+1])
				return 2
			}
			opts.MemoryMB = v
			i += 2
		case "--cpus":
			if i+1 >= len(args) {
				fmt.Fprintln(stderr, "lab start: --cpus requires a value")
				return 2
			}
			v, err := strconv.Atoi(args[i+1])
			if err != nil {
				fmt.Fprintf(stderr, "lab start: invalid --cpus value %q\n", args[i+1])
				return 2
			}
			opts.CPUs = v
			i += 2
		default:
			fmt.Fprintf(stderr, "lab start: unexpected argument %q\n", args[i])
			return 2
		}
	}

	run, err := lab.NewRun(opts)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	if err := lab.Start(run); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	fmt.Fprintf(stdout, "lab VM started\n")
	fmt.Fprintf(stdout, "  run ID:   %s\n", run.ID)
	fmt.Fprintf(stdout, "  profile:  %s\n", run.ProfileName)
	fmt.Fprintf(stdout, "  ssh port: %d\n", run.SSHPort)
	fmt.Fprintf(stdout, "  vnc port: %d\n", run.VNCPort)
	fmt.Fprintf(stdout, "  work dir: %s\n", run.WorkDir)
	fmt.Fprintf(stdout, "  ssh cmd:  %s\n", lab.SSHCommand(run, "user"))
	return 0
}

func runLabStop(args []string, stdout, stderr io.Writer) int {
	if len(args) != 1 {
		fmt.Fprintln(stderr, "usage: bootwrangler lab stop <run-id>")
		return 2
	}
	runID := args[0]
	run := &lab.Run{
		ID:      runID,
		WorkDir: lab.WorkDir(runID),
	}
	if err := lab.Stop(run); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	fmt.Fprintf(stdout, "lab VM stopped: %s\n", runID)
	return 0
}

func runLabStatus(args []string, stdout, stderr io.Writer) int {
	if len(args) != 1 {
		fmt.Fprintln(stderr, "usage: bootwrangler lab status <run-id>")
		return 2
	}
	runID := args[0]
	run := &lab.Run{
		ID:      runID,
		WorkDir: lab.WorkDir(runID),
	}
	running := lab.IsRunning(run)
	state := "stopped"
	if running {
		state = "running"
	}
	fmt.Fprintf(stdout, "run ID: %s\n", runID)
	fmt.Fprintf(stdout, "state:  %s\n", state)
	return 0
}

func runLabConsole(args []string, stdout, stderr io.Writer) int {
	if len(args) != 1 {
		fmt.Fprintln(stderr, "usage: bootwrangler lab console <run-id>")
		return 2
	}
	runID := args[0]
	run := &lab.Run{
		ID:        runID,
		WorkDir:   lab.WorkDir(runID),
		SerialLog: lab.WorkDir(runID) + "/serial.log",
	}
	content, err := lab.ReadSerialLog(run)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	fmt.Fprint(stdout, content)
	return 0
}

func runLabSSH(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "usage: bootwrangler lab ssh <run-id> [--user <username>]")
		return 2
	}
	runID := args[0]
	user := "user"

	i := 1
	for i < len(args) {
		switch args[i] {
		case "--user":
			if i+1 >= len(args) {
				fmt.Fprintln(stderr, "lab ssh: --user requires a value")
				return 2
			}
			user = args[i+1]
			i += 2
		default:
			fmt.Fprintf(stderr, "lab ssh: unexpected argument %q\n", args[i])
			return 2
		}
	}

	run := &lab.Run{
		ID:      runID,
		WorkDir: lab.WorkDir(runID),
	}
	fmt.Fprintln(stdout, lab.SSHCommand(run, user))
	return 0
}

func runLabPlan(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "usage: bootwrangler lab plan <profile-name>")
		return 2
	}
	profileName := args[0]
	opts := lab.StartOptions{
		ProfileName: profileName,
		MemoryMB:    2048,
		CPUs:        2,
		SSHPort:     12200,
		VNCPort:     15900,
	}
	run, err := lab.NewRun(opts)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	planArgs, err := lab.PlanQEMU(run)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	fmt.Fprintf(stdout, "qemu-system-x86_64")
	for _, a := range planArgs {
		fmt.Fprintf(stdout, " %s", a)
	}
	fmt.Fprintln(stdout)
	return 0
}

func runLabSnapshot(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "usage:")
		fmt.Fprintln(stderr, "  bootwrangler lab snapshot create <run-id> <name>")
		fmt.Fprintln(stderr, "  bootwrangler lab snapshot list   <run-id>")
		fmt.Fprintln(stderr, "  bootwrangler lab snapshot revert <run-id> <name>")
		return 2
	}

	switch args[0] {
	case "create":
		if len(args) != 3 {
			fmt.Fprintln(stderr, "usage: bootwrangler lab snapshot create <run-id> <name>")
			return 2
		}
		runID, name := args[1], args[2]
		run := runFromID(runID)
		if err := lab.CreateSnapshot(run, name); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		fmt.Fprintf(stdout, "snapshot created: %s\n", name)
		return 0

	case "list":
		if len(args) != 2 {
			fmt.Fprintln(stderr, "usage: bootwrangler lab snapshot list <run-id>")
			return 2
		}
		runID := args[1]
		run := runFromID(runID)
		names, err := lab.ListSnapshots(run)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		if len(names) == 0 {
			fmt.Fprintln(stdout, "no snapshots")
			return 0
		}
		for _, n := range names {
			fmt.Fprintln(stdout, n)
		}
		return 0

	case "revert":
		if len(args) != 3 {
			fmt.Fprintln(stderr, "usage: bootwrangler lab snapshot revert <run-id> <name>")
			return 2
		}
		runID, name := args[1], args[2]
		run := runFromID(runID)
		if err := lab.RevertSnapshot(run, name); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		fmt.Fprintf(stdout, "reverted to snapshot: %s\n", name)
		return 0

	default:
		fmt.Fprintf(stderr, "unknown snapshot command %q\n", args[0])
		return 2
	}
}

// runFromID builds a minimal Run from a run ID so CLI sub-commands can
// reference disk paths without needing the full in-memory state.
func runFromID(id string) *lab.Run {
	wd := lab.WorkDir(id)
	return &lab.Run{
		ID:        id,
		WorkDir:   wd,
		DiskPath:  wd + "/disk.qcow2",
		SerialLog: wd + "/serial.log",
	}
}

func printLabUsage(w io.Writer) {
	fmt.Fprintln(w, "usage:")
	fmt.Fprintln(w, "  bootwrangler lab start <profile-name> [--memory 2048] [--cpus 2]")
	fmt.Fprintln(w, "  bootwrangler lab stop <run-id>")
	fmt.Fprintln(w, "  bootwrangler lab status <run-id>")
	fmt.Fprintln(w, "  bootwrangler lab console <run-id>")
	fmt.Fprintln(w, "  bootwrangler lab ssh <run-id> [--user <username>]")
	fmt.Fprintln(w, "  bootwrangler lab plan <profile-name>")
	fmt.Fprintln(w, "  bootwrangler lab snapshot create <run-id> <name>")
	fmt.Fprintln(w, "  bootwrangler lab snapshot list <run-id>")
	fmt.Fprintln(w, "  bootwrangler lab snapshot revert <run-id> <name>")
}
