import { useState, useEffect } from "react";
import { renderIPXEMenu, renderGRUBMenu, writeTextFile, listNetbootImages, formatNetbootCmdline, type EntryKind, type MenuEntry, type NetbootImage } from "../api/backend";

type Format = "iPXE" | "GRUB";

const KIND_LABELS: Record<EntryKind, string> = {
  netboot: "netboot",
  "local-profile": "local-profile",
  "local-image": "local-image",
  shell: "shell",
  reboot: "reboot",
};

const ALL_KINDS: EntryKind[] = ["netboot", "local-profile", "local-image", "shell", "reboot"];

function showURL(kind: EntryKind): boolean {
  return kind === "netboot" || kind === "local-profile" || kind === "local-image";
}

function showKernelFields(kind: EntryKind): boolean {
  return kind === "local-profile" || kind === "local-image";
}

const DEFAULT_ENTRY: MenuEntry = {
  Kind: "netboot",
  Label: "Boot from network (netboot.xyz)",
  URL: "https://boot.netboot.xyz",
  Kernel: "",
  Initrd: "",
  Cmdline: "",
};

function emptyEntry(): MenuEntry {
  return { Kind: "netboot", Label: "", URL: "", Kernel: "", Initrd: "", Cmdline: "" };
}

export default function BootMenu() {
  const [title, setTitle] = useState("BootWrangler Boot Menu");
  const [format, setFormat] = useState<Format>("iPXE");
  const [entries, setEntries] = useState<MenuEntry[]>([{ ...DEFAULT_ENTRY }]);
  const [preview, setPreview] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [copied, setCopied] = useState(false);
  const [savePath, setSavePath] = useState("");
  const [saved, setSaved] = useState(false);
  const [netbootImages, setNetbootImages] = useState<NetbootImage[]>([]);
  const [autofillIndex, setAutofillIndex] = useState<number | null>(null);
  const [autofillConfigURL, setAutofillConfigURL] = useState(
    () => localStorage.getItem("bw_server_base_url") ?? "",
  );

  useEffect(() => {
    void listNetbootImages().then(setNetbootImages).catch(() => {});
  }, []);

  function updateEntry(index: number, patch: Partial<MenuEntry>) {
    setEntries((prev) =>
      prev.map((e, i) => {
        if (i !== index) return e;
        const updated = { ...e, ...patch };
        // Clear irrelevant fields when kind changes
        if ("Kind" in patch) {
          if (!showURL(patch.Kind!)) updated.URL = "";
          if (!showKernelFields(patch.Kind!)) {
            updated.Kernel = "";
            updated.Initrd = "";
            updated.Cmdline = "";
          }
        }
        return updated;
      }),
    );
  }

  function removeEntry(index: number) {
    setEntries((prev) => prev.filter((_, i) => i !== index));
  }

  function addEntry() {
    setEntries((prev) => [...prev, emptyEntry()]);
  }

  async function handleAutofill(entryIndex: number, image: NetbootImage) {
    const cmdline = await formatNetbootCmdline(image.Family, autofillConfigURL).catch(() => "");
    updateEntry(entryIndex, {
      Kernel: image.KernelURL,
      Initrd: image.InitrdURL,
      Cmdline: cmdline,
    });
    setAutofillIndex(null);
  }

  async function handlePreview() {
    setLoading(true);
    setError(null);
    try {
      const result =
        format === "iPXE"
          ? await renderIPXEMenu(title, entries)
          : await renderGRUBMenu(title, entries);
      setPreview(result);
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err));
      setPreview(null);
    } finally {
      setLoading(false);
    }
  }

  async function handleCopy() {
    if (!preview) return;
    await navigator.clipboard.writeText(preview);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  }

  async function handleSaveToFile() {
    if (!preview || !savePath.trim()) return;
    try {
      await writeTextFile(savePath.trim(), preview);
      setSaved(true);
      setTimeout(() => setSaved(false), 2000);
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err));
    }
  }

  return (
    <div className="boot-menu-page">
      <div className="boot-menu-columns">
        {/* Left panel — Menu Builder */}
        <div className="panel boot-menu-builder">
          <span className="panel-label">Menu Builder</span>

          <div className="boot-menu-field">
            <span className="field-label">Menu Title</span>
            <input
              type="text"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder="BootWrangler Boot Menu"
            />
          </div>

          <div className="boot-menu-field">
            <span className="field-label">Format</span>
            <div className="boot-menu-toggle">
              {(["iPXE", "GRUB"] as Format[]).map((f) => (
                <button
                  key={f}
                  type="button"
                  className={
                    format === f
                      ? "boot-menu-toggle-btn boot-menu-toggle-btn--active"
                      : "boot-menu-toggle-btn"
                  }
                  onClick={() => setFormat(f)}
                >
                  {f}
                </button>
              ))}
            </div>
          </div>

          <div className="boot-menu-entries-header section-header">
            <span className="field-label">Entries</span>
          </div>

          <div className="boot-menu-entry-list">
            {entries.map((entry, index) => (
              <div key={index} className="boot-menu-entry-card">
                <div className="boot-menu-entry-row">
                  <div className="boot-menu-field">
                    <span className="field-label">Kind</span>
                    <select
                      value={entry.Kind}
                      onChange={(e) =>
                        updateEntry(index, { Kind: e.target.value as EntryKind })
                      }
                    >
                      {ALL_KINDS.map((k) => (
                        <option key={k} value={k}>
                          {KIND_LABELS[k]}
                        </option>
                      ))}
                    </select>
                  </div>

                  <div className="boot-menu-field boot-menu-field--grow">
                    <span className="field-label">Label</span>
                    <input
                      type="text"
                      value={entry.Label}
                      onChange={(e) => updateEntry(index, { Label: e.target.value })}
                      placeholder="Entry label"
                    />
                  </div>

                  <button
                    type="button"
                    className="text-action danger boot-menu-remove"
                    onClick={() => removeEntry(index)}
                    aria-label="Remove entry"
                  >
                    Remove
                  </button>
                </div>

                {showURL(entry.Kind) && (
                  <div className="boot-menu-field">
                    <span className="field-label">URL</span>
                    <input
                      type="text"
                      value={entry.URL}
                      onChange={(e) => updateEntry(index, { URL: e.target.value })}
                      placeholder="https://..."
                    />
                  </div>
                )}

                {showKernelFields(entry.Kind) && (
                  <div className="boot-menu-kernel-fields">
                    <div className="boot-menu-autofill-row">
                      <button
                        type="button"
                        className="link-btn"
                        onClick={() => setAutofillIndex(autofillIndex === index ? null : index)}
                      >
                        {autofillIndex === index ? "Cancel auto-fill" : "Auto-fill from OS"}
                      </button>
                    </div>
                    {autofillIndex === index && (
                      <div className="boot-menu-autofill-panel">
                        <div className="boot-menu-field">
                          <span className="field-label">Config server URL</span>
                          <input
                            type="text"
                            value={autofillConfigURL}
                            onChange={(e) => setAutofillConfigURL(e.target.value)}
                            placeholder="http://192.168.1.1:8080"
                          />
                        </div>
                        <div className="boot-menu-autofill-options">
                          {netbootImages.map((img) => (
                            <button
                              key={`${img.ID}-${img.Version}-${img.Arch}`}
                              type="button"
                              className="boot-menu-autofill-btn"
                              onClick={() => void handleAutofill(index, img)}
                            >
                              {img.Name} {img.Version} ({img.Arch})
                            </button>
                          ))}
                          {netbootImages.length === 0 && (
                            <p className="boot-menu-autofill-empty">No direct netboot images available — use netboot.xyz.</p>
                          )}
                        </div>
                      </div>
                    )}
                    <div className="boot-menu-field">
                      <span className="field-label">Kernel</span>
                      <input
                        type="text"
                        value={entry.Kernel}
                        onChange={(e) => updateEntry(index, { Kernel: e.target.value })}
                        placeholder="/vmlinuz"
                      />
                    </div>
                    <div className="boot-menu-field">
                      <span className="field-label">Initrd</span>
                      <input
                        type="text"
                        value={entry.Initrd}
                        onChange={(e) => updateEntry(index, { Initrd: e.target.value })}
                        placeholder="/initrd.img"
                      />
                    </div>
                    <div className="boot-menu-field">
                      <span className="field-label">Cmdline</span>
                      <input
                        type="text"
                        value={entry.Cmdline}
                        onChange={(e) => updateEntry(index, { Cmdline: e.target.value })}
                        placeholder="quiet splash"
                      />
                    </div>
                  </div>
                )}
              </div>
            ))}
          </div>

          <div className="boot-menu-actions">
            <button type="button" className="secondary-action" onClick={addEntry}>
              Add Entry
            </button>
            <button
              type="button"
              className="primary-action"
              onClick={() => void handlePreview()}
              disabled={loading}
            >
              {loading ? "Rendering..." : "Preview"}
            </button>
          </div>
        </div>

        {/* Right panel — Preview */}
        <div className="panel boot-menu-preview-panel">
          <div className="preview-header">
            <span className="panel-label">Preview — {format}</span>
            {preview && (
              <button type="button" className="secondary-action" onClick={() => void handleCopy()}>
                {copied ? "Copied!" : "Copy to clipboard"}
              </button>
            )}
          </div>

          {error && (
            <p className="render-error">{error}</p>
          )}

          {preview ? (
            <>
              <pre className="boot-menu-output">{preview}</pre>
              <div className="boot-menu-save-row">
                <input
                  type="text"
                  className="boot-menu-save-input"
                  placeholder="Save to file path, e.g. ~/boot.ipxe"
                  value={savePath}
                  onChange={(e) => setSavePath(e.target.value)}
                />
                <button
                  type="button"
                  className="secondary-action"
                  onClick={() => void handleSaveToFile()}
                  disabled={!savePath.trim()}
                >
                  {saved ? "Saved!" : "Save to File"}
                </button>
              </div>
            </>
          ) : (
            <p className="boot-menu-placeholder">
              Configure your menu entries and click Preview to render the{" "}
              {format} output.
            </p>
          )}
        </div>
      </div>
    </div>
  );
}
