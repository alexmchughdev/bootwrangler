import { useState } from "react";

const YAML_PLACEHOLDER = `# Recipe YAML format:
# label: my-usb-stick
# partitions:
#   - label: efi
#     fs: fat32
#     size: 512M
#     content: ./efi-files/
#   - label: ubuntu
#     fs: ext4
#     size: 8G
#     image: ./ubuntu-24.04.iso
#   - label: data
#     fs: exfat
#     size: fill`;

type PlanState =
  | { kind: "idle" }
  | { kind: "error"; message: string }
  | { kind: "info"; message: string };

export default function MediaBuilder() {
  const [yaml, setYaml] = useState("");
  const [plan, setPlan] = useState<PlanState>({ kind: "idle" });

  function handleValidate() {
    if (yaml.trim() === "") {
      setPlan({ kind: "error", message: "Recipe is empty. Paste or type a YAML recipe first." });
      return;
    }
    if (!yaml.includes("partitions:")) {
      setPlan({
        kind: "error",
        message:
          'Invalid recipe: missing required "partitions:" key. Check your YAML structure.',
      });
      return;
    }
    setPlan({
      kind: "info",
      message:
        "Plan preview not available in this release; use the CLI:\n\n  bootwrangler recipe plan <recipe.yaml>",
    });
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

      <div className="media-builder-panels">
        <div className="panel media-builder-input">
          <span className="panel-label">Recipe YAML</span>
          <p className="media-builder-hint">
            Paste an existing recipe or write one from scratch. The recipe must contain a{" "}
            <code className="inline-code">partitions:</code> key with at least one partition entry.
          </p>
          <textarea
            className="media-builder-textarea"
            placeholder={YAML_PLACEHOLDER}
            value={yaml}
            onChange={(e) => setYaml(e.target.value)}
            spellCheck={false}
          />
          <div className="media-builder-actions">
            <button className="primary-action" type="button" onClick={handleValidate}>
              Validate &amp; Plan
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
          {plan.kind === "error" && (
            <p className="render-error media-builder-result-text">{plan.message}</p>
          )}
          {plan.kind === "info" && (
            <pre className="media-builder-plan-output">{plan.message}</pre>
          )}
        </div>
      </div>
    </div>
  );
}
