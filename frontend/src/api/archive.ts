import { fetchJSON } from "./utils";

function cleanPath(p: string): string {
  if (!p) return "/";
  if (p.startsWith("/files/")) return p.substring(6);
  if (p === "/files") return "/";
  if (!p.startsWith("/")) return "/" + p;
  return p;
}

export function extractArchive(source: string, destination?: string) {
  return fetchJSON<{ success: boolean; path: string }>("/api/archive/extract", {
    method: "POST",
    body: JSON.stringify({
      source: cleanPath(source),
      destination: destination ? cleanPath(destination) : "",
    }),
  });
}

export function compressItems(
  items: string[],
  destination: string,
  archiveName?: string
) {
  return fetchJSON<{ success: boolean; path: string }>("/api/archive/compress", {
    method: "POST",
    body: JSON.stringify({
      items: items.map(cleanPath),
      destination: cleanPath(destination),
      archiveName: archiveName || "archive.zip",
    }),
  });
}

export interface ArchiveEntry {
  name: string;
  path: string;
  size: number;
  compressedSize: number;
  isDir: boolean;
  modified: string;
}

export function listArchiveEntries(path: string) {
  const cleaned = cleanPath(path);
  return fetchJSON<ArchiveEntry[]>(`/api/archive/list?path=${encodeURIComponent(cleaned)}`);
}
