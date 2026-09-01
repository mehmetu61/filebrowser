<template>
  <div class="card floating modal-box">
    <div class="card-title">
      <h2>🏷️ Batch Rename ({{ items.length }} Files)</h2>
    </div>

    <div class="card-content">
      <!-- Mode Selection Tabs -->
      <div class="rename-tabs">
        <button
          type="button"
          :class="['tab-btn', { active: mode === 'replace' }]"
          @click="mode = 'replace'"
        >
          Replace Text
        </button>
        <button
          type="button"
          :class="['tab-btn', { active: mode === 'affix' }]"
          @click="mode = 'affix'"
        >
          Prefix / Suffix
        </button>
        <button
          type="button"
          :class="['tab-btn', { active: mode === 'sequence' }]"
          @click="mode = 'sequence'"
        >
          Numbering
        </button>
      </div>

      <!-- Replace Form -->
      <div v-if="mode === 'replace'" class="tab-content">
        <div class="input-group">
          <label>Find Text:</label>
          <input
            v-model="findText"
            type="text"
            class="input input--block"
            placeholder="Text to find..."
          />
        </div>
        <div class="input-group">
          <label>Replace With:</label>
          <input
            v-model="replaceText"
            type="text"
            class="input input--block"
            placeholder="Replacement text..."
          />
        </div>
      </div>

      <!-- Affix Form -->
      <div v-else-if="mode === 'affix'" class="tab-content">
        <div class="input-group">
          <label>Add Prefix:</label>
          <input
            v-model="prefixText"
            type="text"
            class="input input--block"
            placeholder="e.g. 2026_"
          />
        </div>
        <div class="input-group">
          <label>Add Suffix:</label>
          <input
            v-model="suffixText"
            type="text"
            class="input input--block"
            placeholder="e.g. _final"
          />
        </div>
      </div>

      <!-- Sequence Form -->
      <div v-else-if="mode === 'sequence'" class="tab-content">
        <div class="input-group">
          <label>Pattern (use {n} for index):</label>
          <input
            v-model="sequencePattern"
            type="text"
            class="input input--block"
            placeholder="e.g. Vacation_{n}"
          />
        </div>
        <div class="input-row">
          <div class="input-group">
            <label>Start Number:</label>
            <input
              v-model.number="startNumber"
              type="number"
              min="0"
              class="input input--block"
            />
          </div>
          <div class="input-group">
            <label>Padding Digits:</label>
            <input
              v-model.number="paddingDigits"
              type="number"
              min="1"
              max="6"
              class="input input--block"
            />
          </div>
        </div>
      </div>

      <!-- Preview Table -->
      <div class="preview-section">
        <h3>Preview:</h3>
        <div class="preview-table-wrapper">
          <table class="preview-table">
            <thead>
              <tr>
                <th>Original</th>
                <th>→</th>
                <th>New Name</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="(item, idx) in previewItems" :key="idx">
                <td class="orig-name">{{ item.originalName }}</td>
                <td class="arrow">→</td>
                <td :class="['new-name', { changed: item.originalName !== item.newName }]">
                  {{ item.newName }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <div class="card-action">
      <button
        type="button"
        class="button button--flat"
        @click="$emit('close')"
      >
        Cancel
      </button>
      <button
        type="button"
        class="button button--primary"
        :disabled="loading || !hasChanges"
        @click="executeRename"
      >
        {{ loading ? "Renaming..." : "Apply Rename" }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";
import * as batchApi from "@/api/batch";

interface Props {
  items: Array<{ name: string; path: string; isDir?: boolean }>;
  currentPath: string;
}

const props = defineProps<Props>();
const emit = defineEmits<{
  (e: "close"): void;
  (e: "done"): void;
}>();

const mode = ref<"replace" | "affix" | "sequence">("replace");
const findText = ref("");
const replaceText = ref("");
const prefixText = ref("");
const suffixText = ref("");
const sequencePattern = ref("File_{n}");
const startNumber = ref(1);
const paddingDigits = ref(3);
const loading = ref(false);

const splitExt = (filename: string, isDir?: boolean) => {
  if (isDir) return { base: filename, ext: "" };
  const lastDot = filename.lastIndexOf(".");
  if (lastDot <= 0) return { base: filename, ext: "" };
  return {
    base: filename.substring(0, lastDot),
    ext: filename.substring(lastDot),
  };
};

const previewItems = computed(() => {
  return props.items.map((item, index) => {
    const { base, ext } = splitExt(item.name, item.isDir);
    let newBase = base;

    if (mode.value === "replace") {
      if (findText.value) {
        newBase = base.split(findText.value).join(replaceText.value);
      }
    } else if (mode.value === "affix") {
      newBase = `${prefixText.value}${base}${suffixText.value}`;
    } else if (mode.value === "sequence") {
      const num = startNumber.value + index;
      const padded = String(num).padStart(paddingDigits.value, "0");
      newBase = sequencePattern.value.replace(/\{n\}/g, padded);
    }

    const newName = `${newBase}${ext}`;
    const basePath = item.path.substring(0, item.path.lastIndexOf("/"));
    const newPath = `${basePath}/${newName}`;

    return {
      from: item.path,
      to: newPath,
      originalName: item.name,
      newName,
    };
  });
});

const hasChanges = computed(() => {
  return previewItems.value.some((i) => i.originalName !== i.newName);
});

const executeRename = async () => {
  loading.value = true;
  try {
    const payload = previewItems.value
      .filter((i) => i.originalName !== i.newName)
      .map((i) => ({ from: i.from, to: i.to }));

    if (payload.length > 0) {
      await batchApi.batchRename(payload);
    }
    emit("done");
    emit("close");
  } catch (err) {
    console.error("Batch rename failed", err);
  } finally {
    loading.value = false;
  }
};
</script>

<style scoped>
.modal-box {
  max-width: 650px;
  width: 90vw;
  max-height: 85vh;
  display: flex;
  flex-direction: column;
}

.rename-tabs {
  display: flex;
  gap: 8px;
  margin-bottom: 16px;
  border-bottom: 1px solid var(--border-color, #e2e8f0);
  padding-bottom: 8px;
}

.tab-btn {
  background: none;
  border: none;
  padding: 8px 16px;
  font-weight: 600;
  cursor: pointer;
  border-radius: 6px;
  color: var(--textSecondary, #64748b);
  transition: all 0.2s;
}

.tab-btn.active {
  background: var(--surfacePrimary, #2563eb);
  color: #fff;
}

.input-group {
  margin-bottom: 12px;
}

.input-group label {
  display: block;
  font-size: 0.85em;
  font-weight: 600;
  margin-bottom: 4px;
  color: var(--textPrimary, #1e293b);
}

.input-row {
  display: flex;
  gap: 16px;
}

.preview-section {
  margin-top: 16px;
  border-top: 1px solid var(--border-color, #e2e8f0);
  padding-top: 12px;
}

.preview-section h3 {
  font-size: 0.95em;
  margin-bottom: 8px;
}

.preview-table-wrapper {
  max-height: 220px;
  overflow-y: auto;
  border: 1px solid var(--border-color, #e2e8f0);
  border-radius: 6px;
}

.preview-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.88em;
}

.preview-table th,
.preview-table td {
  padding: 6px 10px;
  text-align: left;
  border-bottom: 1px solid var(--border-color, #e2e8f0);
}

.preview-table th {
  background: var(--surfaceSecondary, #f8fafc);
  font-weight: 600;
  position: sticky;
  top: 0;
}

.arrow {
  color: var(--textSecondary, #94a3b8);
  width: 20px;
  text-align: center;
}

.new-name.changed {
  color: #10b981;
  font-weight: 600;
}
</style>
