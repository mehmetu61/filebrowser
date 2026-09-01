<template>
  <base-modal v-if="modal != null" :prompt="currentPromptName" @closed="close">
    <keep-alive>
      <component
        :is="modal"
        v-bind="layoutStore.currentPrompt?.props"
        @close="close"
        @done="fileStore.reload = true"
      />
    </keep-alive>
  </base-modal>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { storeToRefs } from "pinia";
import { useLayoutStore } from "@/stores/layout";
import { useFileStore } from "@/stores/file";

import BaseModal from "./BaseModal.vue";
import Help from "./Help.vue";
import Info from "./Info.vue";
import Delete from "./Delete.vue";
import DeleteUser from "./DeleteUser.vue";
import Download from "./Download.vue";
import Rename from "./Rename.vue";
import Move from "./Move.vue";
import Copy from "./Copy.vue";
import NewFile from "./NewFile.vue";
import NewDir from "./NewDir.vue";
import Replace from "./Replace.vue";
import Share from "./Share.vue";
import ShareDelete from "./ShareDelete.vue";
import Upload from "./Upload.vue";
import DiscardEditorChanges from "./DiscardEditorChanges.vue";
import ResolveConflict from "./ResolveConflict.vue";
import CurrentPassword from "./CurrentPassword.vue";
import BatchRenameModal from "@/components/files/BatchRenameModal.vue";
import ChecksumModal from "@/components/files/ChecksumModal.vue";
import ArchiveExtractModal from "@/components/files/ArchiveExtractModal.vue";
import ArchiveCompressModal from "@/components/files/ArchiveCompressModal.vue";

const layoutStore = useLayoutStore();
const fileStore = useFileStore();

const { currentPromptName } = storeToRefs(layoutStore);

const components = new Map<string, any>([
  ["info", Info],
  ["help", Help],
  ["delete", Delete],
  ["rename", Rename],
  ["batchRename", BatchRenameModal],
  ["checksum", ChecksumModal],
  ["archiveExtract", ArchiveExtractModal],
  ["archiveCompress", ArchiveCompressModal],
  ["move", Move],
  ["copy", Copy],
  ["newFile", NewFile],
  ["newDir", NewDir],
  ["download", Download],
  ["replace", Replace],
  ["share", Share],
  ["upload", Upload],
  ["share-delete", ShareDelete],
  ["deleteUser", DeleteUser],
  ["discardEditorChanges", DiscardEditorChanges],
  ["resolve-conflict", ResolveConflict],
  ["current-password", CurrentPassword],
]);

const modal = computed(() => {
  const modal = components.get(currentPromptName.value!);
  if (!modal) null;

  return modal;
});

const close = () => {
  if (!layoutStore.currentPrompt) return;
  layoutStore.closeHovers();
};
</script>
