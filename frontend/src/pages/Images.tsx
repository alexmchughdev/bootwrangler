import { useCallback, useEffect, useState } from "react";
import {
  type CacheStatus,
  type CatalogueEntry,
  type CustomImage,
  customImagesPath,
  downloadImage,
  imageCacheStatus,
  listCustomImages,
  listImages,
} from "../api/backend";

interface VersionArchKey {
  id: string;
  version: string;
  arch: string;
}

type CacheMap = Record<string, CacheStatus>;

interface CacheStatusPresentation {
  className: string;
  label: string;
  showDownload: boolean;
  downloadLabel: string;
  downloadingLabel: string;
}

function cacheKey(id: string, version: string, arch: string): string {
  return `${id}/${version}/${arch}`;
}

function cacheStatusPresentation(status: CacheStatus): CacheStatusPresentation {
  if (!status.Cached) {
    return {
      className: "cache-badge cache-badge--missing",
      label: "Not downloaded",
      showDownload: true,
      downloadLabel: "Download",
      downloadingLabel: "Downloading…",
    };
  }

  if (!status.Verified) {
    return {
      className: "cache-badge cache-badge--unverified",
      label: "Downloaded, not verified",
      showDownload: true,
      downloadLabel: "Download & verify",
      downloadingLabel: "Verifying…",
    };
  }

  return {
    className: "cache-badge cache-badge--cached",
    label: "Verified",
    showDownload: false,
    downloadLabel: "",
    downloadingLabel: "",
  };
}

export default function Images() {
  const [entries, setEntries] = useState<CatalogueEntry[]>([]);
  const [customImages, setCustomImages] = useState<CustomImage[]>([]);
  const [customPath, setCustomPath] = useState("");
  const [cacheMap, setCacheMap] = useState<CacheMap>({});
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [error, setError] = useState("");
  const [downloading, setDownloading] = useState<Record<string, boolean>>({});
  const [downloadErrors, setDownloadErrors] = useState<Record<string, string>>({});

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
        const [catalogue, custom, path] = await Promise.all([
          listImages(),
          listCustomImages(),
          customImagesPath(),
        ]);
        setEntries(catalogue);
        setCustomImages(custom);
        setCustomPath(path);
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
      const [catalogue, custom, path] = await Promise.all([
        listImages(),
        listCustomImages(),
        customImagesPath(),
      ]);
      setEntries(catalogue);
      setCustomImages(custom);
      setCustomPath(path);
      await fetchCacheStatuses(catalogue);
    } catch (err) {
      setError(String(err));
    } finally {
      setRefreshing(false);
    }
  }

  async function handleDownload(id: string, version: string, arch: string) {
    const key = cacheKey(id, version, arch);
    setDownloading((prev) => ({ ...prev, [key]: true }));
    setDownloadErrors((prev) => { const n = { ...prev }; delete n[key]; return n; });
    try {
      await downloadImage(id, version, arch);
      // Refresh cache status for this entry
      const status = await imageCacheStatus(id, version, arch);
      setCacheMap((prev) => ({ ...prev, [key]: status }));
    } catch (err) {
      setDownloadErrors((prev) => ({ ...prev, [key]: String(err) }));
    } finally {
      setDownloading((prev) => { const n = { ...prev }; delete n[key]; return n; });
    }
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
      ) : entries.length === 0 && customImages.length === 0 ? (
        <p className="library-empty">No images available.</p>
      ) : (
        <>
          <div className="image-list">
            {entries.map((entry) => (
              <article className="panel image-card" key={entry.ID}>
                <div className="image-card-header">
                  <div className="image-card-title">
                    <span className="panel-label">{entry.Family}</span>
                    <strong className="image-name">{entry.Name}</strong>
                    <code className="image-id">{entry.ID}</code>
                  </div>
                </div>

                <div className="image-versions">
                  {entry.Versions.map((ver) => (
                    <div className="image-version-row" key={ver.Version}>
                      <span className="image-version-label">v{ver.Version}</span>
                      <div className="image-arch-list">
                        {ver.Architectures.map((archEntry) => {
                          const key = cacheKey(entry.ID, ver.Version, archEntry.Arch);
                          const status = cacheMap[key];
                          const statusPresentation =
                            status !== undefined ? cacheStatusPresentation(status) : undefined;
                          const firstImg = archEntry.Images[0];
                          const isDownloading = !!downloading[key];
                          const dlError = downloadErrors[key];
                          return (
                            <div className="image-arch-entry" key={archEntry.Arch}>
                              <span className="image-arch-badge">{archEntry.Arch}</span>

                              {statusPresentation !== undefined ? (
                                <span className={statusPresentation.className}>
                                  {statusPresentation.label}
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

                              {statusPresentation?.showDownload && (
                                <button
                                  className="secondary-action image-download-btn"
                                  disabled={isDownloading}
                                  onClick={() =>
                                    void handleDownload(entry.ID, ver.Version, archEntry.Arch)
                                  }
                                  type="button"
                                >
                                  {isDownloading
                                    ? statusPresentation.downloadingLabel
                                    : statusPresentation.downloadLabel}
                                </button>
                              )}

                              {dlError && (
                                <span className="image-dl-error" title={dlError}>
                                  Failed
                                </span>
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

          <section className="panel custom-image-panel">
            <div className="image-card-header">
              <div className="image-card-title">
                <span className="panel-label">custom</span>
                <strong className="image-name">Custom Images</strong>
                {customPath && <code className="image-id">{customPath}</code>}
              </div>
            </div>

            {customImages.length === 0 ? (
              <p className="library-empty">No custom images configured.</p>
            ) : (
              <div className="custom-image-list">
                {customImages.map((img) => (
                  <div className="custom-image-entry" key={img.ID}>
                    <div className="image-card-title">
                      <strong className="image-name">{img.Name}</strong>
                      <code className="image-id">{img.ID}</code>
                    </div>
                    <span className="cache-badge cache-badge--cached">
                      {img.Source.Type}
                    </span>
                    <div className="image-compat-badges">
                      {img.Compatibility.WholeDrive && (
                        <span className="compat-badge">Whole Drive</span>
                      )}
                      {img.Compatibility.Partition && (
                        <span className="compat-badge">Partition</span>
                      )}
                      {img.Compatibility.ISOFileBoot && (
                        <span className="compat-badge">ISO File Boot</span>
                      )}
                    </div>
                    <code className="custom-image-source">
                      {img.Source.Type === "local-file" ? img.Source.Path : img.Source.URL}
                    </code>
                  </div>
                ))}
              </div>
            )}
          </section>
        </>
      )}
    </div>
  );
}
