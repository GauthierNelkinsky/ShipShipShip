<script lang="ts">
    import { onMount } from "svelte";
    import { page } from "$app/stores";
    import { goto } from "$app/navigation";
    import { api } from "$lib/api";
    import { Button } from "$lib/components/ui";
    import {
        Mail,
        Users,
        Settings,
        AlertCircle,
        CheckCircle,
    } from "lucide-svelte";

    let loading = true;
    let error = "";
    let success = "";

    // Newsletter settings
    let newsletterEnabled = false;
    let mailConfigured = false;

    // Current tab
    let currentTab = "home";

    // Navigation items
    const navItems = [
        { id: "home", label: "Home", icon: Users },
        { id: "settings", label: "Settings", icon: Settings },
    ];

    onMount(async () => {
        await loadData();
        // Set default tab based on URL or default to home
        const urlParams = new URLSearchParams(window.location.search);
        currentTab = urlParams.get("tab") || "home";
    });

    async function loadData() {
        loading = true;
        error = "";

        try {
            // Load mail settings to check if configured
            await loadMailSettings();
            // Load newsletter enabled status from settings
            await loadNewsletterSettings();
        } catch (err) {
            console.error("Error loading data:", err);
            error = "Failed to load newsletter data";
        } finally {
            loading = false;
        }
    }

    async function loadMailSettings() {
        try {
            const settings = await api.getMailSettings();
            mailConfigured = !!(settings?.smtp_host && settings?.from_email);
        } catch (err) {
            console.log("No mail settings found");
            mailConfigured = false;
        }
    }

    async function loadNewsletterSettings() {
        try {
            const settings = await api.getSettings();
            newsletterEnabled = !!settings?.newsletter_enabled;
        } catch (err) {
            console.error("Failed to load newsletter settings:", err);
            newsletterEnabled = false;
        }
    }

    function switchTab(tabId: string) {
        currentTab = tabId;
        const url = new URL(window.location.href);
        url.searchParams.set("tab", tabId);
        goto(url.pathname + url.search, { replaceState: true });
    }
</script>

<svelte:head>
    <title>Newsletter Management - Admin</title>
</svelte:head>

<div class="max-w-6xl mx-auto">
    <!-- Header -->
    <div class="mb-8">
        <h1 class="text-xl font-semibold mb-1">Newsletter Management</h1>
        <p class="text-muted-foreground text-sm">
            Manage newsletter subscriptions and email settings
        </p>
    </div>

    {#if loading}
        <div class="flex items-center justify-center min-h-32">
            <div
                class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"
            ></div>
        </div>
    {:else}
        {#if success}
            <div
                class="bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800 text-green-800 dark:text-green-200 px-4 py-3 rounded-lg mb-6 flex items-center gap-2"
            >
                <CheckCircle class="h-4 w-4" />
                {success}
            </div>
        {/if}

        {#if error}
            <div
                class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 text-red-800 dark:text-red-200 px-4 py-3 rounded-lg mb-6 flex items-center gap-2"
            >
                <AlertCircle class="h-4 w-4" />
                {error}
            </div>
        {/if}

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
                        <div
                            class="animate-spin rounded-full h-6 w-6 border-b-2 border-primary"
                        ></div>
                    </div>
                {:then { default: HomePage }}
                    <svelte:component
                        this={HomePage}
                        disabled={!newsletterEnabled}
                        {newsletterEnabled}
                    />
                {:catch}
                    <div class="text-center py-8 text-red-600">
                        Failed to load home page
                    </div>
                {/await}
            {:else if currentTab === "settings"}
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
                        Failed to load settings page
                    </div>
                {/await}
            {/if}
        </div>
    {/if}
</div>
