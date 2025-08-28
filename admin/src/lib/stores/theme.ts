import { writable } from "svelte/store";
import { browser } from "$app/environment";

export type Theme = "light" | "dark";
export type ThemePreference = "light" | "dark";

function createThemeStore() {
  const { subscribe, set, update } = writable<Theme>("light");

  let mediaQuery: MediaQueryList | null = null;
  let mediaQueryHandler: ((e: MediaQueryListEvent) => void) | null = null;

  return {
    subscribe,
    init: () => {
      if (!browser) return;

      // Get stored preference or default to light
      const stored = localStorage.getItem("theme") as ThemePreference | null;

      let initialTheme: Theme;
      if (stored === "light" || stored === "dark") {
        initialTheme = stored;
      } else {
        // No stored preference or was "system", default to light
        initialTheme = "light";
        localStorage.setItem("theme", "light");
      }

      // Apply theme to document
      document.documentElement.classList.toggle(
        "dark",
        initialTheme === "dark",
      );
      set(initialTheme);

      // No longer listening for system theme changes since system option is removed
    },
    toggle: () => {
      update((current) => {
        const newTheme = current === "light" ? "dark" : "light";

        if (browser) {
          document.documentElement.classList.toggle(
            "dark",
            newTheme === "dark",
          );
          localStorage.setItem("theme", newTheme);
        }

        return newTheme;
      });
    },
    setPreference: (preference: ThemePreference) => {
      if (!browser) return;

      // Set explicit theme
      localStorage.setItem("theme", preference);
      document.documentElement.classList.toggle("dark", preference === "dark");
      set(preference);
    },
    getPreference: (): ThemePreference => {
      if (!browser) return "light";
      const stored = localStorage.getItem("theme");
      if (stored === "light" || stored === "dark") {
        return stored;
      }
      // If no valid stored preference or was "system", default to light
      localStorage.setItem("theme", "light");
      return "light";
    },
    set: (theme: Theme) => {
      if (browser) {
        document.documentElement.classList.toggle("dark", theme === "dark");
        localStorage.setItem("theme", theme);
      }
      set(theme);
    },
  };
}

export const theme = createThemeStore();
