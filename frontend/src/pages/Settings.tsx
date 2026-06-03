import { useEffect, useRef, useState } from "react";

const LS_KEY = "bw_server_base_url";

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
  const savedTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    const stored = localStorage.getItem(LS_KEY);
    if (stored) setServerBaseURL(stored);

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
