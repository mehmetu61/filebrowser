<template>
  <div class="breadcrumbs">
    <component
      :is="element"
      :to="base || ''"
      :aria-label="t('files.home')"
      :title="t('files.home')"
    >
      <i class="material-icons">home</i>
    </component>

    <span v-for="(link, index) in items" :key="index">
      <span class="chevron"
        ><i class="material-icons">keyboard_arrow_right</i></span
      >
      <component :is="element" :to="link.url">{{ link.name }}</component>
    </span>

    <button
      v-if="currentFolderPath"
      type="button"
      class="fav-toggle-btn"
      :class="{ isFav: favoritesStore.isFavorite(currentFolderPath) }"
      @click="favoritesStore.toggleFavorite(currentFolderPath)"
      :title="favoritesStore.isFavorite(currentFolderPath) ? 'Remove from favorites' : 'Add to favorites'"
    >
      <i class="material-icons">{{ favoritesStore.isFavorite(currentFolderPath) ? 'star' : 'star_border' }}</i>
    </button>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "vue-i18n";
import { useRoute } from "vue-router";
import { useFavoritesStore } from "@/stores/favorites";

const { t } = useI18n();
const route = useRoute();
const favoritesStore = useFavoritesStore();

const props = defineProps<{
  base: string;
  noLink?: boolean;
}>();

const currentFolderPath = computed(() => {
  if (!route.path.startsWith(props.base)) return "";
  const rel = route.path.replace(props.base, "");
  return rel || "/";
});

const items = computed(() => {
  const relativePath = route.path.replace(props.base, "");
  const parts = relativePath.split("/");

  if (parts[0] === "") {
    parts.shift();
  }

  if (parts[parts.length - 1] === "") {
    parts.pop();
  }

  const breadcrumbs: BreadCrumb[] = [];

  for (let i = 0; i < parts.length; i++) {
    if (i === 0) {
      breadcrumbs.push({
        name: decodeURIComponent(parts[i]),
        url: props.base + "/" + parts[i] + "/",
      });
    } else {
      breadcrumbs.push({
        name: decodeURIComponent(parts[i]),
        url: breadcrumbs[i - 1].url + parts[i] + "/",
      });
    }
  }

  if (breadcrumbs.length > 3) {
    while (breadcrumbs.length !== 4) {
      breadcrumbs.shift();
    }

    breadcrumbs[0].name = "...";
  }

  return breadcrumbs;
});

const element = computed(() => {
  if (props.noLink) {
    return "span";
  }

  return "router-link";
});
</script>

<style scoped>
.fav-toggle-btn {
  background: none;
  border: none;
  cursor: pointer;
  padding: 0 4px;
  display: inline-flex;
  align-items: center;
  color: var(--textSecondary, #94a3b8);
  transition: all 0.2s;
  vertical-align: middle;
}

.fav-toggle-btn:hover {
  color: #f59e0b;
  transform: scale(1.15);
}

.fav-toggle-btn.isFav {
  color: #f59e0b;
}

.fav-toggle-btn i {
  font-size: 1.25em;
}
</style>
