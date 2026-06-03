import { useState } from "react";
import { formatMediaBuildPlan, listDevices, planMediaBuild } from "../api/backend";

const YAML_PLACEHOLDER = `# Recipe YAML format:
# label: my-usb-stick
# partition_table: gpt
# partitions:
#   - label: efi
#     filesystem: fat32
#     size: 512MiB
#     content_type: efi
#   - label: ubuntu
#     filesystem: ext4
#     size: 8GiB
#     content_type: installer-assets
#     source: ./ubuntu-server/
#   - label: data
#     filesystem: ext4
#     size: remaining
#     content_type: data`;

type PlanState =
  | { kind: "idle" }
  | { kind: "loading" }
  | { kind: "error"; message: string }
  | { kind: "plan"; text: string };

export default function MediaBuilder() {
  const [yaml, setYaml] = useState("");
  const [devicePath, setDevicePath] = useState("/dev/sdb");
  const [plan, setPlan] = useState<PlanState>({ kind: "idle" });

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
      setPlan({ kind: "plan", text });
    } catch (err) {
      setPlan({ kind: "error", message: String(err) });
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
            <pre className="media-builder-plan-output">{plan.text}</pre>
          )}
        </div>
      </div>
    </div>
  );
}
