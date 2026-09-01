<template>
  <div id="editor-container">
    <header-bar>
      <action icon="close" :label="t('buttons.close')" @action="close()" />
      <title>{{ fileStore.req?.name ?? "" }}</title>

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

        <div>
          <button
            :disabled="isSelectionEmpty"
            @click="executeEditorCommand('copy')"
          >
            <span><i class="material-icons">content_copy</i></span>
          </button>
          <button
            :disabled="isSelectionEmpty"
            @click="executeEditorCommand('cut')"
          >
            <span><i class="material-icons">content_cut</i></span>
          </button>
          <button @click="executeEditorCommand('paste')">
            <span><i class="material-icons">content_paste</i></span>
          </button>
          <button @click="executeEditorCommand('openCommandPalette')">
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
    </template>
  </div>
</template>

<script setup lang="ts">
import { files as api } from "@/api";
import buttons from "@/utils/buttons";
import url from "@/utils/url";
import ace, { Ace, version as ace_version } from "ace-builds";
import "ace-builds/src-noconflict/ext-language_tools";
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

const fileStore = useFileStore();
const authStore = useAuthStore();
const layoutStore = useLayoutStore();

const { t } = useI18n();

const route = useRoute();
const router = useRouter();

const editor = ref<Ace.Editor | null>(null);
const fontSize = ref(parseInt(localStorage.getItem("editorFontSize") || "14"));

const viewMode = ref<"edit" | "split" | "preview">("edit");
const previewContent = ref("");
const isMarkdownFile = computed(() => {
  const name = fileStore.req?.name?.toLowerCase() || "";
  return name.endsWith(".md") || name.endsWith(".markdown");
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

  editor.value = ace.edit("editor", {
    value: fileContent,
    showPrintMargin: false,
    readOnly: fileStore.req?.type === "textImmutable",
    theme: getEditorTheme(authStore.user?.aceEditorTheme ?? ""),
    mode: modelist.getModeForPath(fileName).mode,
    wrap: true,
    enableBasicAutocompletion: true,
    enableLiveAutocompletion: true,
    enableSnippets: true,
  });

  editor.value.setFontSize(fontSize.value);
  editor.value.focus();

  const selection = editor.value?.getSelection();
  selection.on("changeSelection", function () {
    isSelectionEmpty.value = selection.isEmpty();
  });

  editor.value.session.on("change", () => {
    if (viewMode.value === "split" || viewMode.value === "preview") {
      updateMarkdownPreview();
    }
  });

  if (isMarkdownFile.value) {
    updateMarkdownPreview();
  }
};

const revert = () => {
  if (editor.value && fileStore.req) {
    editor.value.setValue(fileStore.req.content || "", -1);
    editor.value.session.getUndoManager().markClean();
  }
};

const keyEvent = (event: KeyboardEvent) => {
  if (event.code === "Escape") {
    close();
  }

  if (!event.ctrlKey && !event.metaKey) {
    return;
  }

  if (event.key !== "s") {
    return;
  }

  event.preventDefault();
  save();
};

const handlePageChange = (event: BeforeUnloadEvent) => {
  if (!editor.value?.session.getUndoManager().isClean()) {
    event.preventDefault();
    // returnValue is now depecrated, though keeping in for legacy browser support
    // https://developer.mozilla.org/en-US/docs/Web/API/BeforeUnloadEvent/returnValue
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
}

.editor-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.editor-header > div > button {
  background: transparent;
  color: var(--action);
  border: none;
  outline: none;
  opacity: 0.8;
  cursor: pointer;
}

.editor-header > div > button:hover:not(:disabled) {
  opacity: 1;
}

.editor-header > div > button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.editor-header > div > button > span > i {
  font-size: 1.2rem;
}

.editor-workspace {
  height: calc(100vh - 8em);
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
</style>
