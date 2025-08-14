<script lang="ts">
    import { onMount } from "svelte";
    import { api } from "$lib/api";
    import {
        parseEvent,
        groupEventsByStatus,
        formatDate,
        markdownToHtml,
    } from "$lib/utils";
    import { settings } from "$lib/stores/settings";
    import type { ParsedEvent } from "$lib/types";
    import { Calendar, Tag, Send, ThumbsUp } from "lucide-svelte";
    import { Button, Card, Badge, Input, Textarea } from "$lib/components/ui";
    import ThemeSelector from "$lib/components/ThemeSelector.svelte";

    function generateTagColor(tag: string): string {
        // Generate a consistent color based on tag name
        const colors = [
            "#3b82f6",
            "#ef4444",
            "#10b981",
            "#f59e0b",
            "#8b5cf6",
            "#ec4899",
            "#06b6d4",
            "#84cc16",
            "#f97316",
            "#6366f1",
        ];

        let hash = 0;
        for (let i = 0; i < tag.length; i++) {
            hash = tag.charCodeAt(i) + ((hash << 5) - hash);
        }

        return colors[Math.abs(hash) % colors.length];
    }

    let events: ParsedEvent[] = [];
    let loading = true;
    let error = "";
    let groupedEvents: ReturnType<typeof groupEventsByStatus> = {
        backlogs: [],
        doing: [],
        release: [],
        upcoming: [],
        archived: [],
    };

    // Feedback form
    let feedbackTitle = "";
    let feedbackDescription = "";
    let submittingFeedback = false;
    let feedbackSuccess = false;

    // Vote tracking
    let votedEvents = new Set<number>();
    let voteErrors: Record<number, string> = {};

    onMount(async () => {
        await loadEvents();
    });

    async function loadVoteStatuses() {
        // Check vote status for all upcoming events
        const upcomingEvents = groupedEvents.upcoming || [];
        for (const event of upcomingEvents) {
            try {
                const status = await api.checkVoteStatus(event.id);
                if (status.voted) {
                    votedEvents.add(event.id);
                }
            } catch (err) {
                console.error("Failed to check vote status:", err);
            }
        }
        votedEvents = new Set(votedEvents); // Trigger reactivity
    }

    async function loadEvents() {
        try {
            loading = true;
            const data = await api.getEvents();
            events = data.map(parseEvent);
            groupedEvents = groupEventsByStatus(events);
            // Load vote statuses after events are loaded
            await loadVoteStatuses();
        } catch (err) {
            error =
                err instanceof Error ? err.message : "Failed to load events";
        } finally {
            loading = false;
        }
    }

    async function handleVote(eventId: number) {
        try {
            const result = await api.voteEvent(eventId);
            // Update the event in our local state
            events = events.map((event) =>
                event.id === eventId
                    ? { ...event, votes: result.votes }
                    : event,
            );
            groupedEvents = groupEventsByStatus(events);

            // Update vote status
            if (result.voted) {
                votedEvents.add(eventId);
            } else {
                votedEvents.delete(eventId);
            }
            votedEvents = new Set(votedEvents); // Trigger reactivity

            // Clear any previous error for this event
            delete voteErrors[eventId];
            voteErrors = { ...voteErrors };
        } catch (err) {
            console.error("Failed to vote:", err);
            const errorMessage =
                err instanceof Error ? err.message : "Failed to vote";
            voteErrors[eventId] = errorMessage;
            voteErrors = { ...voteErrors };

            // Clear error after 3 seconds
            setTimeout(() => {
                delete voteErrors[eventId];
                voteErrors = { ...voteErrors };
            }, 3000);
        }
    }

    async function submitFeedback() {
        if (!feedbackTitle.trim() || !feedbackDescription.trim()) {
            return;
        }

        submittingFeedback = true;

        try {
            await api.submitFeedback(
                feedbackTitle.trim(),
                feedbackDescription.trim(),
            );

            feedbackTitle = "";
            feedbackDescription = "";
            feedbackSuccess = true;

            setTimeout(() => {
                feedbackSuccess = false;
            }, 3000);

            // Reload events to show the new feedback
            await loadEvents();
        } catch (err) {
            console.error("Failed to submit feedback:", err);
        } finally {
            submittingFeedback = false;
        }
    }

    // Combine doing and release events for timeline, sorted by date
    $: timelineEvents = [
        ...groupedEvents.doing,
        ...groupedEvents.release.sort((a, b) => {
            if (!a.date && !b.date) return 0;
            if (!a.date) return 1;
            if (!b.date) return -1;
            return new Date(b.date).getTime() - new Date(a.date).getTime();
        }),
    ];
</script>

<svelte:head>
    <title>{$settings?.title || "Changelog"}</title>
    <meta name="description" content="Product changelog and feature updates" />
</svelte:head>

<!-- Main Container -->
<div
    class="min-h-screen bg-white dark:bg-neutral-950 text-gray-900 dark:text-neutral-100"
>
    <!-- Header -->
    <header
        class="border-b border-gray-200 dark:border-neutral-800 bg-white/80 dark:bg-neutral-950/80 backdrop-blur-sm sticky top-0 z-40"
    >
        <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
            <div class="flex items-center justify-between h-16 min-w-0">
                <!-- Logo/Title -->
                {#if $settings?.website_url}
                    <a
                        href={$settings.website_url}
                        target="_blank"
                        rel="noopener noreferrer"
                        class="flex items-center space-x-2 sm:space-x-3 hover:opacity-80 transition-opacity min-w-0 flex-1"
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
                        class="flex items-center space-x-2 sm:space-x-3 min-w-0 flex-1"
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

                <!-- Theme Toggle -->
                <ThemeSelector />
            </div>
        </div>
    </header>

    {#if loading}
        <div class="flex items-center justify-center min-h-[50vh]">
            <div
                class="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900 dark:border-neutral-100"
            ></div>
        </div>
    {:else if error}
        <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
            <div
                class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 text-red-800 dark:text-red-200 px-4 py-3 rounded-lg"
            >
                {error}
            </div>
        </div>
    {:else}
        <!-- Main Content -->
        <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6 sm:py-12">
            <div class="flex flex-col lg:flex-row gap-6 lg:gap-12">
                <!-- Left Column - Timeline -->
                <div class="flex-1 min-w-0">
                    <div class="max-w-4xl">
                        <!-- Changelog Timeline -->
                        {#if groupedEvents.doing.length > 0 || groupedEvents.release.length > 0}
                            {@const sortedDoingEvents =
                                groupedEvents.doing.sort((a, b) => {
                                    const dateA = a.date || "";
                                    const dateB = b.date || "";
                                    return (
                                        new Date(dateB).getTime() -
                                        new Date(dateA).getTime()
                                    );
                                })}
                            {@const sortedReleaseEvents =
                                groupedEvents.release.sort((a, b) => {
                                    const dateA = a.date || "";
                                    const dateB = b.date || "";
                                    return (
                                        new Date(dateB).getTime() -
                                        new Date(dateA).getTime()
                                    );
                                })}
                            <div class="space-y-8">
                                <!-- Doing Events -->
                                {#each sortedDoingEvents as event}
                                    <article
                                        class="flex flex-col md:flex-row gap-4 md:gap-8 pb-8"
                                    >
                                        <!-- Left Column - Date and Tags -->
                                        <div
                                            class="w-full md:w-[250px] md:flex-shrink-0 text-left md:text-right"
                                        >
                                            <!-- Date -->
                                            {#if event.date}
                                                <div class="mb-3">
                                                    {#if event.status === "Doing"}
                                                        <div
                                                            class="flex items-center justify-start md:justify-end gap-2 mb-1"
                                                        >
                                                            <span
                                                                class="text-amber-600 dark:text-amber-400 font-medium flex items-center gap-2 text-xs px-2 py-1 bg-amber-50 dark:bg-amber-900/20 rounded"
                                                            >
                                                                <span
                                                                    class="relative flex size-2"
                                                                >
                                                                    <span
                                                                        class="absolute inline-flex h-full w-full animate-ping rounded-full bg-amber-400 opacity-75"
                                                                    ></span>
                                                                    <span
                                                                        class="relative inline-flex size-2 rounded-full bg-amber-500"
                                                                    ></span>
                                                                </span>
                                                                Estimated
                                                            </span>
                                                            <time
                                                                class="text-sm font-medium text-gray-600 dark:text-neutral-400"
                                                            >
                                                                {formatDate(
                                                                    event.date,
                                                                ).replace(
                                                                    /^\d+\s/,
                                                                    "",
                                                                )}
                                                            </time>
                                                        </div>
                                                    {:else}
                                                        <time
                                                            class="block text-sm font-medium text-gray-600 dark:text-neutral-400 text-left md:text-right"
                                                        >
                                                            {formatDate(
                                                                event.date,
                                                            )}
                                                        </time>
                                                    {/if}
                                                </div>
                                            {/if}

                                            <!-- Tags -->
                                            {#if event.tags.length > 0}
                                                <div
                                                    class="flex flex-wrap gap-1 justify-start md:justify-end"
                                                >
                                                    {#each event.tags as tag}
                                                        {@const tagColor =
                                                            generateTagColor(
                                                                tag,
                                                            )}
                                                        <Badge
                                                            variant="outline"
                                                            class="text-xs"
                                                            style="border-color: {tagColor}; background-color: {tagColor}20; color: {tagColor};"
                                                        >
                                                            {tag}
                                                        </Badge>
                                                    {/each}
                                                </div>
                                            {/if}
                                        </div>

                                        <!-- Right Column - Title and Content -->
                                        <div class="flex-1 min-w-0">
                                            <!-- Event Title -->
                                            <div class="mb-3 sm:mb-4">
                                                <h2
                                                    class="text-xl sm:text-2xl font-bold text-gray-900 dark:text-neutral-100 leading-tight"
                                                >
                                                    {event.title}
                                                </h2>
                                            </div>

                                            <!-- Event Content -->
                                            {#if event.content}
                                                <div
                                                    class="prose prose-sm sm:prose prose-gray dark:prose-invert max-w-none mb-4 sm:mb-6"
                                                >
                                                    {@html markdownToHtml(
                                                        event.content,
                                                    )}
                                                </div>
                                            {/if}

                                            <!-- Event Media -->
                                            {#if event.media.length > 0}
                                                <div
                                                    class="grid grid-cols-1 md:grid-cols-2 gap-4"
                                                >
                                                    {#each event.media as mediaUrl}
                                                        <img
                                                            src={mediaUrl}
                                                            alt="Update media"
                                                            class="rounded-lg border border-gray-200 dark:border-neutral-700 max-w-full h-auto"
                                                            loading="lazy"
                                                        />
                                                    {/each}
                                                </div>
                                            {/if}
                                        </div>
                                    </article>
                                {/each}

                                <!-- Separator between Doing and Released events -->
                                {#if sortedDoingEvents.length > 0 && sortedReleaseEvents.length > 0}
                                    <div class="py-8">
                                        <div
                                            class="border-t border-dashed border-gray-300 dark:border-neutral-600"
                                        ></div>
                                    </div>
                                {/if}

                                <!-- Released Events -->
                                {#each sortedReleaseEvents as event}
                                    <article
                                        class="flex flex-col md:flex-row gap-4 md:gap-8 pb-8"
                                    >
                                        <!-- Left Column - Date and Tags -->
                                        <div
                                            class="w-full md:w-[250px] md:flex-shrink-0 text-left md:text-right"
                                        >
                                            <!-- Date -->
                                            {#if event.date}
                                                <div class="mb-3">
                                                    {#if event.status === "Doing"}
                                                        <div
                                                            class="flex items-center justify-start md:justify-end gap-2 mb-1"
                                                        >
                                                            <span
                                                                class="text-amber-600 dark:text-amber-400 font-medium flex items-center gap-2 text-xs px-2 py-1 bg-amber-50 dark:bg-amber-900/20 rounded"
                                                            >
                                                                <span
                                                                    class="relative flex size-2"
                                                                >
                                                                    <span
                                                                        class="absolute inline-flex h-full w-full animate-ping rounded-full bg-amber-400 opacity-75"
                                                                    ></span>
                                                                    <span
                                                                        class="relative inline-flex size-2 rounded-full bg-amber-500"
                                                                    ></span>
                                                                </span>
                                                                Estimated
                                                            </span>
                                                            <time
                                                                class="text-sm font-medium text-gray-600 dark:text-neutral-400"
                                                            >
                                                                {formatDate(
                                                                    event.date,
                                                                ).replace(
                                                                    /^\d+\s/,
                                                                    "",
                                                                )}
                                                            </time>
                                                        </div>
                                                    {:else}
                                                        <time
                                                            class="block text-sm font-medium text-gray-600 dark:text-neutral-400 text-left md:text-right"
                                                        >
                                                            {formatDate(
                                                                event.date,
                                                            )}
                                                        </time>
                                                    {/if}
                                                </div>
                                            {/if}

                                            <!-- Tags -->
                                            {#if event.tags.length > 0}
                                                <div
                                                    class="flex flex-wrap gap-1 justify-start md:justify-end"
                                                >
                                                    {#each event.tags as tag}
                                                        {@const tagColor =
                                                            generateTagColor(
                                                                tag,
                                                            )}
                                                        <Badge
                                                            variant="outline"
                                                            class="text-xs"
                                                            style="border-color: {tagColor}; background-color: {tagColor}20; color: {tagColor};"
                                                        >
                                                            {tag}
                                                        </Badge>
                                                    {/each}
                                                </div>
                                            {/if}
                                        </div>

                                        <!-- Right Column - Title and Content -->
                                        <div class="flex-1 min-w-0">
                                            <!-- Event Title -->
                                            <div class="mb-3 sm:mb-4">
                                                <h2
                                                    class="text-xl sm:text-2xl font-bold text-gray-900 dark:text-neutral-100 leading-tight"
                                                >
                                                    {event.title}
                                                </h2>
                                            </div>

                                            <!-- Event Content -->
                                            {#if event.content}
                                                <div
                                                    class="prose prose-sm sm:prose prose-gray dark:prose-invert max-w-none mb-4 sm:mb-6"
                                                >
                                                    {@html markdownToHtml(
                                                        event.content,
                                                    )}
                                                </div>
                                            {/if}

                                            <!-- Event Media -->
                                            {#if event.media.length > 0}
                                                <div
                                                    class="grid grid-cols-1 md:grid-cols-2 gap-4"
                                                >
                                                    {#each event.media as mediaUrl}
                                                        <img
                                                            src={mediaUrl}
                                                            alt="Update media"
                                                            class="rounded-lg border border-gray-200 dark:border-neutral-700 max-w-full h-auto"
                                                            loading="lazy"
                                                        />
                                                    {/each}
                                                </div>
                                            {/if}
                                        </div>
                                    </article>
                                {/each}
                            </div>
                        {:else}
                            <div class="text-center py-16">
                                <div
                                    class="mx-auto h-24 w-24 text-gray-300 dark:text-neutral-600 mb-6"
                                >
                                    <svg
                                        fill="none"
                                        stroke="currentColor"
                                        viewBox="0 0 24 24"
                                        class="w-full h-full"
                                    >
                                        <path
                                            stroke-linecap="round"
                                            stroke-linejoin="round"
                                            stroke-width="1.5"
                                            d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
                                        />
                                    </svg>
                                </div>
                                <h3
                                    class="text-xl sm:text-2xl font-bold text-gray-900 dark:text-neutral-100 mb-4"
                                >
                                    No Updates Yet
                                </h3>
                                <p
                                    class="text-gray-500 dark:text-neutral-400 max-w-sm mx-auto"
                                >
                                    We haven't published any updates yet. Check
                                    back soon for the latest features and
                                    improvements!
                                </p>
                            </div>
                        {/if}
                    </div>
                </div>

                <!-- Right Sidebar -->
                <div
                    class="w-full lg:w-[250px] lg:flex-shrink-0 space-y-6 lg:space-y-8 order-first lg:order-last"
                >
                    <!-- Feedback Card -->
                    <div
                        class="bg-gray-50 dark:bg-neutral-800/50 rounded-lg p-4 sm:p-6"
                    >
                        <h3
                            class="text-lg sm:text-xl font-semibold mb-4 sm:mb-6 text-gray-900 dark:text-neutral-100"
                        >
                            Share Your Ideas
                        </h3>

                        {#if feedbackSuccess}
                            <div
                                class="bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800 text-green-800 dark:text-green-200 px-3 py-2 rounded-lg text-sm mb-4"
                            >
                                Thanks for your feedback! We'll review it soon.
                            </div>
                        {/if}

                        <form
                            on:submit|preventDefault={submitFeedback}
                            class="space-y-4"
                        >
                            <div>
                                <label
                                    for="feedback-title"
                                    class="block text-sm font-medium text-gray-700 dark:text-neutral-300 mb-2"
                                >
                                    Title
                                </label>
                                <input
                                    id="feedback-title"
                                    type="text"
                                    bind:value={feedbackTitle}
                                    placeholder="What's your idea?"
                                    class="w-full px-4 py-3 border border-gray-300 dark:border-neutral-700 rounded-lg bg-white dark:bg-neutral-800 text-gray-900 dark:text-neutral-100 placeholder-gray-500 dark:placeholder-neutral-400 focus:outline-none focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-400 focus:border-transparent transition-colors"
                                    disabled={submittingFeedback}
                                />
                            </div>

                            <div>
                                <label
                                    for="feedback-description"
                                    class="block text-sm font-medium text-gray-700 dark:text-neutral-300 mb-2"
                                >
                                    Description
                                </label>
                                <textarea
                                    id="feedback-description"
                                    bind:value={feedbackDescription}
                                    placeholder="Tell us more about your idea..."
                                    rows="4"
                                    class="w-full px-4 py-3 border border-gray-300 dark:border-neutral-700 rounded-lg bg-white dark:bg-neutral-800 text-gray-900 dark:text-neutral-100 placeholder-gray-500 dark:placeholder-neutral-400 focus:outline-none focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-400 focus:border-transparent resize-none transition-colors"
                                    disabled={submittingFeedback}
                                ></textarea>
                            </div>

                            <button
                                type="submit"
                                disabled={submittingFeedback ||
                                    !feedbackTitle.trim() ||
                                    !feedbackDescription.trim()}
                                class="w-full flex items-center justify-center gap-2 px-3 py-2 bg-blue-600 hover:bg-blue-700 disabled:bg-gray-400 dark:disabled:bg-neutral-700 text-white rounded-lg transition-all duration-200 disabled:cursor-not-allowed font-medium shadow-sm hover:shadow-md"
                            >
                                {#if submittingFeedback}
                                    <div
                                        class="animate-spin rounded-full h-4 w-4 border-b-2 border-white"
                                    ></div>
                                    Submitting...
                                {:else}
                                    <Send class="h-4 w-4" />
                                    Submit Idea
                                {/if}
                            </button>
                        </form>
                    </div>

                    <!-- Voting Section -->
                    <div
                        class="bg-gray-50 dark:bg-neutral-800/50 rounded-lg p-4 sm:p-6"
                    >
                        <h3
                            class="text-lg font-semibold mb-4 sm:mb-6 text-gray-900 dark:text-neutral-100"
                        >
                            Vote for Next Features
                        </h3>

                        {#if groupedEvents.upcoming.length > 0}
                            <div class="space-y-3">
                                {#each groupedEvents.upcoming as event}
                                    <div class="group">
                                        <div class="mb-2">
                                            <h4
                                                class="font-semibold text-foreground text-base leading-tight mb-2"
                                            >
                                                {event.title}
                                            </h4>

                                            {#if event.tags.length > 0}
                                                <div
                                                    class="flex flex-wrap gap-1"
                                                >
                                                    {#each event.tags as tag}
                                                        {@const tagColor =
                                                            generateTagColor(
                                                                tag,
                                                            )}
                                                        <Badge
                                                            variant="outline"
                                                            class="text-xs"
                                                            style="border-color: {tagColor}; background-color: {tagColor}20; color: {tagColor};"
                                                        >
                                                            {tag}
                                                        </Badge>
                                                    {/each}
                                                </div>
                                            {/if}
                                        </div>

                                        {#if voteErrors[event.id]}
                                            <div
                                                class="mb-2 p-2 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 text-red-800 dark:text-red-200 rounded text-xs"
                                            >
                                                {voteErrors[event.id]}
                                            </div>
                                        {/if}

                                        <button
                                            on:click={() =>
                                                handleVote(event.id)}
                                            class="w-full flex items-center justify-center gap-2 px-3 py-2 text-sm font-medium rounded-md transition-all duration-200 border {!votedEvents.has(
                                                event.id,
                                            )
                                                ? 'bg-white dark:bg-neutral-800 hover:bg-gray-50 dark:hover:bg-neutral-700 border-gray-300 dark:border-neutral-700 text-gray-700 dark:text-neutral-300 hover:border-gray-400 dark:hover:border-neutral-600'
                                                : 'bg-green-50 dark:bg-green-900/20 border-green-200 dark:border-green-800 text-green-700 dark:text-green-300 hover:bg-red-50 dark:hover:bg-red-900/20 hover:border-red-200 dark:hover:border-red-800 hover:text-red-700 dark:hover:text-red-300'}"
                                            title={votedEvents.has(event.id)
                                                ? "Click to remove your vote"
                                                : "Click to vote for this feature"}
                                        >
                                            <ThumbsUp class="h-4 w-4" />
                                            {votedEvents.has(event.id)
                                                ? "Remove Vote"
                                                : "Vote"} ({event.votes})
                                        </button>
                                    </div>

                                    {#if event !== groupedEvents.upcoming[groupedEvents.upcoming.length - 1]}
                                        <hr
                                            class="border-gray-200 dark:border-neutral-800"
                                        />
                                    {/if}
                                {/each}
                            </div>
                        {:else}
                            <div class="text-center py-8">
                                <div
                                    class="mx-auto h-12 w-12 text-gray-300 dark:text-neutral-600 mb-3"
                                >
                                    <svg
                                        fill="none"
                                        stroke="currentColor"
                                        viewBox="0 0 24 24"
                                        class="w-full h-full"
                                    >
                                        <path
                                            stroke-linecap="round"
                                            stroke-linejoin="round"
                                            stroke-width="1.5"
                                            d="M7 11.5V14m0-2.5v-6a1.5 1.5 0 113 0m-3 6a1.5 1.5 0 00-3 0v2a7.5 7.5 0 0015 0v-5a1.5 1.5 0 00-3 0m-6-3V11m0-5.5v-1a1.5 1.5 0 013 0v1m0 0V11m0-5.5T6.5 15a2.5 2.5 0 002.5 2.5zm0 0V11.5m0 9.5a2.5 2.5 0 01-2.5-2.5v-2m0 0V15"
                                        />
                                    </svg>
                                </div>
                                <p
                                    class="text-sm text-gray-500 dark:text-neutral-400"
                                >
                                    No features available for voting yet.
                                </p>
                            </div>
                        {/if}
                    </div>
                </div>
            </div>
        </div>
    {/if}

    <!-- Footer -->
    <footer
        class="border-t border-gray-200 dark:border-neutral-800 bg-white/80 dark:bg-neutral-950/80 backdrop-blur-sm"
    >
        <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
            <div class="flex items-center justify-center h-16">
                <p
                    class="text-sm text-gray-500 dark:text-neutral-400 flex items-center gap-2"
                >
                    Shipped with
                    <a
                        href="https://github.com/GauthierNelkinsky/ShipShipShip"
                        target="_blank"
                        rel="noopener noreferrer"
                        class="font-medium text-gray-700 dark:text-neutral-300 hover:text-gray-900 dark:hover:text-neutral-100 transition-colors flex items-center gap-1"
                    >
                        ShipShipShip 🚢
                    </a>
                </p>
            </div>
        </div>
    </footer>
</div>
