import { useEffect, useState } from "react";
import type { Profile } from "../types/profile";
import { loadProfile } from "../api/backend";

interface LibraryEntry {
  name: string;
  filename: string;
  os_family: string;
  os_version: string;
  updated_at: string;
}

interface LibraryService {
  LibraryInit(): Promise<void>;
  LibraryList(): Promise<LibraryEntry[]>;
  LibraryGet(name: string): Promise<Profile>;
  LibraryRemove(name: string): Promise<void>;
  LibraryAdd(value: Profile): Promise<string>;
}

function getLibraryService(): LibraryService | undefined {
  return (
    window.go?.app?.Service as unknown as LibraryService | undefined
  );
}

interface Props {
  onOpen: (profile: Profile, filename: string) => void;
}

export default function ProfileLibrary({ onOpen }: Props) {
  const [entries, setEntries] = useState<LibraryEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [importPath, setImportPath] = useState("");

  useEffect(() => {
    void refresh();
  }, []);

  async function refresh() {
    setLoading(true);
    setError("");
    const svc = getLibraryService();
    if (!svc) {
      setLoading(false);
      return;
    }
    try {
      await svc.LibraryInit();
      const list = await svc.LibraryList();
      setEntries(list ?? []);
    } catch (err) {
      setError(String(err));
    } finally {
      setLoading(false);
    }
  }

  async function handleOpen(name: string) {
    const svc = getLibraryService();
    if (!svc) return;
    try {
      const p = await svc.LibraryGet(name);
      onOpen(p, name + ".yaml");
    } catch (err) {
      setError(String(err));
    }
  }

  async function handleRemove(name: string) {
    const svc = getLibraryService();
    if (!svc) return;
    if (!confirm(`Remove profile "${name}" from the library?`)) return;
    try {
      await svc.LibraryRemove(name);
      await refresh();
    } catch (err) {
      setError(String(err));
    }
  }

  async function handleImport() {
    if (!importPath.trim()) return;
    const svc = getLibraryService();
    if (!svc) return;
    try {
      const p = await loadProfile(importPath.trim());
      await svc.LibraryAdd(p);
      setImportPath("");
      await refresh();
    } catch (err) {
      setError(String(err));
    }
  }

  return (
    <div className="profile-library">
      <div className="library-header">
        <h2>Profile Library</h2>
        <button type="button" className="link-btn" onClick={() => void refresh()}>
          Refresh
        </button>
      </div>

      {error && <p className="render-error">{error}</p>}

      {loading ? (
        <p className="library-empty">Loading…</p>
      ) : entries.length === 0 ? (
        <p className="library-empty">No profiles in library. Save a profile to add it.</p>
      ) : (
        <table className="file-table library-table">
          <thead>
            <tr>
              <th>Name</th>
              <th>OS</th>
              <th>Updated</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            {entries.map((e) => (
              <tr key={e.name}>
                <td className="file-path">{e.name}</td>
                <td>
                  {e.os_family} {e.os_version}
                </td>
                <td className="checksum">
                  {e.updated_at ? new Date(e.updated_at).toLocaleDateString() : "—"}
                </td>
                <td className="library-actions">
                  <button
                    type="button"
                    className="link-btn"
                    onClick={() => void handleOpen(e.name)}
                  >
                    Open
                  </button>
                  <button
                    type="button"
                    className="link-btn danger"
                    onClick={() => void handleRemove(e.name)}
                  >
                    Remove
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}

      <div className="library-import">
        <label>
          Import profile from path
          <div className="import-row">
            <input
              type="text"
              placeholder="e.g. /path/to/profile.yaml"
              value={importPath}
              onChange={(e) => setImportPath(e.target.value)}
            />
            <button
              type="button"
              className="secondary-action"
              onClick={() => void handleImport()}
            >
              Import
            </button>
          </div>
        </label>
      </div>
    </div>
  );
}
