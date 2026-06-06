import { useEffect, useState } from "react";
import type { Profile } from "../types/profile";
import {
  libraryCommit,
  libraryDiff,
  libraryExportBundle,
  libraryHistory,
  libraryImportBundle,
  libraryRestoreVersion,
  loadProfile,
  renderProfile,
  saveProfile,
  type LibraryEntry,
  type VersionEntry,
} from "../api/backend";

interface BatchRenderResult {
  name: string;
  status: "success" | "error";
  message: string;
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
  const [search, setSearch] = useState("");
  const [exportName, setExportName] = useState<string | null>(null);
  const [exportPath, setExportPath] = useState("");
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const [batchOutDir, setBatchOutDir] = useState("");
  const [batchResults, setBatchResults] = useState<BatchRenderResult[]>([]);
  const [batchRunning, setBatchRunning] = useState(false);
  const [historyName, setHistoryName] = useState<string | null>(null);
  const [historyEntries, setHistoryEntries] = useState<VersionEntry[]>([]);
  const [historyLoading, setHistoryLoading] = useState(false);
  const [commitMsg, setCommitMsg] = useState<Record<string, string>>({});
  const [diffText, setDiffText] = useState<string | null>(null);
  const [renameName, setRenameName] = useState<string | null>(null);
  const [renameValue, setRenameValue] = useState("");
  const [sortKey, setSortKey] = useState<"name" | "os" | "updated">("name");
  const [sortAsc, setSortAsc] = useState(true);

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

  async function handleDuplicate(name: string) {
    const svc = getLibraryService();
    if (!svc) return;
    try {
      const p = await svc.LibraryGet(name);
      const cloned = { ...p, name: `${p.name}-copy` };
      await svc.LibraryInit();
      await svc.LibraryAdd(cloned);
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

  async function handleImportBundle() {
    if (!importPath.trim()) return;
    try {
      await libraryImportBundle(importPath.trim());
      setImportPath("");
      await refresh();
    } catch (err) {
      setError(String(err));
    }
  }

  async function handleToggleHistory(name: string) {
    if (historyName === name) {
      setHistoryName(null);
      setHistoryEntries([]);
      setDiffText(null);
      return;
    }
    setHistoryName(name);
    setHistoryLoading(true);
    setDiffText(null);
    try {
      const entries = await libraryHistory(name);
      setHistoryEntries(entries ?? []);
    } catch {
      setHistoryEntries([]);
    } finally {
      setHistoryLoading(false);
    }
  }

  async function handleCommit(name: string) {
    const msg = (commitMsg[name] ?? "").trim();
    if (!msg) return;
    try {
      await libraryCommit(name, msg);
      setCommitMsg((prev) => ({ ...prev, [name]: "" }));
      const entries = await libraryHistory(name);
      setHistoryEntries(entries ?? []);
    } catch (err) {
      setError(String(err));
    }
  }

  async function handleRestore(name: string, hash: string) {
    if (!confirm(`Restore "${name}" to version ${hash.slice(0, 8)}? This overwrites the current file.`)) return;
    try {
      await libraryRestoreVersion(name, hash);
      await refresh();
    } catch (err) {
      setError(String(err));
    }
  }

  async function handleDiff(name: string, fromHash: string, toHash: string) {
    try {
      const diff = await libraryDiff(name, fromHash, toHash);
      setDiffText(diff);
    } catch (err) {
      setError(String(err));
    }
  }

  function handleRenameToggle(name: string) {
    if (renameName === name) {
      setRenameName(null);
      setRenameValue("");
    } else {
      setRenameName(name);
      setRenameValue(name);
    }
  }

  async function handleRenameSubmit(oldName: string) {
    const newName = renameValue.trim();
    if (!newName || newName === oldName) {
      setRenameName(null);
      return;
    }
    const svc = getLibraryService();
    if (!svc) return;
    try {
      const p = await svc.LibraryGet(oldName);
      const renamed = { ...p, name: newName };
      await svc.LibraryAdd(renamed);
      await svc.LibraryRemove(oldName);
      setRenameName(null);
      setRenameValue("");
      await refresh();
    } catch (err) {
      setError(String(err));
    }
  }

  function toggleSort(key: "name" | "os" | "updated") {
    if (sortKey === key) {
      setSortAsc((a) => !a);
    } else {
      setSortKey(key);
      setSortAsc(true);
    }
  }

  function handleExportToggle(name: string) {
    if (exportName === name) {
      setExportName(null);
      setExportPath("");
    } else {
      setExportName(name);
      setExportPath("");
    }
  }

  async function handleExportSave(asBundle = false) {
    if (!exportName || !exportPath.trim()) return;
    const svc = getLibraryService();
    if (!svc) return;
    try {
      if (asBundle) {
        await libraryExportBundle(exportName, exportPath.trim());
      } else {
        const p = await svc.LibraryGet(exportName);
        await saveProfile(exportPath.trim(), p);
      }
      setExportName(null);
      setExportPath("");
    } catch (err) {
      setError(String(err));
    }
  }

  function toggleSelect(name: string) {
    setSelected((prev) => {
      const next = new Set(prev);
      if (next.has(name)) next.delete(name);
      else next.add(name);
      return next;
    });
  }

  function toggleSelectAll() {
    if (selected.size === filteredEntries.length) {
      setSelected(new Set());
    } else {
      setSelected(new Set(filteredEntries.map((e) => e.name)));
    }
  }

  async function handleBatchRender() {
    if (selected.size === 0 || !batchOutDir.trim()) return;
    const svc = getLibraryService();
    if (!svc) return;
    setBatchRunning(true);
    setBatchResults([]);
    const serverBaseURL = localStorage.getItem("bw_server_base_url") ?? "";
    const results: BatchRenderResult[] = [];
    for (const name of selected) {
      try {
        const p = await svc.LibraryGet(name);
        const outDir = `${batchOutDir.trim()}/${name}`;
        await renderProfile(p, outDir, serverBaseURL);
        results.push({ name, status: "success", message: outDir });
      } catch (err) {
        results.push({ name, status: "error", message: String(err) });
      }
      setBatchResults([...results]);
    }
    setBatchRunning(false);
  }

  const q = search.trim().toLowerCase();
  const filteredEntries = [...(q
    ? entries.filter(
        (e) =>
          e.name.toLowerCase().includes(q) ||
          e.os_family.toLowerCase().includes(q) ||
          e.os_version.toLowerCase().includes(q),
      )
    : entries)].sort((a, b) => {
    let cmp = 0;
    if (sortKey === "name") cmp = a.name.localeCompare(b.name);
    else if (sortKey === "os") cmp = (a.os_family + a.os_version).localeCompare(b.os_family + b.os_version);
    else cmp = (a.updated_at ?? "").localeCompare(b.updated_at ?? "");
    return sortAsc ? cmp : -cmp;
  });

  return (
    <div className="profile-library">
      <div className="library-header">
        <h2>Profile Library</h2>
        <div className="library-header-right">
          <input
            className="library-search"
            placeholder="Filter profiles…"
            type="search"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
          <button type="button" className="link-btn" onClick={() => void refresh()}>
            Refresh
          </button>
        </div>
      </div>

      {error && <p className="render-error">{error}</p>}

      {loading ? (
        <p className="library-empty">Loading…</p>
      ) : filteredEntries.length === 0 ? (
        <p className="library-empty">
          {entries.length === 0
            ? "No profiles in library. Save a profile to add it."
            : "No profiles match the filter."}
        </p>
      ) : (
        <table className="file-table library-table">
          <thead>
            <tr>
              <th>
                <input
                  type="checkbox"
                  checked={filteredEntries.length > 0 && selected.size === filteredEntries.length}
                  onChange={toggleSelectAll}
                  title="Select all"
                />
              </th>
              <th className="sortable-th" onClick={() => toggleSort("name")}>
                Name {sortKey === "name" ? (sortAsc ? "↑" : "↓") : ""}
              </th>
              <th className="sortable-th" onClick={() => toggleSort("os")}>
                OS {sortKey === "os" ? (sortAsc ? "↑" : "↓") : ""}
              </th>
              <th className="sortable-th" onClick={() => toggleSort("updated")}>
                Updated {sortKey === "updated" ? (sortAsc ? "↑" : "↓") : ""}
              </th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            {filteredEntries.map((e) => (
              <>
                <tr key={e.name}>
                  <td>
                    <input
                      type="checkbox"
                      checked={selected.has(e.name)}
                      onChange={() => toggleSelect(e.name)}
                    />
                  </td>
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
                      className="link-btn"
                      onClick={() => void handleDuplicate(e.name)}
                    >
                      Duplicate
                    </button>
                    <button
                      type="button"
                      className={exportName === e.name ? "link-btn active" : "link-btn"}
                      onClick={() => handleExportToggle(e.name)}
                    >
                      Export
                    </button>
                    <button
                      type="button"
                      className={renameName === e.name ? "link-btn active" : "link-btn"}
                      onClick={() => handleRenameToggle(e.name)}
                    >
                      Rename
                    </button>
                    <button
                      type="button"
                      className={historyName === e.name ? "link-btn active" : "link-btn"}
                      onClick={() => void handleToggleHistory(e.name)}
                    >
                      History
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
                {exportName === e.name && (
                  <tr key={`${e.name}-export`} className="library-export-row">
                    <td colSpan={4}>
                      <div className="library-export-form">
                        <span className="field-label">Export to path</span>
                        <div className="import-row">
                          <input
                            type="text"
                            placeholder="e.g. /home/user/profiles/my-profile.yaml"
                            value={exportPath}
                            onChange={(ev) => setExportPath(ev.target.value)}
                            autoFocus
                          />
                          <button
                            type="button"
                            className="secondary-action"
                            onClick={() => void handleExportSave(false)}
                            disabled={!exportPath.trim()}
                          >
                            Save YAML
                          </button>
                          <button
                            type="button"
                            className="secondary-action"
                            onClick={() => void handleExportSave(true)}
                            disabled={!exportPath.trim()}
                          >
                            Save Bundle (.zip)
                          </button>
                          <button
                            type="button"
                            className="link-btn"
                            onClick={() => handleExportToggle(e.name)}
                          >
                            Cancel
                          </button>
                        </div>
                      </div>
                    </td>
                  </tr>
                )}
                {renameName === e.name && (
                  <tr key={`${e.name}-rename`} className="library-export-row">
                    <td colSpan={5}>
                      <div className="library-export-form">
                        <span className="field-label">Rename profile</span>
                        <div className="import-row">
                          <input
                            type="text"
                            value={renameValue}
                            onChange={(ev) => setRenameValue(ev.target.value)}
                            onKeyDown={(ev) => { if (ev.key === "Enter") void handleRenameSubmit(e.name); }}
                            autoFocus
                          />
                          <button
                            type="button"
                            className="secondary-action"
                            onClick={() => void handleRenameSubmit(e.name)}
                            disabled={!renameValue.trim() || renameValue.trim() === e.name}
                          >
                            Rename
                          </button>
                          <button
                            type="button"
                            className="link-btn"
                            onClick={() => handleRenameToggle(e.name)}
                          >
                            Cancel
                          </button>
                        </div>
                      </div>
                    </td>
                  </tr>
                )}
                {historyName === e.name && (
                  <tr key={`${e.name}-history`} className="library-export-row">
                    <td colSpan={5}>
                      <div className="library-history-panel">
                        <div className="library-history-header">
                          <span className="panel-label">Version History — {e.name}</span>
                          <div className="import-row">
                            <input
                              type="text"
                              placeholder="Commit message"
                              value={commitMsg[e.name] ?? ""}
                              onChange={(ev) => setCommitMsg((prev) => ({ ...prev, [e.name]: ev.target.value }))}
                            />
                            <button
                              type="button"
                              className="secondary-action"
                              onClick={() => void handleCommit(e.name)}
                              disabled={!(commitMsg[e.name] ?? "").trim()}
                            >
                              Commit
                            </button>
                          </div>
                        </div>
                        {historyLoading ? (
                          <p className="library-empty">Loading history…</p>
                        ) : historyEntries.length === 0 ? (
                          <p className="library-empty">No versions yet. Click Commit to create the first version.</p>
                        ) : (
                          <div className="library-history-list">
                            {historyEntries.map((v, i) => (
                              <div key={v.hash} className="library-history-entry">
                                <code className="history-hash">{v.hash.slice(0, 8)}</code>
                                <span className="history-msg">{v.message}</span>
                                <span className="history-date">{new Date(v.timestamp).toLocaleString()}</span>
                                <div className="history-actions">
                                  {i < historyEntries.length - 1 && (
                                    <button
                                      type="button"
                                      className="link-btn"
                                      onClick={() => void handleDiff(e.name, historyEntries[i + 1].hash, v.hash)}
                                    >
                                      Diff
                                    </button>
                                  )}
                                  <button
                                    type="button"
                                    className="link-btn"
                                    onClick={() => void handleRestore(e.name, v.hash)}
                                  >
                                    Restore
                                  </button>
                                </div>
                              </div>
                            ))}
                          </div>
                        )}
                        {diffText && (
                          <pre className="library-diff-output">{diffText}</pre>
                        )}
                      </div>
                    </td>
                  </tr>
                )}
              </>
            ))}
          </tbody>
        </table>
      )}

      {selected.size > 0 && (
        <div className="library-batch-render">
          <span className="panel-label">Batch Render — {selected.size} profile{selected.size !== 1 ? "s" : ""} selected</span>
          <div className="import-row">
            <input
              type="text"
              placeholder="Output base directory (e.g. /tmp/batch-render)"
              value={batchOutDir}
              onChange={(e) => setBatchOutDir(e.target.value)}
            />
            <button
              type="button"
              className="primary-action"
              onClick={() => void handleBatchRender()}
              disabled={batchRunning || !batchOutDir.trim()}
            >
              {batchRunning ? "Rendering…" : "Render All"}
            </button>
          </div>
          {batchResults.length > 0 && (
            <div className="batch-render-results">
              {batchResults.map((r) => (
                <div key={r.name} className={`batch-result-row batch-result--${r.status}`}>
                  <span className="batch-result-name">{r.name}</span>
                  <span className="batch-result-msg">{r.message}</span>
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      <div className="library-import">
        <label>
          Import profile from path (YAML or .zip bundle)
          <div className="import-row">
            <input
              type="text"
              placeholder="e.g. /path/to/profile.yaml or profile.zip"
              value={importPath}
              onChange={(e) => setImportPath(e.target.value)}
            />
            <button
              type="button"
              className="secondary-action"
              onClick={() => void handleImport()}
            >
              Import YAML
            </button>
            <button
              type="button"
              className="secondary-action"
              onClick={() => void handleImportBundle()}
            >
              Import Bundle
            </button>
          </div>
        </label>
      </div>
    </div>
  );
}
