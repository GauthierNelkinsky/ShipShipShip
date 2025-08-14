<script lang="ts">
    import "../app.css";
    import { onMount } from "svelte";
    import { settings, loadSettings } from "$lib/stores/settings";
    import { theme } from "$lib/stores/theme";

    let faviconUrl = "";

    onMount(async () => {
        // Initialize theme
        theme.init();

        // Load project settings
        await loadSettings();
    });

    // Reactive statement to update favicon when settings change
    $: if ($settings.logo_url) {
        faviconUrl = $settings.logo_url;
        updateFavicon(faviconUrl);
    }

    function updateFavicon(url: string) {
        if (typeof window === "undefined" || !url) return;

        // Remove existing favicon links
        const existingLinks = document.querySelectorAll('link[rel*="icon"]');
        existingLinks.forEach((link) => link.remove());

        // Add new favicon
        const link = document.createElement("link");
        link.rel = "icon";
        link.type = "image/x-icon";
        link.href = url;
        document.head.appendChild(link);
    }
</script>

<svelte:head>
    <title>{$settings.title || "Changelog"}</title>
    <meta name="description" content="Product roadmap and changelog" />
    {#if faviconUrl}
        <link rel="icon" type="image/x-icon" href={faviconUrl} />
    {/if}
</svelte:head>

<div class="min-h-screen bg-background text-foreground">
    <slot />
</div>
