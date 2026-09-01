<template>
  <div class="card floating extract-box">
    <div class="card-title">
      <h2>📦 Extract Archive</h2>
    </div>

    <div class="card-content">
      <p class="archive-name">
        Extracting: <strong>{{ fileName }}</strong>
      </p>

      <div class="input-group">
        <label>Extract to directory:</label>
        <input
          v-model="destination"
          type="text"
          class="input input--block"
          placeholder="e.g. /my-folder (leave empty for current folder)"
        />
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
        :disabled="loading"
        @click="executeExtract"
      >
        {{ loading ? "Extracting..." : "Extract Here" }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from "vue";
import * as archiveApi from "@/api/archive";

interface Props {
  path: string;
  name?: string;
  currentDir?: string;
}

const props = defineProps<Props>();
const emit = defineEmits<{
  (e: "close"): void;
  (e: "done"): void;
}>();

const fileName = ref(props.name || props.path.split("/").pop() || "");
const destination = ref(props.currentDir || "");
const loading = ref(false);
const error = ref("");

const executeExtract = async () => {
  loading.value = true;
  error.value = "";
  try {
    await archiveApi.extractArchive(props.path, destination.value);
    emit("done");
    emit("close");
  } catch (err: any) {
    error.value = err.message || "Failed to extract archive";
  } finally {
    loading.value = false;
  }
};
</script>

<style scoped>
.extract-box {
  max-width: 480px;
  width: 90vw;
}

.archive-name {
  margin-bottom: 16px;
  font-size: 0.95em;
  word-break: break-all;
}

.input-group {
  margin-bottom: 12px;
}

.input-group label {
  display: block;
  font-size: 0.85em;
  font-weight: 600;
  margin-bottom: 6px;
}

.error-msg {
  color: #ef4444;
  margin-top: 8px;
  font-size: 0.9em;
}
</style>
