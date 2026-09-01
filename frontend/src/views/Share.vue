<template>
  <div class="share-page">
    <header-bar showMenu showLogo>
      <title />

      <action
        v-if="fileStore.selectedCount"
        icon="file_download"
        :label="t('buttons.download')"
        @action="download"
        :counter="fileStore.selectedCount"
      />
      <button
        v-if="!req?.isDir || isSingleFile()"
        class="action copy-clipboard"
        :aria-label="t('buttons.copyDownloadLinkToClipboard')"
        :data-title="t('buttons.copyDownloadLinkToClipboard')"
        @click="copyToClipboard(linkSelected() || link)"
      >
        <i class="material-icons">content_paste</i>
      </button>
      <action
        v-if="req?.isDir"
        icon="check_circle"
        :label="t('buttons.selectMultiple')"
        @action="toggleMultipleSelection"
      />
    </header-bar>

    <breadcrumbs :base="'/share/' + hash" />

    <div v-if="layoutStore.loading" class="share-loading">
      <h2 class="message delayed">
        <div class="spinner">
          <div class="bounce1"></div>
          <div class="bounce2"></div>
          <div class="bounce3"></div>
        </div>
        <span>{{ t("files.loading") }}</span>
      </h2>
    </div>

    <div v-else-if="error" class="share-error">
      <div v-if="error.status === 401">
        <div class="card floating" id="password">
          <div v-if="attemptedPasswordLogin" class="share__wrong__password">
            {{ t("login.wrongCredentials") }}
          </div>
          <div class="card-title">
            <h2>{{ t("login.password") }}</h2>
          </div>

          <div class="card-content">
            <input
              v-focus
              class="input input--block"
              type="password"
              :placeholder="t('login.password')"
              v-model="password"
              @keyup.enter="fetchData"
            />
          </div>
          <div class="card-action">
            <button
              class="button button--primary"
              @click="fetchData"
              :aria-label="t('buttons.submit')"
              :data-title="t('buttons.submit')"
            >
              {{ t("buttons.submit") }}
            </button>
          </div>
        </div>
        <div class="overlay" />
      </div>
      <errors v-else :errorCode="error.status" />
    </div>

    <div v-else-if="req !== null" class="share-content">
      <!-- SINGLE FILE SHARE VIEW -->
      <div v-if="!req.isDir" class="single-file-share">
        <div class="single-file-card">
          <!-- Thumbnail / Preview Area -->
          <div class="single-file-preview">
            <img
              v-if="req.type === 'image'"
              :src="raw"
              :alt="req.name"
              class="single-file-image"
            />
            <video
              v-else-if="req.type === 'video'"
              :src="raw"
              controls
              class="single-file-video"
            ></video>
            <audio
              v-else-if="req.type === 'audio'"
              :src="raw"
              controls
              class="single-file-audio"
            ></audio>
            <div v-else class="single-file-icon">
              <i class="material-icons">{{ icon }}</i>
            </div>
          </div>

          <!-- File Info & Actions -->
          <div class="single-file-details">
            <h1 class="single-file-title" :title="req.name">{{ req.name }}</h1>
            <div class="single-file-meta">
              <span class="meta-item">
                <i class="material-icons">data_usage</i>
                {{ humanSize }}
              </span>
              <span class="meta-item" :title="modTime">
                <i class="material-icons">schedule</i>
                {{ humanTime }}
              </span>
            </div>

            <div class="single-file-actions">
              <a :href="link" class="button button--primary download-btn" download>
                <i class="material-icons">file_download</i>
                <span>{{ t("buttons.download") }}</span>
              </a>
              <a
                v-if="inlineLink"
                :href="inlineLink"
                target="_blank"
                class="button button--flat open-btn"
              >
                <i class="material-icons">open_in_new</i>
                <span>{{ t("buttons.openFile") }}</span>
              </a>
              <button
                type="button"
                class="button button--flat copy-btn"
                @click="copyToClipboard(link)"
              >
                <i class="material-icons">link</i>
                <span>Copy Link</span>
              </button>
            </div>
          </div>

          <!-- QR Code section -->
          <div class="single-file-qr">
            <div class="qr-box">
              <qrcode-vue :value="link" :size="120" level="M"></qrcode-vue>
            </div>
            <span class="qr-label">Scan to download on mobile</span>
          </div>
        </div>
      </div>

      <!-- UPLOAD-ONLY / FILE DROP VIEW -->
      <div v-else-if="req.uploadOnly" class="share upload-only-share">
        <div class="upload-drop-card">
          <div class="drop-icon">
            <i class="material-icons">cloud_upload</i>
          </div>
          <h2>Dateianforderung / Briefkasten</h2>
          <p class="drop-subtitle">
            Lade deine Dateien sicher in den Ordner <strong>{{ req.name }}</strong> hoch.<br />
            Vorhandene Dateien bleiben geschützt und sind nicht einsehbar.
          </p>

          <div
            class="drop-zone"
            :class="{ 'is-dragover': isDragOver }"
            @dragover.prevent="isDragOver = true"
            @dragleave.prevent="isDragOver = false"
            @drop.prevent="onDropFiles"
            @click="triggerFileInput"
          >
            <input
              type="file"
              ref="fileInputRef"
              multiple
              style="display: none"
              @change="onFileInputChange"
            />
            <i class="material-icons upload-cloud-icon">move_to_inbox</i>
            <p class="drop-prompt">
              Dateien hierher ziehen oder <span>Durchsuchen</span>
            </p>
            <span class="drop-hint">Mehrere Dateien gleichzeitig möglich</span>
          </div>

          <!-- Upload Status List -->
          <div v-if="uploadedFiles.length > 0" class="upload-status-section">
            <h3>Dateistatus ({{ uploadedFiles.length }})</h3>
            <div class="upload-item-list">
              <div
                v-for="file in uploadedFiles"
                :key="file.name + file.size"
                class="upload-item-row"
                :class="file.status"
              >
                <i class="material-icons status-icon">
                  {{ file.status === 'success' ? 'check_circle' : file.status === 'uploading' ? 'sync' : 'error' }}
                </i>
                <span class="file-name">{{ file.name }}</span>
                <span class="file-size">{{ filesize(file.size) }}</span>
                <span class="status-text">
                  {{ file.status === 'success' ? 'Hochgeladen' : file.status === 'uploading' ? 'Wird übertragen...' : 'Fehler' }}
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- FOLDER SHARE VIEW -->
      <div v-else class="share">
        <div class="share__box share__box__info">
          <div class="share__box__header">
            <i class="material-icons" style="font-size: 1.3em; vertical-align: middle; margin-right: 4px;">folder_shared</i>
            {{ t("download.downloadFolder") }}
          </div>

          <div class="share__box__element">
            <strong>{{ $t("prompts.displayName") }}:</strong> {{ req.name }}
          </div>
          <div class="share__box__element">
            <strong>{{ $t("prompts.size") }}:</strong> {{ req.items.length }} Items
          </div>

          <div class="share__box__element share__box__center">
            <a :href="link" class="button button--primary folder-download-btn">
              <i class="material-icons">archive</i>
              {{ t("buttons.download") }} (ZIP)
            </a>
          </div>

          <div class="share__box__element share__box__center qr-folder-container">
            <qrcode-vue :value="link" :size="110" level="M"></qrcode-vue>
            <p class="qr-subtext">Scan for folder link</p>
          </div>
        </div>

        <div
          id="shareList"
          v-if="req.items.length > 0"
          class="share__box share__box__items"
        >
          <div class="share__box__header">
            {{ t("files.files") }} ({{ req.items.length }})
          </div>
          <div id="listing" class="list file-icons">
            <item
              v-for="item in req.items.slice(0, showLimit)"
              :key="base64(item.name)"
              v-bind:index="item.index"
              v-bind:name="item.name"
              v-bind:isDir="item.isDir"
              v-bind:url="item.url"
              v-bind:modified="item.modified"
              v-bind:type="item.type"
              v-bind:size="item.size"
              readOnly
            >
            </item>
            <div
              v-if="req.items.length > showLimit"
              class="item"
              @click="showLimit += 100"
            >
              <div>
                <p class="name">+ {{ req.items.length - showLimit }} more</p>
              </div>
            </div>
          </div>
        </div>
        <div
          v-else
          class="share__box share__box__items"
        >
          <h2 class="message">
            <i class="material-icons">sentiment_dissatisfied</i>
            <span>{{ t("files.lonely") }}</span>
          </h2>
        </div>
      </div>
    </div>

    <!-- Floating Multiple Selection Bar -->
    <div
      v-if="req?.isDir"
      :class="{ active: fileStore.multiple }"
      id="multiple-selection"
    >
      <p>{{ t("files.multipleSelectionEnabled") }}</p>
      <div
        @click="() => (fileStore.multiple = false)"
        tabindex="0"
        role="button"
        :data-title="t('buttons.clear')"
        :aria-label="t('buttons.clear')"
        class="action"
      >
        <i class="material-icons">clear</i>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { pub as api } from "@/api";
import { filesize } from "@/utils";
import dayjs from "dayjs";
import { Base64 } from "js-base64";
import { createURL } from "@/api/utils";
import HeaderBar from "@/components/header/HeaderBar.vue";
import Action from "@/components/header/Action.vue";
import Breadcrumbs from "@/components/Breadcrumbs.vue";
import Errors from "@/views/Errors.vue";
import QrcodeVue from "qrcode.vue";
import Item from "@/components/files/ListingItem.vue";
import { useFileStore } from "@/stores/file";
import { useLayoutStore } from "@/stores/layout";
import { computed, inject, onMounted, onBeforeUnmount, ref, watch } from "vue";
import { useRoute } from "vue-router";
import { useI18n } from "vue-i18n";
import { StatusError } from "@/api/utils";
import { copy } from "@/utils/clipboard";

const error = ref<StatusError | null>(null);
const showLimit = ref<number>(100);
const password = ref<string>("");
const attemptedPasswordLogin = ref<boolean>(false);
const hash = ref<string>("");
const token = ref<string>("");

interface UploadedFileItem {
  name: string;
  size: number;
  status: "uploading" | "success" | "error";
}

const isDragOver = ref(false);
const fileInputRef = ref<HTMLInputElement | null>(null);
const uploadedFiles = ref<UploadedFileItem[]>([]);

const triggerFileInput = () => {
  fileInputRef.value?.click();
};

const onFileInputChange = (event: Event) => {
  const target = event.target as HTMLInputElement;
  if (target.files && target.files.length > 0) {
    uploadPublicFiles(target.files);
    target.value = "";
  }
};

const onDropFiles = (event: DragEvent) => {
  isDragOver.value = false;
  if (event.dataTransfer?.files && event.dataTransfer.files.length > 0) {
    uploadPublicFiles(event.dataTransfer.files);
  }
};

const uploadPublicFiles = async (files: FileList | File[]) => {
  for (const file of Array.from(files)) {
    const item: UploadedFileItem = {
      name: file.name,
      size: file.size,
      status: "uploading",
    };
    uploadedFiles.value.unshift(item);

    const formData = new FormData();
    formData.append("file", file, file.name);

    try {
      const headers: Record<string, string> = {};
      if (password.value) {
        headers["X-SHARE-PASSWORD"] = encodeURIComponent(password.value);
      }
      const tokenVal = token.value || (route.query.token as string);
      const url = tokenVal
        ? `/api/public/upload/${hash.value}?token=${encodeURIComponent(tokenVal)}`
        : `/api/public/upload/${hash.value}`;

      const res = await fetch(url, {
        method: "POST",
        headers,
        body: formData,
      });

      if (res.ok) {
        item.status = "success";
        $showSuccess(`"${file.name}" erfolgreich übertragen!`);
      } else {
        item.status = "error";
        $showError(`Fehler beim Upload von "${file.name}".`);
      }
    } catch (err: any) {
      item.status = "error";
      $showError(err.message || "Upload-Fehler");
    }
  }
};

const $showError = inject<IToastError>("$showError")!;
const $showSuccess = inject<IToastSuccess>("$showSuccess")!;

const { t } = useI18n({});

const route = useRoute();
const fileStore = useFileStore();
const layoutStore = useLayoutStore();

watch(route, () => {
  showLimit.value = 100;
  fetchData();
});

const req = computed(() => fileStore.req);

const icon = computed(() => {
  if (req.value === null) return "insert_drive_file";
  if (req.value.isDir) return "folder";
  if (req.value.type === "image") return "insert_photo";
  if (req.value.type === "audio") return "volume_up";
  if (req.value.type === "video") return "movie";
  return "insert_drive_file";
});

const link = computed(() => (req.value ? api.getDownloadURL(req.value) : ""));

const raw = computed(() => {
  if (!req.value) return "";
  if (!req.value.isDir) {
    return createURL(`api/public/dl/${hash.value}`, { token: token.value });
  }
  if (req.value.items && req.value.items[fileStore.selected[0]]) {
    return createURL(
      `api/public/dl/${hash.value}${req.value.items[fileStore.selected[0]].path}`,
      { token: token.value }
    );
  }
  return "";
});

const inlineLink = computed(() =>
  req.value ? api.getDownloadURL(req.value, true) : ""
);

const humanSize = computed(() => {
  if (req.value) {
    return req.value.isDir
      ? `${req.value.items.length} items`
      : filesize(req.value.size ?? 0);
  }
  return "";
});

const humanTime = computed(() => dayjs(req.value?.modified).fromNow());
const modTime = computed(() =>
  req.value
    ? new Date(Date.parse(req.value.modified)).toLocaleString()
    : new Date().toLocaleString()
);

const base64 = (name: any) => Base64.encodeURI(name);

const fetchData = async () => {
  fileStore.reload = false;
  fileStore.selected = [];
  fileStore.multiple = false;
  layoutStore.closeHovers();

  layoutStore.loading = true;
  error.value = null;
  if (password.value !== "") {
    attemptedPasswordLogin.value = true;
  }

  let url = route.path;
  if (url === "") url = "/";
  if (url[0] !== "/") url = "/" + url;

  try {
    const file = await api.fetch(url, password.value);
    file.hash = hash.value;
    token.value = file.token || "";

    fileStore.updateRequest(file);
    document.title = `${file.name} - ${document.title}`;
  } catch (err) {
    if (err instanceof Error) {
      error.value = err;
    }
  } finally {
    layoutStore.loading = false;
  }
};

const keyEvent = (event: KeyboardEvent) => {
  if (event.key === "Escape") {
    if (fileStore.selectedCount > 0) {
      fileStore.selected = [];
    }
  }
};

const toggleMultipleSelection = () => {
  fileStore.toggleMultiple();
};

const isSingleFile = () =>
  fileStore.selectedCount === 1 &&
  !req.value?.items[fileStore.selected[0]]?.isDir;

const download = () => {
  if (!req.value) return false;

  if (isSingleFile()) {
    api.download(
      null,
      hash.value,
      token.value,
      req.value.items[fileStore.selected[0]].path
    );
    return true;
  }

  layoutStore.showHover({
    prompt: "download",
    confirm: (format: DownloadFormat) => {
      if (req.value === null) return false;
      layoutStore.closeHovers();

      const files: string[] = [];
      for (const i of fileStore.selected) {
        files.push(req.value.items[i].path);
      }

      api.download(format, hash.value, token.value, ...files);
      return true;
    },
  });

  return true;
};

const linkSelected = () => {
  return isSingleFile() && req.value
    ? api.getDownloadURL({
        ...req.value,
        hash: hash.value,
        path: req.value.items[fileStore.selected[0]].path,
      })
    : "";
};

const copyToClipboard = (text: string) => {
  copy({ text }).then(
    () => {
      $showSuccess(t("success.linkCopied"));
    },
    () => {
      copy({ text }, { permission: true }).then(
        () => {
          $showSuccess(t("success.linkCopied"));
        },
        (e) => {
          $showError(e);
        }
      );
    }
  );
};

onMounted(async () => {
  hash.value = route.params.path[0];
  window.addEventListener("keydown", keyEvent);
  await fetchData();
});

onBeforeUnmount(() => {
  window.removeEventListener("keydown", keyEvent);
});
</script>

<style scoped>
.share-page {
  min-height: 100vh;
  background: var(--background, #0f172a);
}

/* SINGLE FILE HERO CARD */
.single-file-share {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 3rem 1.5rem;
  max-width: 900px;
  margin: 0 auto;
}

.single-file-card {
  width: 100%;
  background: var(--surfacePrimary, #1e293b);
  border: 1px solid var(--borderPrimary, rgba(255, 255, 255, 0.08));
  border-radius: 16px;
  padding: 2.5rem;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.25);
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  gap: 1.5rem;
}

.single-file-preview {
  width: 100%;
  max-height: 320px;
  display: flex;
  justify-content: center;
  align-items: center;
  background: var(--surfaceSecondary, rgba(0, 0, 0, 0.2));
  border-radius: 12px;
  overflow: hidden;
  padding: 1rem;
}

.single-file-image {
  max-height: 300px;
  max-width: 100%;
  object-fit: contain;
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.single-file-video {
  max-height: 300px;
  max-width: 100%;
  border-radius: 8px;
}

.single-file-audio {
  width: 100%;
  max-width: 450px;
}

.single-file-icon i {
  font-size: 6rem;
  color: var(--blue, #3b82f6);
}

.single-file-details {
  width: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1rem;
}

.single-file-title {
  font-size: 1.6rem;
  font-weight: 700;
  color: var(--textPrimary, #f8fafc);
  word-break: break-word;
  margin: 0;
  max-width: 90%;
}

.single-file-meta {
  display: flex;
  gap: 1.5rem;
  font-size: 0.95rem;
  color: var(--textSecondary, #94a3b8);
}

.meta-item {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.meta-item i {
  font-size: 1.1rem;
  opacity: 0.8;
}

.single-file-actions {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 12px;
  margin-top: 0.5rem;
}

.download-btn {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 0.75rem 1.8rem;
  font-size: 1rem;
  font-weight: 600;
  background: var(--blue, #2563eb);
  color: #fff !important;
  border-radius: 8px;
  box-shadow: 0 4px 14px rgba(37, 99, 235, 0.4);
  transition: all 0.2s ease;
}

.download-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(37, 99, 235, 0.5);
}

.open-btn,
.copy-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 0.75rem 1.2rem;
  font-size: 0.95rem;
  border: 1px solid var(--borderPrimary, rgba(255, 255, 255, 0.15));
  border-radius: 8px;
}

.single-file-qr {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  margin-top: 0.5rem;
  padding-top: 1rem;
  border-top: 1px solid var(--borderPrimary, rgba(255, 255, 255, 0.08));
}

.qr-box {
  background: #fff;
  padding: 10px;
  border-radius: 10px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.qr-label,
.qr-subtext {
  font-size: 0.8rem;
  color: var(--textSecondary, #94a3b8);
  margin: 0;
}

.folder-download-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-weight: 600;
}

.qr-folder-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
}

#listing.list {
  height: auto;
}

#shareList {
  overflow-y: scroll;
}

@media (min-width: 930px) {
  #shareList {
    height: calc(100vh - 9.8em);
    overflow-y: auto;
  }
}

/* Upload-Only View Styles */
.upload-only-share {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 2.5rem 1rem;
  width: 100%;
}

.upload-drop-card {
  max-width: 600px;
  width: 100%;
  background: var(--surfacePrimary, #1e293b);
  border: 1px solid var(--borderPrimary, rgba(255, 255, 255, 0.08));
  border-radius: 16px;
  padding: 2.5rem 2rem;
  text-align: center;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.25);
}

.drop-icon {
  width: 64px;
  height: 64px;
  margin: 0 auto 1.2rem;
  border-radius: 50%;
  background: rgba(33, 150, 243, 0.15);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--primary, #2196f3);
}

.drop-icon i {
  font-size: 2.4rem;
}

.drop-subtitle {
  color: var(--textSecondary, #94a3b8);
  font-size: 0.95rem;
  margin-bottom: 2rem;
  line-height: 1.5;
}

.drop-zone {
  border: 2px dashed rgba(33, 150, 243, 0.4);
  border-radius: 12px;
  padding: 2.5rem 1.5rem;
  background: rgba(33, 150, 243, 0.03);
  cursor: pointer;
  transition: all 0.2s ease;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.6rem;
}

.drop-zone:hover,
.drop-zone.is-dragover {
  border-color: var(--primary, #2196f3);
  background: rgba(33, 150, 243, 0.1);
  transform: scale(1.01);
}

.upload-cloud-icon {
  font-size: 3rem;
  color: var(--primary, #2196f3);
  opacity: 0.9;
}

.drop-prompt {
  font-size: 1.05rem;
  font-weight: 500;
  margin: 0;
}

.drop-prompt span {
  color: var(--primary, #2196f3);
  text-decoration: underline;
}

.drop-hint {
  font-size: 0.82rem;
  opacity: 0.6;
}

.upload-status-section {
  margin-top: 2rem;
  text-align: left;
}

.upload-status-section h3 {
  font-size: 1rem;
  margin-bottom: 0.8rem;
  opacity: 0.9;
}

.upload-item-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  max-height: 240px;
  overflow-y: auto;
}

.upload-item-row {
  display: flex;
  align-items: center;
  gap: 0.8rem;
  padding: 0.6rem 0.8rem;
  border-radius: 6px;
  background: rgba(255, 255, 255, 0.04);
  font-size: 0.88rem;
}

.upload-item-row.success {
  border-left: 3px solid #22c55e;
}

.upload-item-row.success .status-icon {
  color: #22c55e;
}

.upload-item-row.uploading {
  border-left: 3px solid #3b82f6;
}

.upload-item-row.uploading .status-icon {
  color: #3b82f6;
  animation: spin 1.2s linear infinite;
}

.upload-item-row.error {
  border-left: 3px solid #ef4444;
}

.upload-item-row.error .status-icon {
  color: #ef4444;
}

.file-name {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.file-size {
  opacity: 0.7;
  font-size: 0.8rem;
}

.status-text {
  font-size: 0.78rem;
  font-weight: 500;
  opacity: 0.85;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}
</style>
