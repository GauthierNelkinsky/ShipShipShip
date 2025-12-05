<script lang="ts">
    import { onMount } from "svelte";
    import { api } from "$lib/api";
    import { Button, Input } from "$lib/components/ui";
    import { Loader2, Save, AlertCircle, X, ChevronDown } from "lucide-svelte";
    import { toast } from "svelte-sonner";
    import * as m from "$lib/paraglide/messages";

    interface ThemeSetting {
        id: string;
        label: string;
        description: string;
        type: string;
        default: any;
        options?: { value: string; label: string }[];
        items?: ThemeSetting[];
    }

    interface SettingGroup {
        group: string;
        description: string;
        settings: ThemeSetting[];
    }

    interface ThemeManifest {
        id: string;
        name: string;
        version: string;
        settings: SettingGroup[];
    }

    let loading = true;
    let saving = false;
    let manifest: ThemeManifest | null = null;
    let settingsValues: Record<string, any> = {};
    let error: string | null = null;

    let activeGroup = "";
    let sidebarTop: number | null = null;
    let sidebarElement: HTMLElement;

    // Track expanded state for array settings
    let expandedArrays: Record<string, boolean> = {};

    function handleScroll() {
        if (!sidebarElement) return;

        const scrollTop =
            window.pageYOffset || document.documentElement.scrollTop;

        // Get initial position of sidebar from its parent container
        const initialTop = sidebarElement.offsetTop || 0;

        // If scrolled past initial position, stick to top with padding
        if (scrollTop > initialTop - 24) {
            sidebarTop = 24; // 1.5rem top padding from viewport top
        } else {
            sidebarTop = initialTop - scrollTop + 24;
        }
    }

    function updateActiveGroupOnScroll() {
        if (!manifest || !manifest.settings) return;

        // Find which group is currently most visible
        let closestGroup = "";
        let closestDistance = Infinity;

        manifest.settings.forEach((group) => {
            const element = document.getElementById(`group-${group.group}`);
            if (element) {
                const rect = element.getBoundingClientRect();
                // Distance from top of viewport (accounting for some offset)
                const distance = Math.abs(rect.top - 100);

                if (
                    distance < closestDistance &&
                    rect.top < window.innerHeight
                ) {
                    closestDistance = distance;
                    closestGroup = group.group;
                }
            }
        });

        if (closestGroup && closestGroup !== activeGroup) {
            activeGroup = closestGroup;
        }
    }

    onMount(() => {
        loadSettings().then(() => {
            // Set first group as active by default
            if (manifest && manifest.settings && manifest.settings.length > 0) {
                activeGroup = manifest.settings[0].group;
            }

            // Wait for DOM to be fully rendered before calculating position
            setTimeout(() => {
                // Combined scroll handler
                const onScroll = () => {
                    handleScroll();
                    updateActiveGroupOnScroll();
                };

                // Add scroll listener
                window.addEventListener("scroll", onScroll);
                onScroll(); // Initial call

                // Cleanup
                return () => {
                    window.removeEventListener("scroll", onScroll);
                };
            }, 100);
        });
    });

    async function loadSettings() {
        loading = true;
        error = null;

        try {
            // Load theme manifest
            const manifestData = await api.getThemeManifest();
            manifest = manifestData.manifest as any;

            // Load current settings values
            const settingsData = await api.getThemeSettings();

            // Initialize settingsValues with current values or defaults
            if (manifest && manifest.settings) {
                manifest.settings.forEach((group) => {
                    group.settings.forEach((setting) => {
                        const savedSetting = settingsData.settings?.find(
                            (s: any) => s.id === setting.id,
                        );
                        let value = savedSetting?.value ?? setting.default;

                        // Handle array values
                        if (setting.type === "array") {
                            // Already an array, use as-is
                            if (Array.isArray(value)) {
                                settingsValues[setting.id] = value;
                            }
                            // String that needs parsing
                            else if (typeof value === "string") {
                                // Empty string or "[]" should be empty array
                                if (value === "" || value === "[]") {
                                    settingsValues[setting.id] = [];
                                } else {
                                    try {
                                        const parsed = JSON.parse(value);
                                        settingsValues[setting.id] =
                                            Array.isArray(parsed) ? parsed : [];
                                    } catch {
                                        // Reset corrupted Go-format data to empty array
                                        settingsValues[setting.id] = [];
                                        // Immediately save the corrected value
                                        updateSettingValue(setting.id, []);
                                    }
                                }
                            }
                            // Fallback to default
                            else {
                                settingsValues[setting.id] =
                                    setting.default ?? [];
                            }
                        } else {
                            settingsValues[setting.id] = value;
                        }
                    });
                });
            }
        } catch (err) {
            console.error("Failed to load theme settings:", err);
            error =
                err instanceof Error
                    ? err.message
                    : "Failed to load theme settings";
            toast.error("Failed to load theme settings");
        } finally {
            loading = false;
        }
    }

    function updateSettingValue(settingId: string, value: any) {
        settingsValues[settingId] = value;
    }

    function addArrayItem(settingId: string) {
        if (!settingsValues[settingId]) {
            settingsValues[settingId] = [];
        }

        // Find the setting definition to get the items structure
        const setting = manifest?.settings
            .flatMap((group) => group.settings)
            .find((s) => s.id === settingId);

        // Initialize empty object with field IDs from items array
        const newItem: Record<string, any> = {};
        if (setting?.items) {
            setting.items.forEach((fieldSetting) => {
                newItem[fieldSetting.id] = fieldSetting.default || "";
            });
        }

        settingsValues[settingId] = [...settingsValues[settingId], newItem];
    }

    function removeArrayItem(settingId: string, index: number) {
        const currentValue = settingsValues[settingId];
        if (!Array.isArray(currentValue)) {
            return;
        }
        settingsValues[settingId] = currentValue.filter(
            (_: any, i: number) => i !== index,
        );
    }

    function updateArrayItemField(
        settingId: string,
        index: number,
        field: string,
        value: string,
    ) {
        const currentValue = settingsValues[settingId];
        if (!Array.isArray(currentValue)) {
            return;
        }
        const items = [...currentValue];
        items[index] = { ...items[index], [field]: value };
        settingsValues[settingId] = items;
    }

    async function handleImageUpload(
        settingId: string,
        event: Event & { currentTarget: HTMLInputElement },
    ) {
        const file = event.currentTarget.files?.[0];
        if (!file) return;

        try {
            const result = await api.uploadImage(file);
            updateSettingValue(settingId, result.url);
            toast.success("Image uploaded successfully");
        } catch (err) {
            toast.error(
                err instanceof Error ? err.message : "Failed to upload image",
            );
        }
    }

    async function saveSettings() {
        if (!manifest) return;

        saving = true;
        try {
            await api.updateThemeSettings(settingsValues);
            toast.success("Settings saved successfully");
        } catch (err) {
            console.error("Failed to save settings:", err);
            const errorMessage =
                err instanceof Error ? err.message : "Failed to save settings";
            toast.error("Failed to save settings", {
                description: errorMessage,
            });
        } finally {
            saving = false;
        }
    }

    function scrollToGroup(groupName: string) {
        activeGroup = groupName;
        const element = document.getElementById(`group-${groupName}`);
        if (element) {
            element.scrollIntoView({ behavior: "smooth", block: "start" });
        }
    }
</script>

<svelte:head>
    <title>{m.branding_page_title()}</title>
</svelte:head>

<div class="w-full flex gap-6">
    {#if loading}
        <div class="flex-1 flex items-center justify-center py-16">
            <div class="flex items-center gap-2 text-sm">
                <Loader2 class="h-4 w-4 animate-spin" />
                <span class="text-muted-foreground">Loading settings...</span>
            </div>
        </div>
    {:else if error}
        <div class="flex-1 p-6">
            <div
                class="flex items-start gap-3 text-destructive bg-destructive/10 p-4 rounded-lg"
            >
                <AlertCircle class="h-5 w-5 flex-shrink-0 mt-0.5" />
                <div>
                    <p class="font-medium mb-1">Failed to Load Settings</p>
                    <p class="text-sm">{error}</p>
                </div>
            </div>
        </div>
    {:else if !manifest || !manifest.settings || manifest.settings.length === 0}
        <div class="flex-1 text-center py-16 px-6">
            <AlertCircle class="h-8 w-8 text-muted-foreground mx-auto mb-3" />
            <p class="text-sm font-medium mb-1">No Settings Available</p>
            <p class="text-xs text-muted-foreground">
                This theme doesn't have any configurable settings.
            </p>
        </div>
    {:else}
        <!-- Fixed Sidebar Navigation -->
        <aside class="w-48 flex-shrink-0" bind:this={sidebarElement}>
            <div
                class="fixed w-48 transition-opacity duration-200 {sidebarTop ===
                null
                    ? 'opacity-0'
                    : 'opacity-100'}"
                style="top: {sidebarTop !== null ? `${sidebarTop}px` : '0'};"
            >
                <nav class="space-y-1">
                    {#each manifest.settings as group}
                        <button
                            on:click={() => scrollToGroup(group.group)}
                            class="w-full text-left px-3 py-2 rounded-md text-sm transition-colors {activeGroup ===
                            group.group
                                ? 'bg-accent text-accent-foreground font-medium'
                                : 'text-muted-foreground hover:text-foreground hover:bg-accent/50'}"
                        >
                            {group.group}
                        </button>
                    {/each}
                </nav>
            </div>
        </aside>

        <!-- Main Content -->
        <div class="flex-1 min-w-0 space-y-12 ml-6">
            {#each manifest.settings as group, index}
                <div
                    id={`group-${group.group}`}
                    class={index > 0 ? "pt-12 border-t" : ""}
                >
                    <!-- Group Header -->
                    <div class="mb-6">
                        <h3 class="text-base font-semibold">{group.group}</h3>
                        <p class="text-sm text-muted-foreground mt-1.5">
                            {group.description}
                        </p>
                    </div>

                    <!-- Settings List -->
                    <div class="space-y-6">
                        {#each group.settings as setting}
                            {#if setting.type === "array" && setting.items}
                                <!-- Array type: full width layout -->
                                <div class="space-y-4">
                                    <div
                                        class="flex items-center justify-between"
                                    >
                                        <div class="flex-1">
                                            <label
                                                class="text-sm font-medium block mb-1.5"
                                            >
                                                {setting.label}
                                            </label>
                                            <p
                                                class="text-xs text-muted-foreground"
                                            >
                                                {setting.description}
                                            </p>
                                        </div>
                                        <button
                                            type="button"
                                            on:click={() =>
                                                (expandedArrays[setting.id] =
                                                    !expandedArrays[
                                                        setting.id
                                                    ])}
                                            class="flex items-center gap-2 px-3 py-1.5 text-sm hover:bg-muted rounded transition-colors"
                                        >
                                            <ChevronDown
                                                class="w-4 h-4 transition-transform {expandedArrays[
                                                    setting.id
                                                ]
                                                    ? ''
                                                    : '-rotate-90'}"
                                            />
                                            <span>
                                                {(
                                                    settingsValues[
                                                        setting.id
                                                    ] || []
                                                ).length}
                                                {(
                                                    settingsValues[
                                                        setting.id
                                                    ] || []
                                                ).length === 1
                                                    ? "item"
                                                    : "items"}
                                            </span>
                                        </button>
                                    </div>
                                    {#if expandedArrays[setting.id] === true}
                                        <div class="space-y-4">
                                            {#each settingsValues[setting.id] || [] as item, index}
                                                <div
                                                    class="space-y-4 p-4 border rounded-lg bg-muted/30"
                                                >
                                                    {#each setting.items as fieldSetting, fieldIndex}
                                                        <div
                                                            class="flex items-start gap-6"
                                                        >
                                                            <div
                                                                class="flex-1 min-w-0"
                                                            >
                                                                <label
                                                                    for={`${setting.id}-${index}-${fieldSetting.id}`}
                                                                    class="text-sm font-medium block mb-1.5"
                                                                >
                                                                    {fieldSetting.label}
                                                                </label>
                                                                <p
                                                                    class="text-xs text-muted-foreground"
                                                                >
                                                                    {fieldSetting.description}
                                                                </p>
                                                            </div>
                                                            <div
                                                                class="flex-1 flex items-start gap-2"
                                                            >
                                                                <Input
                                                                    id={`${setting.id}-${index}-${fieldSetting.id}`}
                                                                    type={fieldSetting.type ===
                                                                    "url"
                                                                        ? "url"
                                                                        : "text"}
                                                                    value={item[
                                                                        fieldSetting
                                                                            .id
                                                                    ] || ""}
                                                                    on:input={(
                                                                        e,
                                                                    ) => {
                                                                        const target =
                                                                            e.target as HTMLInputElement;
                                                                        updateArrayItemField(
                                                                            setting.id,
                                                                            index,
                                                                            fieldSetting.id,
                                                                            target.value,
                                                                        );
                                                                    }}
                                                                    placeholder={fieldSetting.description ||
                                                                        `Enter ${fieldSetting.label}`}
                                                                    class="flex-1"
                                                                />
                                                                {#if fieldIndex === 0}
                                                                    <button
                                                                        on:click={() =>
                                                                            removeArrayItem(
                                                                                setting.id,
                                                                                index,
                                                                            )}
                                                                        class="p-2 text-destructive hover:bg-destructive/10 rounded transition-colors"
                                                                        type="button"
                                                                        aria-label="Remove item"
                                                                    >
                                                                        <X
                                                                            class="w-4 h-4"
                                                                        />
                                                                    </button>
                                                                {/if}
                                                            </div>
                                                        </div>
                                                    {/each}
                                                </div>
                                            {/each}
                                            <Button
                                                variant="outline"
                                                size="sm"
                                                on:click={() =>
                                                    addArrayItem(setting.id)}
                                            >
                                                Add Item
                                            </Button>
                                        </div>
                                    {/if}
                                </div>
                            {:else}
                                <!-- Other types: standard two-column layout -->
                                <div class="flex items-start gap-6">
                                    <div class="flex-1 min-w-0">
                                        <label
                                            for={setting.id}
                                            class="text-sm font-medium block mb-1.5"
                                        >
                                            {setting.label}
                                        </label>
                                        <p
                                            class="text-xs text-muted-foreground"
                                        >
                                            {setting.description}
                                        </p>
                                    </div>

                                    <div class="flex-1">
                                        {#if setting.type === "text"}
                                            <Input
                                                id={setting.id}
                                                type="text"
                                                value={settingsValues[
                                                    setting.id
                                                ] || ""}
                                                on:input={(e) => {
                                                    const target =
                                                        e.target as HTMLInputElement;
                                                    updateSettingValue(
                                                        setting.id,
                                                        target.value,
                                                    );
                                                }}
                                                placeholder={setting.default ||
                                                    ""}
                                            />
                                        {:else if setting.type === "boolean"}
                                            <!-- Switch Toggle -->
                                            <label
                                                class="relative inline-flex items-center cursor-pointer"
                                            >
                                                <input
                                                    type="checkbox"
                                                    checked={settingsValues[
                                                        setting.id
                                                    ] ?? setting.default}
                                                    on:change={(e) =>
                                                        updateSettingValue(
                                                            setting.id,
                                                            e.currentTarget
                                                                .checked,
                                                        )}
                                                    class="sr-only peer"
                                                />
                                                <div
                                                    class="w-11 h-6 bg-muted peer-focus:outline-none peer-focus:ring-2 peer-focus:ring-primary rounded-full peer peer-checked:after:translate-x-full rtl:peer-checked:after:-translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:start-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-primary"
                                                ></div>
                                                <span
                                                    class="ms-3 text-sm text-foreground"
                                                >
                                                    {(settingsValues[
                                                        setting.id
                                                    ] ?? setting.default)
                                                        ? "Enabled"
                                                        : "Disabled"}
                                                </span>
                                            </label>
                                        {:else if setting.type === "color"}
                                            <div
                                                class="flex items-center gap-2"
                                            >
                                                <input
                                                    type="color"
                                                    value={settingsValues[
                                                        setting.id
                                                    ] || setting.default}
                                                    on:input={(e) => {
                                                        const target =
                                                            e.target as HTMLInputElement;
                                                        updateSettingValue(
                                                            setting.id,
                                                            target.value,
                                                        );
                                                    }}
                                                    class="w-12 h-9 rounded border cursor-pointer"
                                                />
                                                <Input
                                                    type="text"
                                                    value={settingsValues[
                                                        setting.id
                                                    ] || setting.default}
                                                    on:input={(e) => {
                                                        const target =
                                                            e.target as HTMLInputElement;
                                                        updateSettingValue(
                                                            setting.id,
                                                            target.value,
                                                        );
                                                    }}
                                                    class="flex-1"
                                                />
                                            </div>
                                        {:else if setting.type === "select"}
                                            <select
                                                value={settingsValues[
                                                    setting.id
                                                ] || setting.default}
                                                on:change={(e) => {
                                                    const target =
                                                        e.target as HTMLSelectElement;
                                                    updateSettingValue(
                                                        setting.id,
                                                        target.value,
                                                    );
                                                }}
                                                class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                                            >
                                                {#if setting.options}
                                                    {#each setting.options as option}
                                                        <option
                                                            value={option.value}
                                                        >
                                                            {option.label}
                                                        </option>
                                                    {/each}
                                                {/if}
                                            </select>
                                        {:else if setting.type === "number"}
                                            <Input
                                                id={setting.id}
                                                type="number"
                                                value={settingsValues[
                                                    setting.id
                                                ] ?? setting.default}
                                                on:input={(e) => {
                                                    const target =
                                                        e.target as HTMLInputElement;
                                                    updateSettingValue(
                                                        setting.id,
                                                        parseFloat(
                                                            target.value,
                                                        ),
                                                    );
                                                }}
                                            />
                                        {:else if setting.type === "image"}
                                            <div class="space-y-3">
                                                <div
                                                    class="p-3 border rounded-lg bg-muted/30"
                                                >
                                                    {#if settingsValues[setting.id] && settingsValues[setting.id] !== ""}
                                                        {@const imageUrl =
                                                            settingsValues[
                                                                setting.id
                                                            ].startsWith("http")
                                                                ? settingsValues[
                                                                      setting.id
                                                                  ]
                                                                : `http://localhost:8080${settingsValues[setting.id]}`}
                                                        <div class="space-y-2">
                                                            <img
                                                                src={imageUrl}
                                                                alt={setting.label}
                                                                class="max-h-24 w-auto rounded border"
                                                                on:error={(
                                                                    e,
                                                                ) => {
                                                                    const target =
                                                                        e.currentTarget as HTMLImageElement;
                                                                    target.src =
                                                                        "data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='100' height='100'%3E%3Crect fill='%23ddd' width='100' height='100'/%3E%3Ctext x='50%25' y='50%25' text-anchor='middle' dy='.3em' fill='%23999'%3ENo Image%3C/text%3E%3C/svg%3E";
                                                                }}
                                                            />
                                                            <button
                                                                on:click={() =>
                                                                    updateSettingValue(
                                                                        setting.id,
                                                                        "",
                                                                    )}
                                                                class="text-xs text-destructive hover:underline"
                                                                type="button"
                                                            >
                                                                Remove
                                                            </button>
                                                        </div>
                                                    {:else}
                                                        <div
                                                            class="text-sm text-muted-foreground py-8 text-center"
                                                        >
                                                            No image selected
                                                        </div>
                                                    {/if}
                                                </div>
                                                <div>
                                                    <input
                                                        type="file"
                                                        accept="image/*,.ico"
                                                        on:change={(e) =>
                                                            handleImageUpload(
                                                                setting.id,
                                                                e,
                                                            )}
                                                        class="hidden"
                                                        id={`file-${setting.id}`}
                                                    />
                                                    <Button
                                                        variant="outline"
                                                        size="sm"
                                                        on:click={() =>
                                                            document
                                                                .getElementById(
                                                                    `file-${setting.id}`,
                                                                )
                                                                ?.click()}
                                                        type="button"
                                                    >
                                                        {settingsValues[
                                                            setting.id
                                                        ] &&
                                                        settingsValues[
                                                            setting.id
                                                        ] !== ""
                                                            ? "Change Image"
                                                            : "Upload Image"}
                                                    </Button>
                                                </div>
                                            </div>
                                        {/if}
                                    </div>
                                </div>
                            {/if}
                        {/each}
                    </div>
                </div>
            {/each}

            <!-- Save Button -->
            <div class="flex justify-end pt-6 mt-6 border-t">
                <Button on:click={saveSettings} disabled={saving}>
                    {#if saving}
                        <Loader2 class="h-4 w-4 animate-spin mr-2" />
                        Saving...
                    {:else}
                        <Save class="h-4 w-4 mr-2" />
                        Save Settings
                    {/if}
                </Button>
            </div>
        </div>
    {/if}
</div>
