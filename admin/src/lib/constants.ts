/**
 * Application constants
 *
 * APP_VERSION is injected at build time from package.json via Vite config
 */

/**
 * Application version
 * Sourced from package.json and injected at build time
 */
export const APP_VERSION = __APP_VERSION__;

/**
 * API configuration
 */
export const API_CONFIG = {
  POCKETBASE_URL: "https://api.shipshipship.io",
  BACKEND_URL: "/api",
} as const;

/**
 * Theme-related constants
 */
export const THEME_CONFIG = {
  MAX_SCREENSHOT_SIZE: 5 * 1024 * 1024, // 5MB
  SUPPORTED_IMAGE_FORMATS: ["image/jpeg", "image/png", "image/webp"],
  DEFAULT_THEME_ID: "shipshipship-template-default",
} as const;
