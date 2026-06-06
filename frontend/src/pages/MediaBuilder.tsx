import { useState } from "react";
import {
  type BuildPlan,
  formatMediaBuildPlan,
  listDevices,
  planMediaBuild,
  readRenderedFile,
  writeTextFile,
} from "../api/backend";

const YAML_PLACEHOLDER = `name: lab-usb
device:
  partition_table: gpt
boot:
  mode: uefi-bios
  menu: ipxe
partitions:
  - label: BOOTWRANGLER
    size: 2G
    filesystem: fat32
    content:
      type: boot-menu
  - label: UBUNTU_24
    size: 6G
    filesystem: exfat
    content:
      type: catalogue-image
      image: ubuntu-server
      version: "24.04"
  - label: STORAGE
    size: remaining
    filesystem: exfat
    content:
      type: empty`;

type PlanState =
  | { kind: "idle" }
  | { kind: "loading" }
  | { kind: "error"; message: string }
  | { kind: "plan"; text: string; buildPlan: BuildPlan };

type PlanAction = BuildPlan["Actions"][number];

function formatBytes(bytes: number): string {
  if (!Number.isFinite(bytes) || bytes < 0) return "0 B";
  const units = ["B", "KB", "MB", "GB", "TB"];
  let value = bytes;
  let unit = 0;
  while (value >= 1024 && unit < units.length - 1) {
    value /= 1024;
    unit += 1;
  }
  const digits = unit <= 1 ? 0 : 1;
  return `${value.toFixed(digits)} ${units[unit]}`;
}

function contentLabel(action: PlanAction): string {
  const content = action.Content;
  switch (content.Type) {
    case "catalogue-image":
    case "custom-image":
    case "image-file":
      return [content.Image, content.Version].filter(Boolean).join(" ");
    case "rendered-profile":
      return content.Profile ?? "";
    case "profile-bundle":
      return content.Bundle ?? "";
    default:
      return content.Type;
  }
}

function statusClass(status: string): string {
  switch (status) {
    case "ready":
      return "cache-badge cache-badge--cached";
    case "warning":
      return "cache-badge cache-badge--unverified";
    case "blocked":
      return "cache-badge cache-badge--missing";
    default:
      return "cache-badge cache-badge--unknown";
  }
}

export default function MediaBuilder() {
  const [yaml, setYaml] = useState("");
  const [devicePath, setDevicePath] = useState("/dev/sdb");
  const [plan, setPlan] = useState<PlanState>({ kind: "idle" });
  const [recipeSavePath, setRecipeSavePath] = useState("");
  const [recipeLoadPath, setRecipeLoadPath] = useState("");
  const [recipeSaved, setRecipeSaved] = useState(false);

  async function handleValidate() {
    if (yaml.trim() === "") {
      setPlan({ kind: "error", message: "Recipe is empty. Paste or type a YAML recipe first." });
      return;
    }
    if (!yaml.includes("partitions:")) {
      setPlan({
        kind: "error",
        message: 'Invalid recipe: missing required "partitions:" key.',
      });
      return;
    }

    setPlan({ kind: "loading" });
    try {
      const builtPlan = await planMediaBuild(yaml, devicePath);
      const text = await formatMediaBuildPlan(builtPlan);
      setPlan({ kind: "plan", text, buildPlan: builtPlan });
    } catch (err) {
      setPlan({ kind: "error", message: String(err) });
    }
  }

  async function handleSaveRecipe() {
    if (!recipeSavePath.trim() || !yaml.trim()) return;
    try {
      await writeTextFile(recipeSavePath.trim(), yaml);
      setRecipeSaved(true);
      setTimeout(() => setRecipeSaved(false), 2000);
    } catch {
      // ignore
    }
  }

  async function handleLoadRecipe() {
    if (!recipeLoadPath.trim()) return;
    try {
      const content = await readRenderedFile(recipeLoadPath.trim());
      setYaml(content);
      setRecipeLoadPath("");
    } catch {
      // ignore
    }
  }

  async function handleLoadDevices() {
    try {
      const devices = await listDevices();
      const removable = devices.find((d) => d.Safe && d.Removable);
      if (removable) {
        setDevicePath(removable.Path);
      }
    } catch {
      // ignore
    }
  }

  return (
    <div className="media-builder">
      <div className="panel panel-wide media-builder-intro">
        <span className="panel-label">About Media Builder</span>
        <h2>Multi-partition USB media from a recipe</h2>
        <p>
          Recipes describe the complete layout of a USB drive: partition labels, filesystems,
          sizes, and content sources. You can embed multiple OS images, a custom EFI partition,
          and a shared data volume — all defined in a single YAML file and reproduced exactly
          the same way every time.
        </p>
      </div>

      <div className="media-builder-device-row panel">
        <span className="panel-label">Target Device</span>
        <div className="media-builder-device-input">
          <input
            type="text"
            className="field-input"
            value={devicePath}
            onChange={(e) => setDevicePath(e.target.value)}
            placeholder="/dev/sdb"
          />
          <button className="secondary-action" type="button" onClick={() => void handleLoadDevices()}>
            Auto-detect
          </button>
        </div>
        <p className="media-builder-hint">
          Block device path. The planner needs the device size to validate partition totals.
        </p>
      </div>

      <div className="media-builder-panels">
        <div className="panel media-builder-input">
          <span className="panel-label">Recipe YAML</span>
          <textarea
            className="media-builder-textarea"
            placeholder={YAML_PLACEHOLDER}
            value={yaml}
            onChange={(e) => setYaml(e.target.value)}
            spellCheck={false}
          />
          <div className="media-builder-actions">
            <button
              className="primary-action"
              type="button"
              onClick={() => void handleValidate()}
              disabled={plan.kind === "loading"}
            >
              {plan.kind === "loading" ? "Planning…" : "Validate & Plan"}
            </button>
            {yaml.trim() !== "" && (
              <button
                className="secondary-action"
                type="button"
                onClick={() => {
                  setYaml("");
                  setPlan({ kind: "idle" });
                }}
              >
                Clear
              </button>
            )}
          </div>
          <div className="media-builder-file-row">
            <input
              type="text"
              placeholder="Load recipe from path…"
              value={recipeLoadPath}
              onChange={(e) => setRecipeLoadPath(e.target.value)}
            />
            <button className="link-btn" type="button" onClick={() => void handleLoadRecipe()} disabled={!recipeLoadPath.trim()}>
              Load
            </button>
          </div>
          <div className="media-builder-file-row">
            <input
              type="text"
              placeholder="Save recipe to path…"
              value={recipeSavePath}
              onChange={(e) => setRecipeSavePath(e.target.value)}
            />
            <button className="link-btn" type="button" onClick={() => void handleSaveRecipe()} disabled={!recipeSavePath.trim() || !yaml.trim()}>
              {recipeSaved ? "Saved!" : "Save"}
            </button>
          </div>
        </div>

        <div className="panel media-builder-output">
          <span className="panel-label">Plan Output</span>
          {plan.kind === "idle" && (
            <p className="library-empty">
              Enter a recipe and click "Validate &amp; Plan" to see the partition plan.
            </p>
          )}
          {plan.kind === "loading" && <p className="library-empty">Planning…</p>}
          {plan.kind === "error" && (
            <p className="render-error media-builder-result-text">{plan.message}</p>
          )}
          {plan.kind === "plan" && (
            <div className="media-builder-plan">
              <div className="media-builder-plan-summary">
                <span className={plan.buildPlan.Ready ? "badge badge-safe" : "badge badge-unsafe"}>
                  {plan.buildPlan.Ready ? "Ready" : "Blocked"}
                </span>
                <span>{plan.buildPlan.RecipeName}</span>
                <span>{formatBytes(plan.buildPlan.TotalBytes)} planned</span>
                <span>{formatBytes(plan.buildPlan.DeviceSize)} device</span>
              </div>

              <table className="file-table media-builder-plan-table">
                <thead>
                  <tr>
                    <th>Partition</th>
                    <th>Size</th>
                    <th>Filesystem</th>
                    <th>Content</th>
                    <th>Status</th>
                  </tr>
                </thead>
                <tbody>
                  {plan.buildPlan.Actions.map((action) => {
                    const resolution = action.ContentResolution;
                    const warnings = resolution.Warnings ?? [];
                    const errors = resolution.Errors ?? [];
                    return (
                      <tr key={action.Label}>
                        <td>
                          <strong>{action.Label}</strong>
                        </td>
                        <td>{formatBytes(action.SizeBytes)}</td>
                        <td>{action.Filesystem}</td>
                        <td>
                          <div className="media-builder-content-cell">
                            <span>{action.Content.Type}</span>
                            {contentLabel(action) !== action.Content.Type && (
                              <span className="media-builder-content-ref">{contentLabel(action)}</span>
                            )}
                            {resolution.SourcePath && (
                              <span className="file-path">{resolution.SourcePath}</span>
                            )}
                          </div>
                        </td>
                        <td>
                          <div className="media-builder-content-cell">
                            <span className={statusClass(resolution.Status)}>
                              {resolution.Status || "planned"}
                            </span>
                            {resolution.Message && <span>{resolution.Message}</span>}
                            {warnings.map((warning) => (
                              <span className="media-builder-warning" key={warning}>
                                {warning}
                              </span>
                            ))}
                            {errors.map((error) => (
                              <span className="render-error" key={error}>
                                {error}
                              </span>
                            ))}
                          </div>
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>

              {(plan.buildPlan.Warnings?.length ?? 0) > 0 && (
                <div className="media-builder-message-list">
                  {plan.buildPlan.Warnings.map((warning) => (
                    <span className="media-builder-warning" key={warning}>
                      {warning}
                    </span>
                  ))}
                </div>
              )}

              {(plan.buildPlan.Errors?.length ?? 0) > 0 && (
                <div className="media-builder-message-list">
                  {plan.buildPlan.Errors.map((error) => (
                    <span className="render-error" key={error}>
                      {error}
                    </span>
                  ))}
                </div>
              )}

              <details className="media-builder-plan-details">
                <summary>Text summary</summary>
                <pre className="media-builder-plan-output">{plan.text}</pre>
              </details>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
