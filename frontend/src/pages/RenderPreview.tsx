import { useState } from "react";
import type { Profile } from "../types/profile";
import {
  loadProfile,
  renderProfile,
  validateProfile,
  type RenderManifest,
} from "../api/backend";

type RenderState = "idle" | "loading" | "done" | "error";

export default function RenderPreview() {
  const [profilePath, setProfilePath] = useState("");
  const [outDir, setOutDir] = useState("");
  const [dryRunNote] = useState(
    "Render writes files to the selected output directory. Leave blank to use the current directory.",
  );
  const [state, setState] = useState<RenderState>("idle");
  const [manifest, setManifest] = useState<RenderManifest | null>(null);
  const [errorMsg, setErrorMsg] = useState("");
  const [validationProblems, setValidationProblems] = useState<string[]>([]);
  const [previewFile, setPreviewFile] = useState<string | null>(null);
  const [previewContent, setPreviewContent] = useState("");

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

    const validation = await validateProfile(profile);
    if (!validation.valid) {
      setValidationProblems(validation.problems);
      setState("error");
      return;
    }

    try {
      const serverBaseURL = localStorage.getItem("bw_server_base_url") ?? "";
      const result = await renderProfile(profile, outDir.trim() || ".", serverBaseURL);
      setManifest(result);
      setState("done");
    } catch (err) {
      setErrorMsg(`Render failed: ${String(err)}`);
      setState("error");
    }
  }

  function handlePreviewFile(path: string, outDirectory: string) {
    const fullPath = `${outDirectory.trim() || "."}/${path}`;
    // In browser preview mode we can't read files; show the path
    setPreviewFile(path);
    setPreviewContent(
      `File: ${fullPath}\n\n(Open in Neovim or a file manager to view generated content.)`,
    );
  }

  return (
    <div className="render-preview">
      <section className="render-controls panel">
        <h2>Render Profile</h2>
        <p className="render-hint">{dryRunNote}</p>

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
                        onClick={() => handlePreviewFile(f.path, outDir)}
                      >
                        Preview
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
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
          <pre className="preview-content">{previewContent}</pre>
        </section>
      )}
    </div>
  );
}
