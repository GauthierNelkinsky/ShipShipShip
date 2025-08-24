<script lang="ts">
    import { onMount } from "svelte";
    import { browser } from "$app/environment";
    import { goto } from "$app/navigation";
    import { api } from "$lib/api";
    import {
        parseEvent,
        groupEventsByStatus,
        formatDate,
        markdownToHtml,
    } from "$lib/utils";
    import { settings } from "$lib/stores/settings";

    import type { ParsedEvent } from "$lib/types";
    import {
        Calendar,
        Tag,
        Send,
        ThumbsUp,
        MessageSquare,
        Mail,
        Vote,
    } from "lucide-svelte";
    import { Button, Card, Badge, Input, Textarea } from "$lib/components/ui";
    import ThemeSelector from "$lib/components/ThemeSelector.svelte";
    import NewsletterSubscription from "$lib/components/NewsletterSubscription.svelte";
    import Footer from "$lib/components/Footer.svelte";

    let events: ParsedEvent[] = [];
    let loading = true;
    let error = "";
    let groupedEvents: ReturnType<typeof groupEventsByStatus> = {
        backlogs: [],
        proposed: [],
        release: [],
        upcoming: [],
        archived: [],
    };

    // Feedback form
    let feedbackTitle = "";
    let feedbackDescription = "";
    let submittingFeedback = false;
    let feedbackSuccess = false;
    let formStartTime = Date.now();
    let lastSubmissionTime = 0;
    let submissionCount = 0;
    let feedbackError = "";

    // Vote tracking
    let votedEvents = new Set<number>();
    let voteErrors: Record<number, string> = {};

    // Newsletter settings
    let newsletterEnabled = false;

    onMount(async () => {
        // Load rate limiting data from localStorage
        if (browser) {
            const stored = localStorage.getItem("feedback_rate_limit");
            if (stored) {
                try {
                    const data = JSON.parse(stored);
                    lastSubmissionTime = data.lastSubmissionTime || 0;
                    submissionCount = data.submissionCount || 0;

                    // Reset count if it's been more than 24 hours
                    if (Date.now() - lastSubmissionTime > 24 * 60 * 60 * 1000) {
                        submissionCount = 0;
                    }
                } catch (err) {
                    console.error("Failed to parse rate limit data:", err);
                }
            }
        }

        await loadEvents();
        await loadNewsletterSettings();
        await loadVoteStatuses();
    });

    async function loadNewsletterSettings() {
        try {
            const settings = await api.getSettings();
            newsletterEnabled = !!settings?.newsletter_enabled;
        } catch (err) {
            newsletterEnabled = false;
        }
    }

    async function loadVoteStatuses() {
        // Check vote status for all proposed events
        const proposedEvents = groupedEvents.proposed || [];
        for (const event of proposedEvents) {
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

        // Time validation
        const now = Date.now();
        const formDuration = now - formStartTime;
        const minimumTime = 3000; // 3 seconds

        if (formDuration < minimumTime) {
            feedbackError =
                "Please take your time to fill out the form properly.";
            return;
        }

        // Client-side rate limiting
        const rateLimitWindow = 60 * 1000; // 1 minute
        if (now - lastSubmissionTime < rateLimitWindow) {
            const remainingTime = Math.ceil(
                (rateLimitWindow - (now - lastSubmissionTime)) / 1000,
            );
            feedbackError = `Please wait ${remainingTime} seconds before submitting again.`;
            return;
        }

        // Progressive rate limiting for frequent submitters
        if (submissionCount >= 3) {
            const extendedLimit = 5 * 60 * 1000; // 5 minutes
            if (now - lastSubmissionTime < extendedLimit) {
                const remainingTime = Math.ceil(
                    (extendedLimit - (now - lastSubmissionTime)) / 1000,
                );
                feedbackError = `Too many submissions. Please wait ${Math.ceil(remainingTime / 60)} minutes.`;
                return;
            }
        }

        submittingFeedback = true;
        feedbackError = "";

        try {
            await api.submitFeedback(
                feedbackTitle.trim(),
                feedbackDescription.trim(),
                formStartTime,
            );

            // Update rate limiting data
            lastSubmissionTime = now;
            submissionCount++;

            // Save to localStorage
            if (browser) {
                localStorage.setItem(
                    "feedback_rate_limit",
                    JSON.stringify({
                        lastSubmissionTime: lastSubmissionTime,
                        submissionCount: submissionCount,
                    }),
                );
            }

            feedbackTitle = "";
            feedbackDescription = "";
            feedbackSuccess = true;
            feedbackError = ""; // Clear any error messages
            formStartTime = Date.now(); // Reset form start time

            setTimeout(() => {
                feedbackSuccess = false;
            }, 3000);

            // Reload events to show the new feedback
            await loadEvents();
        } catch (err) {
            console.error("Failed to submit feedback:", err);
            // Handle rate limiting errors from server
            const errorMessage =
                err instanceof Error ? err.message : String(err);
            if (
                errorMessage.includes("Rate limit exceeded") ||
                errorMessage.includes("Too many submissions")
            ) {
                feedbackError = errorMessage;
            } else if (
                errorMessage.includes("take your time") ||
                errorMessage.includes("session expired")
            ) {
                feedbackError = errorMessage;
                formStartTime = Date.now(); // Reset form start time
            } else {
                feedbackError = "Failed to submit feedback. Please try again.";
            }
        } finally {
            submittingFeedback = false;
        }
    }

    // Combine doing and release events for timeline, sorted by date
    $: timelineEvents = [
        ...groupedEvents.upcoming,
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
                        {#if groupedEvents.upcoming.length > 0 || groupedEvents.release.length > 0}
                            {@const sortedUpcomingEvents =
                                groupedEvents.upcoming.sort((a, b) => {
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
                                <!-- Upcoming Events -->
                                {#each sortedUpcomingEvents as event, index}
                                    <article
                                        class="flex flex-col md:flex-row gap-4 md:gap-8 pb-8 relative"
                                    >
                                        <!-- Left Column - Date and Tags -->
                                        <div
                                            class="sticky-column w-full md:w-[250px] md:flex-shrink-0 text-left md:text-right"
                                        >
                                            <!-- Date -->
                                            {#if event.date}
                                                <div class="mb-3">
                                                    {#if event.status === "Upcoming"}
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
                                            {#if event.tags && Array.isArray(event.tags) && event.tags.length > 0}
                                                <div
                                                    class="flex flex-wrap gap-1 justify-start md:justify-end"
                                                >
                                                    {#each event.tags as tag}
                                                        <Badge
                                                            variant="outline"
                                                            class="text-xs"
                                                            style="background-color: {tag.color}20; color: {tag.color}; border-color: {tag.color}"
                                                        >
                                                            {tag.name}
                                                        </Badge>
                                                    {/each}
                                                </div>
                                            {/if}
                                        </div>

                                        <!-- Right Column - Title and Content -->
                                        <div class="flex-1 min-w-0">
                                            <!-- Event Title -->
                                            <div class="mb-2 sm:mb-3">
                                                <h1
                                                    class="text-2xl sm:text-3xl font-bold text-gray-900 dark:text-neutral-100 leading-tight cursor-pointer hover:text-primary transition-colors"
                                                    on:click={() =>
                                                        goto(`/${event.slug}`)}
                                                >
                                                    {event.title}
                                                </h1>
                                            </div>

                                            <!-- Event Content -->
                                            {#if event.content}
                                                <div
                                                    class="prose prose-sm prose-gray dark:prose-invert max-w-none mb-3 sm:mb-4"
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

                                <!-- Newsletter Section (Mobile) - After upcoming events if no release events -->
                                {#if sortedUpcomingEvents.length > 0 && sortedReleaseEvents.length === 0 && newsletterEnabled}
                                    <div class="lg:hidden mb-8">
                                        <NewsletterSubscription
                                            variant="inline"
                                        />
                                    </div>
                                {/if}

                                <!-- Separator between Doing and Released events -->
                                {#if sortedUpcomingEvents.length > 0 && sortedReleaseEvents.length > 0}
                                    <div class="py-8">
                                        <div
                                            class="border-t border-dashed border-gray-300 dark:border-neutral-600"
                                        ></div>
                                    </div>
                                {/if}

                                <!-- Released Events -->
                                {#each sortedReleaseEvents as event, index}
                                    {@const adjustedIndex =
                                        sortedUpcomingEvents.length + index}
                                    <!-- Vote for Next Features (Mobile) - After Doing Events -->
                                    {#if index === 0 && groupedEvents.proposed.length > 0}
                                        <div class="lg:hidden mb-8">
                                            <div
                                                class="bg-gray-50 dark:bg-neutral-800/50 rounded-lg p-4"
                                            >
                                                <h3
                                                    class="text-lg font-semibold mb-3 text-gray-900 dark:text-neutral-100 flex items-center gap-2"
                                                >
                                                    <Vote
                                                        class="h-5 w-5 text-primary"
                                                    />
                                                    Feature Voting
                                                </h3>
                                                <div class="space-y-2.5">
                                                    {#each groupedEvents.proposed as proposedEvent}
                                                        <div class="group">
                                                            <div class="mb-1.5">
                                                                <h4
                                                                    class="font-semibold text-foreground text-sm leading-tight mb-1.5 cursor-pointer hover:text-primary transition-colors"
                                                                    on:click={() =>
                                                                        goto(
                                                                            `/${proposedEvent.slug}`,
                                                                        )}
                                                                >
                                                                    {proposedEvent.title}
                                                                </h4>
                                                                {#if proposedEvent.tags && Array.isArray(proposedEvent.tags) && proposedEvent.tags.length > 0}
                                                                    <div
                                                                        class="flex flex-wrap gap-1"
                                                                    >
                                                                        {#each proposedEvent.tags as tag}
                                                                            <Badge
                                                                                variant="outline"
                                                                                class="text-xs"
                                                                                style="background-color: {tag.color}20; color: {tag.color}; border-color: {tag.color}"
                                                                            >
                                                                                {tag.name}
                                                                            </Badge>
                                                                        {/each}
                                                                    </div>
                                                                {/if}
                                                            </div>
                                                            {#if voteErrors[proposedEvent.id]}
                                                                <div
                                                                    class="mb-1.5 p-2 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 text-red-800 dark:text-red-200 rounded text-xs"
                                                                >
                                                                    {voteErrors[
                                                                        proposedEvent
                                                                            .id
                                                                    ]}
                                                                </div>
                                                            {/if}
                                                            <button
                                                                on:click={() =>
                                                                    handleVote(
                                                                        proposedEvent.id,
                                                                    )}
                                                                class="w-full flex items-center justify-center gap-1.5 px-2.5 py-1.5 text-xs font-medium rounded-md transition-all duration-200 border {!votedEvents.has(
                                                                    proposedEvent.id,
                                                                )
                                                                    ? 'bg-white dark:bg-neutral-800 hover:bg-gray-50 dark:hover:bg-neutral-700 border-gray-300 dark:border-neutral-700 text-gray-700 dark:text-neutral-300 hover:border-gray-400 dark:hover:border-neutral-600'
                                                                    : 'bg-green-50 dark:bg-green-900/20 border-green-200 dark:border-green-800 text-green-700 dark:text-green-300 hover:bg-red-50 dark:hover:bg-red-900/20 hover:border-red-200 dark:hover:border-red-800 hover:text-red-700 dark:hover:text-red-300'}"
                                                                title={votedEvents.has(
                                                                    proposedEvent.id,
                                                                )
                                                                    ? "Click to remove your vote"
                                                                    : "Click to vote for this feature"}
                                                            >
                                                                <ThumbsUp
                                                                    class="h-3.5 w-3.5"
                                                                />
                                                                {votedEvents.has(
                                                                    proposedEvent.id,
                                                                )
                                                                    ? "Remove Vote"
                                                                    : "Vote"} ({proposedEvent.votes})
                                                            </button>
                                                        </div>
                                                        {#if proposedEvent !== groupedEvents.proposed[groupedEvents.proposed.length - 1]}
                                                            <hr
                                                                class="border-gray-200 dark:border-neutral-800 my-2"
                                                            />
                                                        {/if}
                                                    {/each}
                                                </div>
                                            </div>
                                        </div>
                                    {/if}

                                    <article
                                        class="flex flex-col md:flex-row gap-4 md:gap-8 pb-8 relative"
                                    >
                                        <!-- Left Column - Date and Tags -->
                                        <div
                                            class="sticky-column w-full md:w-[250px] md:flex-shrink-0 text-left md:text-right"
                                        >
                                            <!-- Date -->
                                            {#if event.date}
                                                <div class="mb-3">
                                                    {#if event.status === "Upcoming"}
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
                                            {#if event.tags && Array.isArray(event.tags) && event.tags.length > 0}
                                                <div
                                                    class="flex flex-wrap gap-1 justify-start md:justify-end"
                                                >
                                                    {#each event.tags as tag}
                                                        <Badge
                                                            variant="outline"
                                                            class="text-xs"
                                                            style="background-color: {tag.color}20; color: {tag.color}; border-color: {tag.color}"
                                                        >
                                                            {tag.name}
                                                        </Badge>
                                                    {/each}
                                                </div>
                                            {/if}
                                        </div>

                                        <!-- Right Column - Title and Content -->
                                        <div class="flex-1 min-w-0">
                                            <!-- Event Title -->
                                            <div class="mb-2 sm:mb-3">
                                                <h1
                                                    class="text-2xl sm:text-3xl font-bold text-gray-900 dark:text-neutral-100 leading-tight cursor-pointer hover:text-primary transition-colors"
                                                    on:click={() =>
                                                        goto(`/${event.slug}`)}
                                                >
                                                    {event.title}
                                                </h1>
                                            </div>

                                            <!-- Event Content -->
                                            {#if event.content}
                                                <div
                                                    class="prose prose-sm prose-gray dark:prose-invert max-w-none mb-3 sm:mb-4"
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

                                    <!-- Newsletter Section (Mobile) - After 1st Released Event -->
                                    {#if index === 0 && newsletterEnabled}
                                        <div class="lg:hidden mb-8">
                                            <NewsletterSubscription
                                                variant="inline"
                                            />
                                        </div>
                                    {/if}

                                    <!-- Share Your Ideas Section (Mobile) - After 6th Released Event -->
                                    {#if index === 5}
                                        <div class="lg:hidden mb-8">
                                            <div
                                                class="bg-gray-50 dark:bg-neutral-800/50 rounded-lg p-4"
                                            >
                                                <h3
                                                    class="text-lg font-semibold mb-3 text-gray-900 dark:text-neutral-100"
                                                >
                                                    Share Your Ideas
                                                </h3>
                                                {#if feedbackSuccess}
                                                    <div
                                                        class="bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800 text-green-800 dark:text-green-200 px-3 py-2 rounded-lg text-sm mb-3"
                                                    >
                                                        Thanks for your
                                                        feedback! We'll review
                                                        it soon.
                                                    </div>
                                                {/if}
                                                {#if feedbackError}
                                                    <div
                                                        class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 text-red-800 dark:text-red-200 px-3 py-2 rounded-lg text-sm mb-3"
                                                    >
                                                        {feedbackError}
                                                    </div>
                                                {/if}
                                                <form
                                                    on:submit|preventDefault={submitFeedback}
                                                    class="space-y-3"
                                                >
                                                    <div>
                                                        <label
                                                            for="feedback-title-mobile"
                                                            class="block text-sm font-medium text-gray-700 dark:text-neutral-300 mb-1.5"
                                                        >
                                                            Title
                                                        </label>
                                                        <Input
                                                            id="feedback-title-mobile"
                                                            type="text"
                                                            placeholder="What's your idea?"
                                                            bind:value={
                                                                feedbackTitle
                                                            }
                                                            disabled={submittingFeedback}
                                                            class="text-sm"
                                                        />
                                                    </div>
                                                    <div>
                                                        <label
                                                            for="feedback-description-mobile"
                                                            class="block text-sm font-medium text-gray-700 dark:text-neutral-300 mb-1.5"
                                                        >
                                                            Description
                                                        </label>
                                                        <Textarea
                                                            id="feedback-description-mobile"
                                                            placeholder="Tell us more about your idea..."
                                                            bind:value={
                                                                feedbackDescription
                                                            }
                                                            disabled={submittingFeedback}
                                                            rows={3}
                                                            class="text-sm"
                                                        />
                                                    </div>
                                                    <button
                                                        type="submit"
                                                        disabled={submittingFeedback ||
                                                            !feedbackTitle.trim() ||
                                                            !feedbackDescription.trim()}
                                                        class="w-full flex items-center justify-center gap-2 px-3 py-2 text-sm font-medium rounded-md transition-all duration-200 bg-primary hover:bg-primary/90 disabled:opacity-50 disabled:cursor-not-allowed text-primary-foreground"
                                                    >
                                                        {#if submittingFeedback}
                                                            <div
                                                                class="animate-spin rounded-full h-3.5 w-3.5 border-b-2 border-white"
                                                            ></div>
                                                            Submitting...
                                                        {:else}
                                                            <Send
                                                                class="h-3.5 w-3.5"
                                                            />
                                                            Submit Idea
                                                        {/if}
                                                    </button>
                                                </form>
                                            </div>
                                        </div>
                                    {/if}
                                {/each}
                            </div>
                        {:else}
                            <!-- Newsletter Section (Mobile) - Fallback when no events -->
                            {#if newsletterEnabled}
                                <div class="lg:hidden mb-8">
                                    <NewsletterSubscription variant="inline" />
                                </div>
                            {/if}

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
                    class="w-full lg:w-[250px] lg:flex-shrink-0 space-y-4 lg:space-y-5 order-first lg:order-last"
                >
                    <!-- Feedback Card -->
                    <div
                        class="bg-gray-50 dark:bg-neutral-800/50 rounded-lg p-4"
                    >
                        <h3
                            class="text-lg font-semibold mb-3 text-gray-900 dark:text-neutral-100 flex items-center gap-2"
                        >
                            <MessageSquare class="h-5 w-5 text-primary" />
                            Share Your Ideas
                        </h3>

                        {#if feedbackSuccess}
                            <div
                                class="bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800 text-green-800 dark:text-green-200 px-3 py-2 rounded-lg text-sm mb-4"
                            >
                                Thanks for your feedback! We'll review it soon.
                            </div>
                        {/if}

                        {#if feedbackError}
                            <div
                                class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 text-red-800 dark:text-red-200 px-3 py-2 rounded-lg text-sm mb-4"
                            >
                                {feedbackError}
                            </div>
                        {/if}

                        <form
                            on:submit|preventDefault={submitFeedback}
                            class="space-y-3"
                        >
                            <div>
                                <label
                                    for="feedback-title"
                                    class="block text-sm font-medium text-gray-700 dark:text-neutral-300 mb-1.5"
                                >
                                    Title
                                </label>
                                <input
                                    id="feedback-title"
                                    type="text"
                                    bind:value={feedbackTitle}
                                    placeholder="What's your idea?"
                                    class="w-full px-3 py-2 border border-gray-300 dark:border-neutral-700 rounded-lg bg-white dark:bg-neutral-800 text-gray-900 dark:text-neutral-100 placeholder-gray-500 dark:placeholder-neutral-400 focus:outline-none focus:ring-2 focus:ring-primary dark:focus:ring-primary focus:border-transparent transition-colors text-sm"
                                    disabled={submittingFeedback}
                                />
                            </div>

                            <div>
                                <label
                                    for="feedback-description"
                                    class="block text-sm font-medium text-gray-700 dark:text-neutral-300 mb-1.5"
                                >
                                    Description
                                </label>
                                <textarea
                                    id="feedback-description"
                                    bind:value={feedbackDescription}
                                    placeholder="Tell us more about your idea..."
                                    rows="3"
                                    class="w-full px-3 py-2 border border-gray-300 dark:border-neutral-700 rounded-lg bg-white dark:bg-neutral-800 text-gray-900 dark:text-neutral-100 placeholder-gray-500 dark:placeholder-neutral-400 focus:outline-none focus:ring-2 focus:ring-primary dark:focus:ring-primary focus:border-transparent resize-none transition-colors text-sm"
                                    disabled={submittingFeedback}
                                ></textarea>
                            </div>

                            <button
                                type="submit"
                                disabled={submittingFeedback ||
                                    !feedbackTitle.trim() ||
                                    !feedbackDescription.trim()}
                                class="w-full flex items-center justify-center gap-2 px-3 py-2 text-sm font-medium rounded-md transition-all duration-200 bg-primary hover:bg-primary/90 disabled:opacity-50 disabled:cursor-not-allowed text-primary-foreground"
                            >
                                {#if submittingFeedback}
                                    <div
                                        class="animate-spin rounded-full h-3.5 w-3.5 border-b-2 border-white"
                                    ></div>
                                    Submitting...
                                {:else}
                                    <Send class="h-3.5 w-3.5" />
                                    Submit Idea
                                {/if}
                            </button>
                        </form>
                    </div>

                    <!-- Newsletter Section - Desktop/Sidebar -->
                    {#if newsletterEnabled}
                        <div class="hidden lg:block">
                            <NewsletterSubscription variant="sidebar" />
                        </div>
                    {/if}

                    <!-- Voting Section -->
                    {#if groupedEvents.proposed.length > 0}
                        <div
                            class="bg-gray-50 dark:bg-neutral-800/50 rounded-lg p-4 hidden lg:block"
                        >
                            <h3
                                class="text-lg font-semibold mb-3 text-gray-900 dark:text-neutral-100 flex items-center gap-2"
                            >
                                <Vote class="h-5 w-5 text-primary" />
                                Feature Voting
                            </h3>
                            <div class="space-y-2.5">
                                {#each groupedEvents.proposed as event}
                                    <div class="group">
                                        <div class="mb-1.5">
                                            <h4
                                                class="font-semibold text-foreground text-sm leading-tight mb-1.5 cursor-pointer hover:text-primary transition-colors"
                                                on:click={() =>
                                                    goto(`/${event.slug}`)}
                                            >
                                                {event.title}
                                            </h4>
                                            {#if event.tags && Array.isArray(event.tags) && event.tags.length > 0}
                                                <div
                                                    class="flex flex-wrap gap-1"
                                                >
                                                    {#each event.tags as tag}
                                                        <Badge
                                                            variant="outline"
                                                            class="text-xs"
                                                            style="background-color: {tag.color}20; color: {tag.color}; border-color: {tag.color}"
                                                        >
                                                            {tag.name}
                                                        </Badge>
                                                    {/each}
                                                </div>
                                            {/if}
                                        </div>
                                        {#if voteErrors[event.id]}
                                            <div
                                                class="mb-1.5 p-2 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 text-red-800 dark:text-red-200 rounded text-xs"
                                            >
                                                {voteErrors[event.id]}
                                            </div>
                                        {/if}
                                        <button
                                            on:click={() =>
                                                handleVote(event.id)}
                                            class="w-full flex items-center justify-center gap-1.5 px-2.5 py-1.5 text-xs font-medium rounded-md transition-all duration-200 border {!votedEvents.has(
                                                event.id,
                                            )
                                                ? 'bg-white dark:bg-neutral-800 hover:bg-gray-50 dark:hover:bg-neutral-700 border-gray-300 dark:border-neutral-700 text-gray-700 dark:text-neutral-300 hover:border-gray-400 dark:hover:border-neutral-600'
                                                : 'bg-green-50 dark:bg-green-900/20 border-green-200 dark:border-green-800 text-green-700 dark:text-green-300 hover:bg-red-50 dark:hover:bg-red-900/20 hover:border-red-200 dark:hover:border-red-800 hover:text-red-700 dark:hover:text-red-300'}"
                                            title={votedEvents.has(event.id)
                                                ? "Click to remove your vote"
                                                : "Click to vote for this feature"}
                                        >
                                            <ThumbsUp class="h-3.5 w-3.5" />
                                            {votedEvents.has(event.id)
                                                ? "Remove Vote"
                                                : "Vote"} ({event.votes})
                                        </button>
                                    </div>
                                    {#if event !== groupedEvents.proposed[groupedEvents.proposed.length - 1]}
                                        <hr
                                            class="border-gray-200 dark:border-neutral-800 my-2"
                                        />
                                    {/if}
                                {/each}
                            </div>
                        </div>
                    {/if}
                </div>
            </div>
        </div>
    {/if}

    <!-- Footer -->
    <Footer />
</div>

<style>
    /* Simple CSS sticky behavior */
    @media (min-width: 768px) {
        .sticky-column {
            position: sticky;
            top: 80px;
            align-self: flex-start;
        }
    }

    /* Ensure proper positioning context */
    article {
        position: relative;
    }
</style>
