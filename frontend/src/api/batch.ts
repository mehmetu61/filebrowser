import { fetchJSON } from "./utils";

export interface BatchRenameItem {
  from: string;
  to: string;
}

export interface BatchRenameResult {
  from: string;
  to: string;
  error?: string;
}

export function batchRename(items: BatchRenameItem[]) {
  return fetchJSON<{ success: boolean; results: BatchRenameResult[] }>(
    "/api/resources/batch-rename",
    {
      method: "POST",
      body: JSON.stringify({ items }),
    }
  );
}
