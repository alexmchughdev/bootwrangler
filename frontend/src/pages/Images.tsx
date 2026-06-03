import { useCallback, useEffect, useState } from "react";
import {
  type CacheStatus,
  type CatalogueEntry,
  imageCacheStatus,
  listImages,
} from "../api/backend";

interface VersionArchKey {
  id: string;
  version: string;
  arch: string;
}

type CacheMap = Record<string, CacheStatus>;

function cacheKey(id: string, version: string, arch: string): string {
  return `${id}/${version}/${arch}`;
}

export default function Images() {
  const [entries, setEntries] = useState<CatalogueEntry[]>([]);
  const [cacheMap, setCacheMap] = useState<CacheMap>({});
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [error, setError] = useState("");

  function collectKeys(catalogue: CatalogueEntry[]): VersionArchKey[] {
    const keys: VersionArchKey[] = [];
    for (const entry of catalogue) {
      for (const ver of entry.Versions) {
        for (const arch of ver.Architectures) {
          keys.push({ id: entry.ID, version: ver.Version, arch: arch.Arch });
        }
      }
    }
    return keys;
  }

  const fetchCacheStatuses = useCallback(
    async (catalogue: CatalogueEntry[]) => {
      const keys = collectKeys(catalogue);
      const results = await Promise.all(
        keys.map(async (k) => {
          try {
            const status = await imageCacheStatus(k.id, k.version, k.arch);
            return { key: cacheKey(k.id, k.version, k.arch), status };
          } catch {
            return {
              key: cacheKey(k.id, k.version, k.arch),
              status: { Cached: false, Path: "", Verified: false },
            };
          }
        }),
      );
      const map: CacheMap = {};
      for (const r of results) {
        map[r.key] = r.status;
      }
      setCacheMap(map);
    },
    [],
  );

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

  async function handleRefresh() {
    setRefreshing(true);
    setError("");
    try {
      const catalogue = await listImages();
      setEntries(catalogue);
      await fetchCacheStatuses(catalogue);
    } catch (err) {
      setError(String(err));
    } finally {
      setRefreshing(false);
    }
  }

  function handleDownload(name: string) {
    alert(`Download not yet implemented in GUI (${name})`);
  }

  return (
    <div className="image-catalogue">
      <div className="library-header">
        <h2>Image Catalogue</h2>
        <button
          className="secondary-action"
          disabled={loading || refreshing}
          onClick={() => void handleRefresh()}
          type="button"
        >
          {refreshing ? "Refreshing…" : "Refresh Status"}
        </button>
      </div>

      {error && <p className="render-error">{error}</p>}

      {loading ? (
        <p className="library-empty">Loading catalogue…</p>
      ) : entries.length === 0 ? (
        <p className="library-empty">No images available.</p>
      ) : (
        <div className="image-list">
          {entries.map((entry) => (
            <article className="panel image-card" key={entry.ID}>
              <div className="image-card-header">
                <div className="image-card-title">
                  <span className="panel-label">{entry.Family}</span>
                  <strong className="image-name">{entry.Name}</strong>
                  <code className="image-id">{entry.ID}</code>
                </div>
                <button
                  className="secondary-action"
                  onClick={() => handleDownload(entry.Name)}
                  type="button"
                >
                  Download
                </button>
              </div>

              <div className="image-versions">
                {entry.Versions.map((ver) => (
                  <div className="image-version-row" key={ver.Version}>
                    <span className="image-version-label">v{ver.Version}</span>
                    <div className="image-arch-list">
                      {ver.Architectures.map((archEntry) => {
                        const key = cacheKey(entry.ID, ver.Version, archEntry.Arch);
                        const status = cacheMap[key];
                        const firstImg = archEntry.Images[0];
                        return (
                          <div className="image-arch-entry" key={archEntry.Arch}>
                            <span className="image-arch-badge">{archEntry.Arch}</span>

                            {status !== undefined ? (
                              <span
                                className={
                                  status.Cached
                                    ? "cache-badge cache-badge--cached"
                                    : "cache-badge cache-badge--missing"
                                }
                              >
                                {status.Cached ? "Downloaded" : "Not cached"}
                              </span>
                            ) : (
                              <span className="cache-badge cache-badge--unknown">
                                Checking…
                              </span>
                            )}

                            {firstImg && (
                              <div className="image-compat-badges">
                                {firstImg.Compatibility.WholeDrive && (
                                  <span className="compat-badge">Whole Drive</span>
                                )}
                                {firstImg.Compatibility.Partition && (
                                  <span className="compat-badge">Partition</span>
                                )}
                                {firstImg.Compatibility.ISOFileBoot && (
                                  <span className="compat-badge">ISO File Boot</span>
                                )}
                                {firstImg.BootMode.map((mode) => (
                                  <span className="boot-badge" key={mode}>
                                    {mode.toUpperCase()}
                                  </span>
                                ))}
                              </div>
                            )}
                          </div>
                        );
                      })}
                    </div>
                  </div>
                ))}
              </div>
            </article>
          ))}
        </div>
      )}
    </div>
  );
}
