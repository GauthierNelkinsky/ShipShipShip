import { writable } from "svelte/store";
import type { ProjectSettings } from "$lib/types";
import { api } from "$lib/api";

// Default settings
const defaultSettings: ProjectSettings = {
  title: "Changelog",
  favicon_url: "",
  website_url: "",
  created_at: "",
  updated_at: "",
};

// Create writable store
export const settings = writable<ProjectSettings>(defaultSettings);

// Load settings from API
export async function loadSettings() {
  try {
    const projectSettings = await api.getSettings();
    settings.set(projectSettings);

    return projectSettings;
  } catch (error) {
    console.error("Failed to load settings:", error);
    return defaultSettings;
  }
}

// Update settings
export async function updateSettings(updates: Partial<ProjectSettings>) {
  try {
    const updatedSettings = await api.updateSettings(updates);
    settings.set(updatedSettings);

    return updatedSettings;
  } catch (error) {
    console.error("Failed to update settings:", error);
    throw error;
  }
}

// Initialize settings on app load
if (typeof window !== "undefined") {
  loadSettings();
}
