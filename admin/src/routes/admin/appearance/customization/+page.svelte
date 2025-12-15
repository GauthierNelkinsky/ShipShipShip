<script lang="ts">
    import { onMount } from "svelte";
    import { goto } from "$app/navigation";
    import { Palette, Map, AlertTriangle } from "lucide-svelte";
    import * as m from "$lib/paraglide/messages";
    import { emptyCategoriesStore } from "$lib/stores/emptyCategories";
    import { api } from "$lib/api";
    import ThemeErrorMessage from "$lib/components/ThemeErrorMessage.svelte";

    let loading = true;
    let currentTab = "settings";
    let themeError = false;
    let currentThemeId: string | null = null;

    $: hasEmptyCategories = $emptyCategoriesStore.hasEmptyCategories;

    // Navigation items
    $: navItems = [
        { id: "settings", label: m.branding_tab_settings(), icon: Palette },
        {
            id: "status-mapping",
            label: m.branding_tab_status_mapping(),
            icon: Map,
        },
    ];

    async function checkThemeManifest() {
        try {
            const manifestData = await api.getThemeManifest();
            const manifest = manifestData.manifest;

            // Validate manifest structure
            if (
                !manifest ||
                !manifest.settings ||
                !Array.isArray(manifest.settings)
            ) {
                themeError = true;
                await checkCurrentTheme();
                return false;
            }
            themeError = false;
            return true;
        } catch (err) {
            console.error("Failed to load theme manifest:", err);
            const errorMessage = err instanceof Error ? err.message : "";
            if (
                errorMessage.includes("Failed to load theme manifest") ||
                errorMessage.includes("theme manifest")
            ) {
                themeError = true;
                await checkCurrentTheme();
            }
            return false;
        }
    }

    async function checkCurrentTheme() {
        try {
            const settingsResponse = await api.getSettings();
            currentThemeId = settingsResponse.current_theme_id || null;
        } catch (err) {
            console.error("Failed to get current theme:", err);
            currentThemeId = null;
        }
    }

    onMount(async () => {
        loading = false;
        // Set default tab based on URL or default to settings
        const urlParams = new URLSearchParams(window.location.search);
        currentTab = urlParams.get("tab") || "settings";

        // Check theme manifest
        await checkThemeManifest();

        // Check for empty categories
        emptyCategoriesStore.check();
    });

    function switchTab(tabId: string) {
        currentTab = tabId;
        const url = new URL(window.location.href);
        url.searchParams.set("tab", tabId);
        goto(url.pathname + url.search, { replaceState: true });
    }
</script>

<svelte:head>
    <title>{m.branding_page_title()}</title>
</svelte:head>

<div class="max-w-6xl mx-auto">
    <!-- Header -->
    <div class="mb-4">
        <h1 class="text-xl font-semibold mb-1">{m.branding_heading()}</h1>
        <p class="text-muted-foreground text-sm">
            {m.branding_subheading()}
        </p>
    </div>

    {#if loading}
        <div class="flex items-center justify-center min-h-32">
            <div
                class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"
            ></div>
        </div>
    {:else if themeError}
        <ThemeErrorMessage {currentThemeId} />
    {:else}
        <!-- Navigation Menu -->
        <nav class="mb-6">
            <div class="border-b border-border">
                <div class="flex space-x-8">
                    {#each navItems as item}
                        <button
                            on:click={() => switchTab(item.id)}
                            class="flex items-center gap-2 py-4 px-1 border-b-2 font-medium text-sm transition-colors {currentTab ===
                            item.id
                                ? 'border-primary text-primary'
                                : 'border-transparent text-muted-foreground hover:text-foreground hover:border-border'}"
                        >
                            <svelte:component
                                this={item.icon}
                                class="h-4 w-4"
                            />
                            {item.label}
                            {#if item.id === "status-mapping" && hasEmptyCategories}
                                <AlertTriangle
                                    class="h-3.5 w-3.5 text-amber-500 opacity-80"
                                    title={m.customization_empty_categories_tooltip()}
                                />
                            {/if}
                        </button>
                    {/each}
                </div>
            </div>
        </nav>

        <!-- Tab Content -->
        <div class="space-y-6">
            {#if currentTab === "settings"}
                {#await import("./settings/+page.svelte")}
                    <div class="flex items-center justify-center py-8">
                        <div
                            class="animate-spin rounded-full h-6 w-6 border-b-2 border-primary"
                        ></div>
                    </div>
                {:then { default: SettingsPage }}
                    <svelte:component this={SettingsPage} />
                {:catch}
                    <div class="text-center py-8 text-red-600">
                        {m.customization_failed_load_settings()}
                    </div>
                {/await}
            {:else if currentTab === "status-mapping"}
                {#await import("./status-mapping/+page.svelte")}
                    <div class="flex items-center justify-center py-8">
                        <div
                            class="animate-spin rounded-full h-6 w-6 border-b-2 border-primary"
                        ></div>
                    </div>
                {:then { default: StatusMappingPage }}
                    <svelte:component this={StatusMappingPage} />
                {:catch}
                    <div class="text-center py-8 text-red-600">
                        {m.customization_failed_load_status_mapping()}
                    </div>
                {/await}
            {/if}
        </div>
    {/if}
</div>
