<script lang="ts">
    import { createEventDispatcher } from "svelte";
    import { fly } from "svelte/transition";
    import { quintOut } from "svelte/easing";
    import type { ParsedEvent } from "$lib/types";
    import { markdownToHtml, formatDate } from "$lib/utils";

    import { Trash2, Calendar, Archive, Share2 } from "lucide-svelte";
    import { Button, Badge } from "$lib/components/ui";

    const dispatch = createEventDispatcher();

    export let event: ParsedEvent;
    export let isBeingDragged: boolean = false;
    export const draggable = true;

    let isDragging = false;
    let isDragStarted = false;

    function handleCardClick(_e: any) {
        if (!isDragStarted) {
            dispatch("edit", event);
        }
    }

    function handleKeyDown(e: any) {
        if (e.key === "Enter") {
            handleCardClick(e);
        }
    }

    function _startDrag() {
        // Store drag data in parent component
        dispatch("carddragstart", {
            eventId: event.id,
            sourceStatus: event.status,
        });

        isDragging = true;
    }

    function handleDelete(e: Event) {
        e.stopPropagation();
        e.preventDefault();
        dispatch("delete", event.id);
    }

    function handleMoveToArchived(e: Event) {
        e.stopPropagation();
        e.preventDefault();
        dispatch("statusChange", { eventId: event.id, newStatus: "Archived" });
    }

    function handleShare(e: Event) {
        e.stopPropagation();
        e.preventDefault();
        dispatch("publish", event);
    }

    function handleDragStart(e: DragEvent) {
        e.stopPropagation();

        if (e.dataTransfer) {
            e.dataTransfer.effectAllowed = "move";
            e.dataTransfer.dropEffect = "move";
            isDragging = true;

            // Store drag data in parent component instead of dataTransfer
            dispatch("carddragstart", {
                eventId: event.id,
                sourceStatus: event.status,
            });
        }
    }

    function handleDragEnd() {
        isDragging = false;

        dispatch("carddragend");

        // Reset after drag
        setTimeout(() => {
            // Reset any state if needed
        }, 100);
    }
</script>

<div transition:fly={{ y: 20, duration: 300, easing: quintOut }}>
    <div
        class="p-3 cursor-pointer transition-all duration-200 hover:shadow-md hover:scale-[1.01] transform {isDragging ||
        isBeingDragged
            ? 'opacity-50 rotate-2 scale-95 ring-2 ring-primary'
            : ''} bg-card text-card-foreground rounded-lg border border-border shadow-sm w-full group/card"
        draggable="true"
        on:dragstart={handleDragStart}
        on:dragend={handleDragEnd}
        on:click={handleCardClick}
        on:keydown={handleKeyDown}
        role="button"
        tabindex={0}
    >
        <!-- Card Header -->
        <div class="flex items-start justify-between mb-2">
            <h3
                class="font-semibold text-xs sm:text-sm leading-tight flex-1 pr-1 break-words overflow-wrap-anywhere"
            >
                {event.title}
            </h3>
            <div
                class="flex items-center gap-1 opacity-0 group-hover/card:opacity-100 transition-opacity"
            >
                {#if event.status !== "Backlogs" && event.status !== "Archived"}
                    <Button
                        variant="ghost"
                        size="icon"
                        on:click={handleShare}
                        class="h-6 w-6 hover:bg-green-500 hover:text-white"
                        title="Share"
                    >
                        <Share2 class="h-3 w-3" />
                    </Button>
                {/if}

                <Button
                    variant="ghost"
                    size="icon"
                    on:click={handleMoveToArchived}
                    class="h-6 w-6 hover:bg-secondary hover:text-secondary-foreground"
                    title="Move to Archived"
                >
                    <Archive class="h-3 w-3" />
                </Button>
                <Button
                    variant="ghost"
                    size="icon"
                    on:click={handleDelete}
                    class="h-6 w-6 hover:bg-destructive hover:text-destructive-foreground"
                    title="Delete"
                >
                    <Trash2 class="h-3 w-3" />
                </Button>
            </div>
        </div>

        <!-- Tags -->
        {#if event.tags && event.tags.length > 0}
            <div class="flex flex-wrap gap-1 mb-2 min-w-0">
                {#each event.tags.slice(0, 2) as tag}
                    <Badge
                        variant="outline"
                        class="text-xs truncate max-w-32"
                        style="background-color: {tag.color}20; color: {tag.color}; border-color: {tag.color}"
                    >
                        <span class="truncate">{tag.name}</span>
                    </Badge>
                {/each}
                {#if event.tags.length > 2}
                    <Badge variant="secondary" class="text-xs">
                        +{event.tags.length - 2}
                    </Badge>
                {/if}
            </div>
        {/if}

        <!-- Content Preview -->
        {#if event.content}
            <div
                class="text-xs text-muted-foreground mb-2 line-clamp-2 break-words overflow-hidden"
            >
                {@html markdownToHtml(
                    event.content.slice(0, 80) +
                        (event.content.length > 80 ? "..." : ""),
                )}
            </div>
        {/if}

        <!-- Footer Info -->
        <div class="space-y-1 relative min-w-0">
            <!-- Date -->
            {#if event.date}
                <div
                    class="flex items-center gap-1 text-xs text-muted-foreground min-w-0"
                >
                    <Calendar class="h-3 w-3 flex-shrink-0" />
                    <span class="truncate">{formatDate(event.date)}</span>
                </div>
            {/if}

            <!-- Votes (for Upcoming status) -->
            {#if event.status === "Upcoming"}
                <div
                    class="flex items-center gap-1 text-xs text-muted-foreground min-w-0"
                >
                    <span>{event.votes} votes</span>
                </div>
            {/if}

            <!-- Media indicator -->
            {#if event.media.length > 0}
                <div class="text-xs text-muted-foreground break-words">
                    📎 {event.media.length} media file{event.media.length > 1
                        ? "s"
                        : ""}
                </div>
            {/if}
        </div>
    </div>
</div>

<style>
    .line-clamp-2 {
        display: -webkit-box;
        -webkit-line-clamp: 2;
        line-clamp: 2;
        -webkit-box-orient: vertical;
        overflow: hidden;
    }

    /* Ensure proper text wrapping and prevent overflow */
    .overflow-wrap-anywhere {
        overflow-wrap: anywhere;
        word-wrap: break-word;
        word-break: break-word;
        hyphens: auto;
    }
</style>
