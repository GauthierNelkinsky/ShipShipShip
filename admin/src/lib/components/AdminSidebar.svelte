<script lang="ts">
    import { page } from "$app/stores";
    import { goto } from "$app/navigation";
    import { authStore } from "$lib/stores/auth";
    import { theme } from "$lib/stores/theme";
    import { api } from "$lib/api";
    import { onMount } from "svelte";

    import {
        Calendar,
        LogOut,
        ChevronLeft,
        ChevronRight,
        Palette,
        Building2,
        Tag,
        ChevronDown,
        ChevronRight as ChevronRightIcon,
        Mail,
        Monitor,
        Globe,
        Github,
        Sun,
        Moon,
    } from "lucide-svelte";

    export let collapsed = false;

    let customizationExpanded = false;
    let themeUpdateAvailable = false;
    let currentThemeId = null;
    let currentThemeVersion = null;
    let selectedTheme: "light" | "dark" = "light";

    const menuItems = [
        {
            label: "Events",
            href: "/admin/events",
            icon: Calendar,
        },
        {
            label: "Customization",
            href: "/admin/customization",
            icon: Palette,
            children: [
                {
                    label: "Branding",
                    href: "/admin/customization/branding",
                    icon: Building2,
                },
                {
                    label: "Tags",
                    href: "/admin/customization/tags",
                    icon: Tag,
                },
                {
                    label: "Theme",
                    href: "/admin/customization/theme",
                    icon: Monitor,
                },
            ],
        },
        {
            label: "Newsletter",
            href: "/admin/newsletter",
            icon: Mail,
        },
    ];

    $: currentPath = $page?.url?.pathname || "";

    // Auto-expand customization if we're on a customization page
    $: if (currentPath && currentPath.startsWith("/admin/customization")) {
        customizationExpanded = true;
    }

    function isActive(href: string): boolean {
        if (!currentPath || !href) return false;

        // Handle admin events page (includes redirect from root)
        if (href === "/admin/events") {
            return (
                currentPath === "/" ||
                currentPath === "/admin/events" ||
                currentPath.startsWith("/admin/events/")
            );
        }

        // Handle customization pages
        if (href === "/admin/customization") {
            return currentPath.startsWith("/admin/customization");
        }

        if (href.includes("/admin/customization/branding")) {
            return currentPath.includes("/admin/customization/branding");
        }

        if (href.includes("/admin/customization/tags")) {
            return currentPath.includes("/admin/customization/tags");
        }

        if (href.includes("/admin/customization/theme")) {
            return currentPath.includes("/admin/customization/theme");
        }

        // Exact match for other paths
        return currentPath === href;
    }

    function isChildActive(_children: any[]): boolean {
        return _children.some((child) => isActive(child.href));
    }

    function isParentActive(href: string, _children: any[]): boolean {
        // Only highlight parent if we're on the parent page itself, not on children
        return currentPath === href;
    }

    function toggleCustomization() {
        customizationExpanded = !customizationExpanded;
    }

    function toggleTheme() {
        selectedTheme = selectedTheme === "light" ? "dark" : "light";
        theme.setPreference(selectedTheme);
    }

    function handleLogout() {
        authStore.logout();
        if (!$authStore.isDemoMode) {
            goto("/login");
        }
    }

    function toggleSidebar() {
        collapsed = !collapsed;

        // Re-check for theme updates when toggling sidebar
        // This ensures the notification dot appears correctly
        checkThemeUpdates();
    }

    async function checkThemeUpdates() {
        try {
            // Fetch current theme info
            const data = await api.getCurrentTheme();
            currentThemeId = data.currentThemeId || null;
            currentThemeVersion = data.currentThemeVersion || null;

            if (currentThemeId && currentThemeVersion) {
                // Check environment mode from backend
                const settingsData = await api.getSettings();
                const isDevelopment =
                    settingsData.environment === "development";

                // Build filter based on environment
                let filter = "(submission_status='approved')";
                if (isDevelopment) {
                    filter =
                        "(submission_status='approved'||submission_status='staging')";
                }

                // Fetch available themes to check for updates
                const themesResponse = await fetch(
                    `https://api.shipshipship.io/api/collections/themes/records?filter=${filter}&expand=owner`,
                );

                if (themesResponse.ok) {
                    const themesData = await themesResponse.json();
                    const themes = themesData.items || [];

                    // Find current theme
                    const currentTheme = themes.find(
                        (t) => t.id === currentThemeId,
                    );

                    if (currentTheme) {
                        // Compare versions
                        themeUpdateAvailable =
                            compareVersions(
                                currentTheme.version,
                                currentThemeVersion,
                            ) > 0;
                    }
                }
            }
        } catch (error) {
            console.error("Error checking theme updates:", error);
        }
    }

    function compareVersions(version1, version2) {
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

    // Check for theme updates when the component mounts and periodically
    onMount(() => {
        // Initialize theme preference
        selectedTheme = theme.getPreference();

        checkThemeUpdates();

        // Set up a periodic check for theme updates (every 5 minutes)
        const updateCheckInterval = setInterval(
            checkThemeUpdates,
            5 * 60 * 1000,
        );

        // Clean up interval on component unmount
        return () => {
            clearInterval(updateCheckInterval);
        };
    });
</script>

<aside
    class="flex flex-col h-screen bg-background border-r border-border transition-all duration-300 {collapsed
        ? 'w-16'
        : 'w-64'} fixed left-0 top-0 z-40"
>
    <!-- Header -->
    <div
        class="flex items-center justify-between p-4 h-14 border-b border-border"
    >
        {#if !collapsed}
            <div class="flex items-center gap-2">
                <span class="font-medium text-foreground text-sm"
                    >ShipShipShip</span
                >
            </div>
        {/if}

        <button
            on:click={toggleSidebar}
            class="w-7 h-7 flex items-center justify-center rounded-md hover:bg-accent transition-colors"
            title={collapsed ? "Expand sidebar" : "Collapse sidebar"}
        >
            {#if collapsed}
                <ChevronRight class="h-4 w-4" />
            {:else}
                <ChevronLeft class="h-4 w-4" />
            {/if}
        </button>
    </div>

    <!-- Navigation -->
    <nav class="flex-1 px-3 py-4 flex flex-col">
        <ul class="space-y-1">
            {#each menuItems as item}
                <li>
                    {#if item.children}
                        <!-- Parent item with children -->
                        <div class="space-y-1">
                            {#if item.label === "Customization" && collapsed}
                                <a
                                    href="/admin/customization/theme"
                                    class="flex items-center gap-3 px-3 py-2 rounded-lg text-sm font-medium transition-all duration-200 w-full justify-center {isParentActive(
                                        item.href,
                                        item.children,
                                    )
                                        ? 'bg-primary text-primary-foreground'
                                        : isChildActive(item.children)
                                          ? 'bg-accent text-foreground'
                                          : 'text-muted-foreground hover:text-foreground hover:bg-accent'}"
                                    title={item.label}
                                    data-sveltekit-preload-data="tap"
                                    data-sveltekit-reload
                                >
                                    <div class="relative">
                                        <svelte:component
                                            this={item.icon}
                                            class="h-4 w-4 flex-shrink-0"
                                        />
                                        {#if themeUpdateAvailable}
                                            <span
                                                class="absolute -top-1 -right-1 w-2.5 h-2.5 bg-amber-500 rounded-full"
                                            ></span>
                                        {/if}
                                    </div>
                                </a>
                            {:else}
                                <button
                                    on:click={item.label === "Customization"
                                        ? toggleCustomization
                                        : () => {}}
                                    class="flex items-center gap-3 px-3 py-2 rounded-lg text-sm font-medium transition-all duration-200 w-full {collapsed
                                        ? 'justify-center'
                                        : ''} {isParentActive(
                                        item.href,
                                        item.children,
                                    )
                                        ? 'bg-primary text-primary-foreground'
                                        : isChildActive(item.children)
                                          ? 'bg-accent text-foreground'
                                          : 'text-muted-foreground hover:text-foreground hover:bg-accent'}"
                                    title={collapsed ? item.label : ""}
                                >
                                    <div class="relative">
                                        <svelte:component
                                            this={item.icon}
                                            class="h-4 w-4 flex-shrink-0"
                                        />
                                        {#if item.label === "Customization" && themeUpdateAvailable && collapsed}
                                            <span
                                                class="absolute -top-1 -right-1 w-2.5 h-2.5 bg-amber-500 rounded-full"
                                            ></span>
                                        {/if}
                                    </div>
                                    {#if !collapsed}
                                        <span class="flex-1 text-left"
                                            >{item.label}</span
                                        >
                                        {#if item.label === "Customization" && themeUpdateAvailable}
                                            <span
                                                class="ml-2 inline-flex items-center px-1.5 py-0.5 rounded-full text-xs font-medium bg-amber-100 text-amber-800 dark:bg-amber-900/20 dark:text-amber-400"
                                            >
                                                1
                                            </span>
                                        {/if}
                                        {#if item.label === "Customization"}
                                            <svelte:component
                                                this={customizationExpanded
                                                    ? ChevronDown
                                                    : ChevronRightIcon}
                                                class="h-4 w-4 flex-shrink-0 transition-transform duration-200 ml-2"
                                            />
                                        {/if}
                                    {/if}
                                </button>
                            {/if}

                            {#if !collapsed && item.label === "Customization" && customizationExpanded}
                                <div
                                    class="ml-6 space-y-1 border-l border-border/40 pl-3"
                                >
                                    {#each item.children as child}
                                        <a
                                            href={child.href}
                                            class="flex items-center gap-3 px-3 py-1.5 rounded-md text-sm font-medium transition-all duration-200 {isActive(
                                                child.href,
                                            )
                                                ? 'bg-primary text-primary-foreground'
                                                : 'text-muted-foreground hover:text-foreground hover:bg-accent'}"
                                            data-sveltekit-preload-data="tap"
                                            data-sveltekit-reload
                                        >
                                            <div class="relative">
                                                <svelte:component
                                                    this={child.icon}
                                                    class="h-3.5 w-3.5 flex-shrink-0"
                                                />
                                                {#if child.label === "Theme" && themeUpdateAvailable}
                                                    <span
                                                        class="absolute -top-1 -right-1 w-2 h-2 bg-amber-500 rounded-full"
                                                    ></span>
                                                {/if}
                                            </div>
                                            <span>{child.label}</span>
                                            {#if child.label === "Theme" && themeUpdateAvailable}
                                                <span
                                                    class="ml-2 inline-flex items-center px-1.5 py-0.5 rounded-full text-xs font-medium bg-amber-100 text-amber-800 dark:bg-amber-900/20 dark:text-amber-400"
                                                >
                                                    Update
                                                </span>
                                            {/if}
                                        </a>
                                    {/each}
                                </div>
                            {/if}
                        </div>
                    {:else}
                        <!-- Regular menu item -->
                        <a
                            href={item.href}
                            class="flex items-center gap-3 px-3 py-2 rounded-lg text-sm font-medium transition-all duration-200 {collapsed
                                ? 'justify-center'
                                : ''} {isActive(item.href)
                                ? 'bg-primary text-primary-foreground'
                                : 'text-muted-foreground hover:text-foreground hover:bg-accent'}"
                            title={collapsed ? item.label : ""}
                            data-sveltekit-preload-data="tap"
                            data-sveltekit-reload
                        >
                            <svelte:component
                                this={item.icon}
                                class="h-4 w-4 flex-shrink-0"
                            />
                            {#if !collapsed}
                                <span>{item.label}</span>
                            {/if}
                        </a>
                    {/if}
                </li>
                {#if item.label === "Events"}
                    <div class="my-6"></div>
                {/if}
            {/each}
        </ul>

        <!-- Bottom actions -->
        <div class="mt-auto space-y-1">
            <!-- Theme Toggle -->
            <button
                on:click={toggleTheme}
                class="flex items-center gap-3 px-3 py-2 rounded-lg text-sm font-medium text-muted-foreground hover:text-foreground hover:bg-accent transition-all duration-200 w-full {collapsed
                    ? 'justify-center'
                    : ''}"
                title={collapsed
                    ? selectedTheme === "light"
                        ? "Switch to Dark"
                        : "Switch to Light"
                    : ""}
            >
                {#if selectedTheme === "light"}
                    <Sun class="h-4 w-4 flex-shrink-0" />
                {:else}
                    <Moon class="h-4 w-4 flex-shrink-0" />
                {/if}
                {#if !collapsed}
                    <span>{selectedTheme === "light" ? "Light" : "Dark"}</span>
                {/if}
            </button>

            <!-- GitHub Link -->
            <a
                href="https://github.com/GauthierNelkinsky/ShipShipShip"
                target="_blank"
                class="flex items-center gap-3 px-3 py-2 rounded-lg text-sm font-medium text-muted-foreground hover:text-foreground hover:bg-accent transition-all duration-200 {collapsed
                    ? 'justify-center'
                    : ''}"
                title={collapsed ? "GitHub" : ""}
            >
                <Github class="h-4 w-4 flex-shrink-0" />
                {#if !collapsed}
                    <span>GitHub</span>
                {/if}
            </a>

            <!-- Public Page Link -->
            <a
                href="/"
                target="_blank"
                class="flex items-center gap-3 px-3 py-2 rounded-lg text-sm font-medium text-muted-foreground hover:text-foreground hover:bg-accent transition-all duration-200 {collapsed
                    ? 'justify-center'
                    : ''}"
                title={collapsed ? "Public Page" : ""}
            >
                <Globe class="h-4 w-4 flex-shrink-0" />
                {#if !collapsed}
                    <span>Public Page</span>
                {/if}
            </a>

            <!-- Separator -->
            <div class="border-t border-border my-2"></div>

            <!-- Logout (hidden in demo mode) -->
            {#if !$authStore.isDemoMode}
                <button
                    on:click={handleLogout}
                    class="flex items-center gap-3 px-3 py-2 rounded-lg text-sm font-medium text-muted-foreground hover:text-foreground hover:bg-accent transition-all duration-200 w-full {collapsed
                        ? 'justify-center'
                        : ''}"
                    title={collapsed ? "Logout" : ""}
                >
                    <LogOut class="h-4 w-4 flex-shrink-0" />
                    {#if !collapsed}
                        <span>Logout</span>
                    {/if}
                </button>
            {/if}
        </div>
    </nav>
</aside>
