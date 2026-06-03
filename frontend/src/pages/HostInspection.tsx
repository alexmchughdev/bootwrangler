import { useEffect, useState } from "react";
import { gatherHostInfo, type HostInfo } from "../api/backend";

export default function HostInspection() {
  const [info, setInfo] = useState<HostInfo | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  async function load() {
    setLoading(true);
    setError("");
    try {
      const result = await gatherHostInfo();
      setInfo(result);
    } catch (err) {
      setError(String(err));
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    void load();
  }, []);

  return (
    <div className="host-inspection">
      <div className="host-inspection-toolbar">
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

      {loading && !info && <p className="library-empty">Loading host information…</p>}

      {info && (
        <div className="host-inspection-grid">
          <article className="panel">
            <span className="panel-label">System</span>
            <h2>{info.Hostname}</h2>
            <p>Hostname</p>
          </article>

          <article className="panel">
            <span className="panel-label">CPU</span>
            <h2>{info.CPU.Model || "—"}</h2>
            <div className="host-cpu-meta">
              <span className="host-meta-item">
                <span className="panel-label">Cores</span>
                {info.CPU.Cores}
              </span>
              <span className="host-meta-item">
                <span className="panel-label">Threads</span>
                {info.CPU.Threads}
              </span>
            </div>
          </article>

          <article className="panel">
            <span className="panel-label">Memory</span>
            <strong>{info.Memory.TotalHuman}</strong>
            <p>Total RAM</p>
          </article>

          <article className="panel panel-wide">
            <span className="panel-label">Network Interfaces</span>
            {info.Interfaces.length === 0 ? (
              <p className="library-empty">No network interfaces found.</p>
            ) : (
              <table className="file-table host-iface-table">
                <thead>
                  <tr>
                    <th>Name</th>
                    <th>Addresses</th>
                    <th>MAC Address</th>
                  </tr>
                </thead>
                <tbody>
                  {info.Interfaces.map((iface) => (
                    <tr key={iface.Name}>
                      <td className="file-path">{iface.Name}</td>
                      <td>{iface.Addresses.length > 0 ? iface.Addresses.join(", ") : "—"}</td>
                      <td className="host-hwaddr">{iface.HWAddr || "—"}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </article>
        </div>
      )}
    </div>
  );
}
