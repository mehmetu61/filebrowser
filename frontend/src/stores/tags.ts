import { defineStore } from "pinia";
import { ref } from "vue";

export interface ColorTag {
  id: string;
  name: string;
  color: string;
  bg: string;
}

export const TAG_COLORS: ColorTag[] = [
  { id: "red", name: "Red", color: "#ef4444", bg: "rgba(239, 68, 68, 0.15)" },
  { id: "orange", name: "Orange", color: "#f97316", bg: "rgba(249, 115, 22, 0.15)" },
  { id: "yellow", name: "Yellow", color: "#eab308", bg: "rgba(234, 179, 8, 0.15)" },
  { id: "green", name: "Green", color: "#22c55e", bg: "rgba(34, 197, 94, 0.15)" },
  { id: "blue", name: "Blue", color: "#3b82f6", bg: "rgba(59, 130, 246, 0.15)" },
  { id: "purple", name: "Purple", color: "#a855f7", bg: "rgba(168, 85, 247, 0.15)" },
];

export const useTagsStore = defineStore("tags", () => {
  const STORAGE_KEY = "fb_file_tags";

  const loadTags = (): Record<string, string[]> => {
    try {
      const data = localStorage.getItem(STORAGE_KEY);
      return data ? JSON.parse(data) : {};
    } catch {
      return {};
    }
  };

  const tagsMap = ref<Record<string, string[]>>(loadTags());
  const activeFilterTag = ref<string | null>(null);

  const save = () => {
    try {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(tagsMap.value));
    } catch (e) {
      console.error("Failed to save file tags", e);
    }
  };

  const getTags = (path: string): string[] => {
    return tagsMap.value[path] || [];
  };

  const toggleTag = (path: string, tagId: string) => {
    const current = tagsMap.value[path] || [];
    if (current.includes(tagId)) {
      const updated = current.filter((t) => t !== tagId);
      if (updated.length === 0) {
        delete tagsMap.value[path];
      } else {
        tagsMap.value[path] = updated;
      }
    } else {
      tagsMap.value[path] = [...current, tagId];
    }
    save();
  };

  const removeTag = (path: string, tagId: string) => {
    const current = tagsMap.value[path];
    if (!current) return;
    const updated = current.filter((t) => t !== tagId);
    if (updated.length === 0) {
      delete tagsMap.value[path];
    } else {
      tagsMap.value[path] = updated;
    }
    save();
  };

  const setFilterTag = (tagId: string | null) => {
    activeFilterTag.value = tagId;
  };

  return {
    tagsMap,
    activeFilterTag,
    getTags,
    toggleTag,
    removeTag,
    setFilterTag,
  };
});
