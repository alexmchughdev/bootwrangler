import { useEffect, useState } from "react";
import { getHealth, type HealthStatus } from "./api/backend";
import { navigationItems } from "./navigation";
import ProfileEditor from "./pages/ProfileEditor";

const initialHealth: HealthStatus = {
  status: "connecting",
  version: "unknown",
};

function App() {
  const [activeSection, setActiveSection] = useState("Dashboard");
  const [health, setHealth] = useState(initialHealth);
  const [profileEditorKey, setProfileEditorKey] = useState(0);

  useEffect(() => {
    void getHealth().then(setHealth);
  }, []);

  function handleNewProfile() {
    setActiveSection("Profiles");
    setProfileEditorKey((value) => value + 1);
  }

  return (
    <main className="app-shell">
      <aside className="sidebar">
        <div className="brand">
          <span className="brand-mark">BF</span>
          <div>
            <strong>BootWrangler</strong>
            <span>Provisioning Studio</span>
          </div>
        </div>

        <nav aria-label="Main navigation">
          {navigationItems.map((item) => (
            <button
              className={item === activeSection ? "nav-item active" : "nav-item"}
              key={item}
              onClick={() => setActiveSection(item)}
              type="button"
            >
              {item}
            </button>
          ))}
        </nav>

        <div className="backend-status">
          <span className={health.status === "ok" ? "status-dot ok" : "status-dot"} />
          <div>
            <strong>Backend {health.status}</strong>
            <span>Version {health.version}</span>
          </div>
        </div>
      </aside>

      <section className="workspace">
        <header className="workspace-header">
          <div>
            <span className="eyebrow">BootWrangler Workspace</span>
            <h1>{activeSection}</h1>
          </div>
          <button className="primary-action" onClick={handleNewProfile} type="button">
            New Profile
          </button>
        </header>

        {activeSection === "Profiles" ? <ProfileEditor key={profileEditorKey} /> : <Dashboard />}
      </section>
    </main>
  );
}

function Dashboard() {
  return (
        <div className="dashboard-grid">
          <article className="panel panel-wide">
            <span className="panel-label">Workspace Status</span>
            <h2>Desktop foundation ready</h2>
            <p>
              Profile workflows, image management, media composition, and lab
              controls will appear here as their backend services are added.
            </p>
          </article>

          <article className="panel">
            <span className="panel-label">Profiles</span>
            <strong>0</strong>
            <p>Local profiles</p>
          </article>

          <article className="panel">
            <span className="panel-label">USB Devices</span>
            <strong>0</strong>
            <p>Connected targets</p>
          </article>

          <article className="panel">
            <span className="panel-label">Provisioning Server</span>
            <strong>Stopped</strong>
            <p>No build root selected</p>
          </article>
        </div>
  );
}

export default App;
