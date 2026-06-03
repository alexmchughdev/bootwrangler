import { useState } from "react";

export default function Lab() {
  const [profile, setProfile] = useState("");
  const [memory, setMemory] = useState("2048");
  const [cpus, setCpus] = useState("2");

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
                disabled
                title="Backend wiring not yet available"
              >
                Start VM
              </button>
            </div>
          </div>

          <div className="panel lab-requirement-panel">
            <span className="panel-label">Requirements</span>
            <strong className="lab-req-title">QEMU is required</strong>
            <p>
              The lab feature requires <code className="inline-code">qemu-system-x86_64</code>{" "}
              to be installed and available on your PATH. Install it with your system package
              manager:
            </p>
            <pre className="lab-install-hint">{`# Debian / Ubuntu
sudo apt install qemu-system-x86

# Fedora / RHEL
sudo dnf install qemu-system-x86

# Arch Linux
sudo pacman -S qemu-full`}</pre>
          </div>
        </div>

        <div className="lab-right">
          <div className="panel lab-vms-panel">
            <span className="panel-label">Active VMs</span>
            <p className="library-empty">
              No active VMs. Start one using the form above.
            </p>
          </div>

          <div className="panel lab-console-panel">
            <span className="panel-label">Serial Console</span>
            <textarea
              className="lab-console"
              readOnly
              value="Serial console output will appear here once a VM is running."
              spellCheck={false}
            />
          </div>

          <div className="panel lab-snapshots-panel">
            <span className="panel-label">Snapshots</span>
            <p className="library-empty">Select a running VM to manage snapshots.</p>
          </div>
        </div>
      </div>
    </div>
  );
}
