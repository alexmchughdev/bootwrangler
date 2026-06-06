import { useEffect, useState } from "react";
import { getHealth, listDevices, type HealthStatus } from "./api/backend";
import type { Profile } from "./types/profile";
import BootMenu from "./pages/BootMenu";
import FlashImage from "./pages/FlashImage";
import Policy from "./pages/Policy";
import FlashPartition from "./pages/FlashPartition";
import HostInspection from "./pages/HostInspection";
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

const NAV_GROUPS = [
  {
    label: "Profiles",
    items: ["Dashboard", "Profiles", "Library", "Import"] as const,
  },
  {
    label: "Build",
    items: ["Render", "Boot Menu", "Media Builder"] as const,
  },
  {
    label: "Deploy",
    items: ["Images", "Flash Image", "Flash Partition", "USB Devices"] as const,
  },
  {
    label: "Testing",
    items: ["Lab", "Host Info"] as const,
  },
  {
    label: "Config",
    items: ["Policy", "Provisioning Server", "Settings"] as const,
  },
] as const;

type NavItem = (typeof NAV_GROUPS)[number]["items"][number];

function App() {
  const [activeSection, setActiveSection] = useState<string>("Dashboard");
  const [health, setHealth] = useState(initialHealth);
  const [profileEditorKey, setProfileEditorKey] = useState(0);
  const [libraryKey, setLibraryKey] = useState(0);
  const [importedProfile, setImportedProfile] = useState<Profile | undefined>();

  useEffect(() => {
    void getHealth().then(setHealth);
  }, []);

  function handleNewProfile() {
    setImportedProfile(undefined);
    setActiveSection("Profiles");
    setProfileEditorKey((value) => value + 1);
  }

  function handleOpenImported(profile: Profile) {
    setImportedProfile(profile);
    setActiveSection("Profiles");
    setProfileEditorKey((k) => k + 1);
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

        <nav aria-label="Main navigation" className="sidebar-nav">
          {NAV_GROUPS.map((group) => (
            <div key={group.label} className="nav-group">
              <span className="nav-group-label">{group.label}</span>
              {group.items.map((item) => (
                <button
                  className={item === activeSection ? "nav-item active" : "nav-item"}
                  key={item}
                  onClick={() => setActiveSection(item)}
                  type="button"
                >
                  {item}
                </button>
              ))}
            </div>
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
          <ProfileEditor key={profileEditorKey} initialProfile={importedProfile} onNavigate={setActiveSection} />
        ) : activeSection === "Library" ? (
          <ProfileLibrary
            key={libraryKey}
            onOpen={(profile, _filename) => {
              handleOpenImported(profile);
            }}
          />
        ) : activeSection === "Render" ? (
          <RenderPreview onNavigate={setActiveSection} />
        ) : activeSection === "Policy" ? (
          <Policy />
        ) : activeSection === "Provisioning Server" ? (
          <ProvisioningServer />
        ) : activeSection === "Images" ? (
          <Images />
        ) : activeSection === "Flash Image" ? (
          <FlashImage />
        ) : activeSection === "Flash Partition" ? (
          <FlashPartition />
        ) : activeSection === "USB Devices" ? (
          <USBDevices onNavigate={setActiveSection} />
        ) : activeSection === "Host Info" ? (
          <HostInspection />
        ) : activeSection === "Media Builder" ? (
          <MediaBuilder />
        ) : activeSection === "Boot Menu" ? (
          <BootMenu />
        ) : activeSection === "Lab" ? (
          <Lab />
        ) : activeSection === "Import" ? (
          <Import onNavigate={setActiveSection} onOpenInEditor={handleOpenImported} />
        ) : activeSection === "Settings" ? (
          <Settings />
        ) : (
          <Dashboard onNavigate={setActiveSection} />
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

interface DashboardProps {
  onNavigate: (section: string) => void;
}

function Dashboard({ onNavigate }: DashboardProps) {
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
        <span className="panel-label">Quick Actions</span>
        <h2>Get started</h2>
        <div className="dashboard-actions">
          <button type="button" className="dashboard-action-btn" onClick={() => onNavigate("Profiles")}>
            New Profile
          </button>
          <button type="button" className="dashboard-action-btn" onClick={() => onNavigate("Library")}>
            Open Library
          </button>
          <button type="button" className="dashboard-action-btn" onClick={() => onNavigate("Images")}>
            Manage Images
          </button>
          <button type="button" className="dashboard-action-btn" onClick={() => onNavigate("Lab")}>
            Launch Lab VM
          </button>
          <button type="button" className="dashboard-action-btn" onClick={() => onNavigate("Render")}>
            Render Profile
          </button>
          <button type="button" className="dashboard-action-btn" onClick={() => onNavigate("Boot Menu")}>
            Build Boot Menu
          </button>
        </div>
      </article>

      <article
        className="panel dashboard-stat"
        onClick={() => onNavigate("Library")}
        onKeyDown={(e) => e.key === "Enter" && onNavigate("Library")}
        role="button"
        tabIndex={0}
      >
        <span className="panel-label">Profiles</span>
        <strong className="dashboard-stat-value">{profileLabel}</strong>
        <p>Saved in library</p>
        <span className="dashboard-stat-arrow" aria-hidden="true">→</span>
      </article>

      <article
        className="panel dashboard-stat"
        onClick={() => onNavigate("USB Devices")}
        onKeyDown={(e) => e.key === "Enter" && onNavigate("USB Devices")}
        role="button"
        tabIndex={0}
      >
        <span className="panel-label">USB Devices</span>
        <strong className="dashboard-stat-value">{deviceLabel}</strong>
        <p>Connected targets</p>
        <span className="dashboard-stat-arrow" aria-hidden="true">→</span>
      </article>

      <article
        className="panel dashboard-stat"
        onClick={() => onNavigate("Provisioning Server")}
        onKeyDown={(e) => e.key === "Enter" && onNavigate("Provisioning Server")}
        role="button"
        tabIndex={0}
      >
        <span className="panel-label">Provisioning Server</span>
        <strong className="dashboard-stat-value">{serverLabel}</strong>
        <p>{serverRunning ? "Serving files" : "Click to configure"}</p>
        <span className="dashboard-stat-arrow" aria-hidden="true">→</span>
      </article>
    </div>
  );
}

// Ensure all nav items are typed — compile-time check
type _NavCheck = NavItem;

export default App;
