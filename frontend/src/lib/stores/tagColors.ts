import { writable } from 'svelte/store';
import { browser } from '$app/environment';

// Tag color storage key
const TAG_COLORS_KEY = 'shipshipship_tag_colors';

// Default color palette for tag generation
const DEFAULT_COLORS = [
    '#3b82f6', // Blue
    '#ef4444', // Red
    '#10b981', // Green
    '#f59e0b', // Yellow
    '#8b5cf6', // Purple
    '#ec4899', // Pink
    '#06b6d4', // Cyan
    '#84cc16', // Lime
    '#f97316', // Orange
    '#6366f1', // Indigo
];

// Create the store
function createTagColorStore() {
    const { subscribe, set, update } = writable<Record<string, string>>({});

    return {
        subscribe,

        // Initialize the store by loading from localStorage
        init: () => {
            if (browser) {
                try {
                    const saved = localStorage.getItem(TAG_COLORS_KEY);
                    if (saved) {
                        const colors = JSON.parse(saved);
                        set(colors);
                    }
                } catch (e) {
                    console.error('Error loading tag colors:', e);
                }
            }
        },

        // Save colors to localStorage
        save: (colors: Record<string, string>) => {
            if (browser) {
                try {
                    localStorage.setItem(TAG_COLORS_KEY, JSON.stringify(colors));
                    set(colors);
                } catch (e) {
                    console.error('Error saving tag colors:', e);
                }
            }
        },

        // Get color for a specific tag
        getColor: (tagName: string): string => {
            let currentColors: Record<string, string> = {};
            subscribe(colors => currentColors = colors)();

            if (currentColors[tagName]) {
                return currentColors[tagName];
            }

            // Generate consistent color based on tag name
            return generateTagColor(tagName);
        },

        // Set color for a specific tag
        setColor: (tagName: string, color: string) => {
            update(colors => {
                const newColors = { ...colors, [tagName]: color };
                if (browser) {
                    try {
                        localStorage.setItem(TAG_COLORS_KEY, JSON.stringify(newColors));
                    } catch (e) {
                        console.error('Error saving tag colors:', e);
                    }
                }
                return newColors;
            });
        },

        // Remove color for a specific tag
        removeColor: (tagName: string) => {
            update(colors => {
                const newColors = { ...colors };
                delete newColors[tagName];
                if (browser) {
                    try {
                        localStorage.setItem(TAG_COLORS_KEY, JSON.stringify(newColors));
                    } catch (e) {
                        console.error('Error saving tag colors:', e);
                    }
                }
                return newColors;
            });
        },

        // Update tag name (transfer color from old name to new name)
        renameTag: (oldName: string, newName: string) => {
            update(colors => {
                const newColors = { ...colors };
                if (newColors[oldName]) {
                    newColors[newName] = newColors[oldName];
                    delete newColors[oldName];
                }
                if (browser) {
                    try {
                        localStorage.setItem(TAG_COLORS_KEY, JSON.stringify(newColors));
                    } catch (e) {
                        console.error('Error saving tag colors:', e);
                    }
                }
                return newColors;
            });
        },

        // Bulk update colors
        updateColors: (newColors: Record<string, string>) => {
            update(colors => {
                const mergedColors = { ...colors, ...newColors };
                if (browser) {
                    try {
                        localStorage.setItem(TAG_COLORS_KEY, JSON.stringify(mergedColors));
                    } catch (e) {
                        console.error('Error saving tag colors:', e);
                    }
                }
                return mergedColors;
            });
        }
    };
}

// Generate consistent color based on tag name
function generateTagColor(tag: string): string {
    let hash = 0;
    for (let i = 0; i < tag.length; i++) {
        hash = tag.charCodeAt(i) + ((hash << 5) - hash);
    }
    return DEFAULT_COLORS[Math.abs(hash) % DEFAULT_COLORS.length];
}

// Export the store instance
export const tagColorStore = createTagColorStore();

// Helper function to get tag color (for use in components)
export function getTagColor(tagName: string): string {
    return tagColorStore.getColor(tagName);
}

// Helper function to get tag style (for use in templates)
export function getTagStyle(tagName: string): string {
    const color = tagColorStore.getColor(tagName);
    return `border-color: ${color}; background-color: ${color}20; color: ${color};`;
}
