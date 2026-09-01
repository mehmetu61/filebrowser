import { defineStore } from "pinia";

export type ThemeMode = "system" | "light" | "dark" | "oled";

export const useLayoutStore = defineStore("layout", {
  state: (): {
    loading: boolean;
    prompts: PopupProps[];
    showShell: boolean | null;
    theme: ThemeMode;
  } => ({
    loading: false,
    prompts: [],
    showShell: false,
    theme: (localStorage.getItem("theme") as ThemeMode) || "system",
  }),
  getters: {
    currentPrompt(state) {
      return state.prompts.length > 0
        ? state.prompts[state.prompts.length - 1]
        : null;
    },
    currentPromptName(): string | null | undefined {
      return this.currentPrompt?.prompt;
    },
  },
  actions: {
    setTheme(theme: ThemeMode) {
      this.theme = theme;
      localStorage.setItem("theme", theme);
      this.applyTheme();
    },

    toggleTheme() {
      const nextTheme: Record<ThemeMode, ThemeMode> = {
        system: "dark",
        dark: "oled",
        oled: "light",
        light: "system",
      };
      this.setTheme(nextTheme[this.theme] || "dark");
    },

    applyTheme() {
      const root = document.documentElement;
      root.classList.remove("dark", "oled");

      if (this.theme === "dark") {
        root.classList.add("dark");
      } else if (this.theme === "oled") {
        root.classList.add("oled");
      } else if (this.theme === "system") {
        if (window.matchMedia && window.matchMedia("(prefers-color-scheme: dark)").matches) {
          root.classList.add("dark");
        }
      }
    },
    // no context as first argument, use `this` instead
    toggleShell() {
      this.showShell = !this.showShell;
    },
    setCloseOnPrompt(closeFunction: () => Promise<string>, onPrompt: string) {
      const prompt = this.prompts.find((prompt) => prompt.prompt === onPrompt);
      if (prompt) {
        prompt.close = closeFunction;
      }
    },
    showHover(value: PopupProps | string) {
      if (typeof value !== "object") {
        this.prompts.push({
          prompt: value,
          confirm: null,
          action: undefined,
          saveAction: undefined,
          props: null,
          close: null,
        });
        return;
      }

      this.prompts.push({
        prompt: value.prompt,
        confirm: value?.confirm,
        action: value?.action,
        saveAction: value?.saveAction,
        props: value?.props,
        close: value?.close,
      });
    },
    showError() {
      this.prompts.push({
        prompt: "error",
        confirm: null,
        action: undefined,
        props: null,
        close: null,
      });
    },
    showSuccess() {
      this.prompts.push({
        prompt: "success",
        confirm: null,
        action: undefined,
        props: null,
        close: null,
      });
    },
    closeHovers() {
      this.prompts.pop()?.close?.();
    },
    // easily reset state using `$reset`
    clearLayout() {
      this.$reset();
    },
  },
});
