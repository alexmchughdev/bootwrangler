import { useCallback, useEffect, useState } from "react";
import {
  type CacheStatus,
  type CatalogueEntry,
  type FlashPlan,
  type Partition,
  type UsbDevice,
  executeFlash,
  imageCacheStatus,
  listDevices,
  listImages,
  planFlash,
} from "../api/backend";

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

interface SelectedImage {
  entry: CatalogueEntry;
  version: string;
  arch: string;
  imagePath: string;
}

interface PartitionTarget {
  partition: Partition;
  device: UsbDevice;
}

type Step = "select-image" | "select-partition" | "confirm";

type CacheMap = Record<string, CacheStatus>;

function cacheKey(id: string, version: string, arch: string): string {
  return `${id}/${version}/${arch}`;
}

// ---------------------------------------------------------------------------
// Main page
// ---------------------------------------------------------------------------

export default function FlashPartition() {
  const [step, setStep] = useState<Step>("select-image");
  const [selectedImage, setSelectedImage] = useState<SelectedImage | null>(null);
  const [selectedTarget, setSelectedTarget] = useState<PartitionTarget | null>(null);
  const [plan, setPlan] = useState<FlashPlan | null>(null);

  function handleImageSelected(img: SelectedImage) {
    setSelectedImage(img);
    setSelectedTarget(null);
    setPlan(null);
    setStep("select-partition");
  }

  function handlePartitionAndPlan(target: PartitionTarget, p: FlashPlan) {
    setSelectedTarget(target);
    setPlan(p);
    setStep("confirm");
  }

  function handleStartOver() {
    setStep("select-image");
    setSelectedImage(null);
    setSelectedTarget(null);
    setPlan(null);
  }

  return (
    <div className="flash-page">
      <StepIndicator current={step} />

      {step === "select-image" && (
        <SelectImageStep onNext={handleImageSelected} />
      )}
      {step === "select-partition" && selectedImage !== null && (
        <SelectPartitionStep
          selectedImage={selectedImage}
          onBack={() => setStep("select-image")}
          onNext={handlePartitionAndPlan}
        />
      )}
      {step === "confirm" && plan !== null && selectedTarget !== null && (
        <ConfirmStep
          plan={plan}
          target={selectedTarget}
          onStartOver={handleStartOver}
        />
      )}
    </div>
  );
}

// ---------------------------------------------------------------------------
// Step indicator
// ---------------------------------------------------------------------------

const STEPS: { key: Step; label: string }[] = [
  { key: "select-image", label: "1. Select Image" },
  { key: "select-partition", label: "2. Select Partition" },
  { key: "confirm", label: "3. Confirm & Flash" },
];

function StepIndicator({ current }: { current: Step }) {
  const currentIdx = STEPS.findIndex((s) => s.key === current);
  return (
    <div className="flash-steps">
      {STEPS.map((s, i) => (
        <div
          key={s.key}
          className={
            i < currentIdx
              ? "flash-step flash-step--done"
              : i === currentIdx
                ? "flash-step flash-step--active"
                : "flash-step flash-step--pending"
          }
        >
          {s.label}
        </div>
      ))}
    </div>
  );
}

// ---------------------------------------------------------------------------
// Step 1: Select Image (partition-compatible only)
// ---------------------------------------------------------------------------

interface SelectImageStepProps {
  onNext: (img: SelectedImage) => void;
}

function SelectImageStep({ onNext }: SelectImageStepProps) {
  const [entries, setEntries] = useState<CatalogueEntry[]>([]);
  const [cacheMap, setCacheMap] = useState<CacheMap>({});
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [selected, setSelected] = useState<SelectedImage | null>(null);

  const fetchCacheStatuses = useCallback(async (catalogue: CatalogueEntry[]) => {
    const tasks: Array<{ id: string; version: string; arch: string }> = [];
    for (const entry of catalogue) {
      for (const ver of entry.Versions) {
        for (const arch of ver.Architectures) {
          tasks.push({ id: entry.ID, version: ver.Version, arch: arch.Arch });
        }
      }
    }
    const results = await Promise.all(
      tasks.map(async (t) => {
        try {
          const status = await imageCacheStatus(t.id, t.version, t.arch);
          return { key: cacheKey(t.id, t.version, t.arch), status };
        } catch {
          return {
            key: cacheKey(t.id, t.version, t.arch),
            status: { Cached: false, Path: "", Verified: false } satisfies CacheStatus,
          };
        }
      }),
    );
    const map: CacheMap = {};
    for (const r of results) {
      map[r.key] = r.status;
    }
    setCacheMap(map);
  }, []);

  useEffect(() => {
    void (async () => {
      setLoading(true);
      setError("");
      try {
        const catalogue = await listImages();
        setEntries(catalogue);
        await fetchCacheStatuses(catalogue);
      } catch (err) {
        setError(String(err));
      } finally {
        setLoading(false);
      }
    })();
  }, [fetchCacheStatuses]);

  // Filter: only entries that have at least one Partition-compatible image
  const partitionEntries = entries.filter((entry) =>
    entry.Versions.some((ver) =>
      ver.Architectures.some((arch) =>
        arch.Images.some((img) => img.Compatibility.Partition),
      ),
    ),
  );

  const hasCachedPartitionImage =
    !loading &&
    partitionEntries.some((entry) =>
      entry.Versions.some((ver) =>
        ver.Architectures.some((arch) => {
          const key = cacheKey(entry.ID, ver.Version, arch.Arch);
          return cacheMap[key]?.Cached === true;
        }),
      ),
    );

  return (
    <div className="flash-step-panel">
      <div className="library-header">
        <h2>Select Image</h2>
        <span className="panel-label">Partition-compatible images only</span>
      </div>

      {error && <p className="render-error">{error}</p>}

      {!loading && partitionEntries.length > 0 && !hasCachedPartitionImage && (
        <p className="flash-note">
          No partition-compatible images are cached locally. Download an image
          from the Images page before flashing.
        </p>
      )}

      {loading ? (
        <p className="library-empty">Loading catalogue…</p>
      ) : partitionEntries.length === 0 ? (
        <p className="library-empty">No partition-compatible images available.</p>
      ) : (
        <div className="image-list">
          {partitionEntries.map((entry) =>
            entry.Versions.map((ver) =>
              ver.Architectures.map((archEntry) => {
                const partitionImages = archEntry.Images.filter(
                  (img) => img.Compatibility.Partition,
                );
                if (partitionImages.length === 0) return null;
                const firstImg = partitionImages[0];
                const key = cacheKey(entry.ID, ver.Version, archEntry.Arch);
                const status = cacheMap[key];
                const rowKey = `${entry.ID}/${ver.Version}/${archEntry.Arch}`;
                const isSelected =
                  selected !== null &&
                  selected.entry.ID === entry.ID &&
                  selected.version === ver.Version &&
                  selected.arch === archEntry.Arch;

                return (
                  <button
                    key={rowKey}
                    type="button"
                    className={
                      isSelected
                        ? "flash-image-row flash-image-row--selected"
                        : "flash-image-row"
                    }
                    onClick={() =>
                      setSelected({
                        entry,
                        version: ver.Version,
                        arch: archEntry.Arch,
                        imagePath: firstImg.URL,
                      })
                    }
                  >
                    <div className="flash-image-row-main">
                      <div className="flash-image-row-title">
                        <span className="panel-label">{entry.Family}</span>
                        <strong className="image-name">{entry.Name}</strong>
                        <span className="image-id">{entry.ID}</span>
                      </div>
                      <div className="image-arch-entry">
                        <span className="image-version-label">v{ver.Version}</span>
                        <span className="image-arch-badge">{archEntry.Arch}</span>
                        {status !== undefined ? (
                          <span
                            className={
                              status.Cached
                                ? "cache-badge cache-badge--cached"
                                : "cache-badge cache-badge--missing"
                            }
                          >
                            {status.Cached ? "Cached" : "Not cached"}
                          </span>
                        ) : (
                          <span className="cache-badge cache-badge--unknown">
                            Checking…
                          </span>
                        )}
                        <span className="compat-badge">Partition</span>
                      </div>
                    </div>
                    {isSelected && (
                      <span className="flash-selected-check">&#10003;</span>
                    )}
                  </button>
                );
              }),
            ),
          )}
        </div>
      )}

      <div className="flash-step-actions">
        <button
          type="button"
          className="primary-action"
          disabled={selected === null}
          onClick={() => {
            if (selected !== null) onNext(selected);
          }}
        >
          Next: Select Partition
        </button>
      </div>
    </div>
  );
}

// ---------------------------------------------------------------------------
// Step 2: Select Partition
// ---------------------------------------------------------------------------

interface SelectPartitionStepProps {
  selectedImage: SelectedImage;
  onBack: () => void;
  onNext: (target: PartitionTarget, plan: FlashPlan) => void;
}

function SelectPartitionStep({
  selectedImage,
  onBack,
  onNext,
}: SelectPartitionStepProps) {
  const [devices, setDevices] = useState<UsbDevice[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [selected, setSelected] = useState<PartitionTarget | null>(null);
  const [planning, setPlanning] = useState(false);
  const [plan, setPlan] = useState<FlashPlan | null>(null);
  const [planError, setPlanError] = useState("");

  async function load() {
    setLoading(true);
    setError("");
    try {
      const result = await listDevices();
      setDevices(result);
    } catch (err) {
      setError(String(err));
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    void load();
  }, []);

  // Flatten all partitions across all devices
  const allTargets: PartitionTarget[] = devices.flatMap((dev) =>
    dev.Partitions.map((p) => ({ partition: p, device: dev })),
  );

  async function handlePreviewPlan(target: PartitionTarget) {
    setSelected(target);
    setPlan(null);
    setPlanError("");
    setPlanning(true);
    try {
      const p = await planFlash(target.partition.Path, selectedImage.imagePath);
      setPlan(p);
    } catch (err) {
      setPlanError(String(err));
    } finally {
      setPlanning(false);
    }
  }

  function isDisabled(target: PartitionTarget): boolean {
    return !target.device.Safe || target.partition.MountPoint !== "";
  }

  function disabledReason(target: PartitionTarget): string {
    if (!target.device.Safe) return target.device.SafetyNote;
    if (target.partition.MountPoint !== "")
      return `Partition is mounted at ${target.partition.MountPoint}`;
    return "";
  }

  return (
    <div className="flash-step-panel">
      <div className="library-header">
        <h2>Select Partition</h2>
        <div style={{ display: "flex", gap: "8px" }}>
          <button
            type="button"
            className="secondary-action"
            onClick={() => void load()}
            disabled={loading}
          >
            {loading ? "Refreshing…" : "Refresh"}
          </button>
          <button type="button" className="secondary-action" onClick={onBack}>
            Back
          </button>
        </div>
      </div>

      <div className="flash-image-summary">
        <span className="panel-label">Selected image</span>
        <span>
          {selectedImage.entry.Name} &mdash; v{selectedImage.version} /{" "}
          {selectedImage.arch}
        </span>
      </div>

      {error && <p className="render-error">{error}</p>}

      {!loading && allTargets.length === 0 && !error && (
        <p className="library-empty">
          No partitions found. Connect a device with partitions.
        </p>
      )}

      {allTargets.length > 0 && (
        <table className="file-table flash-partition-table">
          <thead>
            <tr>
              <th>Partition</th>
              <th>Device</th>
              <th>Size</th>
              <th>Filesystem</th>
              <th>Label</th>
              <th>Mount</th>
              <th>Status</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            {allTargets.map((target) => {
              const disabled = isDisabled(target);
              const reason = disabledReason(target);
              const isSelected =
                selected?.partition.Path === target.partition.Path;
              return (
                <>
                  <tr
                    key={target.partition.Path}
                    style={disabled ? { opacity: 0.5 } : undefined}
                    className={isSelected ? "flash-partition-row--selected" : undefined}
                  >
                    <td className="file-path">{target.partition.Path}</td>
                    <td className="file-path">{target.device.Path}</td>
                    <td>{target.partition.SizeHuman}</td>
                    <td>{target.partition.Filesystem || "—"}</td>
                    <td>{target.partition.Label || "—"}</td>
                    <td>{target.partition.MountPoint || "—"}</td>
                    <td>
                      {target.device.Safe ? (
                        target.partition.MountPoint ? (
                          <span className="badge badge-unsafe">Mounted</span>
                        ) : (
                          <span className="badge badge-safe">Available</span>
                        )
                      ) : (
                        <span className="badge badge-unsafe" title={reason}>
                          Unsafe device
                        </span>
                      )}
                    </td>
                    <td>
                      <button
                        type="button"
                        className="secondary-action"
                        disabled={disabled || planning}
                        title={disabled ? reason : undefined}
                        onClick={() => void handlePreviewPlan(target)}
                        style={{ whiteSpace: "nowrap" }}
                      >
                        {planning && isSelected ? "Planning…" : "Preview Plan"}
                      </button>
                    </td>
                  </tr>
                  {isSelected && (planError || plan !== null) && (
                    <tr key={`${target.partition.Path}-plan`}>
                      <td colSpan={8}>
                        {planError && (
                          <p className="render-error" style={{ margin: "8px 0" }}>
                            {planError}
                          </p>
                        )}
                        {plan !== null && (
                          <div className="flash-plan-preview">
                            <span className="panel-label">Flash Plan</span>
                            <div className="flash-plan-rows">
                              <div className="flash-plan-row">
                                <span>Image path</span>
                                <code>{plan.ImagePath}</code>
                              </div>
                              <div className="flash-plan-row">
                                <span>Image size</span>
                                <code>{plan.ImageSizeHuman}</code>
                              </div>
                              <div className="flash-plan-row">
                                <span>Partition path</span>
                                <code>{plan.DevicePath}</code>
                              </div>
                              <div className="flash-plan-row">
                                <span>Partition size</span>
                                <code>{plan.DeviceSizeHuman}</code>
                              </div>
                              <div className="flash-plan-row">
                                <span>Command</span>
                                <code className="flash-plan-cmd">{plan.Command}</code>
                              </div>
                            </div>
                            <div
                              className="flash-step-actions"
                              style={{ marginTop: "16px" }}
                            >
                              <button
                                type="button"
                                className="primary-action"
                                onClick={() => onNext(target, plan)}
                              >
                                Next: Confirm &amp; Flash
                              </button>
                            </div>
                          </div>
                        )}
                      </td>
                    </tr>
                  )}
                </>
              );
            })}
          </tbody>
        </table>
      )}
    </div>
  );
}

// ---------------------------------------------------------------------------
// Step 3: Confirm & Flash
// ---------------------------------------------------------------------------

interface ConfirmStepProps {
  plan: FlashPlan;
  target: PartitionTarget;
  onStartOver: () => void;
}

type FlashState = "idle" | "flashing" | "success" | "error";

function ConfirmStep({ plan, target, onStartOver }: ConfirmStepProps) {
  const [typed, setTyped] = useState("");
  const [flashState, setFlashState] = useState<FlashState>("idle");
  const [flashError, setFlashError] = useState("");

  const confirmed = typed === plan.DevicePath;

  async function handleFlash() {
    if (!confirmed) return;
    setFlashState("flashing");
    setFlashError("");
    try {
      await executeFlash(plan);
      setFlashState("success");
    } catch (err) {
      setFlashError(String(err));
      setFlashState("error");
    }
  }

  if (flashState === "success") {
    return (
      <div className="flash-step-panel">
        <div className="flash-result flash-result--success">
          <strong>Flash complete.</strong>
          <p>
            {plan.ImagePath} was successfully written to {plan.DevicePath}.
            Other partitions on {target.device.Path} were not affected.
          </p>
        </div>
        <div className="flash-step-actions">
          <button type="button" className="secondary-action" onClick={onStartOver}>
            Start Over
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="flash-step-panel">
      <h2>Confirm &amp; Flash</h2>

      <div className="flash-warning">
        <strong>&#9888; Destructive operation</strong>
        <p>
          This will overwrite partition <code>{plan.DevicePath}</code> (
          {target.partition.Filesystem || "unknown FS"}, {plan.DeviceSizeHuman}).
          Other partitions on {target.device.Path} will not be affected.
        </p>
      </div>

      <div className="flash-plan-preview">
        <span className="panel-label">Flash Plan</span>
        <div className="flash-plan-rows">
          <div className="flash-plan-row">
            <span>Image path</span>
            <code>{plan.ImagePath}</code>
          </div>
          <div className="flash-plan-row">
            <span>Image size</span>
            <code>{plan.ImageSizeHuman}</code>
          </div>
          <div className="flash-plan-row">
            <span>Partition path</span>
            <code>{plan.DevicePath}</code>
          </div>
          <div className="flash-plan-row">
            <span>Partition size</span>
            <code>{plan.DeviceSizeHuman}</code>
          </div>
          <div className="flash-plan-row">
            <span>Command</span>
            <code className="flash-plan-cmd">{plan.Command}</code>
          </div>
        </div>
      </div>

      <div className="flash-confirm-field">
        <label htmlFor="flash-partition-confirm-input">
          <span className="panel-label">
            Type <code>{plan.DevicePath}</code> to enable the Flash button
          </span>
          <input
            id="flash-partition-confirm-input"
            type="text"
            value={typed}
            onChange={(e) => setTyped(e.target.value)}
            placeholder={plan.DevicePath}
            disabled={flashState === "flashing"}
            autoComplete="off"
            spellCheck={false}
          />
        </label>
      </div>

      {flashState === "error" && (
        <p className="render-error">{flashError}</p>
      )}

      <div className="flash-step-actions">
        <button
          type="button"
          className="danger-action"
          disabled={!confirmed || flashState === "flashing"}
          onClick={() => void handleFlash()}
        >
          {flashState === "flashing" ? "Flashing…" : "Flash"}
        </button>
        <button
          type="button"
          className="secondary-action"
          onClick={onStartOver}
          disabled={flashState === "flashing"}
        >
          Start Over
        </button>
      </div>
    </div>
  );
}
