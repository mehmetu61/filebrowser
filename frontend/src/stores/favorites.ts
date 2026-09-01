import { defineStore } from "pinia";
import { ref } from "vue";

export const useFavoritesStore = defineStore("favorites", () => {
  const STORAGE_KEY = "fb_favorites";

  const loadFavorites = (): string[] => {
    try {
      const data = localStorage.getItem(STORAGE_KEY);
      return data ? JSON.parse(data) : [];
    } catch {
      return [];
    }
  };

  const favorites = ref<string[]>(loadFavorites());

  const save = () => {
    try {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(favorites.value));
    } catch (e) {
      console.error("Failed to save favorites", e);
    }
  };

  const isFavorite = (path: string): boolean => {
    return favorites.value.includes(path);
  };

  const toggleFavorite = (path: string) => {
    if (isFavorite(path)) {
      favorites.value = favorites.value.filter((p) => p !== path);
    } else {
      favorites.value.push(path);
    }
    save();
  };

  const removeFavorite = (path: string) => {
    favorites.value = favorites.value.filter((p) => p !== path);
    save();
  };

  return {
    favorites,
    isFavorite,
    toggleFavorite,
    removeFavorite,
  };
});
