<script lang="ts">
    import "../app.css";
    import { onMount } from "svelte";
    import { goto } from "$app/navigation";
    import { page } from "$app/stores";
    import { authStore } from "$lib/stores/auth";
    import AdminSidebar from "$lib/components/AdminSidebar.svelte";
    import { Toaster } from "$lib/components/ui/sonner";
    import { theme } from "$lib/stores/theme";
    import * as m from "$lib/paraglide/messages";
    import { localizeUrl, deLocalizeUrl } from "$lib/paraglide/runtime";

    import { loadSettings } from "$lib/stores/settings";

    let sidebarCollapsed = false;

    // Collapse sidebar by default on events page
    $: if (deLocalizeUrl($page.url.href).pathname.startsWith("/admin/events")) {
        sidebarCollapsed = true;
    }

    onMount(async () => {
        // Initialize theme
        theme.init();

        // Initialize authentication
        const isAuthenticated = await authStore.init();

        // Skip login redirect if authenticated or in demo mode
        if (isAuthenticated || $authStore.isDemoMode) {
            // Load project settings to apply global primary color variables
            await loadSettings();
            return;
        }

        // Only redirect to login if not on login page and not in demo mode
        const delocalizedPath = deLocalizeUrl($page.url.href).pathname;
        if (delocalizedPath !== "/login") {
            goto(localizeUrl("/login").toString());
        }
    });
</script>

<svelte:head>
    <title>{m.layout_page_title()}</title>
</svelte:head>

<Toaster
    position="bottom-right"
    richColors={false}
    expand={false}
    closeButton
    visibleToasts={3}
    duration={2500}
    offset="16px"
/>

{#if deLocalizeUrl($page.url.href).pathname === "/login" && !$authStore.isDemoMode}
    <!-- Login page - no layout needed (unless in demo mode) -->
    <slot />
{:else if $authStore.loading}
    <div class="min-h-screen flex items-center justify-center bg-background">
        <div
            class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"
        ></div>
    </div>
{:else if $authStore.isAuthenticated || $authStore.isDemoMode}
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
            <p class="text-muted-foreground">
                {m.layout_redirecting_to_login()}
            </p>
        </div>
    </div>
{/if}
