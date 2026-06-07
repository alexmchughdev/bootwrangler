import { useEffect, useState, type ReactNode } from "react";
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

// ---------------------------------------------------------------------------
// Information architecture: 6 task-oriented sections, each with sub-views.
// `key` values for sub-views match the page-render switch below.
// ---------------------------------------------------------------------------

interface SubView {
  key: string;
  label: string;
}
interface Section {
  key: string;
  label: string;
  tagline: string;
  icon: ReactNode;
  cta?: { label: string; target: string };
  children: SubView[];
}

const SECTIONS: Section[] = [
  {
    key: "Home",
    label: "Home",
    tagline: "Your provisioning workspace",
    icon: <IconHome />,
    children: [{ key: "Dashboard", label: "Overview" }],
  },
  {
    key: "Flash",
    label: "Flash",
    tagline: "Write an OS image to a drive",
    icon: <IconFlash />,
    children: [
      { key: "Flash Image", label: "Image → Drive" },
      { key: "Flash Partition", label: "Image → Partition" },
      { key: "Images", label: "Image Catalogue" },
      { key: "USB Devices", label: "Drives" },
    ],
  },
  {
    key: "Profiles",
    label: "Profiles",
    tagline: "Build unattended install configs",
    icon: <IconProfiles />,
    cta: { label: "New Profile", target: "Profiles" },
    children: [
      { key: "Library", label: "Library" },
      { key: "Profiles", label: "Editor" },
      { key: "Import", label: "Import" },
      { key: "Render", label: "Render" },
      { key: "Policy", label: "Policy" },
    ],
  },
  {
    key: "Media",
    label: "Media",
    tagline: "Compose multi-boot USB media",
    icon: <IconMedia />,
    children: [
      { key: "Media Builder", label: "Builder" },
      { key: "Boot Menu", label: "Boot Menu" },
    ],
  },
  {
    key: "Lab",
    label: "Lab",
    tagline: "Test installs in a VM",
    icon: <IconLab />,
    children: [
      { key: "Lab", label: "VM Lab" },
      { key: "Host Info", label: "Host Info" },
    ],
  },
  {
    key: "Settings",
    label: "Settings",
    tagline: "Preferences & local server",
    icon: <IconSettings />,
    children: [
      { key: "Settings", label: "Preferences" },
      { key: "Provisioning Server", label: "Server" },
    ],
  },
];

// child key -> primary section
const CHILD_TO_PRIMARY: Record<string, Section> = {};
for (const s of SECTIONS) {
  for (const c of s.children) {
    CHILD_TO_PRIMARY[c.key] = s;
  }
}
const primaryDefault = (s: Section) => s.children[0].key;

function App() {
  const [activeSection, setActiveSection] = useState<string>("Dashboard");
  const [health, setHealth] = useState(initialHealth);
  const [profileEditorKey, setProfileEditorKey] = useState(0);
  const [libraryKey, setLibraryKey] = useState(0);
  const [importedProfile, setImportedProfile] = useState<Profile | undefined>();

  useEffect(() => {
    void getHealth().then(setHealth);
  }, []);

  const primary = CHILD_TO_PRIMARY[activeSection] ?? SECTIONS[0];

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
          <div className="brand-row">
            <span className="brand-mark">BW</span>
            <div>
              <strong>BootWrangler</strong>
              <span>Boot Media Studio</span>
            </div>
          </div>
        </div>

        <nav aria-label="Main navigation" className="sidebar-nav">
          {SECTIONS.map((section) => (
            <button
              key={section.key}
              className={section === primary ? "nav-item active" : "nav-item"}
              onClick={() => setActiveSection(primaryDefault(section))}
              type="button"
            >
              <span className="nav-icon">{section.icon}</span>
              {section.label}
            </button>
          ))}
        </nav>

        <div className="backend-status">
          <span className={health.status === "ok" ? "status-dot ok" : "status-dot"} />
          <div>
            <strong>{health.status === "ok" ? "Connected" : health.status}</strong>
            <span>v{health.version}</span>
          </div>
        </div>
      </aside>

      <section className="workspace">
        <header className="workspace-header">
          <div>
            <span className="eyebrow">{primary.tagline}</span>
            <h1>{primary.label}</h1>
          </div>
          {primary.cta && (
            <button className="primary-action" onClick={handleNewProfile} type="button">
              {primary.cta.label}
            </button>
          )}
        </header>

        {primary.children.length > 1 && (
          <div className="subnav" role="tablist">
            {primary.children.map((child) => (
              <button
                key={child.key}
                role="tab"
                aria-selected={child.key === activeSection}
                className={child.key === activeSection ? "subnav-item active" : "subnav-item"}
                onClick={() => setActiveSection(child.key)}
                type="button"
              >
                <span className="subnav-dot" />
                {child.label}
              </button>
            ))}
          </div>
        )}

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

// ---------------------------------------------------------------------------
// Dashboard
// ---------------------------------------------------------------------------

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

  const quickStarts = [
    {
      title: "Flash an OS image",
      body: "Write Ubuntu, Debian, Alpine and more to a USB drive.",
      action: "Flash Image",
      cta: "Start flashing",
      icon: <IconFlash />,
    },
    {
      title: "Create an install profile",
      body: "Build an unattended installer config, no answer-file syntax required.",
      action: "Profiles",
      cta: "New profile",
      icon: <IconProfiles />,
    },
    {
      title: "Test in a VM",
      body: "Boot your profile in a disposable QEMU lab before touching hardware.",
      action: "Lab",
      cta: "Open lab",
      icon: <IconLab />,
    },
  ];

  return (
    <div className="dashboard">
      <div className="quickstart-grid">
        {quickStarts.map((q) => (
          <article
            key={q.action}
            className="quickstart-card"
            role="button"
            tabIndex={0}
            onClick={() => onNavigate(q.action)}
            onKeyDown={(e) => e.key === "Enter" && onNavigate(q.action)}
          >
            <span className="quickstart-icon">{q.icon}</span>
            <h3>{q.title}</h3>
            <p>{q.body}</p>
            <span className="quickstart-cta">{q.cta} →</span>
          </article>
        ))}
      </div>

      <div className="stat-row">
        <button type="button" className="stat-chip" onClick={() => onNavigate("Library")}>
          <span className="stat-value">{profileLabel}</span>
          <span className="stat-label">Saved profiles</span>
        </button>
        <button type="button" className="stat-chip" onClick={() => onNavigate("USB Devices")}>
          <span className="stat-value">{deviceLabel}</span>
          <span className="stat-label">Drives connected</span>
        </button>
        <button type="button" className="stat-chip" onClick={() => onNavigate("Provisioning Server")}>
          <span className={`stat-value ${serverRunning ? "stat-value-on" : ""}`}>{serverLabel}</span>
          <span className="stat-label">Provisioning server</span>
        </button>
      </div>
    </div>
  );
}

// ---------------------------------------------------------------------------
// Inline icons (16px, stroke = currentColor)
// ---------------------------------------------------------------------------

function IconHome() {
  return (
    <svg viewBox="0 0 24 24" width="17" height="17" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
      <path d="M3 11.5 12 4l9 7.5" />
      <path d="M5 10v9h14v-9" />
    </svg>
  );
}
function IconFlash() {
  return (
    <svg viewBox="0 0 24 24" width="17" height="17" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
      <path d="M13 2 4 14h6l-1 8 9-12h-6l1-8Z" />
    </svg>
  );
}
function IconProfiles() {
  return (
    <svg viewBox="0 0 24 24" width="17" height="17" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
      <rect x="5" y="3" width="14" height="18" rx="2" />
      <path d="M9 8h6M9 12h6M9 16h4" />
    </svg>
  );
}
function IconMedia() {
  return (
    <svg viewBox="0 0 24 24" width="17" height="17" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
      <rect x="3" y="5" width="18" height="14" rx="2" />
      <path d="M9 5v14" />
      <circle cx="6" cy="9" r="0.6" fill="currentColor" />
    </svg>
  );
}
function IconLab() {
  return (
    <svg viewBox="0 0 24 24" width="17" height="17" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
      <path d="M9 3h6M10 3v6l-5 8a2 2 0 0 0 1.7 3h10.6A2 2 0 0 0 19 17l-5-8V3" />
    </svg>
  );
}
function IconSettings() {
  return (
    <svg viewBox="0 0 24 24" width="17" height="17" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
      <circle cx="12" cy="12" r="3" />
      <path d="M12 2v3M12 19v3M2 12h3M19 12h3M5 5l2 2M17 17l2 2M19 5l-2 2M7 17l-2 2" />
    </svg>
  );
}

export default App;
