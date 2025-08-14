<script lang="ts">
    import { onMount } from "svelte";
    import { goto } from "$app/navigation";
    import { page } from "$app/stores";
    import { authStore } from "$lib/stores/auth";
    import { settings } from "$lib/stores/settings";
    import AdminSidebar from "$lib/components/AdminSidebar.svelte";
    import ThemeSelector from "$lib/components/ThemeSelector.svelte";

    let sidebarCollapsed = false;

    onMount(async () => {
        // Skip authentication check for login page
        if ($page.url.pathname === "/admin/login") {
            return;
        }

        // Initialize authentication
        const isAuthenticated = await authStore.init();

        if (!isAuthenticated) {
            goto("/admin/login");
        }
    });
</script>

<svelte:head>
    <title>Admin - Changelog</title>
</svelte:head>

{#if $page.url.pathname === "/admin/login"}
    <!-- Login page - no layout needed -->
    <slot />
{:else if $authStore.loading}
    <div class="min-h-screen flex items-center justify-center bg-background">
        <div
            class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"
        ></div>
    </div>
{:else if $authStore.isAuthenticated}
    <div
        class="min-h-screen bg-background text-foreground flex overflow-hidden"
    >
        <!-- Fixed Sidebar -->
        <AdminSidebar bind:collapsed={sidebarCollapsed} />

        <!-- Main Content Area -->
        <div
            class="flex-1 flex flex-col transition-all duration-300 min-w-0"
            style="margin-left: {sidebarCollapsed ? '64px' : '256px'};"
        >
            <!-- Header -->
            <header
                class="h-14 border-b border-border bg-background backdrop-blur-sm sticky top-0 z-30 flex items-center px-6"
                style="background-color: hsl(var(--background) / 0.8);"
            >
                <div class="flex items-center justify-between w-full">
                    <!-- Empty left side for spacing -->
                    <div></div>

                    <!-- Theme Toggle -->
                    <ThemeSelector />
                </div>
            </header>

            <!-- Main Content -->
            <main class="flex-1 overflow-auto min-w-0">
                <div class="w-full px-6 py-6">
                    <slot />
                </div>
            </main>
        </div>
    </div>
{:else}
    <!-- Fallback for unauthenticated state -->
    <div class="min-h-screen flex items-center justify-center bg-background">
        <div class="text-center">
            <div
                class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto mb-4"
            ></div>
            <p class="text-muted-foreground">Redirecting to login...</p>
        </div>
    </div>
{/if}
