<script lang="ts">
    import { onMount } from "svelte";
    import { Card } from "$lib/components/ui";
    import { api } from "$lib/api";
    import {
        ExternalLink,
        Eye,
        Download,
        Star,
        AlertCircle,
        Loader2,
        Check,
        RefreshCw,
    } from "lucide-svelte";

    interface Theme {
        id: string;
        name: string;
        display_name: string;
        description: string;
        version: string;
        demo_url?: string;
        build_file?: string;
        screenshots?: string[];
        features?: string[];
        technologies?: string[];
        stats?: {
            downloads?: number;
            rating?: number;
            reviews?: number;
        };
        submission_status: string;
        owner: string;
        created: string;
        updated: string;
    }

    let themes: Theme[] = [];
    let currentTheme: Theme | null = null;
    let selectedScreenshot = 0;
    let loading = true;
    let error: string | null = null;
    let currentThemeId: string | null = null;
    let currentThemeVersion: string | null = null;
    let applyingTheme = false;

    $: displayScreenshots =
        currentTheme?.screenshots && currentTheme.screenshots.length > 0
            ? currentTheme.screenshots
            : [];

    const API_BASE = "https://api.shipshipship.io";

    async function fetchThemes() {
        try {
            loading = true;
            error = null;

            // Fetch approved themes
            const response = await fetch(
                `${API_BASE}/api/collections/themes/records?filter=(submission_status='approved')&expand=owner`,
            );

            if (!response.ok) {
                throw new Error(
                    `Failed to fetch themes: ${response.status} ${response.statusText}`,
                );
            }

            const data = await response.json();
            themes = data.items || [];

            // Fetch current theme from backend using API client
            const themeData = await api.getCurrentTheme();
            currentThemeId = themeData.theme?.id || null;
            currentThemeVersion = themeData.theme?.version || null;

            // Set the current theme based on the ID, or find default theme
            if (currentThemeId && themes.length > 0) {
                const foundTheme = themes.find((t) => t.id === currentThemeId);
                if (foundTheme) {
                    currentTheme = foundTheme;
                } else {
                    // If current theme ID not found, try to find default theme
                    const defaultTheme = themes.find(
                        (t) => t.name === "default",
                    );
                    currentTheme = defaultTheme || themes[0];
                }
            } else if (themes.length > 0) {
                // If no current theme set, prefer default theme if available
                const defaultTheme = themes.find((t) => t.name === "default");
                currentTheme = defaultTheme || themes[0];
            } else {
                // Fallback theme if no themes are available from PocketBase
                currentTheme = {
                    id: "default-fallback",
                    name: "default",
                    display_name: "Default Theme",
                    description:
                        "The default ShipShipShip theme - please check your internet connection to load themes from the marketplace",
                    version: "1.0.0",
                    submission_status: "approved",
                    owner: "system",
                    created: new Date().toISOString(),
                    updated: new Date().toISOString(),
                    features: [
                        "Responsive design",
                        "Clean typography",
                        "Fast loading",
                        "Offline fallback",
                    ],
                    technologies: ["SvelteKit", "TypeScript", "Tailwind CSS"],
                    stats: { downloads: 0, rating: 0, reviews: 0 },
                };
            }
        } catch (err) {
            console.error("Error fetching themes:", err);
            error =
                err instanceof Error ? err.message : "Failed to load themes";
        } finally {
            loading = false;
        }
    }

    async function applyTheme(theme: Theme) {
        if (!theme.build_file) {
            alert("No build file available for this theme");
            return;
        }

        try {
            applyingTheme = true;

            // Call API client to apply theme
            const data = await api.applyTheme({
                id: theme.id,
                version: theme.version,
                buildFileUrl: getImageUrl("themes", theme.id, theme.build_file),
            });

            // Update current theme
            currentTheme = theme;
            currentThemeId = theme.id;
            currentThemeVersion = theme.version;

            // Show success message based on action
            const action = data.isUpdate ? "updated" : "applied";
            alert(
                `Theme "${theme.display_name}" ${action} successfully! The page will reload to show the new theme.`,
            );

            // Reload the page to show the new theme
            window.location.reload();
        } catch (err) {
            console.error("Error applying theme:", err);
            alert(err instanceof Error ? err.message : "Failed to apply theme");
        } finally {
            applyingTheme = false;
        }
    }

    function getImageUrl(
        collectionId: string,
        recordId: string,
        filename: string,
    ): string {
        if (!filename) return "";
        return `${API_BASE}/api/files/${collectionId}/${recordId}/${filename}`;
    }

    function getPlaceholderImage(text: string): string {
        return `data:image/svg+xml;base64,${btoa(`
            <svg width="800" height="600" xmlns="http://www.w3.org/2000/svg">
                <rect width="100%" height="100%" fill="#f3f4f6"/>
                <text x="50%" y="50%" font-family="Arial, sans-serif" font-size="24" fill="#6b7280" text-anchor="middle" dominant-baseline="middle">${text}</text>
            </svg>
        `)}`;
    }

    onMount(() => {
        fetchThemes();
    });

    function formatStats(stats: Theme["stats"]) {
        return {
            stars: stats?.rating ? Math.round(stats.rating * 10) : 0,
            downloads: stats?.downloads || 0,
        };
    }

    function compareVersions(version1: string, version2: string): number {
        // Simple version comparison - handles semantic versioning like "1.2.3"
        const v1Parts = version1
            .replace(/^v/, "")
            .split(".")
            .map((n) => parseInt(n) || 0);
        const v2Parts = version2
            .replace(/^v/, "")
            .split(".")
            .map((n) => parseInt(n) || 0);

        const maxLength = Math.max(v1Parts.length, v2Parts.length);

        for (let i = 0; i < maxLength; i++) {
            const v1Part = v1Parts[i] || 0;
            const v2Part = v2Parts[i] || 0;

            if (v1Part > v2Part) return 1;
            if (v1Part < v2Part) return -1;
        }

        return 0;
    }

    function getThemeButtonInfo(theme: Theme) {
        // Handle fallback theme case
        if (theme.id === "default-fallback") {
            return { text: "Offline", icon: "alert", variant: "warning" };
        }

        if (currentThemeId !== theme.id) {
            return { text: "Apply", icon: "download", variant: "neutral" };
        }

        if (!currentThemeVersion) {
            return { text: "Applied", icon: "check", variant: "success" };
        }

        const versionComparison = compareVersions(
            theme.version,
            currentThemeVersion,
        );
        if (versionComparison > 0) {
            return { text: "Update", icon: "refresh", variant: "amber" };
        }

        return { text: "Applied", icon: "check", variant: "success" };
    }
</script>

<svelte:head>
    <title>Theme - Customization - Admin</title>
</svelte:head>

<div class="max-w-6xl mx-auto">
    <div class="mb-6">
        <h1 class="text-2xl font-semibold mb-2">Theme</h1>
        <p class="text-muted-foreground">
            Manage your public changelog theme and appearance
        </p>
    </div>

    {#if loading}
        <!-- Loading State -->
        <Card class="p-8 text-center">
            <div class="flex items-center justify-center gap-3">
                <Loader2 class="h-6 w-6 animate-spin" />
                <span class="text-muted-foreground">Loading themes...</span>
            </div>
        </Card>
    {:else if error}
        <!-- Error State -->
        <Card class="p-8 text-center">
            <div class="max-w-md mx-auto space-y-4">
                <div
                    class="w-16 h-16 bg-red-100 dark:bg-red-900/20 rounded-full flex items-center justify-center mx-auto"
                >
                    <AlertCircle
                        class="h-8 w-8 text-red-600 dark:text-red-400"
                    />
                </div>
                <h3 class="text-lg font-medium">Failed to Load Themes</h3>
                <p class="text-muted-foreground text-sm">{error}</p>
                <button
                    on:click={fetchThemes}
                    class="px-4 py-2 bg-primary text-primary-foreground rounded-md hover:bg-primary/90 transition-colors text-sm font-medium"
                >
                    Try Again
                </button>
            </div>
        </Card>
    {:else if currentTheme}
        <!-- Current Theme Section -->
        <div class="mb-8">
            <h2 class="text-lg font-medium mb-4">Current Theme</h2>

            <Card class="p-6">
                <div class="grid grid-cols-1 lg:grid-cols-2 gap-8">
                    <!-- Theme Preview -->
                    <div class="space-y-4">
                        <div
                            class="aspect-video bg-muted rounded-lg overflow-hidden border"
                        >
                            <img
                                src={currentTheme?.screenshots &&
                                currentTheme.screenshots.length > 0
                                    ? getImageUrl(
                                          "themes",
                                          currentTheme.id,
                                          currentTheme.screenshots[
                                              selectedScreenshot
                                          ],
                                      )
                                    : getPlaceholderImage(
                                          currentTheme?.display_name ||
                                              "No Preview Available",
                                      )}
                                alt="Theme preview"
                                class="w-full h-full object-cover"
                            />
                        </div>

                        <!-- Screenshot thumbnails -->
                        {#if currentTheme?.screenshots && currentTheme.screenshots.length > 1}
                            <div class="flex gap-2">
                                {#each currentTheme.screenshots as screenshot, index}
                                    <button
                                        on:click={() =>
                                            (selectedScreenshot = index)}
                                        class="w-16 h-12 rounded border overflow-hidden {selectedScreenshot ===
                                        index
                                            ? 'ring-2 ring-primary'
                                            : 'opacity-60 hover:opacity-100'} transition-all"
                                    >
                                        <img
                                            src={getImageUrl(
                                                "themes",
                                                currentTheme.id,
                                                screenshot,
                                            )}
                                            alt="Screenshot {index + 1}"
                                            class="w-full h-full object-cover"
                                            on:error={(e) => {
                                                let img = e.currentTarget;
                                                if (
                                                    img instanceof
                                                    HTMLImageElement
                                                ) {
                                                    img.src =
                                                        getPlaceholderImage(
                                                            `${index + 1}`,
                                                        );
                                                }
                                            }}
                                        />
                                    </button>
                                {/each}
                            </div>
                        {/if}

                        <!-- Action buttons -->
                        <div class="flex gap-3">
                            {#if currentTheme.demo_url}
                                <a
                                    href={currentTheme.demo_url}
                                    target="_blank"
                                    rel="noopener noreferrer"
                                    class="flex items-center gap-2 px-4 py-2 bg-muted text-muted-foreground rounded-md hover:bg-muted/90 transition-colors text-sm font-medium"
                                >
                                    <Eye class="h-4 w-4" />
                                    Live Preview
                                    <ExternalLink class="h-3 w-3" />
                                </a>
                            {/if}

                            {#if currentTheme.id === "default-fallback"}
                                <span
                                    class="flex items-center gap-2 px-4 py-2 bg-yellow-100 text-yellow-800 dark:bg-yellow-900/20 dark:text-yellow-400 rounded-md text-sm font-medium"
                                >
                                    <AlertCircle class="h-4 w-4" />
                                    Offline Mode
                                </span>
                            {:else if currentThemeId === currentTheme.id && currentThemeVersion && compareVersions(currentTheme.version, currentThemeVersion) > 0}
                                <button
                                    on:click={() => {
                                        if (currentTheme) {
                                            applyTheme(currentTheme);
                                        }
                                    }}
                                    disabled={applyingTheme}
                                    class="flex items-center gap-2 px-4 py-2 bg-amber-600 text-white rounded-md hover:bg-amber-700 transition-colors text-sm font-medium disabled:opacity-50 disabled:cursor-not-allowed"
                                >
                                    {#if applyingTheme}
                                        <Loader2 class="h-4 w-4 animate-spin" />
                                        Updating...
                                    {:else}
                                        <RefreshCw class="h-4 w-4" />
                                        Update Theme
                                    {/if}
                                </button>
                            {:else}
                                <!-- "Currently Applied" button removed as redundant with "Active" badge -->
                            {/if}
                        </div>
                    </div>

                    <!-- Theme Details -->
                    <div class="space-y-6">
                        <div>
                            <div class="flex items-center gap-3 mb-2">
                                <h3 class="text-xl font-semibold">
                                    {currentTheme.display_name}
                                </h3>
                                <span
                                    class="inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium bg-green-100 text-green-800 dark:bg-green-900/20 dark:text-green-400"
                                >
                                    Active
                                </span>
                            </div>
                            {#if currentTheme.description}
                                <p class="text-muted-foreground mb-3">
                                    {currentTheme.description}
                                </p>
                            {/if}
                            <div class="flex items-center gap-2">
                                <p class="text-sm text-muted-foreground">
                                    Version {currentTheme.version}
                                </p>
                                {#if currentThemeId === currentTheme.id && currentThemeVersion && compareVersions(currentTheme.version, currentThemeVersion) > 0}
                                    <span
                                        class="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-amber-100 text-amber-800 dark:bg-amber-900/20 dark:text-amber-400"
                                    >
                                        Update Available
                                    </span>
                                {/if}
                            </div>
                        </div>

                        <!-- Stats -->
                        {#if currentTheme.stats?.downloads}
                            <div class="flex gap-6">
                                <div
                                    class="flex items-center gap-2 text-sm text-muted-foreground"
                                >
                                    <Download class="h-4 w-4" />
                                    {currentTheme.stats.downloads}
                                </div>
                                {#if currentTheme.stats.rating}
                                    <div
                                        class="flex items-center gap-2 text-sm text-muted-foreground"
                                    >
                                        <Star class="h-4 w-4" />
                                        {currentTheme.stats.rating}
                                    </div>
                                {/if}
                            </div>
                        {/if}

                        <!-- Features -->
                        {#if currentTheme.features && currentTheme.features.length > 0}
                            <div>
                                <h4 class="font-medium mb-3">Features</h4>
                                <div class="grid grid-cols-2 gap-2">
                                    {#each currentTheme.features as feature}
                                        <div
                                            class="flex items-center gap-2 text-sm"
                                        >
                                            <div
                                                class="w-1.5 h-1.5 bg-primary rounded-full"
                                            ></div>
                                            {feature}
                                        </div>
                                    {/each}
                                </div>
                            </div>
                        {/if}

                        <!-- Technologies -->
                        {#if currentTheme.technologies && currentTheme.technologies.length > 0}
                            <div>
                                <h4 class="font-medium mb-3">Built with</h4>
                                <div class="flex flex-wrap gap-2">
                                    {#each currentTheme.technologies as tech}
                                        <span
                                            class="inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium bg-muted text-muted-foreground"
                                        >
                                            {tech}
                                        </span>
                                    {/each}
                                </div>
                            </div>
                        {/if}
                    </div>
                </div>
            </Card>
        </div>

        <!-- Available Themes Section -->
        {#if themes.length > 1}
            <div class="mb-8">
                <h2 class="text-lg font-medium mb-4">
                    Available Themes ({themes.length})
                </h2>

                <div
                    class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4"
                >
                    {#each themes as theme}
                        {@const buttonInfo = getThemeButtonInfo(theme)}
                        <Card
                            class="p-4 hover:shadow-lg transition-shadow {theme.id ===
                            currentTheme.id
                                ? 'ring-2 ring-primary'
                                : ''}"
                            on:click={() => {
                                currentTheme = theme;
                                selectedScreenshot = 0;
                            }}
                        >
                            <div class="space-y-3">
                                <!-- Theme preview -->
                                <div
                                    class="aspect-video bg-muted rounded-lg overflow-hidden border"
                                >
                                    <img
                                        src={theme.screenshots &&
                                        theme.screenshots.length > 0
                                            ? getImageUrl(
                                                  "themes",
                                                  theme.id,
                                                  theme.screenshots[0],
                                              )
                                            : getPlaceholderImage(
                                                  theme.display_name,
                                              )}
                                        alt="{theme.display_name} preview"
                                        class="w-full h-full object-cover"
                                        on:error={(e) => {
                                            let img = e.currentTarget;
                                            if (
                                                img instanceof HTMLImageElement
                                            ) {
                                                img.src = getPlaceholderImage(
                                                    theme.display_name,
                                                );
                                            }
                                        }}
                                    />
                                </div>

                                <!-- Theme info -->
                                <div>
                                    <h3 class="font-medium">
                                        {theme.display_name}
                                    </h3>
                                    <p
                                        class="text-sm text-muted-foreground line-clamp-2"
                                    >
                                        {theme.description ||
                                            "No description available"}
                                    </p>
                                    <div
                                        class="flex items-center justify-between mt-1"
                                    >
                                        <p
                                            class="text-xs text-muted-foreground"
                                        >
                                            Version {theme.version}
                                        </p>
                                        {#if currentThemeId === theme.id && currentThemeVersion && compareVersions(theme.version, currentThemeVersion) > 0}
                                            <span
                                                class="inline-flex items-center px-1.5 py-0.5 rounded text-xs font-medium bg-amber-100 text-amber-800 dark:bg-amber-900/20 dark:text-amber-400"
                                            >
                                                Update
                                            </span>
                                        {/if}
                                    </div>
                                </div>

                                <!-- Stats -->
                                {#if theme.stats?.downloads}
                                    <div
                                        class="flex items-center gap-2 text-xs text-muted-foreground"
                                    >
                                        <Download class="h-3 w-3" />
                                        {theme.stats.downloads} downloads
                                    </div>
                                {/if}

                                <!-- Apply Button -->
                                <div class="flex gap-2">
                                    {#if theme.demo_url}
                                        <a
                                            href={theme.demo_url}
                                            target="_blank"
                                            class="flex-1 flex items-center justify-center gap-1 px-3 py-1.5 text-xs border border-border rounded-md hover:bg-accent transition-colors"
                                        >
                                            <Eye class="h-3 w-3" />
                                            Preview
                                        </a>
                                    {/if}
                                    {#if buttonInfo.variant === "success"}
                                        <span
                                            class="flex-1 flex items-center justify-center gap-1 px-3 py-1.5 text-xs bg-green-100 text-green-800 dark:bg-green-900/20 dark:text-green-400 rounded-md"
                                        >
                                            <Check class="h-3 w-3" />
                                            {buttonInfo.text}
                                        </span>
                                    {:else if buttonInfo.variant === "warning"}
                                        <span
                                            class="flex-1 flex items-center justify-center gap-1 px-3 py-1.5 text-xs bg-yellow-100 text-yellow-800 dark:bg-yellow-900/20 dark:text-yellow-400 rounded-md"
                                        >
                                            <AlertCircle class="h-3 w-3" />
                                            {buttonInfo.text}
                                        </span>
                                    {:else if buttonInfo.variant === "amber"}
                                        <button
                                            on:click={() => applyTheme(theme)}
                                            disabled={applyingTheme ||
                                                theme.id === "default-fallback"}
                                            class="flex-1 flex items-center justify-center gap-1 px-3 py-1.5 text-xs bg-amber-600 text-white rounded-md hover:bg-amber-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                                        >
                                            {#if applyingTheme}
                                                <Loader2
                                                    class="h-3 w-3 animate-spin"
                                                />
                                                Updating...
                                            {:else}
                                                <RefreshCw class="h-3 w-3" />
                                                {buttonInfo.text}
                                            {/if}
                                        </button>
                                    {:else if buttonInfo.variant === "neutral"}
                                        <button
                                            on:click={() => applyTheme(theme)}
                                            disabled={applyingTheme ||
                                                theme.id === "default-fallback"}
                                            class="flex-1 flex items-center justify-center gap-1 px-3 py-1.5 text-xs bg-white dark:bg-black text-black dark:text-white border border-border rounded-md hover:bg-accent transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                                        >
                                            {#if applyingTheme}
                                                <Loader2
                                                    class="h-3 w-3 animate-spin"
                                                />
                                                Applying...
                                            {:else}
                                                <Download class="h-3 w-3" />
                                                {buttonInfo.text}
                                            {/if}
                                        </button>
                                    {:else}
                                        <button
                                            on:click={() => applyTheme(theme)}
                                            disabled={applyingTheme ||
                                                theme.id === "default-fallback"}
                                            class="flex-1 flex items-center justify-center gap-1 px-3 py-1.5 text-xs bg-primary text-primary-foreground rounded-md hover:bg-primary/90 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                                        >
                                            {#if applyingTheme}
                                                <Loader2
                                                    class="h-3 w-3 animate-spin"
                                                />
                                                Applying...
                                            {:else}
                                                <Check class="h-3 w-3" />
                                                {buttonInfo.text}
                                            {/if}
                                        </button>
                                    {/if}
                                </div>
                            </div>
                        </Card>
                    {/each}
                </div>
            </div>
        {/if}
    {/if}
</div>

<style>
    .line-clamp-2 {
        display: -webkit-box;
        -webkit-line-clamp: 2;
        -webkit-box-orient: vertical;
        overflow: hidden;
        line-clamp: 2;
    }
</style>
