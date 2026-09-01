<template>
  <div class="card floating archive-viewer-modal">
    <div class="card-title archive-viewer-header">
      <div class="title-with-icon">
        <i class="material-icons">folder_zip</i>
        <h2>{{ fileName }}</h2>
      </div>
      <button class="close-btn" @click="$emit('close')">
        <i class="material-icons">close</i>
      </button>
    </div>

    <div class="card-content archive-viewer-body">
      <div class="archive-stats-bar" v-if="!loading && !error">
        <div class="stat-item">
          <span class="stat-label">Total Files:</span>
          <span class="stat-value">{{ entries.length }}</span>
        </div>
        <div class="stat-item">
          <span class="stat-label">Uncompressed Size:</span>
          <span class="stat-value">{{ formatBytes(totalSize) }}</span>
        </div>
        <div class="stat-item" v-if="totalCompressedSize > 0">
          <span class="stat-label">Compressed Size:</span>
          <span class="stat-value">{{ formatBytes(totalCompressedSize) }}</span>
        </div>
      </div>

      <div class="search-filter-wrapper" v-if="!loading && !error && entries.length > 0">
        <i class="material-icons search-icon">search</i>
        <input
          v-model="searchQuery"
          type="text"
          class="input search-filter-input"
          placeholder="Search files inside archive..."
        />
        <button v-if="searchQuery" class="clear-search-btn" @click="searchQuery = ''">
          <i class="material-icons">clear</i>
        </button>
      </div>

      <div v-if="loading" class="archive-loading">
        <i class="material-icons spin-icon">sync</i>
        <p>Reading archive contents...</p>
      </div>

      <div v-else-if="error" class="archive-error">
        <i class="material-icons">error_outline</i>
        <p>{{ error }}</p>
      </div>

      <div v-else-if="filteredEntries.length === 0" class="archive-empty">
        <p>No matching files found inside archive.</p>
      </div>

      <div v-else class="archive-table-container">
        <table class="archive-table">
          <thead>
            <tr>
              <th>Name</th>
              <th class="th-size">Size</th>
              <th class="th-modified">Modified</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="entry in filteredEntries"
              :key="entry.path"
              :class="{ 'is-dir': entry.isDir }"
            >
              <td class="td-name">
                <i class="material-icons item-icon">
                  {{ entry.isDir ? 'folder' : getFileIcon(entry.name) }}
                </i>
                <span class="entry-path" :title="entry.path">{{ entry.path }}</span>
              </td>
              <td class="td-size">
                {{ entry.isDir ? '-' : formatBytes(entry.size) }}
              </td>
              <td class="td-modified">
                {{ entry.modified ? formatDate(entry.modified) : '-' }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <div class="card-action archive-viewer-actions">
      <button type="button" class="button button--flat" @click="$emit('close')">
        Close
      </button>
      <button
        v-if="canExtract"
        type="button"
        class="button button--primary"
        @click="$emit('extract', props.path)"
      >
        <i class="material-icons" style="font-size: 1.1em; margin-right: 4px;">unarchive</i>
        Extract Archive
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { listArchiveEntries, type ArchiveEntry } from "@/api/archive";
import { filesize } from "filesize";
import dayjs from "dayjs";

interface Props {
  path: string;
  name?: string;
  canExtract?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  canExtract: true,
});

const emit = defineEmits<{
  (e: "close"): void;
  (e: "extract", path: string): void;
}>();

const loading = ref<boolean>(true);
const error = ref<string>("");
const entries = ref<ArchiveEntry[]>([]);
const searchQuery = ref<string>("");

const fileName = computed(() => {
  if (props.name) return props.name;
  const parts = props.path.split("/");
  return parts[parts.length - 1] || props.path;
});

const totalSize = computed(() => {
  return entries.value.reduce((acc, curr) => acc + (curr.isDir ? 0 : curr.size), 0);
});

const totalCompressedSize = computed(() => {
  return entries.value.reduce((acc, curr) => acc + (curr.isDir ? 0 : curr.compressedSize), 0);
});

const filteredEntries = computed(() => {
  if (!searchQuery.value.trim()) return entries.value;
  const q = searchQuery.value.toLowerCase();
  return entries.value.filter((e) => e.path.toLowerCase().includes(q));
});

const formatBytes = (bytes: number): string => {
  if (bytes === 0) return "0 B";
  return filesize(bytes) as string;
};

const formatDate = (dateStr: string): string => {
  return dayjs(dateStr).format("YYYY-MM-DD HH:mm");
};

const getFileIcon = (name: string): string => {
  const ext = name.split(".").pop()?.toLowerCase() || "";
  if (["png", "jpg", "jpeg", "gif", "webp", "svg", "bmp"].includes(ext)) return "image";
  if (["mp4", "webm", "mkv", "avi", "mov"].includes(ext)) return "movie";
  if (["mp3", "flac", "wav", "ogg", "aac"].includes(ext)) return "audiotrack";
  if (["pdf"].includes(ext)) return "picture_as_pdf";
  if (["zip", "tar", "gz", "tgz", "rar", "7z"].includes(ext)) return "folder_zip";
  if (["js", "ts", "json", "html", "css", "vue", "py", "go", "php", "c", "cpp", "java", "rs", "sql", "sh", "yaml", "yml"].includes(ext)) return "code";
  if (["txt", "md", "log", "ini", "conf"].includes(ext)) return "description";
  return "insert_drive_file";
};

onMounted(async () => {
  loading.value = true;
  error.value = "";
  try {
    const data = await listArchiveEntries(props.path);
    entries.value = data || [];
  } catch (err: any) {
    error.value = err.message || "Failed to inspect archive";
  } finally {
    loading.value = false;
  }
});
</script>

<style scoped>
.archive-viewer-modal {
  width: 90vw;
  max-width: 820px;
  max-height: 85vh;
  display: flex;
  flex-direction: column;
  padding: 0;
  overflow: hidden;
}

.archive-viewer-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem 1.25rem;
  border-bottom: 1px solid var(--border, rgba(255, 255, 255, 0.1));
}

.title-with-icon {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  min-width: 0;
}

.title-with-icon h2 {
  margin: 0;
  font-size: 1.2rem;
  font-weight: 600;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.title-with-icon i {
  font-size: 1.5rem;
  color: var(--primary, #2196f3);
}

.close-btn {
  background: transparent;
  border: none;
  cursor: pointer;
  color: var(--fg, currentColor);
  opacity: 0.7;
  padding: 4px;
  display: flex;
  align-items: center;
  border-radius: 4px;
}

.close-btn:hover {
  opacity: 1;
  background: var(--surfaceSecondary, rgba(255, 255, 255, 0.1));
}

.archive-viewer-body {
  padding: 1rem 1.25rem;
  overflow-y: auto;
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.archive-stats-bar {
  display: flex;
  flex-wrap: wrap;
  gap: 1.25rem;
  background: var(--surfaceSecondary, rgba(255, 255, 255, 0.05));
  padding: 0.6rem 1rem;
  border-radius: 6px;
  font-size: 0.9rem;
}

.stat-item {
  display: flex;
  gap: 0.35rem;
}

.stat-label {
  opacity: 0.7;
}

.stat-value {
  font-weight: 600;
}

.search-filter-wrapper {
  position: relative;
  display: flex;
  align-items: center;
}

.search-icon {
  position: absolute;
  left: 0.65rem;
  opacity: 0.6;
  font-size: 1.2rem;
  pointer-events: none;
}

.search-filter-input {
  width: 100%;
  padding-left: 2.2rem;
  padding-right: 2rem;
  height: 36px;
  border-radius: 6px;
}

.clear-search-btn {
  position: absolute;
  right: 0.5rem;
  background: transparent;
  border: none;
  cursor: pointer;
  opacity: 0.6;
  display: flex;
  align-items: center;
}

.clear-search-btn:hover {
  opacity: 1;
}

.archive-loading,
.archive-error,
.archive-empty {
  text-align: center;
  padding: 3rem 1rem;
  opacity: 0.8;
}

.spin-icon {
  animation: spin 1s linear infinite;
  font-size: 2rem;
  margin-bottom: 0.5rem;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.archive-table-container {
  overflow-x: auto;
  max-height: 45vh;
  border: 1px solid var(--border, rgba(255, 255, 255, 0.08));
  border-radius: 6px;
}

.archive-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.875rem;
  text-align: left;
}

.archive-table th {
  position: sticky;
  top: 0;
  background: var(--surfaceSecondary, #252830);
  padding: 0.5rem 0.75rem;
  font-weight: 600;
  border-bottom: 1px solid var(--border, rgba(255, 255, 255, 0.1));
}

.archive-table td {
  padding: 0.45rem 0.75rem;
  border-bottom: 1px solid var(--border, rgba(255, 255, 255, 0.05));
}

.archive-table tr:hover td {
  background: var(--surfaceSecondary, rgba(255, 255, 255, 0.03));
}

.td-name {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  max-width: 450px;
}

.item-icon {
  font-size: 1.1rem;
  opacity: 0.75;
}

.entry-path {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  font-family: monospace;
}

.th-size, .td-size {
  width: 100px;
  text-align: right;
  white-space: nowrap;
}

.th-modified, .td-modified {
  width: 150px;
  text-align: right;
  white-space: nowrap;
  opacity: 0.75;
}

.archive-viewer-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
  padding: 0.75rem 1.25rem;
  border-top: 1px solid var(--border, rgba(255, 255, 255, 0.1));
}
</style>
