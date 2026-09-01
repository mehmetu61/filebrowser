<template>
  <div class="card floating">
    <div class="card-content">
      <p v-if="!this.isListing || selectedCount === 1">
        {{ $t("prompts.deleteMessageSingle") }}
      </p>
      <p v-else>
        {{ $t("prompts.deleteMessageMultiple", { count: selectedCount }) }}
      </p>

      <div style="margin-top: 1em; display: flex; align-items: center; gap: 0.5em; font-size: 0.9em; opacity: 0.85;">
        <input type="checkbox" id="perm-delete" v-model="permanent" />
        <label for="perm-delete" style="cursor: pointer;">Endgültig löschen (nicht in Papierkorb)</label>
      </div>
    </div>
    <div class="card-action">
      <button
        @click="closeHovers"
        class="button button--flat button--grey"
        :aria-label="$t('buttons.cancel')"
        :title="$t('buttons.cancel')"
        tabindex="2"
      >
        {{ $t("buttons.cancel") }}
      </button>
      <button
        id="focus-prompt"
        @click="submit"
        class="button button--flat button--red"
        :aria-label="$t('buttons.delete')"
        :title="$t('buttons.delete')"
        tabindex="1"
      >
        {{ permanent ? "Endgültig löschen" : "In Papierkorb" }}
      </button>
    </div>
  </div>
</template>

<script>
import { mapActions, mapState, mapWritableState } from "pinia";
import { files as api } from "@/api";
import buttons from "@/utils/buttons";
import { useFileStore } from "@/stores/file";
import { useLayoutStore } from "@/stores/layout";

export default {
  name: "delete",
  inject: ["$showError"],
  data() {
    return {
      permanent: false,
    };
  },
  computed: {
    ...mapState(useFileStore, [
      "isListing",
      "selectedCount",
      "req",
      "selected",
    ]),
    ...mapState(useLayoutStore, ["currentPrompt"]),
    ...mapWritableState(useFileStore, ["reload", "preselect"]),
  },
  methods: {
    ...mapActions(useLayoutStore, ["closeHovers"]),
    submit: async function () {
      buttons.loading("delete");
      const suffix = this.permanent ? "?permanent=true" : "";

      try {
        if (!this.isListing) {
          await api.remove(this.$route.path + suffix);
          buttons.success("delete");

          this.currentPrompt?.confirm();
          this.closeHovers();
          return;
        }

        this.closeHovers();

        if (this.selectedCount === 0) {
          return;
        }

        const promises = [];
        for (const index of this.selected) {
          promises.push(api.remove(this.req.items[index].url + suffix));
        }

        await Promise.all(promises);
        buttons.success("delete");

        const nearbyItem =
          this.req.items[Math.max(0, Math.min(this.selected) - 1)];

        this.preselect = nearbyItem?.path;

        this.reload = true;
      } catch (e) {
        buttons.done("delete");
        this.$showError(e);
        if (this.isListing) this.reload = true;
      }
    },
  },
};
</script>
