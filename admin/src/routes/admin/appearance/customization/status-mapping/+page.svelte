<script lang="ts">
    import { onMount } from "svelte";
    import { api } from "$lib/api";
    import { Button } from "$lib/components/ui";
    import {
        Loader2,
        Save,
        AlertCircle,
        X,
        AlertTriangle,
    } from "lucide-svelte";
    import { toast } from "svelte-sonner";
    import * as m from "$lib/paraglide/messages";
    import { emptyCategoriesStore } from "$lib/stores/emptyCategories";

    interface ThemeCategory {
        id: string;
        label: string;
        description: string;
        multiple?: boolean;
        order: number;
    }

    interface StatusMapping {
        status_id: number;
        status_name: string;
        category_id: string;
        category_label: string;
        theme_id: string;
    }

    interface UnmappedStatus {
        status_id: number;
        status_name: string;
        suggested_category: string;
    }

    interface StatusRow {
        status_id: number;
        status_name: string;
        category_id: string | null;
    }

    interface ThemeManifest {
        id: string;
        name: string;
        version: string;
        categories: ThemeCategory[];
    }

    let loading = true;
    let saving = false;
    let manifest: ThemeManifest | null = null;
    let statusRows: StatusRow[] = [];
    let error: string | null = null;
    let activeCategory: string = "";

    let localMappings: Map<number, string | null> = new Map();

    let sidebarTop: number | null = null;
    let sidebarElement: HTMLElement;

    function handleScroll() {
        if (!sidebarElement) return;

        const scrollTop =
            window.pageYOffset || document.documentElement.scrollTop;

        const initialTop = sidebarElement.offsetTop || 0;

        if (scrollTop > initialTop - 24) {
            sidebarTop = 24;
        } else {
            sidebarTop = initialTop - scrollTop + 24;
        }
    }

    onMount(() => {
        loadMappings().then(() => {
            setTimeout(() => {
                window.addEventListener("scroll", handleScroll);
                handleScroll();

                return () => {
                    window.removeEventListener("scroll", handleScroll);
                };
            }, 100);
        });
    });

    async function loadMappings() {
        loading = true;
        error = null;

        try {
            const manifestData = await api.getThemeManifest();
            manifest = manifestData.manifest;

            const mappingsData = await api.getStatusMappings();
            const mappings: StatusMapping[] = mappingsData.mappings || [];
            const unmappedStatuses: UnmappedStatus[] =
                mappingsData.unmapped_statuses || [];

            if (manifest.categories.length > 0 && !activeCategory) {
                activeCategory = manifest.categories.sort(
                    (a, b) => a.order - b.order,
                )[0].id;
            }

            statusRows = [];

            mappings.forEach((m) => {
                statusRows.push({
                    status_id: m.status_id,
                    status_name: m.status_name,
                    category_id: m.category_id,
                });
                localMappings.set(m.status_id, m.category_id);
            });

            unmappedStatuses.forEach((u) => {
                statusRows.push({
                    status_id: u.status_id,
                    status_name: u.status_name,
                    category_id: null,
                });
                localMappings.set(u.status_id, null);
            });

            localMappings = new Map(localMappings);
        } catch (err: any) {
            error = err.message || "Failed to load status mappings";
            console.error("Error loading status mappings:", err);
        } finally {
            loading = false;
        }
    }

    function updateLocalMapping(statusId: number, categoryId: string | null) {
        const value = categoryId === "" ? null : categoryId;
        localMappings.set(statusId, value);
        localMappings = new Map(localMappings);
    }

    function hasChanges(): boolean {
        // Check if any mappings have changed from their original values
        for (const row of statusRows) {
            const currentMapping = localMappings.get(row.status_id);
            // Convert undefined to null for comparison
            const current =
                currentMapping === undefined ? row.category_id : currentMapping;
            if (current !== row.category_id) {
                return true;
            }
        }
        return false;
    }

    async function saveAllMappings() {
        if (!hasChanges()) {
            toast.info(m.status_mapping_modal_no_changes());
            return;
        }

        saving = true;
        try {
            // Process deletions first to avoid conflicts with single-status categories
            for (const row of statusRows) {
                const newCategoryId = localMappings.has(row.status_id)
                    ? localMappings.get(row.status_id)
                    : row.category_id;
                const oldCategoryId = row.category_id;

                if (newCategoryId !== oldCategoryId) {
                    if (newCategoryId === null || newCategoryId === undefined) {
                        // Delete the mapping
                        await api.deleteStatusMapping(row.status_id);
                    }
                }
            }

            // Then process updates and creations
            for (const row of statusRows) {
                const newCategoryId = localMappings.has(row.status_id)
                    ? localMappings.get(row.status_id)
                    : row.category_id;
                const oldCategoryId = row.category_id;

                if (newCategoryId !== oldCategoryId) {
                    if (newCategoryId !== null && newCategoryId !== undefined) {
                        // Update or create the mapping
                        await api.updateStatusMapping(
                            row.status_id,
                            newCategoryId,
                        );
                    }
                }
            }

            // Update statusRows with new mappings
            statusRows = statusRows.map((row) => {
                const newCategoryId = localMappings.get(row.status_id);
                return {
                    ...row,
                    category_id: newCategoryId ?? row.category_id,
                };
            });

            toast.success(m.status_mapping_modal_save_success());

            // Trigger re-check of empty categories to update UI
            emptyCategoriesStore.check();
        } catch (err: any) {
            console.error("Error saving mappings:", err);
            toast.error(err.message || m.status_mapping_modal_save_error());
        } finally {
            saving = false;
        }
    }

    function scrollToCategory(categoryId: string) {
        activeCategory = categoryId;
        const element = document.getElementById(`category-${categoryId}`);
        if (element) {
            element.scrollIntoView({ behavior: "smooth", block: "start" });
        }
    }
</script>

<svelte:head>
    <title>{m.customization_status_mapping_title()}</title>
</svelte:head>

<div class="w-full flex gap-6">
    {#if loading}
        <div class="flex-1 flex items-center justify-center py-16">
            <div class="flex items-center gap-2 text-sm">
                <Loader2 class="h-4 w-4 animate-spin" />
                <span class="text-muted-foreground"
                    >{m.status_mapping_modal_loading()}</span
                >
            </div>
        </div>
    {:else if error}
        <div class="flex-1 p-6">
            <div
                class="p-4 rounded-lg bg-destructive/10 text-destructive border border-destructive/20 flex items-start gap-3"
            >
                <AlertCircle class="h-5 w-5 flex-shrink-0 mt-0.5" />
                <div>
                    <p class="font-medium mb-1">
                        {m.customization_status_mapping_error()}
                    </p>
                    <p class="text-sm">{error}</p>
                </div>
            </div>
        </div>
    {:else if !manifest || !manifest.categories || manifest.categories.length === 0}
        <div class="flex-1 text-center py-16 px-6">
            <AlertCircle class="h-8 w-8 text-muted-foreground mx-auto mb-3" />
            <p class="text-sm font-medium mb-1">
                {m.status_mapping_modal_no_theme()}
            </p>
            <p class="text-xs text-muted-foreground">
                {m.status_mapping_modal_no_theme_description()}
            </p>
        </div>
    {:else}
        <!-- Sidebar Navigation -->
        <aside class="w-48 flex-shrink-0" bind:this={sidebarElement}>
            <div
                class="sticky"
                style="top: {sidebarTop !== null
                    ? `${sidebarTop}px`
                    : 'auto'}; opacity: {sidebarTop !== null ? 1 : 0};"
            >
                <nav class="space-y-1">
                    {#each manifest.categories.sort((a, b) => a.order - b.order) as category}
                        {@const categoryHasStatuses =
                            statusRows.filter((row) => {
                                const currentCategory = localMappings.has(
                                    row.status_id,
                                )
                                    ? localMappings.get(row.status_id)
                                    : row.category_id;
                                return currentCategory === category.id;
                            }).length > 0}
                        <button
                            on:click={() => scrollToCategory(category.id)}
                            class="w-full text-left px-3 py-2 text-sm rounded-md transition-colors flex items-center justify-between gap-2 {activeCategory ===
                            category.id
                                ? 'bg-accent text-accent-foreground font-medium'
                                : 'text-muted-foreground hover:text-foreground hover:bg-accent/50'}"
                        >
                            <span class="truncate">{category.label}</span>
                            {#if !categoryHasStatuses}
                                <AlertTriangle
                                    class="h-3.5 w-3.5 text-amber-500 flex-shrink-0"
                                />
                            {/if}
                        </button>
                    {/each}
                </nav>
            </div>
        </aside>

        <!-- Main Content -->
        <div class="flex-1 min-w-0 space-y-12 ml-6">
            {#each manifest.categories.sort((a, b) => a.order - b.order) as category, index}
                {@const statusesInCategory = statusRows.filter((row) => {
                    const currentCategory = localMappings.has(row.status_id)
                        ? localMappings.get(row.status_id)
                        : row.category_id;
                    return currentCategory === category.id;
                })}
                {@const availableStatuses = statusRows.filter((row) => {
                    const currentCategory = localMappings.has(row.status_id)
                        ? localMappings.get(row.status_id)
                        : row.category_id;
                    return currentCategory !== category.id;
                })}

                <div
                    id="category-{category.id}"
                    class="scroll-mt-6 {index > 0 ? 'pt-12 border-t' : ''}"
                >
                    <div class="mb-6">
                        <h3
                            class="text-base font-semibold flex items-center gap-2"
                        >
                            {category.label}
                            {#if !statusesInCategory.length}
                                <AlertTriangle class="h-4 w-4 text-amber-500" />
                            {/if}
                        </h3>
                        <p class="text-sm text-muted-foreground mt-1.5">
                            {category.description}
                        </p>
                    </div>

                    <div class="space-y-4">
                        <!-- Warning if category is empty -->
                        {#if statusesInCategory.length === 0}
                            <div
                                class="flex items-center gap-2 text-xs p-3 rounded-md bg-amber-500/10 text-amber-700 dark:text-amber-300 border border-amber-500/20"
                            >
                                <AlertTriangle class="h-4 w-4 flex-shrink-0" />
                                <span
                                    >{m.status_mapping_modal_category_empty_warning()}</span
                                >
                            </div>
                        {/if}

                        <!-- Statuses in Category -->
                        <div class="space-y-2">
                            {#each statusesInCategory as status}
                                <div
                                    class="flex items-center justify-between px-4 py-3 rounded-md bg-muted/50 hover:bg-muted transition-colors group"
                                >
                                    <span class="text-sm font-medium">
                                        {status.status_name}
                                    </span>
                                    <button
                                        type="button"
                                        on:click={(e) => {
                                            e.preventDefault();
                                            e.stopPropagation();
                                            updateLocalMapping(
                                                status.status_id,
                                                null,
                                            );
                                        }}
                                        class="text-muted-foreground hover:text-destructive opacity-0 group-hover:opacity-100 transition-all"
                                        title={m.customization_status_mapping_remove_tooltip()}
                                    >
                                        <X class="h-4 w-4" />
                                    </button>
                                </div>
                            {/each}

                            <!-- Add Status Dropdown -->
                            {#if availableStatuses.length > 0 && (category.multiple === true || statusesInCategory.length === 0)}
                                <div class="pt-2">
                                    <select
                                        on:change={(e) => {
                                            const statusId = parseInt(
                                                e.currentTarget.value,
                                            );
                                            if (statusId) {
                                                updateLocalMapping(
                                                    statusId,
                                                    category.id,
                                                );
                                                e.currentTarget.value = "";
                                            }
                                        }}
                                        class="w-full text-sm px-4 py-3 rounded-md border bg-background hover:bg-accent transition-colors focus:outline-none focus:ring-2 focus:ring-ring cursor-pointer"
                                    >
                                        <option value=""
                                            >{m.customization_status_mapping_add_status(
                                                {
                                                    categoryLabel:
                                                        category.label,
                                                },
                                            )}</option
                                        >
                                        {#each statusRows as status}
                                            {@const currentCategory =
                                                localMappings.has(
                                                    status.status_id,
                                                )
                                                    ? localMappings.get(
                                                          status.status_id,
                                                      )
                                                    : status.category_id}
                                            {@const isAlreadyMapped =
                                                currentCategory !== null &&
                                                currentCategory !== category.id}
                                            <option
                                                value={status.status_id}
                                                disabled={isAlreadyMapped}
                                            >
                                                {status.status_name}
                                            </option>
                                        {/each}
                                    </select>
                                </div>
                            {:else if category.multiple === false && statusesInCategory.length > 0}
                                <div
                                    class="pt-2 text-xs text-muted-foreground italic px-4"
                                >
                                    {m.customization_status_mapping_single_only()}
                                </div>
                            {/if}
                        </div>
                    </div>
                </div>
            {/each}

            <!-- Save Button -->
            <div class="flex justify-end pt-6 mt-6 border-t">
                <Button
                    on:click={saveAllMappings}
                    disabled={saving || !hasChanges()}
                >
                    {#if saving}
                        <Loader2 class="h-4 w-4 animate-spin mr-2" />
                        {m.customization_status_mapping_saving()}
                    {:else}
                        <Save class="h-4 w-4 mr-2" />
                        {m.customization_status_mapping_save()}
                    {/if}
                </Button>
            </div>
        </div>
    {/if}
</div>
