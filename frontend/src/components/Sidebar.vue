<template>
  <div v-show="active" @click="closeHovers" class="overlay"></div>
  <nav :class="{ active }">
    <template v-if="isLoggedIn">
      <button @click="toAccountSettings" class="action">
        <i class="material-icons">person</i>
        <span>{{ user.username }}</span>
      </button>
      <button
        class="action"
        @click="toRoot"
        :aria-label="$t('sidebar.myFiles')"
        :title="$t('sidebar.myFiles')"
      >
        <i class="material-icons">folder</i>
        <span>{{ $t("sidebar.myFiles") }}</span>
      </button>
      <button
        class="action"
        @click="toTrash"
        :aria-label="$t('sidebar.trash') || 'Papierkorb'"
        :title="$t('sidebar.trash') || 'Papierkorb'"
      >
        <i class="material-icons">delete_outline</i>
        <span>{{ $t("sidebar.trash") || "Papierkorb" }}</span>
      </button>

      <div v-if="favoritesStore.favorites.length > 0" class="favorites-section">
        <div class="sidebar-heading">⭐ Favorites</div>
        <div
          v-for="fav in favoritesStore.favorites"
          :key="fav"
          class="fav-item-row"
        >
          <button
            class="action fav-action"
            @click="toPath('/files' + fav)"
            :title="fav"
          >
            <i class="material-icons">folder_special</i>
            <span class="fav-name">{{ favoritesStore.getFavoriteName(fav) }}</span>
          </button>
          <div class="fav-actions-group">
            <button
              type="button"
              class="fav-icon-btn"
              title="Favorit umbenennen"
              @click.stop="promptRenameFavorite(fav)"
            >
              <i class="material-icons">edit</i>
            </button>
            <button
              type="button"
              class="fav-icon-btn fav-remove"
              title="Favorit entfernen"
              @click.stop="favoritesStore.removeFavorite(fav)"
            >
              <i class="material-icons">close</i>
            </button>
          </div>
        </div>
      </div>

      <div v-if="user.perm.create">
        <button
          @click="showHover('newDir')"
          class="action"
          :aria-label="$t('sidebar.newFolder')"
          :title="$t('sidebar.newFolder')"
        >
          <i class="material-icons">create_new_folder</i>
          <span>{{ $t("sidebar.newFolder") }}</span>
        </button>

        <button
          @click="showHover('newFile')"
          class="action"
          :aria-label="$t('sidebar.newFile')"
          :title="$t('sidebar.newFile')"
        >
          <i class="material-icons">note_add</i>
          <span>{{ $t("sidebar.newFile") }}</span>
        </button>
      </div>

      <div v-if="user.perm.admin">
        <button
          class="action"
          @click="toGlobalSettings"
          :aria-label="$t('sidebar.settings')"
          :title="$t('sidebar.settings')"
        >
          <i class="material-icons">settings_applications</i>
          <span>{{ $t("sidebar.settings") }}</span>
        </button>
      </div>
      <button
        v-if="canLogout"
        @click="logout"
        class="action"
        id="logout"
        :aria-label="$t('sidebar.logout')"
        :title="$t('sidebar.logout')"
      >
        <i class="material-icons">exit_to_app</i>
        <span>{{ $t("sidebar.logout") }}</span>
      </button>
    </template>
    <template v-else>
      <router-link
        v-if="!hideLoginButton"
        class="action"
        to="/login"
        :aria-label="$t('sidebar.login')"
        :title="$t('sidebar.login')"
      >
        <i class="material-icons">exit_to_app</i>
        <span>{{ $t("sidebar.login") }}</span>
      </router-link>

      <router-link
        v-if="signup"
        class="action"
        to="/login"
        :aria-label="$t('sidebar.signup')"
        :title="$t('sidebar.signup')"
      >
        <i class="material-icons">person_add</i>
        <span>{{ $t("sidebar.signup") }}</span>
      </router-link>
    </template>

    <div
      class="credits"
      v-if="isFiles && !disableUsedPercentage"
      style="width: 90%; margin: 2em 2.5em 3em 2.5em"
    >
      <progress-bar :val="usage.usedPercentage" size="small"></progress-bar>
      <br />
      {{ $t("sidebar.diskUsed", { used: usage.used, total: usage.total }) }}
    </div>

    <p class="credits" style="justify-content: center; text-align: center;">
      <span>
        <a @click="help">{{ $t("sidebar.help") }}</a>
      </span>
    </p>
  </nav>
</template>

<script>
import { reactive } from "vue";
import { mapActions, mapState } from "pinia";
import { useAuthStore } from "@/stores/auth";
import { useFileStore } from "@/stores/file";
import { useLayoutStore } from "@/stores/layout";
import { useFavoritesStore } from "@/stores/favorites";

import * as auth from "@/utils/auth";
import {
  version,
  signup,
  hideLoginButton,
  disableExternal,
  disableUsedPercentage,
  noAuth,
  logoutPage,
  loginPage,
} from "@/utils/constants";
import { files as api } from "@/api";
import ProgressBar from "@/components/ProgressBar.vue";
import prettyBytes from "pretty-bytes";

const USAGE_DEFAULT = { used: "0 B", total: "0 B", usedPercentage: 0 };

export default {
  name: "sidebar",
  setup() {
    const usage = reactive(USAGE_DEFAULT);
    const favoritesStore = useFavoritesStore();
    return { usage, favoritesStore, usageAbortController: new AbortController() };
  },
  components: {
    ProgressBar,
  },
  inject: ["$showError"],
  computed: {
    ...mapState(useAuthStore, ["user", "isLoggedIn"]),
    ...mapState(useFileStore, ["isFiles", "reload"]),
    ...mapState(useLayoutStore, ["currentPromptName"]),
    active() {
      return this.currentPromptName === "sidebar";
    },
    signup: () => signup,
    hideLoginButton: () => hideLoginButton,
    version: () => version,
    disableExternal: () => disableExternal,
    disableUsedPercentage: () => disableUsedPercentage,
    canLogout: () => !noAuth && (loginPage || logoutPage !== "/login"),
  },
  methods: {
    ...mapActions(useLayoutStore, ["closeHovers", "showHover"]),
    abortOngoingFetchUsage() {
      this.usageAbortController.abort();
    },
    async fetchUsage() {
      const path = this.$route.path.endsWith("/")
        ? this.$route.path
        : this.$route.path + "/";
      let usageStats = USAGE_DEFAULT;
      if (this.disableUsedPercentage) {
        return Object.assign(this.usage, usageStats);
      }
      try {
        this.abortOngoingFetchUsage();
        this.usageAbortController = new AbortController();
        const usage = await api.usage(path, this.usageAbortController.signal);
        usageStats = {
          used: prettyBytes(usage.used, { binary: true }),
          total: prettyBytes(usage.total, { binary: true }),
          usedPercentage: Math.round((usage.used / usage.total) * 100),
        };
      } finally {
        return Object.assign(this.usage, usageStats);
      }
    },
    toPath(destPath) {
      this.$router.push({ path: destPath });
      this.closeHovers();
    },
    toRoot() {
      this.$router.push({ path: "/files" });
      this.closeHovers();
    },
    toTrash() {
      this.$router.push({ path: "/trash" });
      this.closeHovers();
    },
    toAccountSettings() {
      this.$router.push({ path: "/settings/profile" });
      this.closeHovers();
    },
    toGlobalSettings() {
      this.$router.push({ path: "/settings/global" });
      this.closeHovers();
    },
    promptRenameFavorite(fav) {
      const currentName = this.favoritesStore.getFavoriteName(fav);
      const newName = window.prompt("Favoriten-Namen anpassen:", currentName);
      if (newName !== null) {
        this.favoritesStore.renameFavorite(fav, newName);
      }
    },
    help() {
      this.showHover("help");
    },
    logout: auth.logout,
  },
  watch: {
    $route: {
      handler(to) {
        if (to.path.includes("/files")) {
          this.fetchUsage();
        }
      },
      immediate: true,
    },
  },
  unmounted() {
    this.abortOngoingFetchUsage();
  },
};
</script>

<style scoped>
.favorites-section {
  padding: 0.2em 0 0.4em 0;
}

.sidebar-heading {
  padding: 0.3em 1.2em;
  font-size: 0.75em;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  opacity: 0.7;
}

.fav-item-row {
  display: flex;
  align-items: center;
  position: relative;
  transition: background 0.15s;
}

.fav-item-row:hover {
  background: var(--hover, rgba(0, 0, 0, 0.04));
}

.fav-action {
  font-size: 0.9em;
  flex: 1;
  min-width: 0;
  text-align: left;
}

.fav-name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.fav-actions-group {
  display: flex;
  align-items: center;
  padding-right: 0.5em;
  opacity: 0;
  transition: opacity 0.15s;
}

.fav-item-row:hover .fav-actions-group {
  opacity: 1;
}

.fav-icon-btn {
  background: transparent;
  border: none;
  cursor: pointer;
  padding: 4px;
  color: var(--action, #64748b);
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.fav-icon-btn:hover {
  background: rgba(0, 0, 0, 0.08);
  color: var(--blue, #2563eb);
}

.fav-icon-btn.fav-remove:hover {
  color: var(--red, #ef4444);
}

.fav-icon-btn i {
  font-size: 1.1em;
}
</style>
