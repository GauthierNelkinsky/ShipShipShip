<script lang="ts">
    import { page } from "$app/stores";
    import { goto } from "$app/navigation";
    import { api } from "$lib/api";
    import { theme } from "$lib/stores/theme";
    import { Button } from "$lib/components/ui";
    import {
        Calendar,
        Settings,
        LogOut,
        ChevronLeft,
        ChevronRight,
        Home,
        Sun,
        Moon,
        ExternalLink,
    } from "lucide-svelte";

    export let collapsed = false;

    const menuItems = [
        {
            label: "Events",
            href: "/admin/events",
            icon: Calendar,
        },
        {
            label: "Settings",
            href: "/admin/settings",
            icon: Settings,
        },
    ];

    $: currentPath = $page?.url?.pathname || "";

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

        // Handle admin settings page
        if (href === "/admin/settings") {
            return (
                currentPath === "/admin/settings" ||
                currentPath.startsWith("/admin/settings/")
            );
        }

        // Exact match for other paths
        return currentPath === href;
    }

    function handleLogout() {
        api.logout();
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
                    <a
                        href={item.href}
                        class="flex items-center gap-3 px-3 py-2 rounded-lg text-sm font-medium transition-all duration-200 {collapsed
                            ? 'justify-center'
                            : ''} {isActive(item.href)
                            ? 'bg-primary text-primary-foreground'
                            : 'text-muted-foreground sidebar-hover-subtle'}"
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
                class="sidebar-hover-subtle flex items-center gap-3 px-3 py-2 rounded-lg text-sm font-medium text-muted-foreground transition-all duration-200 {collapsed
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
                class="sidebar-hover-subtle flex items-center gap-3 px-3 py-2 rounded-lg text-sm font-medium text-muted-foreground transition-all duration-200 {collapsed
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
                class="sidebar-hover-subtle flex items-center gap-3 px-3 py-2 rounded-lg text-sm font-medium text-muted-foreground transition-all duration-200 w-full {collapsed
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
