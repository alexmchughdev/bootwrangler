import { useCallback, useEffect, useRef, useState } from "react";
import {
  labCreateSnapshot,
  labDeleteSnapshot,
  labInstallQEMU,
  labListRuns,
  labListSnapshots,
  labQEMUStatus,
  labRevertSnapshot,
  labSSHCommand,
  labSerialLog,
  labStart,
  labStop,
  type LabRun,
  type QEMUAvailability,
} from "../api/backend";

const LS_LAB_PROFILE = "bw_lab_profile_name";

export default function Lab() {
  const [profile, setProfile] = useState(() => {
    const saved = localStorage.getItem(LS_LAB_PROFILE);
    if (saved) {
      localStorage.removeItem(LS_LAB_PROFILE);
      return saved;
    }
    return "";
  });
  const [memory, setMemory] = useState("2048");
  const [cpus, setCpus] = useState("2");
  const [activeRuns, setActiveRuns] = useState<LabRun[]>([]);
  const [error, setError] = useState("");
  const [selectedRunID, setSelectedRunID] = useState<string | null>(null);
  const [consoleLog, setConsoleLog] = useState("");
  const [snapshotNames, setSnapshotNames] = useState<Record<string, string>>({});
  const [copiedRunID, setCopiedRunID] = useState<string | null>(null);
  const [snapshotLists, setSnapshotLists] = useState<Record<string, string[]>>({});
  const consoleRef = useRef<HTMLDivElement>(null);
  const pollRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const refreshRuns = useCallback(async () => {
    try {
      const runs = await labListRuns();
      setActiveRuns(runs);
    } catch {
      // silent poll failures
    }
  }, []);

  useEffect(() => {
    void refreshRuns();
    pollRef.current = setInterval(() => void refreshRuns(), 5000);
    return () => {
      if (pollRef.current) clearInterval(pollRef.current);
    };
  }, [refreshRuns]);

  async function handleStart() {
    setError("");
    try {
      const run = await labStart(profile, parseInt(memory, 10), parseInt(cpus, 10));
      setActiveRuns((prev) => [...prev, run]);
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : String(err));
    }
  }

  async function handleStop(runID: string) {
    setError("");
    try {
      await labStop(runID);
      setActiveRuns((prev) =>
        prev.map((r) => (r.ID === runID ? { ...r, State: "stopped" } : r)),
      );
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : String(err));
    }
  }

  async function handleConsole(runID: string) {
    setError("");
    setSelectedRunID(runID);
    try {
      const log = await labSerialLog(runID);
      setConsoleLog(log);
      consoleRef.current?.scrollIntoView({ behavior: "smooth" });
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : String(err));
    }
  }

  async function handleRefreshConsole() {
    if (!selectedRunID) return;
    setError("");
    try {
      const log = await labSerialLog(selectedRunID);
      setConsoleLog(log);
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : String(err));
    }
  }

  async function handleSSHCommand(runID: string) {
    setError("");
    try {
      const cmd = await labSSHCommand(runID, "root");
      await navigator.clipboard.writeText(cmd);
      setCopiedRunID(runID);
      setTimeout(() => setCopiedRunID((prev) => (prev === runID ? null : prev)), 2000);
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : String(err));
    }
  }

  async function handleCopyVNC(run: LabRun) {
    await navigator.clipboard.writeText(`localhost:${run.VNCPort}`);
    setCopiedRunID(`vnc-${run.ID}`);
    setTimeout(
      () => setCopiedRunID((prev) => (prev === `vnc-${run.ID}` ? null : prev)),
      2000,
    );
  }

  async function handleCreateSnapshot(runID: string) {
    setError("");
    const name = snapshotNames[runID] ?? "";
    if (!name.trim()) {
      setError("Snapshot name cannot be empty.");
      return;
    }
    try {
      await labCreateSnapshot(runID, name.trim());
      setSnapshotNames((prev) => ({ ...prev, [runID]: "" }));
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : String(err));
    }
  }

  async function handleListSnapshots(runID: string) {
    setError("");
    try {
      const snaps = await labListSnapshots(runID);
      setSnapshotLists((prev) => ({ ...prev, [runID]: snaps }));
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : String(err));
    }
  }

  async function handleRevertSnapshot(runID: string, snap: string) {
    setError("");
    try {
      await labRevertSnapshot(runID, snap);
      setError(`Reverted to snapshot "${snap}"`);
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : String(err));
    }
  }

  async function handleDeleteSnapshot(runID: string, snap: string) {
    setError("");
    try {
      await labDeleteSnapshot(runID, snap);
      const snaps = await labListSnapshots(runID);
      setSnapshotLists((prev) => ({ ...prev, [runID]: snaps }));
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : String(err));
    }
  }

  function stateClass(state: string): string {
    switch (state) {
      case "running":
        return "lab-state-badge lab-state-running";
      case "stopped":
        return "lab-state-badge lab-state-stopped";
      case "failed":
        return "lab-state-badge lab-state-failed";
      default:
        return "lab-state-badge lab-state-pending";
    }
  }

  return (
    <div className="lab-page">
      <div className="panel panel-wide lab-intro">
        <span className="panel-label">QEMU Lab</span>
        <h2>Ephemeral VM testing</h2>
        <p>
          Launch ephemeral VMs from profiles to validate installer configs before flashing
          to hardware. Each VM runs in isolation and is discarded when stopped.
        </p>
      </div>

      {error && <p className="render-error">{error}</p>}

      <div className="lab-columns">
        <div className="lab-left">
          <div className="panel lab-start-panel">
            <span className="panel-label">Start VM</span>
            <div className="lab-form">
              <label className="field">
                <span>Profile Name</span>
                <input
                  type="text"
                  placeholder="e.g. ubuntu-server"
                  value={profile}
                  onChange={(e) => setProfile(e.target.value)}
                />
              </label>
              <label className="field">
                <span>Memory</span>
                <select value={memory} onChange={(e) => setMemory(e.target.value)}>
                  <option value="1024">1024 MB</option>
                  <option value="2048">2048 MB</option>
                  <option value="4096">4096 MB</option>
                </select>
              </label>
              <label className="field">
                <span>CPUs</span>
                <select value={cpus} onChange={(e) => setCpus(e.target.value)}>
                  <option value="1">1</option>
                  <option value="2">2</option>
                  <option value="4">4</option>
                </select>
              </label>
              <button
                className="primary-action lab-start-btn"
                type="button"
                onClick={() => void handleStart()}
                disabled={!profile.trim()}
              >
                Start VM
              </button>
            </div>
          </div>

          <LabSetup />
        </div>

        <div className="lab-right">
          <div className="panel lab-vms-panel">
            <span className="panel-label">Active VMs</span>
            {activeRuns.length === 0 ? (
              <p className="library-empty">No active VMs. Start one using the form above.</p>
            ) : (
              <div className="lab-vm-list">
                {activeRuns.map((run) => (
                  <div key={run.ID} className="lab-vm-card">
                    <div className="lab-vm-header">
                      <div className="lab-vm-title">
                        <code className="inline-code">{run.ID.slice(0, 8)}</code>
                        <span className="lab-vm-profile">{run.ProfileName}</span>
                      </div>
                      <span className={stateClass(run.State)}>{run.State}</span>
                    </div>
                    <div className="lab-vm-ports">
                      <span>SSH: {run.SSHPort}</span>
                      <span>VNC: {run.VNCPort}</span>
                    </div>
                    {run.Error && <p className="render-error lab-vm-error">{run.Error}</p>}
                    <div className="lab-vm-actions">
                      <button
                        className="secondary-action"
                        type="button"
                        onClick={() => void handleStop(run.ID)}
                        disabled={run.State === "stopped" || run.State === "failed"}
                      >
                        Stop
                      </button>
                      <button
                        className="secondary-action"
                        type="button"
                        onClick={() => void handleConsole(run.ID)}
                      >
                        Console
                      </button>
                      <button
                        className="secondary-action"
                        type="button"
                        onClick={() => void handleSSHCommand(run.ID)}
                        title="Copy SSH command to clipboard"
                      >
                        {copiedRunID === run.ID ? "Copied!" : "SSH Command"}
                      </button>
                      <button
                        className="secondary-action"
                        type="button"
                        onClick={() => void handleCopyVNC(run)}
                        title="Copy VNC address"
                      >
                        {copiedRunID === `vnc-${run.ID}` ? "Copied!" : "Copy VNC"}
                      </button>
                    </div>
                    <div className="lab-snapshot-section">
                      <span className="lab-snapshot-label">Snapshots</span>
                      <div className="lab-snapshot-row">
                        <input
                          type="text"
                          placeholder="Snapshot name"
                          value={snapshotNames[run.ID] ?? ""}
                          onChange={(e) =>
                            setSnapshotNames((prev) => ({
                              ...prev,
                              [run.ID]: e.target.value,
                            }))
                          }
                        />
                        <button
                          className="secondary-action"
                          type="button"
                          onClick={() => void handleCreateSnapshot(run.ID)}
                        >
                          Create
                        </button>
                        <button
                          className="secondary-action"
                          type="button"
                          onClick={() => void handleListSnapshots(run.ID)}
                        >
                          List
                        </button>
                      </div>
                      {snapshotLists[run.ID] && (
                        <ul className="lab-snapshot-list">
                          {snapshotLists[run.ID].length === 0 ? (
                            <li className="library-empty">No snapshots.</li>
                          ) : (
                            snapshotLists[run.ID].map((snap) => (
                              <li key={snap} className="lab-snapshot-item">
                                <code className="inline-code">{snap}</code>
                                <button
                                  type="button"
                                  className="link-btn"
                                  onClick={() => void handleRevertSnapshot(run.ID, snap)}
                                >
                                  Revert
                                </button>
                                <button
                                  type="button"
                                  className="link-btn danger"
                                  onClick={() => void handleDeleteSnapshot(run.ID, snap)}
                                >
                                  Delete
                                </button>
                              </li>
                            ))
                          )}
                        </ul>
                      )}
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>

          <div className="panel lab-console-panel" ref={consoleRef}>
            <div className="lab-console-header">
              <span className="panel-label">Serial Console</span>
              {selectedRunID && (
                <span className="lab-console-run-id">
                  Selected run: <code className="inline-code">{selectedRunID.slice(0, 8)}</code>
                </span>
              )}
              <button
                className="secondary-action"
                type="button"
                onClick={() => void handleRefreshConsole()}
                disabled={!selectedRunID}
              >
                Refresh
              </button>
            </div>
            <textarea
              className="lab-console"
              readOnly
              value={
                consoleLog ||
                (selectedRunID
                  ? ""
                  : "Select a VM and click Console to load serial output.")
              }
              spellCheck={false}
            />
          </div>
        </div>
      </div>
    </div>
  );
}

// ---------------------------------------------------------------------------
// Lab setup — QEMU status + one-click install
// ---------------------------------------------------------------------------

function LabSetup() {
  const [status, setStatus] = useState<QEMUAvailability | null>(null);
  const [installing, setInstalling] = useState(false);
  const [log, setLog] = useState("");
  const [installError, setInstallError] = useState("");

  const refresh = useCallback(async () => {
    try {
      setStatus(await labQEMUStatus());
    } catch {
      // ignore — preview mode
    }
  }, []);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  async function handleInstall() {
    setInstalling(true);
    setInstallError("");
    setLog("");
    try {
      const out = await labInstallQEMU();
      setLog(out);
      await refresh();
    } catch (err: unknown) {
      setInstallError(err instanceof Error ? err.message : String(err));
    } finally {
      setInstalling(false);
    }
  }

  if (status === null) {
    return (
      <div className="panel lab-setup-panel">
        <span className="panel-label">Lab Engine</span>
        <p className="library-empty">Checking for QEMU…</p>
      </div>
    );
  }

  if (status.installed) {
    const sourceLabel =
      status.source === "bundled"
        ? "Bundled with BootWrangler"
        : status.source === "installed"
          ? "Installed by BootWrangler"
          : "Found on your system";
    return (
      <div className="panel lab-setup-panel lab-setup-ready">
        <span className="panel-label">Lab Engine</span>
        <div className="lab-ready-row">
          <span className="lab-ready-check">✓</span>
          <div>
            <strong className="lab-ready-title">Lab is ready</strong>
            <p className="lab-ready-sub">
              {status.version || "QEMU"} · {sourceLabel}
            </p>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="panel lab-setup-panel">
      <span className="panel-label">Lab Engine</span>
      <strong className="lab-req-title">One-time setup</strong>
      <p>
        The Lab runs your profiles in a real VM using QEMU. BootWrangler can install
        it for you — no terminal required.
      </p>

      {status.can_auto ? (
        <button
          type="button"
          className="primary-action lab-setup-btn"
          onClick={() => void handleInstall()}
          disabled={installing}
        >
          {installing ? "Setting up the lab…" : "Set Up Lab"}
        </button>
      ) : (
        <div className="lab-setup-manual">
          <p className="lab-setup-manual-note">
            Automatic setup needs a package manager that isn't available here. Run this
            once to finish setup:
          </p>
          <pre className="lab-install-hint">{status.hint}</pre>
          <button
            type="button"
            className="secondary-action"
            onClick={() => void refresh()}
          >
            I've installed it — recheck
          </button>
        </div>
      )}

      {installError && (
        <div className="lab-setup-error">
          <p className="render-error">{installError}</p>
          <pre className="lab-install-hint">{status.hint}</pre>
        </div>
      )}

      {log && <pre className="lab-setup-log">{log}</pre>}
    </div>
  );
}
