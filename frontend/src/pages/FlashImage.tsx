import { useCallback, useEffect, useState } from "react";
import {
  type CacheStatus,
  type CatalogueEntry,
  type FlashPlan,
  type UsbDevice,
  executeFlash,
  imageCacheStatus,
  listDevices,
  listImages,
  planCatalogueFlash,
} from "../api/backend";

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

interface SelectedImage {
  entry: CatalogueEntry;
  version: string;
  arch: string;
}

type Step = "select-image" | "select-device" | "confirm";

type CacheMap = Record<string, CacheStatus>;

function cacheKey(id: string, version: string, arch: string): string {
  return `${id}/${version}/${arch}`;
}

// ---------------------------------------------------------------------------
// Main page
// ---------------------------------------------------------------------------

export default function FlashImage() {
  const [step, setStep] = useState<Step>("select-image");
  const [selectedImage, setSelectedImage] = useState<SelectedImage | null>(null);
  const [selectedDevice, setSelectedDevice] = useState<UsbDevice | null>(null);
  const [plan, setPlan] = useState<FlashPlan | null>(null);

  function handleImageSelected(img: SelectedImage) {
    setSelectedImage(img);
    setSelectedDevice(null);
    setPlan(null);
    setStep("select-device");
  }

  function handleDeviceAndPlan(device: UsbDevice, p: FlashPlan) {
    setSelectedDevice(device);
    setPlan(p);
    setStep("confirm");
  }

  function handleStartOver() {
    setStep("select-image");
    setSelectedImage(null);
    setSelectedDevice(null);
    setPlan(null);
  }

  return (
    <div className="flash-page">
      <StepIndicator current={step} />

      {step === "select-image" && (
        <SelectImageStep onNext={handleImageSelected} />
      )}
      {step === "select-device" && selectedImage !== null && (
        <SelectDeviceStep
          selectedImage={selectedImage}
          onBack={() => setStep("select-image")}
          onNext={handleDeviceAndPlan}
        />
      )}
      {step === "confirm" && plan !== null && selectedDevice !== null && (
        <ConfirmStep
          plan={plan}
          device={selectedDevice}
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
  { key: "select-device", label: "2. Select Device" },
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
// Step 1: Select Image
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

  // Filter: only show entries that have at least one WholeDrive-compatible image
  const wholeDriveEntries = entries.filter((entry) =>
    entry.Versions.some((ver) =>
      ver.Architectures.some((arch) =>
        arch.Images.some((img) => img.Compatibility.WholeDrive),
      ),
    ),
  );

  return (
    <div className="flash-step-panel">
      <div className="library-header">
        <h2>Select Image</h2>
        <span className="panel-label">Whole-drive compatible images only</span>
      </div>

      {error && <p className="render-error">{error}</p>}

      {loading ? (
        <p className="library-empty">Loading catalogue…</p>
      ) : wholeDriveEntries.length === 0 ? (
        <p className="library-empty">No whole-drive compatible images available.</p>
      ) : (
        <div className="image-list">
          {wholeDriveEntries.map((entry) =>
            entry.Versions.map((ver) =>
              ver.Architectures.map((archEntry) => {
                const wholeDriveImages = archEntry.Images.filter(
                  (img) => img.Compatibility.WholeDrive,
                );
                if (wholeDriveImages.length === 0) return null;
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
                        <span className="compat-badge">Whole Drive</span>
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
          Next: Select Device
        </button>
      </div>
    </div>
  );
}

// ---------------------------------------------------------------------------
// Step 2: Select Device
// ---------------------------------------------------------------------------

interface SelectDeviceStepProps {
  selectedImage: SelectedImage;
  onBack: () => void;
  onNext: (device: UsbDevice, plan: FlashPlan) => void;
}

function SelectDeviceStep({ selectedImage, onBack, onNext }: SelectDeviceStepProps) {
  const [devices, setDevices] = useState<UsbDevice[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [selected, setSelected] = useState<UsbDevice | null>(null);
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

  async function handlePreviewPlan(device: UsbDevice) {
    setSelected(device);
    setPlan(null);
    setPlanError("");
    setPlanning(true);
    try {
      const p = await planCatalogueFlash(
        device.Path,
        selectedImage.entry.ID,
        selectedImage.version,
        selectedImage.arch,
      );
      setPlan(p);
    } catch (err) {
      setPlanError(String(err));
    } finally {
      setPlanning(false);
    }
  }

  return (
    <div className="flash-step-panel">
      <div className="library-header">
        <h2>Select Device</h2>
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
          {selectedImage.entry.Name} &mdash; v{selectedImage.version} / {selectedImage.arch}
        </span>
      </div>

      {error && <p className="render-error">{error}</p>}

      {!loading && devices.length === 0 && !error && (
        <p className="library-empty">No block devices found.</p>
      )}

      <div className="usb-device-list">
        {devices.map((dev) => {
          const isSelected = selected?.Path === dev.Path;
          return (
            <article
              key={dev.Path}
              className={
                isSelected
                  ? "usb-device-card panel flash-device-card--selected"
                  : "usb-device-card panel"
              }
              style={!dev.Safe ? { opacity: 0.55 } : undefined}
            >
              <div className="usb-device-header">
                <div className="usb-device-main">
                  <div className="usb-device-path">
                    <code>{dev.Path}</code>
                    {dev.Safe ? (
                      <span className="badge badge-safe">Safe</span>
                    ) : (
                      <span className="badge badge-unsafe" title={dev.SafetyNote}>
                        Unsafe
                      </span>
                    )}
                  </div>
                  <div className="usb-device-meta">
                    <span className="usb-meta-item">
                      <span className="panel-label">Model</span>
                      {dev.Model || "—"}
                    </span>
                    <span className="usb-meta-item">
                      <span className="panel-label">Size</span>
                      {dev.SizeHuman}
                    </span>
                    <span className="usb-meta-item">
                      <span className="panel-label">Transport</span>
                      {dev.Transport || "—"}
                    </span>
                    <span className="usb-meta-item">
                      <span className="panel-label">Removable</span>
                      {dev.Removable ? "Yes" : "No"}
                    </span>
                  </div>
                  {!dev.Safe && dev.SafetyNote && (
                    <p className="usb-safety-note">{dev.SafetyNote}</p>
                  )}
                </div>
                <div className="usb-device-actions">
                  <button
                    type="button"
                    className="secondary-action"
                    disabled={!dev.Safe || planning}
                    title={dev.Safe ? undefined : dev.SafetyNote}
                    onClick={() => void handlePreviewPlan(dev)}
                  >
                    {planning && isSelected ? "Planning…" : "Preview Plan"}
                  </button>
                </div>
              </div>

              {isSelected && planError && (
                <p className="render-error" style={{ marginTop: "12px" }}>
                  {planError}
                </p>
              )}

              {isSelected && plan !== null && (
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
                      <span>Device path</span>
                      <code>{plan.DevicePath}</code>
                    </div>
                    <div className="flash-plan-row">
                      <span>Device size</span>
                      <code>{plan.DeviceSizeHuman}</code>
                    </div>
                    <div className="flash-plan-row">
                      <span>Command</span>
                      <code className="flash-plan-cmd">{plan.Command}</code>
                    </div>
                  </div>
                  <div className="flash-step-actions" style={{ marginTop: "16px" }}>
                    <button
                      type="button"
                      className="primary-action"
                      onClick={() => onNext(dev, plan)}
                    >
                      Next: Confirm &amp; Flash
                    </button>
                  </div>
                </div>
              )}
            </article>
          );
        })}
      </div>
    </div>
  );
}

// ---------------------------------------------------------------------------
// Step 3: Confirm & Flash
// ---------------------------------------------------------------------------

interface ConfirmStepProps {
  plan: FlashPlan;
  device: UsbDevice;
  onStartOver: () => void;
}

type FlashState = "idle" | "flashing" | "success" | "error";

function ConfirmStep({ plan, device, onStartOver }: ConfirmStepProps) {
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
          This will permanently erase <strong>ALL data</strong> on{" "}
          <code>{plan.DevicePath}</code> ({device.Model || "unknown model"},{" "}
          {plan.DeviceSizeHuman}). This cannot be undone.
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
            <span>Device path</span>
            <code>{plan.DevicePath}</code>
          </div>
          <div className="flash-plan-row">
            <span>Device size</span>
            <code>{plan.DeviceSizeHuman}</code>
          </div>
          <div className="flash-plan-row">
            <span>Command</span>
            <code className="flash-plan-cmd">{plan.Command}</code>
          </div>
        </div>
      </div>

      <div className="flash-confirm-field">
        <label htmlFor="flash-confirm-input">
          <span className="panel-label">
            Type <code>{plan.DevicePath}</code> to enable the Flash button
          </span>
          <input
            id="flash-confirm-input"
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
