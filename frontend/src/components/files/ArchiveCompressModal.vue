<template>
  <div class="card floating compress-box">
    <div class="card-title">
      <h2>🗜️ Create ZIP Archive ({{ items.length }} Items)</h2>
    </div>

    <div class="card-content">
      <div class="input-group">
        <label>Archive Name:</label>
        <input
          v-model="archiveName"
          type="text"
          class="input input--block"
          placeholder="e.g. archive.zip"
          :disabled="loading"
          autofocus
        />
      </div>

      <div class="items-summary">
        <p class="summary-label">Items to compress:</p>
        <div class="items-list">
          <span v-for="(item, idx) in items" :key="idx" class="item-chip">
            {{ item.name }}
          </span>
        </div>
      </div>

      <div v-if="loading" class="compressing-state">
        <div class="progress-bar-container">
          <div class="progress-bar-fill"></div>
        </div>
        <p class="status-text">⏳ Compressing files into {{ archiveName }}...</p>
      </div>

      <div v-if="error" class="error-msg">
        {{ error }}
      </div>
    </div>

    <div class="card-action">
      <button
        type="button"
        class="button button--flat"
        :disabled="loading"
        @click="$emit('close')"
      >
        Cancel
      </button>
      <button
        type="button"
        class="button button--primary"
        :disabled="loading || !archiveName.trim()"
        @click="executeCompress"
      >
        {{ loading ? "Compressing..." : "Create ZIP" }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from "vue";
import * as archiveApi from "@/api/archive";

interface Props {
  items: Array<{ name: string; path: string; isDir?: boolean }>;
  currentDir: string;
}

const props = defineProps<Props>();
const emit = defineEmits<{
  (e: "close"): void;
  (e: "done"): void;
}>();

const archiveName = ref("archive.zip");
const loading = ref(false);
const error = ref("");

const executeCompress = async () => {
  loading.value = true;
  error.value = "";

  let name = archiveName.value.trim();
  if (!name.toLowerCase().endsWith(".zip")) {
    name += ".zip";
  }

  try {
    const paths = props.items.map((i) => i.path);
    await archiveApi.compressItems(paths, props.currentDir, name);
    emit("done");
    emit("close");
  } catch (err: any) {
    error.value = err.message || "Failed to create archive";
  } finally {
    loading.value = false;
  }
};
</script>

<style scoped>
.compress-box {
  max-width: 500px;
  width: 90vw;
}

.input-group {
  margin-bottom: 14px;
}

.input-group label {
  display: block;
  font-size: 0.85em;
  font-weight: 600;
  margin-bottom: 6px;
}

.items-summary {
  margin-bottom: 16px;
}

.summary-label {
  font-size: 0.85em;
  color: var(--textSecondary, #64748b);
  margin-bottom: 6px;
}

.items-list {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  max-height: 120px;
  overflow-y: auto;
  padding: 6px;
  background: var(--surfaceSecondary, #f8fafc);
  border-radius: 6px;
  border: 1px solid var(--border-color, #e2e8f0);
}

.item-chip {
  background: var(--surfacePrimary, #fff);
  border: 1px solid var(--border-color, #cbd5e1);
  padding: 2px 8px;
  border-radius: 12px;
  font-size: 0.8em;
  font-weight: 500;
}

.compressing-state {
  margin: 16px 0;
  text-align: center;
}

.progress-bar-container {
  width: 100%;
  height: 6px;
  background: var(--surfaceSecondary, #e2e8f0);
  border-radius: 3px;
  overflow: hidden;
  margin-bottom: 8px;
}

.progress-bar-fill {
  width: 100%;
  height: 100%;
  background: var(--surfacePrimary, #2563eb);
  animation: indeterminate 1.5s infinite linear;
  transform-origin: 0% 50%;
}

@keyframes indeterminate {
  0% { transform: translateX(0) scaleX(0); }
  40% { transform: translateX(0) scaleX(0.4); }
  100% { transform: translateX(100%) scaleX(0.5); }
}

.status-text {
  font-size: 0.85em;
  color: var(--textSecondary, #64748b);
}

.error-msg {
  color: #ef4444;
  margin-top: 8px;
  font-size: 0.9em;
}
</style>
