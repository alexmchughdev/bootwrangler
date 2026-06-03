import { useEffect, useState } from "react";
import { getHealth, listDevices, type HealthStatus } from "./api/backend";
import { navigationItems } from "./navigation";
import FlashImage from "./pages/FlashImage";
import FlashPartition from "./pages/FlashPartition";
import Images from "./pages/Images";
import Import from "./pages/Import";
import Lab from "./pages/Lab";
import MediaBuilder from "./pages/MediaBuilder";
import ProfileEditor from "./pages/ProfileEditor";
import ProfileLibrary from "./pages/ProfileLibrary";
import ProvisioningServer from "./pages/ProvisioningServer";
import RenderPreview from "./pages/RenderPreview";
import Settings from "./pages/Settings";
import USBDevices from "./pages/USBDevices";

const initialHealth: HealthStatus = {
  status: "connecting",
  version: "unknown",
};

function App() {
  const [activeSection, setActiveSection] = useState("Dashboard");
  const [health, setHealth] = useState(initialHealth);
  const [profileEditorKey, setProfileEditorKey] = useState(0);
  const [libraryKey, setLibraryKey] = useState(0);

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
          <span className="brand-mark">BW</span>
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

        {activeSection === "Profiles" ? (
          <ProfileEditor key={profileEditorKey} />
        ) : activeSection === "Library" ? (
          <ProfileLibrary
            key={libraryKey}
            onOpen={(_profile, _filename) => {
              setActiveSection("Profiles");
              setProfileEditorKey((k) => k + 1);
            }}
          />
        ) : activeSection === "Render" ? (
          <RenderPreview />
        ) : activeSection === "Provisioning Server" ? (
          <ProvisioningServer />
        ) : activeSection === "Images" ? (
          <Images />
        ) : activeSection === "Flash Image" ? (
          <FlashImage />
        ) : activeSection === "Flash Partition" ? (
          <FlashPartition />
        ) : activeSection === "USB Devices" ? (
          <USBDevices />
        ) : activeSection === "Media Builder" ? (
          <MediaBuilder />
        ) : activeSection === "Lab" ? (
          <Lab />
        ) : activeSection === "Import" ? (
          <Import onNavigate={setActiveSection} />
        ) : activeSection === "Settings" ? (
          <Settings />
        ) : (
          <Dashboard />
        )}
      </section>
    </main>
  );
}

interface LibraryService {
  LibraryList(): Promise<unknown[]>;
}

interface ServerAddrService {
  ServerAddr(): Promise<string>;
}

function Dashboard() {
  const [profileCount, setProfileCount] = useState<number | null>(null);
  const [deviceCount, setDeviceCount] = useState<number | null>(null);
  const [serverRunning, setServerRunning] = useState<boolean | null>(null);

  useEffect(() => {
    const libSvc = window.go?.app?.Service as unknown as LibraryService | undefined;
    if (libSvc?.LibraryList) {
      void libSvc.LibraryList().then((list) => setProfileCount(list.length)).catch(() => {});
    }

    void listDevices().then((devices) => setDeviceCount(devices.length)).catch(() => {});

    const serverSvc = window.go?.app?.Service as unknown as ServerAddrService | undefined;
    if (serverSvc?.ServerAddr) {
      void serverSvc.ServerAddr().then((addr) => setServerRunning(!!addr)).catch(() => {});
    }
  }, []);

  const profileLabel = profileCount !== null ? String(profileCount) : "—";
  const deviceLabel = deviceCount !== null ? String(deviceCount) : "—";
  const serverLabel = serverRunning === null ? "—" : serverRunning ? "Running" : "Stopped";

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
            <strong>{profileLabel}</strong>
            <p>Local profiles</p>
          </article>

          <article className="panel">
            <span className="panel-label">USB Devices</span>
            <strong>{deviceLabel}</strong>
            <p>Connected targets</p>
          </article>

          <article className="panel">
            <span className="panel-label">Provisioning Server</span>
            <strong>{serverLabel}</strong>
            <p>No build root selected</p>
          </article>
        </div>
  );
}

export default App;
