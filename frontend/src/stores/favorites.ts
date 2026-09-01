import { defineStore } from "pinia";
import { ref } from "vue";

export const useFavoritesStore = defineStore("favorites", () => {
  const STORAGE_KEY = "fb_favorites";
  const ALIAS_KEY = "fb_favorite_aliases";

  const loadFavorites = (): string[] => {
    try {
      const data = localStorage.getItem(STORAGE_KEY);
      return data ? JSON.parse(data) : [];
    } catch {
      return [];
    }
  };

  const loadAliases = (): Record<string, string> => {
    try {
      const data = localStorage.getItem(ALIAS_KEY);
      return data ? JSON.parse(data) : {};
    } catch {
      return {};
    }
  };

  const favorites = ref<string[]>(loadFavorites());
  const aliases = ref<Record<string, string>>(loadAliases());

  const save = () => {
    try {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(favorites.value));
      localStorage.setItem(ALIAS_KEY, JSON.stringify(aliases.value));
    } catch (e) {
      console.error("Failed to save favorites", e);
    }
  };

  const isFavorite = (path: string): boolean => {
    return favorites.value.includes(path);
  };

  const getFavoriteName = (path: string): string => {
    if (aliases.value[path] && aliases.value[path].trim() !== "") {
      return aliases.value[path];
    }
    return path.split("/").filter(Boolean).pop() || "/";
  };

  const renameFavorite = (path: string, newName: string) => {
    if (newName && newName.trim() !== "") {
      aliases.value[path] = newName.trim();
    } else {
      delete aliases.value[path];
    }
    save();
  };

  const toggleFavorite = (path: string) => {
    if (isFavorite(path)) {
      favorites.value = favorites.value.filter((p) => p !== path);
      delete aliases.value[path];
    } else {
      favorites.value.push(path);
    }
    save();
  };

  const removeFavorite = (path: string) => {
    favorites.value = favorites.value.filter((p) => p !== path);
    delete aliases.value[path];
    save();
  };

  return {
    favorites,
    aliases,
    isFavorite,
    getFavoriteName,
    renameFavorite,
    toggleFavorite,
    removeFavorite,
  };
});
