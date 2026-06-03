import { useEffect, useState } from "react";
import { listDevices, type Partition, type UsbDevice } from "../api/backend";

interface Props {
  onNavigate?: (section: string) => void;
}

export default function USBDevices({ onNavigate }: Props) {
  const [devices, setDevices] = useState<UsbDevice[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [expanded, setExpanded] = useState<Set<string>>(new Set());

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

  function toggleExpanded(path: string) {
    setExpanded((prev) => {
      const next = new Set(prev);
      if (next.has(path)) {
        next.delete(path);
      } else {
        next.add(path);
      }
      return next;
    });
  }

  return (
    <div className="usb-devices">
      <div className="usb-toolbar">
        <button
          type="button"
          className="secondary-action"
          onClick={() => void load()}
          disabled={loading}
        >
          {loading ? "Refreshing…" : "Refresh"}
        </button>
      </div>

      {error && <p className="render-error">{error}</p>}

      {!loading && devices.length === 0 && !error && (
        <p className="library-empty">No block devices found.</p>
      )}

      <div className="usb-device-list">
        {devices.map((dev) => (
          <DeviceCard
            key={dev.Path}
            device={dev}
            isExpanded={expanded.has(dev.Path)}
            onToggle={() => toggleExpanded(dev.Path)}
            onFlash={() => onNavigate?.("Flash Image")}
          />
        ))}
      </div>
    </div>
  );
}

interface DeviceCardProps {
  device: UsbDevice;
  isExpanded: boolean;
  onToggle: () => void;
  onFlash: () => void;
}

function DeviceCard({ device, isExpanded, onToggle, onFlash }: DeviceCardProps) {
  return (
    <article className="usb-device-card panel">
      <div className="usb-device-header">
        <div className="usb-device-main">
          <div className="usb-device-path">
            <code>{device.Path}</code>
            {device.Safe ? (
              <span className="badge badge-safe">Safe</span>
            ) : (
              <span className="badge badge-unsafe" title={device.SafetyNote}>
                Unsafe
              </span>
            )}
          </div>
          <div className="usb-device-meta">
            <span className="usb-meta-item">
              <span className="panel-label">Model</span>
              {device.Model || "—"}
            </span>
            <span className="usb-meta-item">
              <span className="panel-label">Size</span>
              {device.SizeHuman}
            </span>
            <span className="usb-meta-item">
              <span className="panel-label">Transport</span>
              {device.Transport || "—"}
            </span>
            <span className="usb-meta-item">
              <span className="panel-label">Removable</span>
              {device.Removable ? "Yes" : "No"}
            </span>
          </div>
          {!device.Safe && device.SafetyNote && (
            <p className="usb-safety-note">{device.SafetyNote}</p>
          )}
        </div>

        <div className="usb-device-actions">
          {device.Partitions.length > 0 && (
            <button
              type="button"
              className="text-action"
              onClick={onToggle}
            >
              {isExpanded ? "Hide partitions" : `Show ${device.Partitions.length} partition${device.Partitions.length !== 1 ? "s" : ""}`}
            </button>
          )}
          {device.Safe ? (
            <button
              type="button"
              className="secondary-action"
              onClick={onFlash}
            >
              Flash Image
            </button>
          ) : (
            <button
              type="button"
              className="secondary-action"
              disabled
              title={device.SafetyNote}
            >
              Flash Image
            </button>
          )}
        </div>
      </div>

      {isExpanded && device.Partitions.length > 0 && (
        <PartitionTable partitions={device.Partitions} />
      )}
    </article>
  );
}

interface PartitionTableProps {
  partitions: Partition[];
}

function PartitionTable({ partitions }: PartitionTableProps) {
  return (
    <div className="usb-partitions">
      <table className="file-table">
        <thead>
          <tr>
            <th>Path</th>
            <th>Size</th>
            <th>Filesystem</th>
            <th>Mount Point</th>
            <th>Label</th>
          </tr>
        </thead>
        <tbody>
          {partitions.map((p) => (
            <tr key={p.Path}>
              <td className="file-path">{p.Path}</td>
              <td>{p.SizeHuman}</td>
              <td>{p.Filesystem || "—"}</td>
              <td>{p.MountPoint || "—"}</td>
              <td>{p.Label || "—"}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
