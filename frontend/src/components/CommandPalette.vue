<template>
  <div v-if="isOpen" class="command-palette-backdrop" @click="close">
    <div class="command-palette-modal" @click.stop>
      <div class="command-palette-header">
        <i class="material-icons">search</i>
        <input
          ref="inputRef"
          v-model="query"
          type="text"
          placeholder="Type a command or search action (e.g. upload, settings, dark mode)..."
          @keydown="handleKeyDown"
        />
        <span class="command-palette-kbd">ESC</span>
      </div>

      <div class="command-palette-list">
        <div
          v-for="(item, index) in filteredCommands"
          :key="item.id"
          class="command-palette-item"
          :class="{ selected: index === selectedIndex }"
          @click="execute(item)"
          @mouseenter="selectedIndex = index"
        >
          <i class="material-icons">{{ item.icon }}</i>
          <div class="command-palette-item-text">
            <span class="title">{{ item.title }}</span>
            <span v-if="item.subtitle" class="subtitle">{{ item.subtitle }}</span>
          </div>
          <span v-if="item.shortcut" class="shortcut">{{ item.shortcut }}</span>
        </div>

        <div v-if="filteredCommands.length === 0" class="command-palette-empty">
          No matching commands found.
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref } from "vue";
import { useRouter } from "vue-router";
import { useLayoutStore } from "@/stores/layout";
import { useAuthStore } from "@/stores/auth";
import { logout } from "@/utils/auth";

const isOpen = ref<boolean>(false);
const query = ref<string>("");
const selectedIndex = ref<number>(0);
const inputRef = ref<HTMLInputElement | null>(null);

const router = useRouter();
const layoutStore = useLayoutStore();
const authStore = useAuthStore();

interface CommandItem {
  id: string;
  title: string;
  subtitle?: string;
  icon: string;
  shortcut?: string;
  action: () => void;
}

const commands = computed<CommandItem[]>(() => [
  {
    id: "nav-files",
    title: "Go to Files",
    subtitle: "Navigate to home directory",
    icon: "folder",
    shortcut: "G F",
    action: () => router.push("/files/"),
  },
  {
    id: "action-upload",
    title: "Upload Files",
    subtitle: "Upload files to current folder",
    icon: "file_upload",
    action: () => document.getElementById("upload-input")?.click(),
  },
  {
    id: "action-new-folder",
    title: "New Folder",
    subtitle: "Create a new folder in current directory",
    icon: "create_new_folder",
    action: () => layoutStore.showHover("newDir"),
  },
  {
    id: "action-new-file",
    title: "New File",
    subtitle: "Create a new text file",
    icon: "note_add",
    action: () => layoutStore.showHover("newFile"),
  },
  {
    id: "action-search",
    title: "Search Files",
    subtitle: "Search files and folders",
    icon: "search",
    shortcut: "Ctrl+F",
    action: () => layoutStore.showHover("search"),
  },
  {
    id: "theme-toggle",
    title: "Toggle Dark / Light Theme",
    subtitle: "Switch between light and dark modes",
    icon: "brightness_4",
    action: () => layoutStore.toggleTheme(),
  },
  {
    id: "nav-profile",
    title: "Profile & Security (2FA)",
    subtitle: "Open account and security settings",
    icon: "person",
    action: () => router.push("/settings/profile"),
  },
  {
    id: "nav-settings",
    title: "Global Settings",
    subtitle: "Open administration panel",
    icon: "settings",
    action: () => router.push("/settings/global"),
  },
  {
    id: "nav-shares",
    title: "Share Management",
    subtitle: "Manage active public links",
    icon: "share",
    action: () => router.push("/settings/shares"),
  },
  {
    id: "action-logout",
    title: "Log out",
    subtitle: "End current session",
    icon: "exit_to_app",
    action: () => logout(),
  },
]);

const filteredCommands = computed(() => {
  const q = query.value.toLowerCase().trim();
  if (!q) return commands.value;
  return commands.value.filter(
    (c) =>
      c.title.toLowerCase().includes(q) ||
      (c.subtitle && c.subtitle.toLowerCase().includes(q))
  );
});

const open = () => {
  isOpen.value = true;
  query.value = "";
  selectedIndex.value = 0;
  nextTick(() => {
    inputRef.value?.focus();
  });
};

const close = () => {
  isOpen.value = false;
};

const execute = (item: CommandItem) => {
  close();
  item.action();
};

const handleKeyDown = (e: KeyboardEvent) => {
  if (e.key === "ArrowDown") {
    e.preventDefault();
    if (selectedIndex.value < filteredCommands.value.length - 1) {
      selectedIndex.value++;
    } else {
      selectedIndex.value = 0;
    }
  } else if (e.key === "ArrowUp") {
    e.preventDefault();
    if (selectedIndex.value > 0) {
      selectedIndex.value--;
    } else {
      selectedIndex.value = filteredCommands.value.length - 1;
    }
  } else if (e.key === "Enter") {
    e.preventDefault();
    const item = filteredCommands.value[selectedIndex.value];
    if (item) {
      execute(item);
    }
  } else if (e.key === "Escape") {
    close();
  }
};

const handleGlobalKeyDown = (e: KeyboardEvent) => {
  if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === "k") {
    e.preventDefault();
    if (isOpen.value) {
      close();
    } else {
      open();
    }
  }
};

onMounted(() => {
  window.addEventListener("keydown", handleGlobalKeyDown);
});

onUnmounted(() => {
  window.removeEventListener("keydown", handleGlobalKeyDown);
});

defineExpose({ open, close });
</script>

<style scoped>
.command-palette-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(4px);
  z-index: 9999;
  display: flex;
  justify-content: center;
  align-items: flex-start;
  padding-top: 15vh;
}

.command-palette-modal {
  background: var(--card-bg, #ffffff);
  color: var(--text-color, #2c3e50);
  border-radius: 12px;
  width: 90%;
  max-width: 580px;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.35);
  overflow: hidden;
  border: 1px solid rgba(255, 255, 255, 0.1);
  display: flex;
  flex-direction: column;
}

.command-palette-header {
  display: flex;
  align-items: center;
  padding: 14px 18px;
  border-bottom: 1px solid rgba(128, 128, 128, 0.2);
  gap: 12px;
}

.command-palette-header i {
  color: #888;
  font-size: 24px;
}

.command-palette-header input {
  flex: 1;
  border: none;
  background: transparent;
  outline: none;
  font-size: 16px;
  color: inherit;
}

.command-palette-kbd {
  background: rgba(128, 128, 128, 0.15);
  border: 1px solid rgba(128, 128, 128, 0.3);
  border-radius: 4px;
  font-size: 11px;
  padding: 2px 6px;
  font-family: monospace;
}

.command-palette-list {
  max-height: 360px;
  overflow-y: auto;
  padding: 6px 0;
}

.command-palette-item {
  display: flex;
  align-items: center;
  padding: 10px 18px;
  cursor: pointer;
  gap: 14px;
  transition: background 0.15s ease;
}

.command-palette-item.selected {
  background: rgba(33, 150, 243, 0.15);
}

.command-palette-item i {
  font-size: 20px;
  color: #2196f3;
}

.command-palette-item-text {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.command-palette-item-text .title {
  font-weight: 500;
  font-size: 14px;
}

.command-palette-item-text .subtitle {
  font-size: 12px;
  opacity: 0.65;
}

.command-palette-item .shortcut {
  font-size: 11px;
  background: rgba(128, 128, 128, 0.15);
  padding: 2px 6px;
  border-radius: 4px;
  font-family: monospace;
}

.command-palette-empty {
  padding: 24px;
  text-align: center;
  opacity: 0.6;
  font-size: 14px;
}
</style>
