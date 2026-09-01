import { fetchJSON, removePrefix } from "./utils";

export interface ChecksumData {
  path: string;
  size: number;
  md5: string;
  sha1: string;
  sha256: string;
}

export function getChecksum(url: string, signal?: AbortSignal) {
  if (url.startsWith("/files/")) {
    url = url.substring(6);
  } else if (url === "/files") {
    url = "/";
  }
  if (!url.startsWith("/")) {
    url = "/" + url;
  }
  return fetchJSON<ChecksumData>(`/api/checksum${url}`, { signal });
}
