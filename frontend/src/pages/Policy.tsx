import { useState } from "react";
import { checkPolicy, loadProfile, type PolicyCheckResult, type PolicySnapshot } from "../api/backend";
import type { Profile } from "../types/profile";

const POLICY_PLACEHOLDER = `# Policy rules — all fields are optional
require_ssh_key_only: true
forbid_root_login: true
require_disk_confirm: true
min_users: 1
# require_users:
#   - deploy
# allowed_distros:
#   - ubuntu
#   - debian
# forbid_packages:
#   - telnet
# require_package_presets:
#   - base`;

function profileToSnapshot(profile: Profile): PolicySnapshot {
  return {
    OSFamily: profile.os.family,
    SSHPasswordAuth: profile.ssh.password_authentication,
    SSHPermitRootLogin: profile.ssh.permit_root_login,
    DiskConfirmDestructive: profile.disk.confirm_destructive,
    Users: profile.users.map((u) => u.name),
    Packages: profile.packages.names,
    PackagePresets: profile.packages.presets,
  };
}

type CheckState =
  | { kind: "idle" }
  | { kind: "loading" }
  | { kind: "error"; message: string }
  | { kind: "done"; result: PolicyCheckResult; snap: PolicySnapshot };

export default function Policy() {
  const [policyYAML, setPolicyYAML] = useState(POLICY_PLACEHOLDER);
  const [profilePath, setProfilePath] = useState("");
  const [profileLoading, setProfileLoading] = useState(false);
  const [loadedProfile, setLoadedProfile] = useState<Profile | null>(null);
  const [loadError, setLoadError] = useState("");
  const [checkState, setCheckState] = useState<CheckState>({ kind: "idle" });

  async function handleLoadProfile() {
    if (!profilePath.trim()) return;
    setProfileLoading(true);
    setLoadError("");
    setLoadedProfile(null);
    setCheckState({ kind: "idle" });
    try {
      const p = await loadProfile(profilePath.trim());
      setLoadedProfile(p);
    } catch (err) {
      setLoadError(String(err));
    } finally {
      setProfileLoading(false);
    }
  }

  async function handleCheck() {
    if (!loadedProfile) return;
    const snap = profileToSnapshot(loadedProfile);
    setCheckState({ kind: "loading" });
    try {
      const result = await checkPolicy(policyYAML, snap);
      setCheckState({ kind: "done", result, snap });
    } catch (err) {
      setCheckState({ kind: "error", message: String(err) });
    }
  }

  const snap = loadedProfile ? profileToSnapshot(loadedProfile) : null;

  return (
    <div className="policy-page">
      <div className="panel panel-wide policy-intro">
        <span className="panel-label">Policy Engine</span>
        <h2>Check profiles against security policy</h2>
        <p>
          Write a YAML policy file and check it against a provisioning profile. Policies define
          rules like requiring SSH key-only auth, forbidding root login, and restricting allowed
          packages or distros.
        </p>
      </div>

      <div className="policy-columns">
        <div className="policy-left">
          <div className="panel policy-editor-panel">
            <span className="panel-label">Policy YAML</span>
            <textarea
              className="policy-textarea"
              value={policyYAML}
              onChange={(e) => setPolicyYAML(e.target.value)}
              spellCheck={false}
              placeholder={POLICY_PLACEHOLDER}
            />
          </div>

          <div className="panel policy-profile-panel">
            <span className="panel-label">Target Profile</span>
            <p className="policy-hint">Load a profile from disk to check it against the policy above.</p>
            <div className="policy-load-row">
              <input
                type="text"
                placeholder="Profile path, e.g. ~/.bootwrangler/profiles/edge.yaml"
                value={profilePath}
                onChange={(e) => setProfilePath(e.target.value)}
                onKeyDown={(e) => { if (e.key === "Enter") void handleLoadProfile(); }}
              />
              <button
                type="button"
                className="secondary-action"
                onClick={() => void handleLoadProfile()}
                disabled={!profilePath.trim() || profileLoading}
              >
                {profileLoading ? "Loading…" : "Load"}
              </button>
            </div>
            {loadError && <p className="render-error policy-load-error">{loadError}</p>}

            {loadedProfile && (
              <div className="policy-snapshot-preview">
                <span className="panel-label">Snapshot extracted from profile</span>
                <div className="policy-snapshot-grid">
                  <SnapRow label="OS Family" value={snap?.OSFamily ?? "—"} />
                  <SnapRow label="SSH Password Auth" value={snap?.SSHPasswordAuth ? "Yes" : "No"} warn={snap?.SSHPasswordAuth} />
                  <SnapRow label="Permit Root Login" value={snap?.SSHPermitRootLogin ? "Yes" : "No"} warn={snap?.SSHPermitRootLogin} />
                  <SnapRow label="Disk Confirm Destructive" value={snap?.DiskConfirmDestructive ? "Yes" : "No"} />
                  <SnapRow label="Users" value={snap?.Users.join(", ") || "—"} />
                  <SnapRow label="Packages" value={snap?.Packages.length ? `${snap.Packages.length} package(s)` : "—"} />
                  <SnapRow label="Package Presets" value={snap?.PackagePresets.join(", ") || "—"} />
                </div>
                <div className="policy-check-actions">
                  <button
                    type="button"
                    className="primary-action"
                    onClick={() => void handleCheck()}
                    disabled={checkState.kind === "loading"}
                  >
                    {checkState.kind === "loading" ? "Checking…" : "Check Policy"}
                  </button>
                </div>
              </div>
            )}
          </div>
        </div>

        <div className="policy-right">
          <div className="panel policy-result-panel">
            <span className="panel-label">Result</span>

            {checkState.kind === "idle" && (
              <p className="library-empty">Load a profile and click "Check Policy" to see results.</p>
            )}

            {checkState.kind === "loading" && (
              <p className="library-empty">Checking…</p>
            )}

            {checkState.kind === "error" && (
              <p className="render-error">{checkState.message}</p>
            )}

            {checkState.kind === "done" && (
              <div className={`policy-result ${checkState.result.Passed ? "policy-pass" : "policy-fail"}`}>
                <div className="policy-result-header">
                  <strong className="policy-verdict">
                    {checkState.result.Passed ? "PASSED" : "FAILED"}
                  </strong>
                  <span className="policy-verdict-sub">
                    {checkState.result.Passed
                      ? "Profile satisfies all policy rules."
                      : `${checkState.result.Violations.length} violation${checkState.result.Violations.length !== 1 ? "s" : ""} found.`}
                  </span>
                </div>
                {checkState.result.Violations.length > 0 && (
                  <ul className="policy-violations">
                    {checkState.result.Violations.map((v, i) => (
                      <li key={i} className="policy-violation-item">
                        <code className="policy-violation-rule">{v.Rule}</code>
                        <span>{v.Message}</span>
                      </li>
                    ))}
                  </ul>
                )}
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}

interface SnapRowProps {
  label: string;
  value: string;
  warn?: boolean;
}

function SnapRow({ label, value, warn }: SnapRowProps) {
  return (
    <div className="policy-snap-row">
      <span className="policy-snap-label">{label}</span>
      <span className={`policy-snap-value${warn ? " policy-snap-warn" : ""}`}>{value}</span>
    </div>
  );
}
