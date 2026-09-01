<template>
  <div class="card floating checksum-box">
    <div class="card-title">
      <h2>🔒 Checksums & Hashes</h2>
    </div>

    <div class="card-content">
      <p class="file-label">
        File: <strong>{{ fileName }}</strong>
      </p>

      <div v-if="loading" class="loading-state">
        <span class="spinner">⏳</span> Calculating hashes...
      </div>

      <div v-else-if="data" class="hash-list">
        <div class="hash-item">
          <div class="hash-header">
            <span class="hash-name">SHA-256</span>
            <button
              type="button"
              class="copy-btn"
              @click="copyToClipboard(data.sha256, 'sha256')"
            >
              {{ copiedField === 'sha256' ? '✓ Copied' : 'Copy' }}
            </button>
          </div>
          <code class="hash-val">{{ data.sha256 }}</code>
        </div>

        <div class="hash-item">
          <div class="hash-header">
            <span class="hash-name">MD5</span>
            <button
              type="button"
              class="copy-btn"
              @click="copyToClipboard(data.md5, 'md5')"
            >
              {{ copiedField === 'md5' ? '✓ Copied' : 'Copy' }}
            </button>
          </div>
          <code class="hash-val">{{ data.md5 }}</code>
        </div>

        <div class="hash-item">
          <div class="hash-header">
            <span class="hash-name">SHA-1</span>
            <button
              type="button"
              class="copy-btn"
              @click="copyToClipboard(data.sha1, 'sha1')"
            >
              {{ copiedField === 'sha1' ? '✓ Copied' : 'Copy' }}
            </button>
          </div>
          <code class="hash-val">{{ data.sha1 }}</code>
        </div>
      </div>

      <div v-else-if="error" class="error-state">
        {{ error }}
      </div>
    </div>

    <div class="card-action">
      <button
        type="button"
        class="button button--flat"
        @click="$emit('close')"
      >
        Close
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from "vue";
import * as checksumApi from "@/api/checksum";

interface Props {
  path: string;
  name?: string;
}

const props = defineProps<Props>();
defineEmits<{
  (e: "close"): void;
}>();

const fileName = ref(props.name || props.path.split("/").pop() || "");
const data = ref<checksumApi.ChecksumData | null>(null);
const loading = ref(true);
const error = ref("");
const copiedField = ref("");

onMounted(async () => {
  try {
    loading.value = true;
    data.value = await checksumApi.getChecksum(props.path);
  } catch (err: any) {
    error.value = err.message || "Failed to calculate checksum";
  } finally {
    loading.value = false;
  }
});

const copyToClipboard = async (text: string, field: string) => {
  try {
    await navigator.clipboard.writeText(text);
    copiedField.value = field;
    setTimeout(() => {
      if (copiedField.value === field) copiedField.value = "";
    }, 2000);
  } catch {
    // Fallback
  }
};
</script>

<style scoped>
.checksum-box {
  max-width: 520px;
  width: 90vw;
}

.file-label {
  margin-bottom: 16px;
  font-size: 0.95em;
  word-break: break-all;
}

.loading-state {
  text-align: center;
  padding: 24px;
  color: var(--textSecondary, #64748b);
}

.hash-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.hash-item {
  background: var(--surfaceSecondary, #f8fafc);
  padding: 10px 12px;
  border-radius: 8px;
  border: 1px solid var(--border-color, #e2e8f0);
}

.hash-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 6px;
}

.hash-name {
  font-weight: 700;
  font-size: 0.82em;
  text-transform: uppercase;
  color: var(--textSecondary, #64748b);
}

.copy-btn {
  background: none;
  border: 1px solid var(--border-color, #cbd5e1);
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 0.8em;
  cursor: pointer;
  font-weight: 600;
  transition: all 0.2s;
}

.copy-btn:hover {
  background: var(--surfacePrimary, #2563eb);
  color: #fff;
  border-color: var(--surfacePrimary, #2563eb);
}

.hash-val {
  display: block;
  font-family: monospace;
  font-size: 0.85em;
  word-break: break-all;
  user-select: all;
  color: var(--textPrimary, #0f172a);
}

.error-state {
  color: #ef4444;
  padding: 12px;
}
</style>
