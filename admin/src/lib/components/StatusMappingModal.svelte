<script lang="ts">
    import { onMount } from "svelte";
    import { api } from "$lib/api";
    import { Button } from "$lib/components/ui";
    import { X, Loader2, AlertCircle, AlertTriangle } from "lucide-svelte";
    import { toast } from "svelte-sonner";

    export let isOpen = false;
    export let onClose: () => void;

    interface ThemeCategory {
        id: string;
        label: string;
        description: string;
        multiple?: boolean;
        order: number;
    }

    interface ThemeSetting {
        id: string;
        label: string;
        description: string;
        type: string;
        default: any;
    }

    interface ThemeManifest {
        id: string;
        name: string;
        version: string;
        settings?: ThemeSetting[];
        categories: ThemeCategory[];
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

    let loading = true;
    let saving = false;
    let manifest: ThemeManifest | null = null;
    let statusRows: StatusRow[] = [];
    let error = "";
    let activeCategory: string = "";
    let activeTab: "display" | "settings" = "display";

    let localMappings: Map<number, string | null> = new Map();
    let settingsValues: Map<string, any> = new Map();

    onMount(() => {
        if (isOpen) {
            loadMappings();
        }
    });

    $: if (isOpen) {
        loadMappings();
    }

    async function loadMappings() {
        loading = true;
        error = "";

        try {
            const manifestData = await api.getThemeManifest();
            manifest = manifestData.manifest;

            const mappingsData = await api.getStatusMappings();
            const mappings: StatusMapping[] = mappingsData.mappings || [];
            const unmappedStatuses: UnmappedStatus[] =
                mappingsData.unmapped_statuses || [];

            // Set active category to first category
            if (manifest.categories.length > 0 && !activeCategory) {
                activeCategory = manifest.categories[0].id;
            }

            // Load settings values from backend
            if (manifest.settings && manifest.settings.length > 0) {
                try {
                    const settingsData = await api.getThemeSettings();
                    if (settingsData.settings) {
                        settingsData.settings.forEach((setting: any) => {
                            settingsValues.set(setting.id, setting.value);
                        });
                        settingsValues = new Map(settingsValues);
                    }
                } catch {
                    // If loading settings fails, use default values
                    manifest.settings.forEach((setting: ThemeSetting) => {
                        settingsValues.set(setting.id, setting.default);
                    });
                    settingsValues = new Map(settingsValues);
                }
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

            statusRows.sort((a, b) =>
                a.status_name.localeCompare(b.status_name),
            );
        } catch (err) {
            error =
                err instanceof Error ? err.message : "Failed to load mappings";
            toast.error(error);
        } finally {
            loading = false;
        }
    }

    function updateLocalMapping(statusId: number, categoryId: string | null) {
        const value = categoryId === "" ? null : categoryId;
        localMappings.set(statusId, value);
        localMappings = new Map(localMappings);
        // Force re-render by reassigning statusRows
        statusRows = [...statusRows];
    }

    function updateSettingValue(settingId: string, value: any) {
        settingsValues.set(settingId, value);
        settingsValues = new Map(settingsValues);
    }

    function hasChanges(): boolean {
        // Always return true to allow saving at any time
        return true;
    }

    async function saveAllMappings() {
        saving = true;

        try {
            // Save status mappings
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

            // Save theme settings
            if (manifest?.settings && manifest.settings.length > 0) {
                const settingsToSave: Record<string, any> = {};
                manifest.settings.forEach((setting: ThemeSetting) => {
                    if (settingsValues.has(setting.id)) {
                        settingsToSave[setting.id] = settingsValues.get(
                            setting.id,
                        );
                    }
                });

                if (Object.keys(settingsToSave).length > 0) {
                    await api.updateThemeSettings(settingsToSave);
                }
            }

            toast.success("Settings saved successfully");
            await loadMappings();
            onClose();
        } catch (err) {
            toast.error(
                err instanceof Error ? err.message : "Failed to save settings",
            );
        } finally {
            saving = false;
        }
    }

    function handleClose() {
        onClose();
    }
</script>

{#if isOpen}
    <div
        class="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4"
        on:click={handleClose}
        on:keydown={(e) => e.key === "Escape" && handleClose()}
        role="button"
        tabindex="0"
    >
        <div
            class="bg-background rounded-lg shadow-lg w-full max-w-4xl max-h-[85vh] overflow-hidden flex flex-col border"
            on:click={(e) => e.stopPropagation()}
            on:keydown={(e) => e.stopPropagation()}
            role="dialog"
            tabindex="-1"
        >
            <!-- Header -->
            <div class="px-6 py-4 border-b">
                <div class="flex items-start justify-between">
                    <div class="space-y-1">
                        <h2 class="text-base font-semibold">
                            Public Page Configuration
                        </h2>
                        <p class="text-xs text-muted-foreground">
                            Configure what visitors see on your public page
                        </p>
                    </div>
                    <button
                        on:click={handleClose}
                        class="text-muted-foreground hover:text-foreground transition-colors"
                    >
                        <X class="h-4 w-4" />
                    </button>
                </div>
            </div>

            <!-- Content with Sidebar -->
            <div class="flex-1 overflow-hidden flex">
                <!-- Sidebar -->
                <div class="w-48 border-r bg-muted/20 flex flex-col">
                    {#if !loading && !error && manifest}
                        <nav class="flex-1 p-3 space-y-1">
                            <button
                                on:click={() => (activeTab = "display")}
                                class="w-full text-left px-3 py-2 text-sm rounded-md transition-colors {activeTab ===
                                'display'
                                    ? 'text-foreground font-medium'
                                    : 'text-muted-foreground hover:text-foreground'}"
                            >
                                Status Display
                            </button>
                            {#if manifest.settings && manifest.settings.length > 0}
                                <button
                                    on:click={() => (activeTab = "settings")}
                                    class="w-full text-left px-3 py-2 text-sm rounded-md transition-colors {activeTab ===
                                    'settings'
                                        ? 'text-foreground font-medium'
                                        : 'text-muted-foreground hover:text-foreground'}"
                                >
                                    Theme Settings
                                </button>
                            {/if}
                        </nav>
                    {/if}
                </div>

                <!-- Main Content -->
                <div class="flex-1 overflow-y-auto">
                    {#if loading}
                        <div class="flex items-center justify-center py-16">
                            <div class="flex items-center gap-2 text-sm">
                                <Loader2 class="h-4 w-4 animate-spin" />
                                <span class="text-muted-foreground">
                                    Loading...
                                </span>
                            </div>
                        </div>
                    {:else if error}
                        <div class="p-6">
                            <div
                                class="flex items-center gap-2 text-sm p-3 rounded-md bg-destructive/10 text-destructive"
                            >
                                <AlertCircle class="h-4 w-4 flex-shrink-0" />
                                <span class="text-xs text-destructive">
                                    {error}
                                </span>
                            </div>
                        </div>
                    {:else if !manifest}
                        <div class="text-center py-16 px-6">
                            <AlertCircle
                                class="h-12 w-12 mx-auto mb-3 text-muted-foreground"
                            />
                            <p class="text-sm font-medium mb-1">
                                No Theme Applied
                            </p>
                            <p class="text-xs text-muted-foreground">
                                Apply a theme to configure display
                            </p>
                        </div>
                    {:else if statusRows.length === 0}
                        <div class="text-center py-16 px-6">
                            <p class="text-xs text-muted-foreground">
                                No statuses found. Create statuses in your
                                kanban first.
                            </p>
                        </div>
                    {:else if activeTab === "display"}
                        <!-- Category Sub-tabs -->
                        <div class="border-b px-6 pt-4">
                            <div class="flex gap-1 overflow-x-auto">
                                {#each manifest.categories.sort((a, b) => a.order - b.order) as category}
                                    {@const categoryHasStatuses =
                                        statusRows.some((row) => {
                                            const currentCategory =
                                                localMappings.has(row.status_id)
                                                    ? localMappings.get(
                                                          row.status_id,
                                                      )
                                                    : row.category_id;
                                            return (
                                                currentCategory === category.id
                                            );
                                        })}
                                    <button
                                        on:click={() =>
                                            (activeCategory = category.id)}
                                        class="px-4 py-2 text-xs font-medium rounded-t-md transition-colors whitespace-nowrap flex items-center gap-1.5 {activeCategory ===
                                        category.id
                                            ? 'bg-background border-t border-x text-foreground'
                                            : 'text-muted-foreground hover:text-foreground hover:bg-muted/50'}"
                                    >
                                        <span>{category.label}</span>
                                        {#if !categoryHasStatuses}
                                            <span
                                                title="No statuses in this category"
                                            >
                                                <AlertTriangle
                                                    class="h-3 w-3 text-amber-500"
                                                />
                                            </span>
                                        {/if}
                                    </button>
                                {/each}
                            </div>
                        </div>

                        <!-- Active Category Content -->
                        {#each manifest.categories.sort((a, b) => a.order - b.order) as category}
                            {#if activeCategory === category.id}
                                {@const statusesInCategory = statusRows.filter(
                                    (row) => {
                                        const currentCategory =
                                            localMappings.has(row.status_id)
                                                ? localMappings.get(
                                                      row.status_id,
                                                  )
                                                : row.category_id;
                                        return currentCategory === category.id;
                                    },
                                )}
                                {@const availableStatuses = statusRows.filter(
                                    (row) => {
                                        const currentCategory =
                                            localMappings.has(row.status_id)
                                                ? localMappings.get(
                                                      row.status_id,
                                                  )
                                                : row.category_id;
                                        return currentCategory !== category.id;
                                    },
                                )}

                                <div class="p-6">
                                    <!-- Category Description -->
                                    <div class="mb-4 pb-4 border-b">
                                        <p
                                            class="text-xs text-muted-foreground"
                                        >
                                            {category.description}
                                        </p>
                                        {#if statusesInCategory.length === 0}
                                            <div
                                                class="flex items-center gap-2 mt-3 text-xs p-2 rounded-md bg-amber-500/10 text-amber-700 dark:text-amber-300 border border-amber-500/20"
                                            >
                                                <AlertTriangle
                                                    class="h-3.5 w-3.5 flex-shrink-0"
                                                />
                                                <span
                                                    >This category is empty and
                                                    won't appear on your public
                                                    page</span
                                                >
                                            </div>
                                        {/if}
                                    </div>

                                    <!-- Statuses in Category -->
                                    <div class="space-y-2">
                                        {#each statusesInCategory as status}
                                            <div
                                                class="flex items-center justify-between px-3 py-2 rounded-md bg-muted/50 hover:bg-muted transition-colors group"
                                            >
                                                <span class="text-sm">
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
                                                    title="Remove from category"
                                                >
                                                    <X class="h-3.5 w-3.5" />
                                                </button>
                                            </div>
                                        {/each}

                                        <!-- Add Status Dropdown -->
                                        {#if availableStatuses.length > 0 && (category.multiple !== false || statusesInCategory.length === 0)}
                                            <div class="pt-2">
                                                <select
                                                    on:change={(e) => {
                                                        const statusId =
                                                            parseInt(
                                                                e.currentTarget
                                                                    .value,
                                                            );
                                                        if (statusId) {
                                                            updateLocalMapping(
                                                                statusId,
                                                                category.id,
                                                            );
                                                            e.currentTarget.value =
                                                                "";
                                                        }
                                                    }}
                                                    class="w-full text-xs px-3 py-2 rounded-md border bg-background hover:bg-accent transition-colors focus:outline-none focus:ring-1 focus:ring-ring cursor-pointer"
                                                >
                                                    <option value="">
                                                        + Add status...
                                                    </option>
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
                                                            currentCategory !==
                                                                null &&
                                                            currentCategory !==
                                                                category.id}
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
                                                class="pt-2 text-xs text-muted-foreground italic"
                                            >
                                                This category already has a
                                                status assigned and does not
                                                allow multiple statuses.
                                            </div>
                                        {/if}
                                    </div>
                                </div>
                            {/if}
                        {/each}
                    {:else if activeTab === "settings" && manifest}
                        <!-- Settings Content -->
                        <div class="p-6">
                            {#if manifest.settings && manifest.settings.length > 0}
                                <div class="space-y-6">
                                    {#each manifest.settings as setting}
                                        <div
                                            class="flex items-start justify-between gap-6 py-3 border-b last:border-b-0"
                                        >
                                            <div class="flex-1">
                                                <label
                                                    for={setting.id}
                                                    class="text-sm font-medium block mb-1"
                                                >
                                                    {setting.label}
                                                </label>
                                                <p
                                                    class="text-xs text-muted-foreground"
                                                >
                                                    {setting.description}
                                                </p>
                                            </div>
                                            <div
                                                class="flex items-center flex-shrink-0"
                                            >
                                                {#if setting.type === "boolean"}
                                                    <label
                                                        class="relative inline-flex items-center cursor-pointer"
                                                    >
                                                        <input
                                                            id={setting.id}
                                                            type="checkbox"
                                                            checked={settingsValues.get(
                                                                setting.id,
                                                            ) ??
                                                                setting.default}
                                                            on:change={(e) =>
                                                                updateSettingValue(
                                                                    setting.id,
                                                                    e
                                                                        .currentTarget
                                                                        .checked,
                                                                )}
                                                            class="sr-only peer"
                                                        />
                                                        <div
                                                            class="w-11 h-6 bg-muted peer-focus:outline-none peer-focus:ring-2 peer-focus:ring-ring rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-primary"
                                                        ></div>
                                                    </label>
                                                {:else if setting.type === "text"}
                                                    <input
                                                        id={setting.id}
                                                        type="text"
                                                        value={settingsValues.get(
                                                            setting.id,
                                                        ) ?? setting.default}
                                                        on:input={(e) =>
                                                            updateSettingValue(
                                                                setting.id,
                                                                e.currentTarget
                                                                    .value,
                                                            )}
                                                        class="text-sm px-3 py-2 rounded-md border bg-background w-48"
                                                    />
                                                {:else if setting.type === "number"}
                                                    <input
                                                        id={setting.id}
                                                        type="number"
                                                        value={settingsValues.get(
                                                            setting.id,
                                                        ) ?? setting.default}
                                                        on:input={(e) =>
                                                            updateSettingValue(
                                                                setting.id,
                                                                parseFloat(
                                                                    e
                                                                        .currentTarget
                                                                        .value,
                                                                ),
                                                            )}
                                                        class="text-sm px-3 py-2 rounded-md border bg-background w-32"
                                                    />
                                                {/if}
                                            </div>
                                        </div>
                                    {/each}
                                </div>
                            {:else}
                                <div class="text-center py-8">
                                    <p class="text-xs text-muted-foreground">
                                        No settings available for this theme.
                                    </p>
                                </div>
                            {/if}
                        </div>
                    {/if}
                </div>
            </div>

            <!-- Footer -->
            {#if !loading && !error && manifest}
                <div class="px-6 py-3 border-t bg-muted/20">
                    <div class="flex items-center justify-end">
                        <div class="flex gap-2">
                            <Button
                                variant="ghost"
                                size="sm"
                                on:click={handleClose}
                                class="h-8 text-xs"
                            >
                                {hasChanges() ? "Cancel" : "Close"}
                            </Button>
                            <Button
                                size="sm"
                                on:click={saveAllMappings}
                                disabled={saving}
                                class="h-8 text-xs"
                            >
                                {#if saving}
                                    <Loader2
                                        class="h-3 w-3 animate-spin mr-1.5"
                                    />
                                {/if}
                                Save
                            </Button>
                        </div>
                    </div>
                </div>
            {/if}
        </div>
    </div>
{/if}
