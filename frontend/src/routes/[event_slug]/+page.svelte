<script lang="ts">
    import { page } from "$app/stores";
    import { onMount } from "svelte";
    import { api } from "$lib/api";
    import { Button, Badge } from "$lib/components/ui";
    import {
        ArrowLeft,
        Calendar,
        Heart,
        Eye,
        ExternalLink,
        Share2,
        Clock,
        Star,
        Users,
        TrendingUp,
    } from "lucide-svelte";
    import { goto } from "$app/navigation";
    import { settings } from "$lib/stores/settings";
    import ThemeSelector from "$lib/components/ThemeSelector.svelte";
    import Footer from "$lib/components/Footer.svelte";
    import type { PageData } from "./$types";
    import type { Tag } from "$lib/types";

    export let data: PageData;

    let event = data.event;
    let eventSettings = data.settings;
    let voting = false;
    let voteStatus = { voted: false, votes: event.votes };
    let error = "";
    let success = "";

    onMount(async () => {
        await loadVoteStatus();
    });

    async function loadVoteStatus() {
        try {
            const status = await api.checkVoteStatus(event.id);
            voteStatus = status;
        } catch (err) {
            console.error("Failed to load vote status:", err);
        }
    }

    async function handleVote() {
        if (voting || voteStatus.voted) return;

        voting = true;
        error = "";
        success = "";

        try {
            const result = await api.voteEvent(event.id);
            voteStatus = { voted: true, votes: result.votes };
            success = "Thank you for your vote!";
        } catch (err) {
            error = err instanceof Error ? err.message : "Failed to vote";
        } finally {
            voting = false;
        }
    }

    function getStatusColor(status: string) {
        switch (status) {
            case "Backlogs":
                return "bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-200";
            case "Proposed":
                return "bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-200";
            case "Upcoming":
                return "bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200";
            case "Release":
                return "bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200";
            default:
                return "bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-200";
        }
    }

    function getStatusIcon(status: string) {
        switch (status) {
            case "Backlogs":
                return Clock;
            case "Proposed":
                return Calendar;
            case "Upcoming":
                return TrendingUp;
            case "Release":
                return Star;
            default:
                return Eye;
        }
    }

    function formatTags(tags: Tag[]) {
        return tags || [];
    }

    function formatMedia(mediaString: string) {
        try {
            return JSON.parse(mediaString || "[]");
        } catch {
            return [];
        }
    }

    function formatDisplayDate(date: string, status: string) {
        if (!date) return "";

        if (status === "Upcoming") {
            const dateObj = new Date(date);
            const months = [
                "Jan",
                "Feb",
                "Mar",
                "Apr",
                "May",
                "Jun",
                "Jul",
                "Aug",
                "Sep",
                "Oct",
                "Nov",
                "Dec",
            ];
            return `${months[dateObj.getMonth()]}. ${dateObj.getFullYear()}`;
        }

        return date;
    }

    function shareEvent() {
        if (navigator.share) {
            navigator.share({
                title: event.title,
                text: `Check out this feature: ${event.title}`,
                url: window.location.href,
            });
        } else {
            // Fallback: copy to clipboard
            navigator.clipboard.writeText(window.location.href);
            success = "Link copied to clipboard!";
        }
    }
</script>

<svelte:head>
    <title>{event.title} - {eventSettings.title}</title>
    <meta
        name="description"
        content={event.content || `Feature: ${event.title}`}
    />
    <meta property="og:title" content={event.title} />
    <meta
        property="og:description"
        content={event.content || `Feature: ${event.title}`}
    />
    <meta property="og:type" content="article" />
    <meta property="og:url" content={$page.url.href} />
    {#if eventSettings.logo_url}
        <meta property="og:image" content={eventSettings.logo_url} />
    {/if}
</svelte:head>

<div class="min-h-screen bg-background">
    <!-- Header -->
    <header
        class="border-b border-border/40 bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60 sticky top-0 z-40"
    >
        <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
            <div class="flex items-center justify-between h-16 min-w-0">
                <!-- Back Button + Logo/Title -->
                <div
                    class="flex items-center space-x-2 sm:space-x-3 min-w-0 flex-1"
                >
                    <button
                        on:click={() => goto("/")}
                        class="flex items-center justify-center w-8 h-8 rounded-md hover:bg-accent/50 transition-colors flex-shrink-0"
                        title="Back to changelog"
                    >
                        <ArrowLeft class="h-4 w-4" />
                    </button>

                    {#if $settings?.website_url}
                        <a
                            href={$settings.website_url}
                            target="_blank"
                            rel="noopener noreferrer"
                            class="flex items-center space-x-2 sm:space-x-3 hover:opacity-80 transition-opacity min-w-0"
                        >
                            {#if $settings?.logo_url}
                                <img
                                    src={$settings.logo_url}
                                    alt="Logo"
                                    class="h-8 w-8 object-contain dark:hidden"
                                />
                                {#if $settings?.dark_logo_url}
                                    <img
                                        src={$settings.dark_logo_url}
                                        alt="Logo"
                                        class="h-8 w-8 object-contain hidden dark:block"
                                    />
                                {:else}
                                    <img
                                        src={$settings.logo_url}
                                        alt="Logo"
                                        class="h-8 w-8 object-contain hidden dark:block"
                                    />
                                {/if}
                            {/if}
                            <h1
                                class="text-lg sm:text-xl font-semibold text-foreground truncate"
                            >
                                {$settings?.title || "Changelog"}
                            </h1>
                        </a>
                    {:else}
                        <div
                            class="flex items-center space-x-2 sm:space-x-3 min-w-0"
                        >
                            {#if $settings?.logo_url}
                                <img
                                    src={$settings.logo_url}
                                    alt="Logo"
                                    class="h-8 w-8 object-contain dark:hidden"
                                />
                                {#if $settings?.dark_logo_url}
                                    <img
                                        src={$settings.dark_logo_url}
                                        alt="Logo"
                                        class="h-8 w-8 object-contain hidden dark:block"
                                    />
                                {:else}
                                    <img
                                        src={$settings.logo_url}
                                        alt="Logo"
                                        class="h-8 w-8 object-contain hidden dark:block"
                                    />
                                {/if}
                            {/if}
                            <h1
                                class="text-lg sm:text-xl font-semibold text-foreground truncate"
                            >
                                {$settings?.title || "Changelog"}
                            </h1>
                        </div>
                    {/if}
                </div>

                <!-- Theme Toggle -->
                <ThemeSelector />
            </div>
        </div>
    </header>

    <!-- Main Content -->
    <div class="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8 sm:py-16">
        {#if success}
            <div
                class="bg-green-50 dark:bg-green-900/20 text-green-800 dark:text-green-200 px-4 py-3 rounded-lg mb-8 text-center"
            >
                {success}
            </div>
        {/if}

        {#if error}
            <div
                class="bg-red-50 dark:bg-red-900/20 text-red-800 dark:text-red-200 px-4 py-3 rounded-lg mb-8 text-center"
            >
                {error}
            </div>
        {/if}

        <!-- Event Content -->
        <article class="text-center space-y-8">
            <!-- Title -->
            <h1
                class="text-3xl sm:text-4xl lg:text-5xl font-bold tracking-tight text-foreground"
            >
                {event.title}
            </h1>

            <!-- Status, Date and Tags -->
            <div
                class="flex flex-wrap items-center justify-center gap-2 text-sm text-muted-foreground"
            >
                <!-- Status Badge -->
                {#if event.status}
                    {@const StatusIcon = getStatusIcon(event.status)}
                    <Badge
                        class="{getStatusColor(
                            event.status,
                        )} flex items-center gap-1.5 text-xs font-medium"
                    >
                        <StatusIcon class="h-3.5 w-3.5" />
                        {event.status}
                    </Badge>
                {/if}

                <!-- Separator -->
                {#if event.status && (event.date || (event.tags && Array.isArray(event.tags) && event.tags.length > 0))}
                    <span>•</span>
                {/if}

                <!-- Date with Estimated badge for Upcoming -->
                {#if event.date}
                    <div class="flex items-center gap-2">
                        {#if event.status === "Upcoming"}
                            <Badge
                                variant="outline"
                                class="bg-orange-50 dark:bg-orange-900/20 text-orange-600 dark:text-orange-400 border-orange-200 dark:border-orange-800 text-xs font-medium"
                            >
                                Estimated
                            </Badge>
                        {/if}
                        <time datetime={event.date}>
                            {formatDisplayDate(event.date, event.status)}
                        </time>
                    </div>
                {/if}

                <!-- Separator -->
                {#if event.date && event.tags && Array.isArray(event.tags) && event.tags.length > 0}
                    <span>•</span>
                {/if}

                <!-- Tags -->
                {#if event.tags && Array.isArray(event.tags) && event.tags.length > 0}
                    {#each event.tags as tag}
                        <Badge
                            variant="outline"
                            class="text-xs font-medium"
                            style="background-color: {tag.color}20; color: {tag.color}; border-color: {tag.color}40"
                        >
                            {tag.name}
                        </Badge>
                    {/each}
                {/if}
            </div>

            <!-- Event Content -->
            {#if event.content}
                <div
                    class="prose prose-lg max-w-none dark:prose-invert mx-auto text-left prose-headings:font-semibold prose-headings:tracking-tight prose-p:text-muted-foreground prose-p:leading-relaxed prose-a:text-primary prose-a:no-underline hover:prose-a:underline prose-strong:text-foreground"
                >
                    {@html event.content}
                </div>
            {/if}

            <!-- Media Gallery -->
            {#if event.media}
                {@const media = formatMedia(event.media)}
                {#if media.length > 0}
                    <div class="space-y-6">
                        <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                            {#each media as mediaUrl}
                                <div class="relative group">
                                    <img
                                        src={mediaUrl}
                                        alt="Feature media"
                                        class="w-full h-64 object-cover rounded-lg"
                                        loading="lazy"
                                    />
                                    <div
                                        class="absolute inset-0 bg-black/0 group-hover:bg-black/10 transition-colors rounded-lg"
                                    >
                                        <Button
                                            variant="secondary"
                                            size="sm"
                                            on:click={() =>
                                                window.open(mediaUrl, "_blank")}
                                            class="absolute top-2 right-2 opacity-0 group-hover:opacity-100 transition-opacity"
                                        >
                                            <ExternalLink class="h-4 w-4" />
                                        </Button>
                                    </div>
                                </div>
                            {/each}
                        </div>
                    </div>
                {/if}
            {/if}
        </article>
    </div>

    <!-- Footer -->
    <Footer />
</div>
