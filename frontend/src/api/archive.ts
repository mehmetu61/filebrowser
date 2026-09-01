import { fetchJSON } from "./utils";

export function extractArchive(source: string, destination?: string) {
  return fetchJSON<{ success: boolean; path: string }>("/api/archive/extract", {
    method: "POST",
    body: JSON.stringify({ source, destination }),
  });
}

export function compressItems(
  items: string[],
  destination: string,
  archiveName?: string
) {
  return fetchJSON<{ success: boolean; path: string }>("/api/archive/compress", {
    method: "POST",
    body: JSON.stringify({ items, destination, archiveName }),
  });
}
