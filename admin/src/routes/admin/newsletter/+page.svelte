<script lang="ts">
    import { onMount } from "svelte";
    import { goto } from "$app/navigation";
    import { api } from "$lib/api";
    import { Users, Settings, Loader2 } from "lucide-svelte";
    import * as m from "$lib/paraglide/messages";

    let loading = true;
    let currentTab = "home";

    // Navigation items
    const navItems = [
        { id: "home", label: m.newsletter_nav_home(), icon: Users },
        { id: "settings", label: m.newsletter_nav_settings(), icon: Settings },
    ];

    onMount(async () => {
        loading = false;
        // Set default tab based on URL or default to home
        const urlParams = new URLSearchParams(window.location.search);
        currentTab = urlParams.get("tab") || "home";
    });

    function switchTab(tabId: string) {
        currentTab = tabId;
        const url = new URL(window.location.href);
        url.searchParams.set("tab", tabId);
        goto(url.pathname + url.search, { replaceState: true });
    }
</script>

<svelte:head>
    <title>{m.newsletter_page_title()}</title>
</svelte:head>

<div class="max-w-6xl mx-auto">
    <!-- Header -->
    <div class="mb-4">
        <h1 class="text-xl font-semibold mb-1">{m.newsletter_heading()}</h1>
        <p class="text-muted-foreground text-sm">
            {m.newsletter_subheading()}
        </p>
    </div>

    {#if loading}
        <div class="flex items-center justify-center min-h-32">
            <div class="flex items-center gap-2 text-sm">
                <Loader2 class="h-4 w-4 animate-spin" />
                <span class="text-muted-foreground">Loading...</span>
            </div>
        </div>
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
                        </button>
                    {/each}
                </div>
            </div>
        </nav>

        <!-- Tab Content -->
        <div class="space-y-6">
            {#if currentTab === "home"}
                {#await import("./home/+page.svelte")}
                    <div class="flex items-center justify-center py-8">
                        <div class="flex items-center gap-2 text-sm">
                            <Loader2 class="h-4 w-4 animate-spin" />
                            <span class="text-muted-foreground">Loading...</span
                            >
                        </div>
                    </div>
                {:then { default: HomePage }}
                    <svelte:component this={HomePage} />
                {:catch}
                    <div class="text-center py-8 text-red-600">
                        {m.newsletter_home_load_failed()}
                    </div>
                {/await}
            {:else if currentTab === "settings"}
                {#await import("./settings/+page.svelte")}
                    <div class="flex items-center justify-center py-8">
                        <div class="flex items-center gap-2 text-sm">
                            <Loader2 class="h-4 w-4 animate-spin" />
                            <span class="text-muted-foreground">Loading...</span
                            >
                        </div>
                    </div>
                {:then { default: SettingsPage }}
                    <svelte:component this={SettingsPage} />
                {:catch}
                    <div class="text-center py-8 text-red-600">
                        {m.newsletter_settings_load_failed()}
                    </div>
                {/await}
            {/if}
        </div>
    {/if}
</div>
