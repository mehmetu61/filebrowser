import { fetchJSON, fetchURL } from "./utils";

export interface TrashItem {
  id: string;
  name: string;
  originalPath: string;
  deletedAt: string;
  size: number;
  isDir: boolean;
  trashPath: string;
}

export function listTrash() {
  return fetchJSON<TrashItem[]>("/api/trash");
}

export function restoreTrash(ids: string[]) {
  return fetchJSON<{ restored: string[] }>("/api/trash/restore", {
    method: "POST",
    body: JSON.stringify({ ids }),
  });
}

export function deleteTrash(ids: string[]) {
  return fetchURL("/api/trash", {
    method: "DELETE",
    body: JSON.stringify({ ids }),
  });
}

export function emptyTrash() {
  return fetchURL("/api/trash?all=true", {
    method: "DELETE",
  });
}
