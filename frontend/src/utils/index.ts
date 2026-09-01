import { partial } from "filesize";

/**
 * Formats filesize as KiB/MiB/...
 */
export const filesize = partial({ base: 2 });

export const vClickOutside = {
  created(el: HTMLElement, binding: any) {
    el.clickOutsideEvent = (event: Event) => {
      const target = event.target;

      if (target instanceof Node) {
        if (!el.contains(target)) {
          binding.value(event);
        }
      }
    };

    document.addEventListener("click", el.clickOutsideEvent);
  },

  unmounted(el: HTMLElement) {
    if (el.clickOutsideEvent) {
      document.removeEventListener("click", el.clickOutsideEvent);
    }
  },
};

const TEXT_EXTENSIONS = new Set([
  ".json", ".json5", ".jsonc", ".jsonl", ".geojson",
  ".php", ".phtml", ".php3", ".php4", ".php5", ".php7", ".phps",
  ".html", ".htm", ".xhtml", ".vue", ".svelte", ".jsx", ".tsx",
  ".js", ".mjs", ".cjs", ".ts", ".mts", ".cts",
  ".css", ".scss", ".sass", ".less", ".styl",
  ".yaml", ".yml", ".toml", ".xml", ".sql", ".env", ".ini", ".conf", ".cfg", ".cnf",
  ".properties", ".prefs", ".log", ".txt", ".md", ".markdown", ".rst", ".tex",
  ".py", ".pyw", ".go", ".rs", ".c", ".cpp", ".cc", ".cxx", ".h", ".hpp", ".hxx",
  ".java", ".kt", ".kts", ".cs", ".swift", ".rb", ".lua", ".sh", ".bash", ".zsh",
  ".fish", ".bat", ".cmd", ".ps1", ".pl", ".pm", ".r", ".dart", ".scala", ".groovy",
  ".erl", ".ex", ".exs", ".hs", ".clj", ".nim", ".v", ".zig", ".csv", ".tsv",
  ".graphql", ".proto", ".diff", ".patch", ".dockerfile", ".makefile", ".lock",
]);

const TEXT_FILENAMES = new Set([
  "dockerfile", "makefile", "rakefile", "gemfile", "procfile", "vagrantfile", "caddyfile",
  ".gitignore", ".gitattributes", ".dockerignore", ".editorconfig", ".npmrc", ".nvmrc", ".env",
]);

export function isTextFile(fileName: string, ext?: string): boolean {
  if (!fileName) return false;
  const lowerName = fileName.toLowerCase();
  if (TEXT_FILENAMES.has(lowerName)) return true;
  const lastDot = lowerName.lastIndexOf(".");
  const fileExt = ext ? ext.toLowerCase() : (lastDot !== -1 ? lowerName.substring(lastDot) : "");
  return TEXT_EXTENSIONS.has(fileExt);
}

