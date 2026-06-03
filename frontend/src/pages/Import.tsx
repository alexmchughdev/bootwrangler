import { useState } from "react";
import { detectConfigFormat, importConfig, type ImportResult } from "../api/backend";
import type { Profile } from "../types/profile";

interface Props {
  onNavigate: (section: string) => void;
  onOpenInEditor?: (profile: Profile) => void;
}

export default function Import({ onNavigate, onOpenInEditor }: Props) {
  const [content, setContent] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [detectedFormat, setDetectedFormat] = useState("");
  const [result, setResult] = useState<ImportResult | null>(null);

  async function handleDetectAndImport() {
    if (!content.trim()) return;
    setError("");
    setLoading(true);
    try {
      const [fmt, importedResult] = await Promise.all([
        detectConfigFormat(content),
        importConfig(content),
      ]);
      setDetectedFormat(fmt);
      setResult(importedResult);
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : String(err));
      setResult(null);
      setDetectedFormat("");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="import-page">
      <div className="panel panel-wide import-intro">
        <span className="panel-label">Import Config</span>
        <h2>Import existing installer configs</h2>
        <p>
          Paste an existing installer config to convert it into a BootWrangler profile.
          The importer will detect the format automatically and map supported fields.
        </p>
      </div>

      {error && <p className="render-error">{error}</p>}

      <div className="import-columns">
        <div className="panel import-input-panel">
          <span className="panel-label">Paste Config</span>
          <textarea
            className="import-textarea"
            value={content}
            onChange={(e) => setContent(e.target.value)}
            spellCheck={false}
            placeholder={`Paste an existing installer config here…\n\nSupported formats:\n• Alpine answerfile\n• Ubuntu autoinstall\n• Debian preseed\n• Kickstart (Rocky/Fedora)\n• AutoYaST (openSUSE)`}
          />
          <div className="import-actions">
            <button
              className="primary-action"
              type="button"
              onClick={() => void handleDetectAndImport()}
              disabled={loading || !content.trim()}
            >
              {loading ? "Importing…" : "Detect & Import"}
            </button>
          </div>
        </div>

        <div className="panel import-result-panel">
          <span className="panel-label">Result</span>
          {!result ? (
            <p className="library-empty">
              Paste a config and click "Detect &amp; Import" to see the result.
            </p>
          ) : (
            <div className="import-result">
              <div className="import-format-row">
                <span className="import-format-label">Detected format</span>
                <span className="import-format-badge">{detectedFormat || "unknown"}</span>
              </div>

              <div className="import-profile-summary">
                <span className="panel-label">Profile Summary</span>
                <pre className="import-summary-pre">{[
                  `name:      ${result.Profile.name}`,
                  `os.family: ${result.Profile.os.family}`,
                  `os.version: ${result.Profile.os.version}`,
                  `hostname:  ${result.Profile.system.hostname}`,
                  `timezone:  ${result.Profile.system.timezone}`,
                  `keyboard:  ${result.Profile.system.keyboard}`,
                ].join("\n")}</pre>
              </div>

              {result.Warnings.length > 0 && (
                <div className="render-warnings">
                  <strong>Warnings</strong>
                  <ul>
                    {result.Warnings.map((w, i) => (
                      <li key={i}>{w}</li>
                    ))}
                  </ul>
                </div>
              )}

              {result.UnsupportedFields.length > 0 && (
                <div className="import-unsupported">
                  <strong>Unsupported fields</strong>
                  <ul className="import-unsupported-list">
                    {result.UnsupportedFields.map((f, i) => (
                      <li key={i}>
                        <code className="inline-code">{f}</code>
                      </li>
                    ))}
                  </ul>
                </div>
              )}

              <button
                className="primary-action"
                type="button"
                onClick={() => {
                  if (onOpenInEditor) {
                    onOpenInEditor(result.Profile as unknown as Profile);
                  } else {
                    onNavigate("Profiles");
                  }
                }}
              >
                Open in Profile Editor
              </button>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
