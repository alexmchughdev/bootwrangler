import { useEffect, useRef, useState } from "react";

interface RequestLog {
  Method: string;
  Path: string;
  StatusCode: number;
  RemoteAddr: string;
}

interface CallbackEvent {
  ProfileName: string;
  Status: string;
  Message: string;
}

interface ServerService {
  ServerStart(root: string, addr: string): Promise<string>;
  ServerStop(): Promise<void>;
  ServerLogs(): Promise<RequestLog[]>;
  ServerEvents(): Promise<CallbackEvent[]>;
  ServerAddr(): Promise<string>;
}

function getServerService(): ServerService | undefined {
  return window.go?.app?.Service as unknown as ServerService | undefined;
}

export default function ProvisioningServer() {
  const [running, setRunning] = useState(false);
  const [addr, setAddr] = useState("");
  const [root, setRoot] = useState("");
  const [listenAddr, setListenAddr] = useState("0.0.0.0:8088");
  const [logs, setLogs] = useState<RequestLog[]>([]);
  const [events, setEvents] = useState<CallbackEvent[]>([]);
  const [error, setError] = useState("");
  const pollRef = useRef<ReturnType<typeof setInterval> | null>(null);

  function startPolling() {
    pollRef.current = setInterval(() => void poll(), 2000);
  }

  function stopPolling() {
    if (pollRef.current) {
      clearInterval(pollRef.current);
      pollRef.current = null;
    }
  }

  async function poll() {
    const svc = getServerService();
    if (!svc) return;
    const [newLogs, newEvents] = await Promise.all([
      svc.ServerLogs(),
      svc.ServerEvents(),
    ]);
    setLogs(newLogs ?? []);
    setEvents(newEvents ?? []);
  }

  useEffect(() => {
    return () => stopPolling();
  }, []);

  async function handleStart() {
    if (!root.trim()) {
      setError("Build root directory is required.");
      return;
    }
    const svc = getServerService();
    if (!svc) {
      setError("Backend unavailable.");
      return;
    }
    setError("");
    try {
      const serverAddr = await svc.ServerStart(root.trim(), listenAddr.trim());
      setAddr(serverAddr);
      setRunning(true);
      startPolling();
    } catch (err) {
      setError(String(err));
    }
  }

  async function handleStop() {
    const svc = getServerService();
    if (!svc) return;
    stopPolling();
    try {
      await svc.ServerStop();
    } catch {
      // ignore stop errors
    }
    setRunning(false);
    setAddr("");
  }

  return (
    <div className="provisioning-server">
      <section className="server-controls panel">
        <h2>Provisioning Server</h2>

        {!running ? (
          <>
            <label>
              Build root directory
              <input
                type="text"
                placeholder="e.g. /tmp/render-out"
                value={root}
                onChange={(e) => setRoot(e.target.value)}
              />
            </label>
            <label>
              Listen address
              <input
                type="text"
                value={listenAddr}
                onChange={(e) => setListenAddr(e.target.value)}
              />
            </label>
            {error && <p className="render-error">{error}</p>}
            <button
              type="button"
              className="primary-action"
              onClick={() => void handleStart()}
            >
              Start Server
            </button>
          </>
        ) : (
          <div className="server-running">
            <div className="server-status-row">
              <span className="status-dot ok" />
              <strong>Server running</strong>
            </div>
            <div className="server-url">
              <span className="panel-label">Address</span>
              <code>{addr || listenAddr}</code>
            </div>
            <div className="server-url">
              <span className="panel-label">Serving</span>
              <code>{root}</code>
            </div>
            <button
              type="button"
              className="danger-action"
              onClick={() => void handleStop()}
            >
              Stop Server
            </button>
          </div>
        )}
      </section>

      {running && (
        <>
          <section className="server-logs panel">
            <div className="logs-header">
              <strong>Request log</strong>
              <span className="log-count">{logs.length} requests</span>
            </div>
            {logs.length === 0 ? (
              <p className="library-empty">No requests yet.</p>
            ) : (
              <div className="log-list">
                {logs
                  .slice()
                  .reverse()
                  .map((l, i) => (
                    <div
                      key={i}
                      className={`log-entry ${l.StatusCode >= 400 ? "log-error" : ""}`}
                    >
                      <span className="log-method">{l.Method}</span>
                      <span className="log-path">{l.Path}</span>
                      <span className="log-status">{l.StatusCode}</span>
                      <span className="log-remote">{l.RemoteAddr}</span>
                    </div>
                  ))}
              </div>
            )}
          </section>

          {events.length > 0 && (
            <section className="server-events panel">
              <strong>Install callbacks</strong>
              <div className="log-list">
                {events.map((e, i) => (
                  <div
                    key={i}
                    className={`log-entry ${e.Status === "failure" ? "log-error" : e.Status === "success" ? "log-success" : ""}`}
                  >
                    <span className="log-method">{e.Status.toUpperCase()}</span>
                    <span className="log-path">{e.ProfileName}</span>
                    <span className="log-remote">{e.Message}</span>
                  </div>
                ))}
              </div>
            </section>
          )}
        </>
      )}
    </div>
  );
}
