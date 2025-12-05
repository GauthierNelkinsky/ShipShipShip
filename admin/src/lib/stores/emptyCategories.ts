import { writable } from 'svelte/store';
import { api } from '$lib/api';

interface EmptyCategoriesStore {
    hasEmptyCategories: boolean;
    loading: boolean;
    lastChecked: number | null;
}

function createEmptyCategoriesStore() {
    const { subscribe, set, update } = writable<EmptyCategoriesStore>({
        hasEmptyCategories: false,
        loading: false,
        lastChecked: null
    });

    return {
        subscribe,
        check: async () => {
            update(state => ({ ...state, loading: true }));

            try {
                const manifestData = await api.getThemeManifest();
                const manifest = manifestData.manifest;

                if (!manifest || !manifest.categories) {
                    set({
                        hasEmptyCategories: false,
                        loading: false,
                        lastChecked: Date.now()
                    });
                    return;
                }

                const mappingsData = await api.getStatusMappings();
                const mappings = mappingsData.mappings || [];

                // Check if any category has no statuses assigned to it
                const emptyCategories = manifest.categories.filter((category: any) => {
                    const categoryHasStatuses = mappings.some(
                        (mapping: any) => mapping.category_id === category.id
                    );
                    return !categoryHasStatuses;
                });

                set({
                    hasEmptyCategories: emptyCategories.length > 0,
                    loading: false,
                    lastChecked: Date.now()
                });
            } catch (err) {
                console.error('Error checking empty categories:', err);
                set({
                    hasEmptyCategories: false,
                    loading: false,
                    lastChecked: Date.now()
                });
            }
        },
        reset: () => {
            set({
                hasEmptyCategories: false,
                loading: false,
                lastChecked: null
            });
        }
    };
}

export const emptyCategoriesStore = createEmptyCategoriesStore();
