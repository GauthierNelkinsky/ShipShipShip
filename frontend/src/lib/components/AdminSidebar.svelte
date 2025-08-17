<script lang="ts">
    import { page } from "$app/stores";
    import { goto } from "$app/navigation";
    import { authStore } from "$lib/stores/auth";
    import { theme } from "$lib/stores/theme";
    import { Button } from "$lib/components/ui";
    import {
        Calendar,
        LogOut,
        ChevronLeft,
        ChevronRight,
        Home,
        Sun,
        Moon,
        ExternalLink,
        Palette,
        Building2,
        Tag,
        ChevronDown,
        ChevronRight as ChevronRightIcon,
    } from "lucide-svelte";

    export let collapsed = false;

    let customizationExpanded = false;

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
            ],
        },
    ];

    $: currentPath = $page?.url?.pathname || "";

    // Auto-expand customization if we're on a customization page
    $: if (currentPath && currentPath.startsWith("/admin/customization")) {
        customizationExpanded = true;
    }

    function isActive(href: string): boolean {
        if (!currentPath || !href) return false;

        // Handle admin events page (includes redirect from /admin)
        if (href === "/admin/events") {
            return (
                currentPath === "/admin" ||
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

        // Exact match for other paths
        return currentPath === href;
    }

    function isChildActive(children: any[]): boolean {
        return children.some((child) => isActive(child.href));
    }

    function isParentActive(href: string, children: any[]): boolean {
        // Only highlight parent if we're on the parent page itself, not on children
        return currentPath === href;
    }

    function toggleCustomization() {
        customizationExpanded = !customizationExpanded;
    }

    function handleLogout() {
        authStore.logout();
        goto("/admin/login");
    }

    function toggleSidebar() {
        collapsed = !collapsed;
    }
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
                                <svelte:component
                                    this={item.icon}
                                    class="h-4 w-4 flex-shrink-0"
                                />
                                {#if !collapsed}
                                    <span class="flex-1 text-left"
                                        >{item.label}</span
                                    >
                                    {#if item.label === "Customization"}
                                        <svelte:component
                                            this={customizationExpanded
                                                ? ChevronDown
                                                : ChevronRightIcon}
                                            class="h-4 w-4 flex-shrink-0 transition-transform duration-200"
                                        />
                                    {/if}
                                {/if}
                            </button>

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
                                            <svelte:component
                                                this={child.icon}
                                                class="h-3.5 w-3.5 flex-shrink-0"
                                            />
                                            <span>{child.label}</span>
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
            <!-- GitHub Link -->
            <a
                href="https://github.com/GauthierNelkinsky/ShipShipShip"
                target="_blank"
                class="flex items-center gap-3 px-3 py-2 rounded-lg text-sm font-medium text-muted-foreground hover:text-foreground hover:bg-accent transition-all duration-200 {collapsed
                    ? 'justify-center'
                    : ''}"
                title={collapsed ? "GitHub" : ""}
            >
                <ExternalLink class="h-4 w-4 flex-shrink-0" />
                {#if !collapsed}
                    <span>GitHub</span>
                {/if}
            </a>

            <!-- Public Site Link -->
            <a
                href="/"
                target="_blank"
                class="flex items-center gap-3 px-3 py-2 rounded-lg text-sm font-medium text-muted-foreground hover:text-foreground hover:bg-accent transition-all duration-200 {collapsed
                    ? 'justify-center'
                    : ''}"
                title={collapsed ? "Public Site" : ""}
            >
                <Home class="h-4 w-4 flex-shrink-0" />
                {#if !collapsed}
                    <span>Public Site</span>
                {/if}
            </a>

            <!-- Separator -->
            <div class="border-t border-border my-2"></div>

            <!-- Logout -->
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
        </div>
    </nav>
</aside>
