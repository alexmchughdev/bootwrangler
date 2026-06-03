import { useEffect, useRef, useState } from "react";
import { validateSSHPublicKey } from "../api/backend";

const LS_KEY = "bw_server_base_url";
const LS_LAST_RENDER_DIR = "bw_last_render_dir";

interface VersionedService {
  Version(): Promise<string>;
}

function getVersionedService(): VersionedService | undefined {
  return window.go?.app?.Service as unknown as VersionedService | undefined;
}

export default function Settings() {
  const [serverBaseURL, setServerBaseURL] = useState("");
  const [saved, setSaved] = useState(false);
  const [version, setVersion] = useState("dev");
  const [lastRenderDir, setLastRenderDir] = useState("");
  const savedTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const [sshKey, setSSHKey] = useState("");
  const [sshResult, setSSHResult] = useState<{ valid: boolean; problems: string[] } | null>(null);
  const [sshChecking, setSSHChecking] = useState(false);

  useEffect(() => {
    const stored = localStorage.getItem(LS_KEY);
    if (stored) setServerBaseURL(stored);
    const dir = localStorage.getItem(LS_LAST_RENDER_DIR);
    if (dir) setLastRenderDir(dir);

    const svc = getVersionedService();
    if (svc) {
      void svc.Version().then((v) => setVersion(v)).catch(() => {});
    }

    return () => {
      if (savedTimerRef.current) clearTimeout(savedTimerRef.current);
    };
  }, []);

  function handleSave() {
    localStorage.setItem(LS_KEY, serverBaseURL);
    setSaved(true);
    if (savedTimerRef.current) clearTimeout(savedTimerRef.current);
    savedTimerRef.current = setTimeout(() => setSaved(false), 2000);
  }

  async function handleValidateSSH() {
    setSSHChecking(true);
    setSSHResult(null);
    try {
      const result = await validateSSHPublicKey(sshKey.trim());
      setSSHResult(result);
    } catch (err) {
      setSSHResult({ valid: false, problems: [String(err)] });
    } finally {
      setSSHChecking(false);
    }
  }

  return (
    <div className="settings-page">
      <article className="panel settings-section">
        <span className="panel-label">Server Configuration</span>
        <h2>Provisioning Server</h2>
        <p className="settings-description">
          Used for self-hosted netboot mode. The provisioning server address
          clients will fetch installer assets from.
        </p>
        <div className="settings-field">
          <span className="field-label">Server Base URL</span>
          <div className="settings-input-row">
            <input
              type="text"
              placeholder="http://192.168.1.1:8088"
              value={serverBaseURL}
              onChange={(e) => setServerBaseURL(e.target.value)}
            />
            <button
              type="button"
              className="primary-action"
              onClick={handleSave}
            >
              Save
            </button>
          </div>
          {saved && <span className="settings-saved">Saved</span>}
        </div>
      </article>

      {lastRenderDir && (
        <article className="panel settings-section">
          <span className="panel-label">Quick Access</span>
          <h2>Last Render Output</h2>
          <p className="settings-description">
            The provisioning server uses this directory as its default root.
          </p>
          <div className="settings-field">
            <span className="field-label">Directory</span>
            <div className="settings-input-row">
              <code className="settings-code">{lastRenderDir}</code>
              <button
                type="button"
                className="secondary-action"
                onClick={() => {
                  localStorage.removeItem(LS_LAST_RENDER_DIR);
                  setLastRenderDir("");
                }}
              >
                Clear
              </button>
            </div>
          </div>
        </article>
      )}

      <article className="panel settings-section">
        <span className="panel-label">Security Tools</span>
        <h2>SSH Key Validator</h2>
        <p className="settings-description">
          Paste an SSH public key to verify it is correctly formatted before
          adding it to a profile's authorized keys.
        </p>
        <div className="settings-field">
          <span className="field-label">SSH Public Key</span>
          <textarea
            className="settings-ssh-input"
            placeholder="ssh-ed25519 AAAA... user@host"
            value={sshKey}
            rows={3}
            onChange={(e) => { setSSHKey(e.target.value); setSSHResult(null); }}
          />
          <div className="settings-input-row settings-ssh-actions">
            <button
              type="button"
              className="primary-action"
              onClick={() => void handleValidateSSH()}
              disabled={sshChecking || !sshKey.trim()}
            >
              {sshChecking ? "Checking…" : "Validate Key"}
            </button>
            {sshKey && (
              <button
                type="button"
                className="link-btn"
                onClick={() => { setSSHKey(""); setSSHResult(null); }}
              >
                Clear
              </button>
            )}
          </div>
          {sshResult && (
            <div className={`policy-result ${sshResult.valid ? "policy-pass" : "policy-fail"}`}>
              <strong>{sshResult.valid ? "Valid key" : "Invalid key"}</strong>
              {sshResult.problems.length > 0 && (
                <ul className="policy-violations">
                  {sshResult.problems.map((p, i) => (
                    <li key={i}>{p}</li>
                  ))}
                </ul>
              )}
            </div>
          )}
        </div>
      </article>

      <article className="panel settings-section">
        <span className="panel-label">About</span>
        <h2>BootWrangler</h2>
        <div className="settings-about">
          <div className="settings-about-row">
            <span className="settings-about-label">Version</span>
            <span className="settings-about-value">{version}</span>
          </div>
          <div className="settings-about-row">
            <span className="settings-about-label">Docs</span>
            <span className="settings-about-value settings-link">
              docs/quickstart.md
            </span>
          </div>
          <div className="settings-about-row">
            <span className="settings-about-label">GitHub</span>
            <span className="settings-about-value settings-link">
              github.com/alexmchughdev/bootwrangler
            </span>
          </div>
        </div>
      </article>
    </div>
  );
}
