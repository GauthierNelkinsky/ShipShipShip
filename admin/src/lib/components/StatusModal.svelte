<script lang="ts">
    import { createEventDispatcher, onMount } from "svelte";
    import { api } from "$lib/api";
    import { X } from "lucide-svelte";
    import { Button, Input } from "$lib/components/ui";

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
        order: number;
    }> = [];
    let _themeId = "";
    let _themeName = "";

    onMount(async () => {
        await loadThemeCategories();
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
            error = "Status name is required";
            return;
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
                err instanceof Error ? err.message : "Failed to create status";
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
                    Create Status
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
                        Status Name
                    </label>
                    <Input
                        id="status-name"
                        type="text"
                        placeholder="e.g., In Progress, Under Review"
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
                            Public Category <span
                                class="text-muted-foreground/60"
                                >(Optional)</span
                            >
                        </label>
                        <select
                            id="status-category"
                            bind:value={selectedCategoryId}
                            disabled={loading || loadingCategories}
                            class="w-full h-9 text-sm rounded-md border border-input bg-background px-3 py-1 ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                        >
                            <option value="">No category (choose later)</option>
                            {#each categories as category}
                                <option value={category.id}>
                                    {category.label}
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
                                Choose how this status appears in your public
                                changelog
                            </p>
                        {/if}
                    </div>
                {:else if loadingCategories}
                    <div class="text-xs text-muted-foreground">
                        Loading categories...
                    </div>
                {:else}
                    <div
                        class="p-2 text-xs bg-muted/50 text-muted-foreground rounded border border-border"
                    >
                        No theme configured. The status will be created without
                        a category mapping.
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
                    Cancel
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
                        Creating...
                    {:else}
                        Create Status
                    {/if}
                </Button>
            </div>
        </div>
    </div>
{/if}
