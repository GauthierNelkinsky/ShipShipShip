import type { Event, ParsedEvent } from "./types";
import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";
import { getLocale as getParaglideLocale } from "./paraglide/runtime";

/**
 * Utility function for merging class names with tailwind-merge
 */
export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

/**
 * Parse an event from the API to a more usable format
 */
export function parseEvent(event: Event): ParsedEvent {
  let media: string[] = [];

  try {
    media = event.media ? JSON.parse(event.media) : [];
  } catch {
    media = [];
  }

  return {
    ...event,
    media,
  };
}

/**
 * Format a date string for display in "10 Aug. 2025" format
 */
export function formatDate(dateString: string | null): string {
  if (!dateString) return "";

  try {
    const date = new Date(dateString);
    if (isNaN(date.getTime())) {
      return "";
    }
    return date
      .toLocaleDateString("en-GB", {
        year: "numeric",
        month: "short",
        day: "numeric",
      })
      .replace(/(\d+)\s+(\w+)\s+(\d+)/, function (match, day, month, year) {
        return day + " " + month + ". " + year;
      });
  } catch {
    return "";
  }
}

/**
 * Get status badge class name
 */
export function getStatusClass(status: string): string {
  // Generate a generic status class based on the status name
  const normalized = status.toLowerCase().replace(/\s+/g, "-");
  return `status-${normalized}`;
}

/**
 * Convert markdown to HTML (basic implementation)
 */
export function markdownToHtml(markdown: string): string {
  return markdown
    .replace(/^### (.*$)/gm, '<h3 class="text-base font-semibold mb-2">$1</h3>')
    .replace(/^## (.*$)/gm, '<h2 class="text-lg font-semibold mb-3">$1</h2>')
    .replace(/^# (.*$)/gm, '<h1 class="text-xl font-semibold mb-4">$1</h1>')
    .replace(/\*\*(.*?)\*\*/g, "<strong>$1</strong>")
    .replace(/\*(.*?)\*/g, "<em>$1</em>")
    .replace(
      /`(.*?)`/g,
      '<code class="bg-muted px-1 py-0.5 rounded text-sm">$1</code>',
    )
    .replace(/\n\n/g, '</p><p class="mb-4">')
    .replace(/\n/g, "<br>")
    .replace(/^/, '<p class="mb-4">')
    .replace(/$/, "</p>");
}

/**
 * Debounce function for search inputs
 */
export function debounce<T extends (...args: unknown[]) => unknown>(
  func: T,
  wait: number,
): (...args: Parameters<T>) => void {
  let timeout: ReturnType<typeof setTimeout>;
  return (...args: Parameters<T>) => {
    clearTimeout(timeout);
    timeout = setTimeout(() => func(...args), wait);
  };
}

/**
 * Validate email format
 */
export function isValidEmail(email: string): boolean {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  return emailRegex.test(email);
}

/**
 * Generate a random ID
 */
export function generateId(): string {
  return Math.random().toString(36).substr(2, 9);
}

/**
 * Toggle theme between light and dark
 */
export function toggleTheme(): void {
  const isDark = document.documentElement.classList.contains("dark");
  const newTheme = isDark ? "light" : "dark";

  document.documentElement.classList.toggle("dark", newTheme === "dark");

  if (typeof window !== "undefined") {
    localStorage.setItem("theme", newTheme);
  }
}

/**
 * Get current theme
 */
export function getCurrentTheme(): "light" | "dark" {
  if (typeof window === "undefined") return "light";

  const stored = localStorage.getItem("theme");
  if (stored === "dark" || stored === "light") return stored;

  return window.matchMedia("(prefers-color-scheme: dark)").matches
    ? "dark"
    : "light";
}

/**
 * Generate a URL-friendly slug from a title
 */
export function generateSlug(title: string): string {
  return title
    .toLowerCase()
    .replace(/[^a-z0-9 -]/g, "") // Remove special characters
    .replace(/\s+/g, "-") // Replace spaces with hyphens
    .replace(/-+/g, "-") // Replace multiple hyphens with single hyphen
    .trim()
    .replace(/^-+|-+$/g, ""); // Remove leading/trailing hyphens
}

/**
 * Initialize theme on page load
 */
export function initializeTheme(): void {
  if (typeof window === "undefined") return;

  const theme = getCurrentTheme();
  document.documentElement.classList.toggle("dark", theme === "dark");
}

/**
 * Get current locale
 * Gets it from paraglide runtime or document
 */
export function getCurrentLocale(): string {
    try {
      return getParaglideLocale();
    } catch {
      if (typeof document !== "undefined") {
        return document.documentElement.lang || "en";
      }
      return "en";
    }
}

/**
 * List of RTL locales
 */
const RTL_LOCALES = ["fa", "ar", "he", "ur"];

/**
 * Check if a locale is RTL (Right-to-Left)
 * If no locale is provided, automatically gets it from paraglide runtime or document
 */
export function isRTL(locale?: string): boolean {
  locale = locale || getCurrentLocale();
  return RTL_LOCALES.includes(locale.toLowerCase());
}

/**
 * Get text direction for a locale
 * If no locale is provided, automatically gets it from paraglide runtime or document
 */
export function getTextDirection(locale?: string): "ltr" | "rtl" {
  return isRTL(locale) ? "rtl" : "ltr";
}
