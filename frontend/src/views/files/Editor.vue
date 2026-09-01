<template>
  <div id="editor-container">
    <header-bar>
      <action icon="close" :label="t('buttons.close')" @action="close()" />
      <title>{{ fileStore.req?.name ?? "" }}</title>

      <!-- Language / Syntax Mode Selector -->
      <div class="editor-mode-select-wrapper" title="Syntax-Hervorhebung ändern">
        <i class="material-icons" style="font-size: 1.1rem; opacity: 0.7;">code</i>
        <select v-model="selectedMode" @change="onModeChange" class="editor-select">
          <option v-for="mode in availableModes" :key="mode.value" :value="mode.value">
            {{ mode.label }}
          </option>
        </select>
      </div>

      <!-- Format Document -->
      <action
        v-if="authStore.user?.perm.modify"
        icon="auto_fix_high"
        label="Format Document (Shift+Alt+F)"
        @action="formatDocument"
      />

      <!-- Word Wrap Toggle -->
      <action
        :icon="isWordWrap ? 'wrap_text' : 'format_align_left'"
        :label="isWordWrap ? 'Word Wrap: An' : 'Word Wrap: Aus'"
        @action="toggleWordWrap"
      />

      <!-- Find & Replace -->
      <action
        icon="search"
        label="Suchen & Ersetzen (Ctrl+F / Ctrl+H)"
        @action="openSearchBox"
      />

      <!-- Font Size Controls -->
      <action
        icon="add"
        @action="increaseFontSize"
        :label="t('buttons.increaseFontSize')"
      />
      <span class="editor-font-size">{{ fontSize }}px</span>
      <action
        icon="remove"
        @action="decreaseFontSize"
        :label="t('buttons.decreaseFontSize')"
      />

      <!-- Save and Revert -->
      <action
        v-if="authStore.user?.perm.modify"
        id="save-button"
        icon="save"
        :label="t('buttons.save')"
        @action="save()"
      />
      <action
        v-if="authStore.user?.perm.modify"
        icon="undo"
        label="Revert"
        @action="revert()"
      />

      <!-- Markdown View Toggle -->
      <template v-if="isMarkdownFile">
        <action
          :icon="viewMode === 'split' ? 'vertical_split' : viewMode === 'preview' ? 'visibility' : 'edit'"
          :label="viewMode === 'split' ? 'Split View' : viewMode === 'preview' ? 'Preview Only' : 'Editor Only'"
          @action="cycleViewMode"
        />
      </template>
    </header-bar>

    <!-- preview container -->
    <div class="loading delayed" v-if="layoutStore.loading">
      <div class="spinner">
        <div class="bounce1"></div>
        <div class="bounce2"></div>
        <div class="bounce3"></div>
      </div>
    </div>
    <template v-else>
      <div class="editor-header">
        <Breadcrumbs base="/files" noLink />

        <div class="editor-quick-actions">
          <button
            :disabled="isSelectionEmpty"
            @click="executeEditorCommand('copy')"
            title="Kopieren"
          >
            <span><i class="material-icons">content_copy</i></span>
          </button>
          <button
            :disabled="isSelectionEmpty"
            @click="executeEditorCommand('cut')"
            title="Ausschneiden"
          >
            <span><i class="material-icons">content_cut</i></span>
          </button>
          <button @click="executeEditorCommand('paste')" title="Einfügen">
            <span><i class="material-icons">content_paste</i></span>
          </button>
          <button @click="formatDocument" title="Dokument formatieren">
            <span><i class="material-icons">auto_fix_high</i></span>
          </button>
          <button @click="openSearchBox" title="Suchen (Ctrl+F)">
            <span><i class="material-icons">search</i></span>
          </button>
          <button @click="executeEditorCommand('openCommandPalette')" title="Befehlspalette">
            <span><i class="material-icons">more_vert</i></span>
          </button>
        </div>
      </div>

      <div :class="['editor-workspace', { 'is-split': viewMode === 'split' && isMarkdownFile }]">
        <form
          v-show="viewMode !== 'preview' || !isMarkdownFile"
          id="editor"
          :class="{ 'split-pane': viewMode === 'split' && isMarkdownFile }"
        ></form>
        <div
          v-show="(viewMode === 'preview' || viewMode === 'split') && isMarkdownFile"
          id="preview-container"
          class="md_preview"
          :class="{ 'split-pane': viewMode === 'split' && isMarkdownFile }"
          v-html="previewContent"
        ></div>
      </div>

      <!-- Editor Status Bar -->
      <div class="editor-status-bar">
        <div class="status-left">
          <span class="status-item" title="Cursor-Position">
            <i class="material-icons status-icon">pin_drop</i>
            Ln {{ cursorRow }}, Col {{ cursorCol }}
          </span>
          <span v-if="selectedChars > 0" class="status-item highlight">
            ({{ selectedChars }} ausgewählt)
          </span>
          <span class="status-item" title="Gesamtgröße">
            {{ totalLines }} Zeilen, {{ totalChars }} Zeichen
          </span>
        </div>

        <div class="status-right">
          <!-- Tab Size Selector -->
          <div class="status-item tab-size-item">
            <label>Spaces:</label>
            <select v-model="tabSize" @change="onTabSizeChange" class="status-select">
              <option :value="2">2</option>
              <option :value="4">4</option>
              <option :value="8">8</option>
            </select>
          </div>

          <span class="status-item" title="Zeichenkodierung">UTF-8</span>
          <span class="status-item mode-badge" :title="'Syntax: ' + selectedMode">
            {{ currentModeLabel }}
          </span>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { files as api } from "@/api";
import buttons from "@/utils/buttons";
import url from "@/utils/url";
import ace, { Ace, version as ace_version } from "ace-builds";
import "ace-builds/src-noconflict/ext-language_tools";
import "ace-builds/src-noconflict/ext-searchbox";
import modelist from "ace-builds/src-noconflict/ext-modelist";
import DOMPurify from "dompurify";

import Breadcrumbs from "@/components/Breadcrumbs.vue";
import Action from "@/components/header/Action.vue";
import HeaderBar from "@/components/header/HeaderBar.vue";
import { useAuthStore } from "@/stores/auth";
import { useFileStore } from "@/stores/file";
import { useLayoutStore } from "@/stores/layout";
import { getEditorTheme } from "@/utils/theme";
import { marked } from "marked";
import markedKatex from "marked-katex-extension";
import { computed, inject, onBeforeUnmount, onMounted, ref, watch, watchEffect } from "vue";
import { useI18n } from "vue-i18n";
import { onBeforeRouteUpdate, useRoute, useRouter } from "vue-router";
import { read, copy } from "@/utils/clipboard";

const $showError = inject<IToastError>("$showError")!;
const $showSuccess = inject<IToastSuccess>("$showSuccess")!;

const fileStore = useFileStore();
const authStore = useAuthStore();
const layoutStore = useLayoutStore();

const { t } = useI18n();

const route = useRoute();
const router = useRouter();

const editor = ref<Ace.Editor | null>(null);
const fontSize = ref(parseInt(localStorage.getItem("editorFontSize") || "14"));
const tabSize = ref<number>(parseInt(localStorage.getItem("editorTabSize") || "2"));
const isWordWrap = ref<boolean>(localStorage.getItem("editorWordWrap") !== "false");

const cursorRow = ref<number>(1);
const cursorCol = ref<number>(1);
const selectedChars = ref<number>(0);
const totalLines = ref<number>(1);
const totalChars = ref<number>(0);

const viewMode = ref<"edit" | "split" | "preview">("edit");
const previewContent = ref("");

const availableModes = [
  { label: "Auto / Plain Text", value: "ace/mode/plain_text" },
  { label: "JSON", value: "ace/mode/json" },
  { label: "PHP", value: "ace/mode/php" },
  { label: "HTML", value: "ace/mode/html" },
  { label: "CSS", value: "ace/mode/css" },
  { label: "SCSS / SASS", value: "ace/mode/scss" },
  { label: "JavaScript", value: "ace/mode/javascript" },
  { label: "TypeScript", value: "ace/mode/typescript" },
  { label: "Python", value: "ace/mode/python" },
  { label: "Go", value: "ace/mode/golang" },
  { label: "Rust", value: "ace/mode/rust" },
  { label: "C / C++", value: "ace/mode/c_cpp" },
  { label: "Java", value: "ace/mode/java" },
  { label: "SQL", value: "ace/mode/sql" },
  { label: "YAML", value: "ace/mode/yaml" },
  { label: "XML / SVG", value: "ace/mode/xml" },
  { label: "Markdown", value: "ace/mode/markdown" },
  { label: "Shell / Bash", value: "ace/mode/sh" },
  { label: "Dockerfile", value: "ace/mode/dockerfile" },
  { label: "Nginx", value: "ace/mode/nginx" },
  { label: "INI / Config", value: "ace/mode/ini" },
];

const selectedMode = ref<string>("ace/mode/plain_text");

const currentModeLabel = computed(() => {
  const m = availableModes.find((item) => item.value === selectedMode.value);
  if (m) return m.label;
  const parts = selectedMode.value.split("/");
  return parts[parts.length - 1].toUpperCase();
});

const isMarkdownFile = computed(() => {
  const name = fileStore.req?.name?.toLowerCase() || "";
  return name.endsWith(".md") || name.endsWith(".markdown") || selectedMode.value === "ace/mode/markdown";
});

const katexOptions = {
  output: "mathml" as const,
  throwOnError: false,
};
marked.use(markedKatex(katexOptions));

const isSelectionEmpty = ref(true);

const cycleViewMode = () => {
  if (viewMode.value === "edit") {
    viewMode.value = "split";
  } else if (viewMode.value === "split") {
    viewMode.value = "preview";
  } else {
    viewMode.value = "edit";
  }
  updateMarkdownPreview();
  setTimeout(() => {
    editor.value?.resize();
  }, 50);
};

const updateMarkdownPreview = async () => {
  if (!isMarkdownFile.value) return;
  const val = editor.value?.getValue() || "";
  try {
    previewContent.value = DOMPurify.sanitize(await marked(val));
  } catch (error) {
    console.error("Failed to convert markdown:", error);
  }
};

const executeEditorCommand = (name: string) => {
  if (name == "paste") {
    read()
      .then((data) => {
        editor.value?.execCommand("paste", {
          text: data,
        });
      })
      .catch((e) => {
        if (
          document.queryCommandSupported &&
          document.queryCommandSupported("paste")
        ) {
          document.execCommand("paste");
        } else {
          console.warn("the clipboard api is not supported", e);
        }
      });
    return;
  }
  if (name == "copy" || name == "cut") {
    const selectedText = editor.value?.getCopyText();
    copy({ text: selectedText });
  }
  editor.value?.execCommand(name);
};

const openSearchBox = () => {
  if (!editor.value) return;
  editor.value.execCommand("find");
};

const toggleWordWrap = () => {
  isWordWrap.value = !isWordWrap.value;
  editor.value?.getSession().setUseWrapMode(isWordWrap.value);
  localStorage.setItem("editorWordWrap", isWordWrap.value.toString());
};

const onTabSizeChange = () => {
  if (!editor.value) return;
  editor.value.getSession().setTabSize(tabSize.value);
  localStorage.setItem("editorTabSize", tabSize.value.toString());
};

const onModeChange = () => {
  if (!editor.value) return;
  editor.value.getSession().setMode(selectedMode.value);
  if (isMarkdownFile.value) {
    updateMarkdownPreview();
  }
};

const formatDocument = () => {
  if (!editor.value) return;
  const mode = selectedMode.value;
  const content = editor.value.getValue();

  if (mode.includes("json")) {
    try {
      const parsed = JSON.parse(content);
      const formatted = JSON.stringify(parsed, null, tabSize.value);
      editor.value.setValue(formatted, -1);
      if ($showSuccess) $showSuccess("JSON sauber formatiert");
    } catch (err: any) {
      $showError("Fehler beim Formatieren von JSON: " + (err.message || err));
    }
    return;
  }

  // Format generic code / indentation
  try {
    const lines = content.split("\n");
    let indentLevel = 0;
    const indentStr = " ".repeat(tabSize.value);
    const formattedLines: string[] = [];

    for (let line of lines) {
      const trimmed = line.trim();
      if (!trimmed) {
        formattedLines.push("");
        continue;
      }

      // Closing braces decrease indent before line
      if (trimmed.startsWith("}") || trimmed.startsWith("]") || trimmed.startsWith("</")) {
        indentLevel = Math.max(0, indentLevel - 1);
      }

      formattedLines.push(indentStr.repeat(indentLevel) + trimmed);

      // Opening braces increase indent for following lines
      if (
        (trimmed.endsWith("{") || trimmed.endsWith("[") || (trimmed.startsWith("<") && !trimmed.endsWith("/>") && !trimmed.startsWith("</") && trimmed.endsWith(">") && !trimmed.includes("</"))) &&
        !trimmed.endsWith("}") && !trimmed.endsWith("]")
      ) {
        indentLevel++;
      }
    }

    editor.value.setValue(formattedLines.join("\n"), -1);
    if ($showSuccess) $showSuccess("Dokument formatiert");
  } catch (e: any) {
    $showError("Formatierung fehlgeschlagen: " + e.message);
  }
};

const updateStatusMetrics = () => {
  if (!editor.value) return;
  const cursor = editor.value.getCursorPosition();
  cursorRow.value = cursor.row + 1;
  cursorCol.value = cursor.column + 1;

  const selText = editor.value.getSelectedText();
  selectedChars.value = selText.length;

  const content = editor.value.getValue();
  totalLines.value = editor.value.getSession().getLength();
  totalChars.value = content.length;
};

onMounted(() => {
  window.addEventListener("keydown", keyEvent);
  window.addEventListener("beforeunload", handlePageChange);

  watchEffect(async () => {
    if (isMarkdownFile.value && viewMode.value !== "edit") {
      const new_value = editor.value?.getValue() || "";
      try {
        previewContent.value = DOMPurify.sanitize(await marked(new_value));
      } catch (error) {
        console.error("Failed to convert content to HTML:", error);
        previewContent.value = "";
      }
    }
  });

  ace.config.set(
    "basePath",
    `https://cdn.jsdelivr.net/npm/ace-builds@${ace_version}/src-min-noconflict/`
  );

  const checkAndInit = () => {
    if (!layoutStore.loading && fileStore.req) {
      if (!editor.value) {
        initEditor();
      } else if (fileStore.req.content !== undefined) {
        if (editor.value.session.getUndoManager().isClean()) {
          editor.value.setValue(fileStore.req.content, -1);
          editor.value.session.getUndoManager().markClean();
          updateStatusMetrics();
        }
      }
    }
  };

  checkAndInit();

  watch(
    () => [layoutStore.loading, fileStore.req?.content],
    () => {
      checkAndInit();
    }
  );
});

onBeforeUnmount(() => {
  window.removeEventListener("keydown", keyEvent);
  window.removeEventListener("beforeunload", handlePageChange);
  editor.value?.destroy();
});

onBeforeRouteUpdate((to, from, next) => {
  if (editor.value?.session.getUndoManager().isClean()) {
    next();
    return;
  }

  layoutStore.showHover({
    prompt: "discardEditorChanges",
    confirm: (event: Event) => {
      event.preventDefault();
      next();
    },
    saveAction: async () => {
      await save();
      next();
    },
  });
});

const initEditor = () => {
  const fileContent = fileStore.req?.content || "";
  const fileName = fileStore.req?.name || "";
  const detectedMode = modelist.getModeForPath(fileName).mode;

  selectedMode.value = detectedMode;

  editor.value = ace.edit("editor", {
    value: fileContent,
    showPrintMargin: false,
    readOnly: fileStore.req?.type === "textImmutable",
    theme: getEditorTheme(authStore.user?.aceEditorTheme ?? ""),
    mode: detectedMode,
    wrap: isWordWrap.value,
    tabSize: tabSize.value,
    useSoftTabs: true,
    enableBasicAutocompletion: true,
    enableLiveAutocompletion: true,
    enableSnippets: true,
  });

  editor.value.setFontSize(fontSize.value);
  editor.value.focus();

  const selection = editor.value.getSelection();
  selection.on("changeSelection", function () {
    isSelectionEmpty.value = selection.isEmpty();
    updateStatusMetrics();
  });

  selection.on("changeCursor", function () {
    updateStatusMetrics();
  });

  editor.value.session.on("change", () => {
    updateStatusMetrics();
    if (viewMode.value === "split" || viewMode.value === "preview") {
      updateMarkdownPreview();
    }
  });

  updateStatusMetrics();

  if (isMarkdownFile.value) {
    updateMarkdownPreview();
  }
};

const revert = () => {
  if (editor.value && fileStore.req) {
    editor.value.setValue(fileStore.req.content || "", -1);
    editor.value.session.getUndoManager().markClean();
    updateStatusMetrics();
  }
};

const keyEvent = (event: KeyboardEvent) => {
  if (event.code === "Escape") {
    close();
  }

  // Format Shortcut: Shift + Alt + F
  if (event.shiftKey && event.altKey && (event.key === "F" || event.key === "f")) {
    event.preventDefault();
    formatDocument();
    return;
  }

  if (!event.ctrlKey && !event.metaKey) {
    return;
  }

  if (event.key === "s" || event.key === "S") {
    event.preventDefault();
    save();
  }
};

const handlePageChange = (event: BeforeUnloadEvent) => {
  if (!editor.value?.session.getUndoManager().isClean()) {
    event.preventDefault();
    event.returnValue = true;
  }
};

const save = async (throwError?: boolean) => {
  const button = "save";
  buttons.loading("save");

  try {
    await api.put(route.path, editor.value?.getValue());
    editor.value?.session.getUndoManager().markClean();
    buttons.success(button);
  } catch (e: any) {
    buttons.done(button);
    $showError(e);
    if (throwError) throw e;
  }
};

const increaseFontSize = () => {
  fontSize.value += 1;
  editor.value?.setFontSize(fontSize.value);
  localStorage.setItem("editorFontSize", fontSize.value.toString());
};

const decreaseFontSize = () => {
  if (fontSize.value > 1) {
    fontSize.value -= 1;
    editor.value?.setFontSize(fontSize.value);
    localStorage.setItem("editorFontSize", fontSize.value.toString());
  }
};

const close = () => {
  if (!editor.value?.session.getUndoManager().isClean()) {
    layoutStore.showHover({
      prompt: "discardEditorChanges",
      confirm: (event: Event) => {
        event.preventDefault();
        editor.value?.session.getUndoManager().reset();
        finishClose();
      },
      saveAction: async () => {
        try {
          await save(true);
          finishClose();
        } catch {}
      },
    });
    return;
  }
  finishClose();
};

const finishClose = () => {
  const uri = url.removeLastDir(route.path) + "/";
  router.push({ path: uri });
};
</script>

<style scoped>
.editor-font-size {
  margin: 0 0.5em;
  color: var(--fg);
  font-size: 0.9rem;
  font-weight: 500;
}

.editor-mode-select-wrapper {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  margin: 0 0.5rem;
  background: var(--surfaceSecondary, rgba(255, 255, 255, 0.08));
  padding: 0.2rem 0.6rem;
  border-radius: 6px;
  border: 1px solid var(--border, rgba(255, 255, 255, 0.1));
}

.editor-select {
  background: transparent;
  color: var(--fg, currentColor);
  border: none;
  outline: none;
  font-size: 0.85rem;
  font-weight: 500;
  cursor: pointer;
}

.editor-select option {
  background: var(--surface, #202328);
  color: var(--fg, #ffffff);
}

.editor-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.4rem 1rem;
}

.editor-quick-actions {
  display: flex;
  align-items: center;
  gap: 0.25rem;
}

.editor-quick-actions button {
  background: transparent;
  color: var(--action, currentColor);
  border: none;
  outline: none;
  opacity: 0.8;
  cursor: pointer;
  padding: 4px 6px;
  border-radius: 4px;
  display: flex;
  align-items: center;
}

.editor-quick-actions button:hover:not(:disabled) {
  opacity: 1;
  background: var(--surfaceSecondary, rgba(255, 255, 255, 0.08));
}

.editor-quick-actions button:disabled {
  opacity: 0.35;
  cursor: not-allowed;
}

.editor-quick-actions button span i {
  font-size: 1.15rem;
}

.editor-workspace {
  height: calc(100vh - 10.5em);
  width: 100%;
  position: relative;
}

.editor-workspace.is-split {
  display: flex;
  flex-direction: row;
  gap: 12px;
}

.editor-workspace.is-split .split-pane {
  flex: 1;
  min-width: 0;
  height: 100%;
  overflow-y: auto;
}

#preview-container {
  padding: 1.5em;
  background: var(--surfaceSecondary, rgba(255, 255, 255, 0.05));
  border-radius: 8px;
  overflow-y: auto;
}

/* Status Bar */
.editor-status-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.35rem 1rem;
  background: var(--surfaceSecondary, #1b1d22);
  border-top: 1px solid var(--border, rgba(255, 255, 255, 0.08));
  font-size: 0.8rem;
  color: var(--fg, currentColor);
  opacity: 0.9;
  user-select: none;
}

.status-left, .status-right {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.status-item {
  display: flex;
  align-items: center;
  gap: 0.3rem;
  opacity: 0.85;
}

.status-icon {
  font-size: 0.95rem;
  opacity: 0.7;
}

.status-item.highlight {
  color: var(--primary, #2196f3);
  font-weight: 600;
}

.tab-size-item {
  display: flex;
  align-items: center;
  gap: 0.25rem;
}

.status-select {
  background: transparent;
  color: var(--fg, currentColor);
  border: 1px solid var(--border, rgba(255, 255, 255, 0.15));
  border-radius: 4px;
  font-size: 0.75rem;
  padding: 1px 4px;
  cursor: pointer;
}

.status-select option {
  background: var(--surface, #202328);
  color: var(--fg, #ffffff);
}

.mode-badge {
  background: rgba(33, 150, 243, 0.15);
  color: var(--primary, #2196f3);
  padding: 2px 6px;
  border-radius: 4px;
  font-weight: 600;
  font-size: 0.75rem;
}
</style>
