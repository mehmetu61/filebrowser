import { fetchJSON, removePrefix } from "./utils";

export interface ChecksumData {
  path: string;
  size: number;
  md5: string;
  sha1: string;
  sha256: string;
}

export function getChecksum(url: string, signal?: AbortSignal) {
  url = removePrefix(url);
  return fetchJSON<ChecksumData>(`/api/checksum${url}`, { signal });
}
