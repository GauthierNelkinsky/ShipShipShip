<script lang="ts">
    import { createEventDispatcher, onMount } from "svelte";
    import { api } from "$lib/api";
    import { X } from "lucide-svelte";
    import { Button, Input } from "$lib/components/ui";
    import * as m from "$lib/paraglide/messages";

    const dispatch = createEventDispatcher();

    export let isOpen = false;

    let loading = false;
    let error = "";
    let loadingCategories = false;

    // Form fields
    let name = "";
    let selectedCategoryId = ""; // Empty string means no category selected

    // Categories from theme
    let categories: Array<{
        id: string;
        label: string;
        description: string;
        multiple: boolean;
        order: number;
    }> = [];
    let _themeId = "";
    let _themeName = "";

    // Status mappings to validate category constraints
    let categoryMappings: Map<string, number[]> = new Map();

    onMount(async () => {
        await loadThemeCategories();
        await loadStatusMappings();
    });

    async function loadThemeCategories() {
        try {
            loadingCategories = true;
            const manifest = await api.getThemeManifest();
            if (manifest.success && manifest.manifest) {
                _themeId = manifest.manifest.id;
                _themeName = manifest.manifest.name;
                categories = manifest.manifest.categories || [];
                // Don't pre-select - leave it optional
            }
        } catch (err) {
            console.error("Failed to load theme categories:", err);
            categories = [];
        } finally {
            loadingCategories = false;
        }
    }

    async function loadStatusMappings() {
        try {
            const mappingsData = await api.getStatusMappings();
            if (mappingsData.success && mappingsData.mappings) {
                // Build a map of category_id -> array of status_ids
                const mappings = new Map<string, number[]>();
                for (const mapping of mappingsData.mappings) {
                    if (!mappings.has(mapping.category_id)) {
                        mappings.set(mapping.category_id, []);
                    }
                    mappings.get(mapping.category_id)?.push(mapping.status_id);
                }
                categoryMappings = mappings;
            }
        } catch (err) {
            console.error("Failed to load status mappings:", err);
        }
    }

    function resetForm() {
        name = "";
        // Reset to no category selected
        selectedCategoryId = "";
        error = "";
    }

    function closeModal() {
        resetForm();
        isOpen = false;
        dispatch("close");
    }

    async function handleSubmit() {
        if (!name.trim()) {
            error = m.status_modal_name_required();
            return;
        }

        // Validate category constraints
        if (selectedCategoryId) {
            const selectedCategory = categories.find(
                (c) => c.id === selectedCategoryId,
            );
            if (selectedCategory && selectedCategory.multiple !== true) {
                // Check if this category already has a status mapped
                const existingMappings =
                    categoryMappings.get(selectedCategoryId) || [];
                if (existingMappings.length > 0) {
                    error = `Category '${selectedCategory.label}' does not allow multiple statuses and already has a status mapped to it.`;
                    return;
                }
            }
        }

        try {
            loading = true;
            error = "";

            const statusData: {
                display_name: string;
                category_id?: string;
            } = {
                display_name: name.trim(),
            };

            // Add category mapping if selected
            if (selectedCategoryId) {
                statusData.category_id = selectedCategoryId;
            }

            const newStatus = await api.createStatus(statusData);
            dispatch("created", newStatus);
            closeModal();
        } catch (err) {
            error =
                err instanceof Error
                    ? err.message
                    : m.status_modal_create_failed();
        } finally {
            loading = false;
        }
    }

    function handleBackdropClick() {
        closeModal();
    }

    function handleKeydown(e: KeyboardEvent) {
        if (e.key === "Escape") {
            closeModal();
        }
        if ((e.metaKey || e.ctrlKey) && e.key === "Enter") {
            handleSubmit();
        }
    }
</script>

<svelte:window on:keydown={handleKeydown} />

{#if isOpen}
    <!-- Modal backdrop -->
    <div
        class="fixed inset-0 bg-black/50 z-[60] flex items-center justify-center p-4"
        on:click={handleBackdropClick}
        on:keydown={(e) => {
            if (e.key === "Escape") closeModal();
        }}
        role="dialog"
        aria-modal="true"
        aria-labelledby="modal-title"
        tabindex="-1"
    >
        <!-- Modal content - more compact -->
        <div
            class="bg-background rounded-lg shadow-xl w-full max-w-md"
            on:click|stopPropagation
            role="none"
        >
            <!-- Modal header - compact -->
            <div
                class="flex items-center justify-between px-4 py-3 border-b border-border"
            >
                <h2 id="modal-title" class="text-base font-semibold">
                    {m.status_modal_title()}
                </h2>
                <button
                    on:click={closeModal}
                    class="text-muted-foreground hover:text-foreground transition-colors rounded-sm p-1 hover:bg-muted"
                    disabled={loading}
                    aria-label="Close"
                >
                    <X class="h-4 w-4" />
                </button>
            </div>

            <!-- Modal body - compact -->
            <div class="p-4 space-y-3">
                {#if error}
                    <div
                        class="p-2 text-xs bg-destructive/10 text-destructive rounded border border-destructive/20"
                    >
                        {error}
                    </div>
                {/if}

                <!-- Name input -->
                <div class="space-y-1.5">
                    <label
                        for="status-name"
                        class="text-xs font-medium block text-muted-foreground"
                    >
                        {m.status_modal_status_name()}
                    </label>
                    <Input
                        id="status-name"
                        type="text"
                        placeholder={m.status_modal_name_placeholder()}
                        bind:value={name}
                        disabled={loading}
                        class="w-full h-9 text-sm"
                        autofocus
                    />
                </div>

                <!-- Category selector -->
                {#if categories.length > 0}
                    <div class="space-y-1.5">
                        <label
                            for="status-category"
                            class="text-xs font-medium block text-muted-foreground"
                        >
                            {m.status_modal_public_category()}{" "}
                            <span class="text-muted-foreground/60"
                                >({m.status_modal_optional()})</span
                            >
                        </label>
                        <select
                            id="status-category"
                            value={selectedCategoryId}
                            on:change={(e) => {
                                selectedCategoryId = e.currentTarget.value;
                            }}
                            disabled={loading || loadingCategories}
                            class="w-full h-9 text-sm rounded-md border border-input bg-background px-3 py-1 ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                        >
                            <option value="">
                                {m.status_modal_no_category()}
                            </option>
                            {#each categories as category}
                                {@const existingMappings =
                                    categoryMappings.get(category.id) || []}
                                {@const isSingleAndOccupied =
                                    category.multiple !== true &&
                                    existingMappings.length > 0}
                                <option
                                    value={category.id}
                                    disabled={isSingleAndOccupied}
                                >
                                    {category.label}{isSingleAndOccupied
                                        ? " (already has a status)"
                                        : ""}
                                </option>
                            {/each}
                        </select>
                        {#if selectedCategoryId}
                            {@const selectedCategory = categories.find(
                                (c) => c.id === selectedCategoryId,
                            )}
                            {#if selectedCategory}
                                <p class="text-xs text-muted-foreground italic">
                                    {selectedCategory.description}
                                </p>
                            {/if}
                        {:else}
                            <p class="text-xs text-muted-foreground">
                                {m.status_modal_category_description()}
                            </p>
                        {/if}
                    </div>
                {:else if loadingCategories}
                    <div class="text-xs text-muted-foreground">
                        {m.status_modal_loading_categories()}
                    </div>
                {:else}
                    <div
                        class="p-2 text-xs bg-muted/50 text-muted-foreground rounded border border-border"
                    >
                        {m.status_modal_no_theme_configured()}
                    </div>
                {/if}
            </div>

            <!-- Modal footer - compact -->
            <div
                class="flex items-center justify-end gap-2 px-4 py-3 bg-muted/20 border-t border-border"
            >
                <Button
                    variant="ghost"
                    size="sm"
                    on:click={closeModal}
                    disabled={loading}
                    class="h-8 text-sm"
                >
                    {m.status_modal_cancel()}
                </Button>
                <Button
                    size="sm"
                    on:click={handleSubmit}
                    disabled={loading || !name.trim()}
                    class="h-8 text-sm"
                >
                    {#if loading}
                        <div
                            class="animate-spin rounded-full h-3 w-3 border-2 border-background border-t-transparent mr-1.5"
                        ></div>
                        {m.status_modal_creating()}
                    {:else}
                        {m.status_modal_create_status()}
                    {/if}
                </Button>
            </div>
        </div>
    </div>
{/if}
