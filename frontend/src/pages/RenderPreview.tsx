import { useEffect, useState } from "react";
import type { Profile } from "../types/profile";
import {
  formatNetbootCmdline,
  libraryGet,
  libraryList,
  listNetbootImages,
  loadProfile,
  readRenderedFile,
  renderProfile,
  validateProfile,
  type NetbootImage,
  type RenderManifest,
} from "../api/backend";

type RenderState = "idle" | "loading" | "done" | "error";

const LS_LAST_RENDER_DIR = "bw_last_render_dir";

interface Props {
  onNavigate?: (section: string) => void;
}

export default function RenderPreview({ onNavigate }: Props) {
  const [profilePath, setProfilePath] = useState("");
  const [outDir, setOutDir] = useState("");
  const [state, setState] = useState<RenderState>("idle");
  const [manifest, setManifest] = useState<RenderManifest | null>(null);
  const [errorMsg, setErrorMsg] = useState("");
  const [validationProblems, setValidationProblems] = useState<string[]>([]);
  const [previewFile, setPreviewFile] = useState<string | null>(null);
  const [previewContent, setPreviewContent] = useState("");
  const [previewLoading, setPreviewLoading] = useState(false);
  const [libraryNames, setLibraryNames] = useState<string[]>([]);
  const [selectedLibName, setSelectedLibName] = useState("");
  const [netbootImages, setNetbootImages] = useState<NetbootImage[]>([]);
  const [netbootCmdline, setNetbootCmdline] = useState<string | null>(null);
  const [netbootCopied, setNetbootCopied] = useState(false);

  useEffect(() => {
    void libraryList().then((entries) => setLibraryNames(entries.map((e) => e.name))).catch(() => {});
    void listNetbootImages().then(setNetbootImages).catch(() => {});
  }, []);

  async function handleRenderFromLibrary() {
    if (!selectedLibName) return;
    setState("loading");
    setErrorMsg("");
    setManifest(null);
    setValidationProblems([]);
    setPreviewFile(null);
    let profile: Profile;
    try {
      profile = await libraryGet(selectedLibName);
    } catch (err) {
      setErrorMsg(`Failed to load library profile: ${String(err)}`);
      setState("error");
      return;
    }
    await doRender(profile);
  }

  async function handleRender() {
    if (!profilePath.trim()) {
      setErrorMsg("Profile path is required.");
      setState("error");
      return;
    }

    setState("loading");
    setErrorMsg("");
    setManifest(null);
    setValidationProblems([]);
    setPreviewFile(null);

    let profile: Profile;
    try {
      profile = await loadProfile(profilePath.trim());
    } catch (err) {
      setErrorMsg(`Failed to load profile: ${String(err)}`);
      setState("error");
      return;
    }

    await doRender(profile);
  }

  async function doRender(profile: Profile) {

    const validation = await validateProfile(profile);
    if (!validation.valid) {
      setValidationProblems(validation.problems);
      setState("error");
      return;
    }

    try {
      const serverBaseURL = localStorage.getItem("bw_server_base_url") ?? "";
      const effectiveOutDir = outDir.trim() || ".";
      const result = await renderProfile(profile, effectiveOutDir, serverBaseURL);
      localStorage.setItem(LS_LAST_RENDER_DIR, effectiveOutDir);
      setManifest(result);
      setState("done");
      if (serverBaseURL) {
        void formatNetbootCmdline(result.os_family, serverBaseURL)
          .then(setNetbootCmdline)
          .catch(() => setNetbootCmdline(null));
      } else {
        setNetbootCmdline(null);
      }
    } catch (err) {
      setErrorMsg(`Render failed: ${String(err)}`);
      setState("error");
    }
  }

  async function handlePreviewFile(path: string, outDirectory: string) {
    const fullPath = `${outDirectory.trim() || "."}/${path}`;
    setPreviewFile(path);
    setPreviewContent("");
    setPreviewLoading(true);
    try {
      const content = await readRenderedFile(fullPath);
      setPreviewContent(content);
    } catch {
      setPreviewContent(
        `File: ${fullPath}\n\n(Open in Neovim or a file manager to view generated content.)`,
      );
    } finally {
      setPreviewLoading(false);
    }
  }

  return (
    <div className="render-preview">
      <section className="render-controls panel">
        <h2>Render Profile</h2>
        <p className="render-hint">
          Render writes installer files to the output directory. Leave output blank to use the
          current directory.
        </p>

        {libraryNames.length > 0 && (
          <div className="render-library-row">
            <span className="field-label">From library</span>
            <div className="render-library-input">
              <select
                value={selectedLibName}
                onChange={(e) => setSelectedLibName(e.target.value)}
              >
                <option value="">— pick a profile —</option>
                {libraryNames.map((n) => (
                  <option key={n} value={n}>
                    {n}
                  </option>
                ))}
              </select>
              <button
                type="button"
                className="secondary-action"
                disabled={!selectedLibName || state === "loading"}
                onClick={() => void handleRenderFromLibrary()}
              >
                Render
              </button>
            </div>
          </div>
        )}

        <label>
          Profile path
          <input
            type="text"
            placeholder="e.g. /home/user/.bootwrangler/profiles/edge-node.yaml"
            value={profilePath}
            onChange={(e) => setProfilePath(e.target.value)}
          />
        </label>

        <label>
          Output directory
          <input
            type="text"
            placeholder="e.g. /tmp/render-out (leave blank for current directory)"
            value={outDir}
            onChange={(e) => setOutDir(e.target.value)}
          />
        </label>

        <button
          className="primary-action"
          type="button"
          onClick={() => void handleRender()}
          disabled={state === "loading"}
        >
          {state === "loading" ? "Rendering…" : "Render"}
        </button>
      </section>

      {state === "error" && (
        <section className="render-result panel panel-error">
          {errorMsg && <p className="render-error">{errorMsg}</p>}
          {validationProblems.length > 0 && (
            <>
              <strong>Validation errors</strong>
              <ul className="problem-list">
                {validationProblems.map((p) => (
                  <li key={p}>{p}</li>
                ))}
              </ul>
            </>
          )}
        </section>
      )}

      {state === "done" && manifest && (
        <section className="render-result panel">
          <div className="manifest-header">
            <div>
              <span className="panel-label">Renderer</span>
              <strong>
                {manifest.renderer} — {manifest.os_family} {manifest.os_version}
              </strong>
            </div>
            <div>
              <span className="panel-label">Profile</span>
              <strong>{manifest.profile_name}</strong>
            </div>
          </div>

          {manifest.warnings.length > 0 && (
            <div className="render-warnings">
              <strong>Warnings</strong>
              <ul>
                {manifest.warnings.map((w) => (
                  <li key={w}>{w}</li>
                ))}
              </ul>
            </div>
          )}

          <div className="manifest-files">
            <strong>Generated files</strong>
            <table className="file-table">
              <thead>
                <tr>
                  <th>File</th>
                  <th>Purpose</th>
                  <th>SHA256</th>
                  <th></th>
                </tr>
              </thead>
              <tbody>
                {manifest.files.map((f) => (
                  <tr key={f.path}>
                    <td className="file-path">{f.path}</td>
                    <td>{f.purpose}</td>
                    <td className="checksum">
                      {f.sha256 ? f.sha256.slice(0, 12) + "…" : "—"}
                    </td>
                    <td>
                      <button
                        type="button"
                        className="link-btn"
                        onClick={() => void handlePreviewFile(f.path, outDir)}
                      >
                        Preview
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {(() => {
            const directImage = netbootImages.find(
              (img) => img.Family === manifest.os_family && img.Version === manifest.os_version,
            ) ?? netbootImages.find((img) => img.Family === manifest.os_family);
            const serverBaseURL = localStorage.getItem("bw_server_base_url") ?? "";
            const chainNote = !directImage
              ? `# ${manifest.os_family} has no standalone netboot image; chain via netboot.xyz\nchain https://boot.netboot.xyz\n\n# Then add this to the installer's 'Additional boot options':\n# ${netbootCmdline ?? "<configure provisioning server URL in Settings>"}`
              : `#!ipxe\nkernel ${directImage.KernelURL}${netbootCmdline ? ` ${netbootCmdline}` : ""}\ninitrd ${directImage.InitrdURL}\nboot`;

            return (
              <div className="render-netboot-panel">
                <span className="panel-label">Netboot Entry</span>
                <p className="render-hint">
                  Paste this iPXE snippet into your PXE server or Boot Menu page. The installer
                  will fetch its config from your provisioning server automatically.
                  {!serverBaseURL && (
                    <> Set the <strong>server base URL</strong> in Settings to fill in the config URL.</>
                  )}
                </p>
                <pre className="render-netboot-snippet">{chainNote}</pre>
                <div className="render-netboot-actions">
                  <button
                    type="button"
                    className="secondary-action"
                    onClick={() => {
                      void navigator.clipboard.writeText(chainNote).then(() => {
                        setNetbootCopied(true);
                        setTimeout(() => setNetbootCopied(false), 2000);
                      });
                    }}
                  >
                    {netbootCopied ? "Copied!" : "Copy"}
                  </button>
                  {directImage && (
                    <span className="render-netboot-source">
                      Kernel/initrd: {directImage.Name} {directImage.Version} ({directImage.Arch})
                    </span>
                  )}
                </div>
              </div>
            );
          })()}

          {onNavigate && (
            <div className="render-result-actions">
              <button
                type="button"
                className="primary-action"
                onClick={() => onNavigate("Provisioning Server")}
              >
                Start Provisioning Server
              </button>
            </div>
          )}
        </section>
      )}

      {previewFile && (
        <section className="file-preview panel">
          <div className="preview-header">
            <strong>{previewFile}</strong>
            <button
              type="button"
              className="link-btn"
              onClick={() => setPreviewFile(null)}
            >
              Close
            </button>
          </div>
          {previewLoading ? (
            <pre className="preview-content">Loading…</pre>
          ) : (
            <pre className="preview-content">{previewContent}</pre>
          )}
        </section>
      )}
    </div>
  );
}
