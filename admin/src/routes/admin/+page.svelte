<script lang="ts">
    import { onMount } from "svelte";
    import { goto } from "$app/navigation";

    onMount(() => {
        // Redirect to events page when accessing /admin
        goto("/admin/events", { replaceState: true });
    });
</script>

<svelte:head>
    <title>Admin - Changelog</title>
</svelte:head>

<!-- This page will redirect to /admin/events -->
<div class="min-h-screen flex items-center justify-center bg-background">
    <div class="text-center">
        <div
            class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto mb-4"
        ></div>
        <p class="text-muted-foreground">Loading admin panel...</p>
    </div>
</div>
