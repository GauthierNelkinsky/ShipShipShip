<script lang="ts">
    import { onMount } from "svelte";
    import { api } from "$lib/api";
    import { Card } from "$lib/components/ui";
    import {
        Eye,
        Download,
        AlertCircle,
        Loader2,
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
    let noThemeInstalled = false;

    $: _displayScreenshots =
        currentTheme?.screenshots && currentTheme.screenshots.length > 0
            ? currentTheme.screenshots
            : [];

    const API_BASE = "https://api.shipshipship.io"; // For PocketBase theme marketplace

    async function fetchThemes() {
        try {
            loading = true;
            error = null;

            // Check environment mode from backend
            const settingsData = await api.getSettings();
            const isDevelopment = settingsData.environment === "development";

            // Build filter based on environment
            let filter = "(submission_status='approved')";
            if (isDevelopment) {
                // In development, fetch approved OR staging themes
                filter =
                    "(submission_status='approved'||submission_status='staging')";
            }

            // Fetch themes
            const response = await fetch(
                `${API_BASE}/api/collections/themes/records?filter=${filter}&expand=owner`,
            );

            if (!response.ok) {
                throw new Error(
                    `Failed to fetch themes: ${response.status} ${response.statusText}`,
                );
            }

            const data = await response.json();
            themes = data.items || [];

            // Fetch current theme ID from backend
            await fetchCurrentTheme();

            // Check if theme files actually exist by calling local backend API
            let themeFilesExist = false;
            try {
                const themeInfo = await api.getThemeInfo();
                themeFilesExist = themeInfo.current?.exists === true;
            } catch (e) {
                console.error("Error checking theme files:", e);
            }

            // If DB says theme installed but files don't exist, treat as no theme
            if (currentThemeId && !themeFilesExist) {
                noThemeInstalled = true;
                currentTheme = null;
                return;
            }

            // Determine theme installation status
            // Priority: files exist = theme is installed (regardless of DB or marketplace)
            if (!currentThemeId && !themeFilesExist) {
                // No theme at all
                noThemeInstalled = true;
                currentTheme = null;
            } else if (currentThemeId && themeFilesExist) {
                // Theme is installed - show it
                noThemeInstalled = false;

                if (themes.length > 0) {
                    const foundTheme = themes.find(
                        (t) => t.id === currentThemeId,
                    );
                    if (foundTheme) {
                        currentTheme = foundTheme;
                    } else {
                        // Theme installed but not in marketplace - use first available
                        const defaultTheme = themes.find(
                            (t) => t.name === "shipshipship-template-default",
                        );
                        currentTheme = defaultTheme || themes[0];
                    }
                } else {
                    // Theme installed but can't fetch marketplace
                    currentTheme = null;
                }
            } else {
                // Edge case: shouldn't happen but handle gracefully
                noThemeInstalled = !themeFilesExist;
                currentTheme = null;
            }
        } catch (err) {
            console.error("Error fetching themes:", err);
            error =
                err instanceof Error ? err.message : "Failed to load themes";
        } finally {
            loading = false;
        }
    }

    async function fetchCurrentTheme() {
        try {
            const data = await api.getCurrentTheme();
            currentThemeId = data.currentThemeId || null;
            currentThemeVersion = data.currentThemeVersion || null;
        } catch (err) {
            console.error("Error fetching current theme:", err);
            // Don't fail the whole operation if we can't get current theme
        }
    }

    async function applyTheme(theme: Theme) {
        if (!theme.build_file) {
            alert("No build file available for this theme");
            return;
        }

        try {
            applyingTheme = true;

            // Call backend API to apply theme
            const data = await api.applyTheme(
                theme.id,
                theme.version,
                getImageUrl("themes", theme.id, theme.build_file),
            );

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

    function _formatStats(stats: Theme["stats"]) {
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

<div class="w-full">
    <div class="mb-8">
        <h1 class="text-xl font-semibold mb-1">Theme</h1>
        <p class="text-muted-foreground text-sm">
            Customize the look and feel of your changelog
        </p>
    </div>

    {#if loading}
        <!-- Loading state -->
        <Card class="p-8 text-center">
            <div class="flex items-center justify-center gap-3">
                <Loader2 class="h-6 w-6 animate-spin" />
                <span class="text-muted-foreground">Loading themes...</span>
            </div>
        </Card>
    {:else if error}
        <!-- Error state -->
        <Card class="p-8 text-center">
            <div class="max-w-md mx-auto space-y-4">
                <div
                    class="mx-auto w-12 h-12 rounded-full bg-destructive/10 flex items-center justify-center"
                >
                    <AlertCircle
                        class="h-6 w-6 text-destructive"
                        strokeWidth={2}
                    />
                </div>
                <h3 class="text-lg font-medium">Failed to load themes</h3>
                <p class="text-muted-foreground text-sm">{error}</p>
                <button
                    on:click={fetchThemes}
                    class="px-4 py-2 bg-primary text-primary-foreground rounded-md hover:bg-primary/90 transition-colors"
                >
                    Try Again
                </button>
            </div>
        </Card>
    {:else if noThemeInstalled}
        <!-- No theme installed state -->
        <div
            class="mb-6 p-6 border border-amber-500/20 bg-amber-500/5 rounded-lg"
        >
            <div class="flex items-start gap-3">
                <div class="flex-shrink-0 mt-0.5">
                    <AlertCircle class="h-5 w-5 text-amber-500" />
                </div>
                <div class="flex-1 min-w-0">
                    <h3 class="text-sm font-medium text-foreground mb-1">
                        No Theme Installed
                    </h3>
                    <p class="text-sm text-muted-foreground">
                        You don't have a theme installed yet. Your changelog is
                        currently showing the admin interface at the root path.
                        {#if themes.length > 0}
                            Select and install a theme below to set up your
                            public changelog.
                        {:else}
                            Please check your internet connection to load themes
                            from the marketplace.
                        {/if}
                    </p>
                </div>
                <button
                    on:click={fetchThemes}
                    class="flex-shrink-0 px-3 py-1.5 text-sm border rounded-md hover:bg-accent transition-colors"
                >
                    Refresh
                </button>
            </div>
        </div>

        {#if themes.length > 0}
            <!-- Show available themes to install -->
            <div class="mb-6">
                <h2 class="text-sm font-medium text-muted-foreground mb-3">
                    Available Themes
                </h2>
                <div
                    class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3"
                >
                    {#each themes as theme}
                        <Card
                            class="overflow-hidden hover:border-primary/50 transition-colors"
                        >
                            <div class="space-y-3 p-4">
                                <!-- Theme screenshot -->
                                <div
                                    class="aspect-video bg-muted relative overflow-hidden rounded-md"
                                >
                                    <img
                                        src={theme.screenshots?.[0]
                                            ? `${API_BASE}/api/files/themes/${theme.id}/${theme.screenshots[0]}`
                                            : getPlaceholderImage(
                                                  theme.display_name,
                                              )}
                                        alt={theme.display_name}
                                        class="w-full h-full object-cover"
                                        on:error={(e) => {
                                            let img = e.target;
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
                                    <h3 class="text-sm font-medium mb-1">
                                        {theme.display_name}
                                    </h3>
                                    <p
                                        class="text-xs text-muted-foreground line-clamp-2 mb-2"
                                    >
                                        {theme.description ||
                                            "No description available"}
                                    </p>
                                    <span
                                        class="inline-flex items-center px-2 py-0.5 text-xs bg-muted rounded"
                                    >
                                        {theme.version}
                                    </span>
                                </div>

                                {#if theme.stats?.downloads}
                                    <div
                                        class="flex items-center gap-1.5 text-xs text-muted-foreground"
                                    >
                                        <Download class="h-3 w-3" />
                                        <span
                                            >{theme.stats.downloads.toLocaleString()}</span
                                        >
                                    </div>
                                {/if}

                                <!-- Actions -->
                                <div class="flex gap-2 pt-2 border-t">
                                    {#if theme.demo_url}
                                        <a
                                            href={theme.demo_url}
                                            target="_blank"
                                            class="flex-1 px-3 py-1.5 text-xs border rounded-md hover:bg-accent transition-colors flex items-center justify-center gap-1.5"
                                        >
                                            <Eye class="h-3 w-3" />
                                            Preview
                                        </a>
                                    {/if}
                                    <button
                                        on:click={() => applyTheme(theme)}
                                        disabled={applyingTheme}
                                        class="flex-1 px-3 py-1.5 text-xs bg-primary text-primary-foreground rounded-md hover:bg-primary/90 disabled:opacity-50 transition-colors flex items-center justify-center gap-1.5"
                                    >
                                        {#if applyingTheme}
                                            <Loader2
                                                class="h-3 w-3 animate-spin"
                                            />
                                            Installing...
                                        {:else}
                                            <Download class="h-3 w-3" />
                                            Install
                                        {/if}
                                    </button>
                                </div>
                            </div>
                        </Card>
                    {/each}
                </div>
            </div>
        {/if}
    {:else if currentTheme}
        <!-- Current Theme Section -->
        <div class="mb-6">
            <h2 class="text-sm font-medium text-muted-foreground mb-3">
                Current Theme
            </h2>

            <Card class="p-6">
                <div class="grid grid-cols-1 lg:grid-cols-2 gap-8">
                    <!-- Theme Preview Column -->
                    <div class="space-y-4">
                        <!-- Main screenshot -->
                        <div
                            class="aspect-video bg-muted rounded-md overflow-hidden border"
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
                                on:error={(e) => {
                                    let img = e.target;
                                    if (img instanceof HTMLImageElement) {
                                        img.src = getPlaceholderImage(
                                            currentTheme?.display_name ||
                                                "Theme",
                                        );
                                    }
                                }}
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
                        {#if currentThemeId === currentTheme.id && currentThemeVersion && compareVersions(currentTheme.version, currentThemeVersion) > 0}
                            <div class="flex gap-2">
                                <button
                                    on:click={() => applyTheme(currentTheme!)}
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
                            </div>
                        {/if}
                    </div>

                    <!-- Theme Details Column -->
                    <div class="space-y-6">
                        <!-- Title and badge -->
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
                                <p class="text-muted-foreground mb-3 text-sm">
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
                                    {currentTheme.stats.downloads.toLocaleString()}
                                </div>
                            </div>
                        {/if}

                        <!-- Features -->
                        {#if currentTheme.features && currentTheme.features.length > 0}
                            <div>
                                <h4 class="font-medium mb-3 text-sm">
                                    Features
                                </h4>
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
                                <h4 class="font-medium mb-3 text-sm">
                                    Built with
                                </h4>
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
            <div class="mb-6">
                <h2 class="text-sm font-medium text-muted-foreground mb-3">
                    Other Available Themes
                </h2>

                <div
                    class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3"
                >
                    {#each themes as theme}
                        {@const buttonInfo = getThemeButtonInfo(theme)}
                        <Card
                            class="overflow-hidden hover:border-primary/50 transition-colors"
                        >
                            <div class="space-y-3 p-4">
                                <!-- Theme screenshot -->
                                <div
                                    class="aspect-video bg-muted relative overflow-hidden rounded-md"
                                >
                                    <img
                                        src={theme.screenshots?.[0]
                                            ? `${API_BASE}/api/files/themes/${theme.id}/${theme.screenshots[0]}`
                                            : getPlaceholderImage(
                                                  theme.display_name,
                                              )}
                                        alt={theme.display_name}
                                        class="w-full h-full object-cover"
                                        on:error={(e) => {
                                            let img = e.target;
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
                                    <h3 class="text-sm font-medium mb-1">
                                        {theme.display_name}
                                    </h3>
                                    <p
                                        class="text-xs text-muted-foreground line-clamp-2 mb-2"
                                    >
                                        {theme.description ||
                                            "No description available"}
                                    </p>
                                    <span
                                        class="inline-flex items-center px-2 py-0.5 text-xs bg-muted rounded"
                                    >
                                        {theme.version}
                                    </span>
                                </div>

                                {#if theme.stats?.downloads}
                                    <div
                                        class="flex items-center gap-1.5 text-xs text-muted-foreground"
                                    >
                                        <Download class="h-3 w-3" />
                                        <span
                                            >{theme.stats.downloads.toLocaleString()}</span
                                        >
                                    </div>
                                {/if}

                                <!-- Actions -->
                                <div class="flex gap-2 pt-2 border-t">
                                    {#if theme.demo_url}
                                        <a
                                            href={theme.demo_url}
                                            target="_blank"
                                            class="flex-1 px-3 py-1.5 text-xs border rounded-md hover:bg-accent transition-colors flex items-center justify-center gap-1.5"
                                        >
                                            <Eye class="h-3 w-3" />
                                            Preview
                                        </a>
                                    {/if}
                                    {#if buttonInfo.variant === "amber"}
                                        <button
                                            on:click={() => applyTheme(theme)}
                                            disabled={applyingTheme}
                                            class="flex-1 px-3 py-1.5 text-xs bg-amber-600 text-white rounded-md hover:bg-amber-700 disabled:opacity-50 transition-colors flex items-center justify-center gap-1.5"
                                        >
                                            {#if applyingTheme}
                                                <Loader2
                                                    class="h-3 w-3 animate-spin"
                                                />
                                                Updating...
                                            {:else}
                                                <RefreshCw class="h-3 w-3" />
                                                Update
                                            {/if}
                                        </button>
                                    {:else if buttonInfo.variant === "neutral"}
                                        <button
                                            on:click={() => applyTheme(theme)}
                                            disabled={applyingTheme}
                                            class="flex-1 px-3 py-1.5 text-xs bg-primary text-primary-foreground rounded-md hover:bg-primary/90 disabled:opacity-50 transition-colors flex items-center justify-center gap-1.5"
                                        >
                                            {#if applyingTheme}
                                                <Loader2
                                                    class="h-3 w-3 animate-spin"
                                                />
                                                Installing...
                                            {:else}
                                                <Download class="h-3 w-3" />
                                                Install
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
