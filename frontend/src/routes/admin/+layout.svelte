<script lang="ts">
    import { onMount } from "svelte";
    import { goto } from "$app/navigation";
    import { page } from "$app/stores";
    import { api } from "$lib/api";
    import { settings } from "$lib/stores/settings";
    import AdminSidebar from "$lib/components/AdminSidebar.svelte";
    import ThemeSelector from "$lib/components/ThemeSelector.svelte";

    let sidebarCollapsed = false;
    let isAuthenticated = false;
    let loading = true;

    onMount(async () => {
        // Skip authentication check for login page
        if ($page.url.pathname === "/admin/login") {
            loading = false;
            return;
        }

        // Check authentication
        if (!api.isAuthenticated()) {
            goto("/admin/login");
            return;
        }

        try {
            await api.validateToken();
            isAuthenticated = true;
        } catch (err) {
            api.clearToken();
            goto("/admin/login");
        } finally {
            loading = false;
        }
    });
</script>

<svelte:head>
    <title>Admin - Changelog</title>
</svelte:head>

{#if $page.url.pathname === "/admin/login"}
    <!-- Login page - no layout needed -->
    <slot />
{:else if loading}
    <div class="min-h-screen flex items-center justify-center bg-background">
        <div
            class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"
        ></div>
    </div>
{:else if isAuthenticated}
    <div class="min-h-screen bg-background text-foreground flex">
        <!-- Fixed Sidebar -->
        <AdminSidebar bind:collapsed={sidebarCollapsed} />

        <!-- Main Content Area -->
        <div
            class="flex-1 flex flex-col transition-all duration-300"
            style="margin-left: {sidebarCollapsed ? '64px' : '256px'};"
        >
            <!-- Header -->
            <header
                class="h-14 border-b border-border bg-background backdrop-blur-sm sticky top-0 z-30 flex items-center px-6"
                style="background-color: hsl(var(--background) / 0.8);"
            >
                <div class="flex items-center justify-between w-full">
                    <!-- Logo/Title -->
                    <div class="flex items-center gap-3">
                        {#if $settings?.logo_url}
                            <img
                                src={$settings.logo_url}
                                alt="Logo"
                                class="h-6 w-6 object-contain dark:hidden"
                            />
                            {#if $settings?.dark_logo_url}
                                <img
                                    src={$settings.dark_logo_url}
                                    alt="Logo"
                                    class="h-6 w-6 object-contain hidden dark:block"
                                />
                            {:else}
                                <img
                                    src={$settings.logo_url}
                                    alt="Logo"
                                    class="h-6 w-6 object-contain hidden dark:block"
                                />
                            {/if}
                        {/if}
                        <h1 class="text-lg font-medium text-foreground">
                            {$settings?.title || "Changelog"} Admin
                        </h1>
                    </div>

                    <!-- Theme Toggle -->
                    <ThemeSelector />
                </div>
            </header>

            <!-- Main Content -->
            <main class="flex-1 overflow-auto">
                <div class="max-w-7xl mx-auto px-6 py-6">
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
