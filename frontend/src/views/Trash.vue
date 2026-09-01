<template>
  <div class="trash-view">
    <header-bar showMenu>
      <title>🗑️ {{ $t("sidebar.trash") || "Papierkorb" }}</title>

      <template #actions>
        <action
          v-if="selectedIds.length > 0"
          icon="restore"
          :label="`Wiederherstellen (${selectedIds.length})`"
          @action="restoreSelected"
        />
        <action
          v-if="selectedIds.length > 0"
          icon="delete_forever"
          :label="`Endgültig löschen (${selectedIds.length})`"
          @action="deleteSelected"
        />
        <action
          v-if="items.length > 0"
          icon="delete_sweep"
          label="Papierkorb leeren"
          @action="promptEmptyTrash"
        />
        <action
          icon="folder"
          label="Dateien"
          @action="router.push('/files/')"
        />
      </template>
    </header-bar>

    <div class="trash-content-container">
      <div v-if="loading" class="trash-loading">
        <div class="spinner">
          <div class="bounce1"></div>
          <div class="bounce2"></div>
          <div class="bounce3"></div>
        </div>
      </div>

      <div v-else-if="items.length === 0" class="trash-empty-state">
        <i class="material-icons trash-empty-icon">delete_outline</i>
        <h2>Der Papierkorb ist leer</h2>
        <p>Gelöschte Dateien und Ordner werden hier zwischengespeichert und können jederzeit wiederhergestellt werden.</p>
        <button class="button button--primary" @click="router.push('/files/')">
          <i class="material-icons" style="margin-right: 4px;">folder</i>
          Zu meinen Dateien
        </button>
      </div>

      <div v-else class="trash-table-wrapper">
        <div class="trash-toolbar">
          <div class="trash-search-box">
            <i class="material-icons">search</i>
            <input
              v-model="searchQuery"
              type="text"
              class="input"
              placeholder="Papierkorb durchsuchen..."
            />
            <button v-if="searchQuery" class="clear-btn" @click="searchQuery = ''">
              <i class="material-icons">clear</i>
            </button>
          </div>

          <div class="trash-selection-info">
            <button class="button button--flat select-all-btn" @click="toggleSelectAll">
              <i class="material-icons">
                {{ isAllSelected ? 'check_box' : selectedIds.length > 0 ? 'indeterminate_check_box' : 'check_box_outline_blank' }}
              </i>
              <span>{{ selectedIds.length }} ausgewählt (von {{ filteredItems.length }})</span>
            </button>
          </div>
        </div>

        <div class="table-responsive">
          <table class="trash-table">
            <thead>
              <tr>
                <th class="th-cb">
                  <input
                    type="checkbox"
                    :checked="isAllSelected"
                    :indeterminate.prop="selectedIds.length > 0 && !isAllSelected"
                    @change="toggleSelectAll"
                  />
                </th>
                <th class="th-name">Name</th>
                <th class="th-orig">Ursprünglicher Ort</th>
                <th class="th-date">Gelöscht am</th>
                <th class="th-size">Größe</th>
                <th class="th-actions">Aktionen</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="item in filteredItems"
                :key="item.id"
                :class="{ 'is-selected': selectedIds.includes(item.id) }"
                @click="toggleSelection(item.id)"
              >
                <td class="td-cb" @click.stop>
                  <input
                    type="checkbox"
                    :value="item.id"
                    v-model="selectedIds"
                  />
                </td>
                <td class="td-name">
                  <i class="material-icons item-type-icon">
                    {{ item.isDir ? 'folder' : getFileIcon(item.name) }}
                  </i>
                  <span class="file-name" :title="item.name">{{ item.name }}</span>
                </td>
                <td class="td-orig" :title="item.originalPath">
                  {{ item.originalPath }}
                </td>
                <td class="td-date" :title="formatDateFull(item.deletedAt)">
                  {{ formatRelativeDate(item.deletedAt) }}
                </td>
                <td class="td-size">
                  {{ item.isDir ? '-' : formatBytes(item.size) }}
                </td>
                <td class="td-actions" @click.stop>
                  <button
                    class="action-icon-btn restore-btn"
                    title="Wiederherstellen"
                    @click="restoreSingle(item.id)"
                  >
                    <i class="material-icons">restore</i>
                  </button>
                  <button
                    class="action-icon-btn delete-btn"
                    title="Endgültig löschen"
                    @click="deleteSingle(item.id)"
                  >
                    <i class="material-icons">delete_forever</i>
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, inject, onMounted, ref } from "vue";
import { useRouter } from "vue-router";
import HeaderBar from "@/components/header/HeaderBar.vue";
import Action from "@/components/header/Action.vue";
import { listTrash, restoreTrash, deleteTrash, emptyTrash, type TrashItem } from "@/api/trash";
import filesize from "filesize";
import dayjs from "dayjs";
import relativeTime from "dayjs/plugin/relativeTime";

dayjs.extend(relativeTime);

const $showError = inject<IToastError>("$showError")!;
const $showSuccess = inject<IToastSuccess>("$showSuccess")!;

const router = useRouter();
const loading = ref<boolean>(true);
const items = ref<TrashItem[]>([]);
const selectedIds = ref<string[]>([]);
const searchQuery = ref<string>("");

const loadData = async () => {
  loading.value = true;
  try {
    const data = await listTrash();
    items.value = data || [];
    selectedIds.value = [];
  } catch (err: any) {
    $showError(err);
  } finally {
    loading.value = false;
  }
};

const filteredItems = computed(() => {
  if (!searchQuery.value.trim()) return items.value;
  const q = searchQuery.value.toLowerCase();
  return items.value.filter(
    (i) =>
      i.name.toLowerCase().includes(q) ||
      i.originalPath.toLowerCase().includes(q)
  );
});

const isAllSelected = computed(() => {
  return (
    filteredItems.value.length > 0 &&
    selectedIds.value.length === filteredItems.value.length
  );
});

const toggleSelectAll = () => {
  if (isAllSelected.value) {
    selectedIds.value = [];
  } else {
    selectedIds.value = filteredItems.value.map((i) => i.id);
  }
};

const toggleSelection = (id: string) => {
  const index = selectedIds.value.indexOf(id);
  if (index >= 0) {
    selectedIds.value.splice(index, 1);
  } else {
    selectedIds.value.push(id);
  }
};

const restoreSingle = async (id: string) => {
  try {
    await restoreTrash([id]);
    if ($showSuccess) $showSuccess("Datei erfolgreich wiederhergestellt");
    await loadData();
  } catch (err: any) {
    $showError(err);
  }
};

const restoreSelected = async () => {
  if (selectedIds.value.length === 0) return;
  try {
    await restoreTrash(selectedIds.value);
    if ($showSuccess) $showSuccess(`${selectedIds.value.length} Element(e) wiederhergestellt`);
    await loadData();
  } catch (err: any) {
    $showError(err);
  }
};

const deleteSingle = async (id: string) => {
  if (!confirm("Element endgültig löschen?")) return;
  try {
    await deleteTrash([id]);
    if ($showSuccess) $showSuccess("Element endgültig gelöscht");
    await loadData();
  } catch (err: any) {
    $showError(err);
  }
};

const deleteSelected = async () => {
  if (selectedIds.value.length === 0) return;
  if (!confirm(`${selectedIds.value.length} Element(e) endgültig löschen?`)) return;
  try {
    await deleteTrash(selectedIds.value);
    if ($showSuccess) $showSuccess(`${selectedIds.value.length} Element(e) endgültig gelöscht`);
    await loadData();
  } catch (err: any) {
    $showError(err);
  }
};

const promptEmptyTrash = async () => {
  if (!confirm("Möchten Sie den gesamten Papierkorb wirklich leeren? Diese Aktion kann nicht rückgängig gemacht werden.")) return;
  try {
    await emptyTrash();
    if ($showSuccess) $showSuccess("Papierkorb erfolgreich geleert");
    await loadData();
  } catch (err: any) {
    $showError(err);
  }
};

const formatBytes = (bytes: number): string => {
  if (bytes === 0) return "0 B";
  return filesize(bytes) as string;
};

const formatDateFull = (dateStr: string): string => {
  return dayjs(dateStr).format("YYYY-MM-DD HH:mm:ss");
};

const formatRelativeDate = (dateStr: string): string => {
  return dayjs(dateStr).fromNow();
};

const getFileIcon = (name: string): string => {
  const ext = name.split(".").pop()?.toLowerCase() || "";
  if (["png", "jpg", "jpeg", "gif", "webp", "svg"].includes(ext)) return "image";
  if (["mp4", "webm", "mkv", "avi"].includes(ext)) return "movie";
  if (["mp3", "flac", "wav", "ogg"].includes(ext)) return "audiotrack";
  if (["pdf"].includes(ext)) return "picture_as_pdf";
  if (["zip", "tar", "gz", "tgz"].includes(ext)) return "folder_zip";
  if (["js", "ts", "json", "html", "css", "py", "go", "php"].includes(ext)) return "code";
  return "insert_drive_file";
};

onMounted(() => {
  loadData();
});
</script>

<style scoped>
.trash-view {
  display: flex;
  flex-direction: column;
  height: 100vh;
  overflow: hidden;
}

.trash-content-container {
  flex: 1;
  overflow-y: auto;
  padding: 1.5rem 2rem;
  max-width: 1200px;
  width: 100%;
  margin: 0 auto;
}

.trash-empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
  padding: 4rem 1rem;
  color: var(--fg, currentColor);
}

.trash-empty-icon {
  font-size: 4.5rem;
  opacity: 0.3;
  margin-bottom: 1rem;
}

.trash-empty-state h2 {
  margin: 0 0 0.5rem 0;
  font-size: 1.5rem;
}

.trash-empty-state p {
  max-width: 480px;
  opacity: 0.7;
  margin-bottom: 1.5rem;
  line-height: 1.5;
}

.trash-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 1rem;
  margin-bottom: 1.25rem;
}

.trash-search-box {
  position: relative;
  display: flex;
  align-items: center;
  min-width: 280px;
}

.trash-search-box i {
  position: absolute;
  left: 0.65rem;
  opacity: 0.6;
  pointer-events: none;
}

.trash-search-box input {
  padding-left: 2.2rem;
  padding-right: 2rem;
  height: 38px;
  width: 100%;
  border-radius: 6px;
}

.clear-btn {
  position: absolute;
  right: 0.5rem;
  background: transparent;
  border: none;
  cursor: pointer;
  opacity: 0.6;
  display: flex;
  align-items: center;
}

.select-all-btn {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.4rem 0.8rem;
  border-radius: 6px;
  background: var(--surfaceSecondary, rgba(255, 255, 255, 0.05));
}

.table-responsive {
  overflow-x: auto;
  border: 1px solid var(--border, rgba(255, 255, 255, 0.1));
  border-radius: 8px;
  background: var(--surface, #1e2124);
}

.trash-table {
  width: 100%;
  border-collapse: collapse;
  text-align: left;
  font-size: 0.9rem;
}

.trash-table th {
  background: var(--surfaceSecondary, #252830);
  padding: 0.75rem 1rem;
  font-weight: 600;
  border-bottom: 1px solid var(--border, rgba(255, 255, 255, 0.1));
}

.trash-table td {
  padding: 0.65rem 1rem;
  border-bottom: 1px solid var(--border, rgba(255, 255, 255, 0.05));
  cursor: pointer;
}

.trash-table tr:hover td {
  background: var(--surfaceSecondary, rgba(255, 255, 255, 0.03));
}

.trash-table tr.is-selected td {
  background: rgba(33, 150, 243, 0.12);
}

.th-cb, .td-cb {
  width: 40px;
  text-align: center;
}

.td-name {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-weight: 500;
}

.item-type-icon {
  font-size: 1.25rem;
  color: var(--primary, #2196f3);
}

.file-name {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 250px;
}

.td-orig {
  max-width: 280px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  opacity: 0.75;
  font-family: monospace;
}

.td-date, .td-size {
  white-space: nowrap;
  opacity: 0.8;
}

.td-actions {
  white-space: nowrap;
  width: 100px;
  text-align: right;
}

.action-icon-btn {
  background: transparent;
  border: none;
  cursor: pointer;
  padding: 6px;
  border-radius: 4px;
  opacity: 0.7;
  transition: opacity 0.15s;
}

.action-icon-btn:hover {
  opacity: 1;
}

.restore-btn:hover {
  color: var(--primary, #2196f3);
  background: rgba(33, 150, 243, 0.15);
}

.delete-btn:hover {
  color: #ef4444;
  background: rgba(239, 68, 68, 0.15);
}

@media (max-width: 768px) {
  .trash-content-container {
    padding: 1rem;
  }
  .td-orig, .th-orig {
    display: none;
  }
}
</style>
