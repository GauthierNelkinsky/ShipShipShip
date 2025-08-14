<script lang="ts">
    import { createEventDispatcher } from "svelte";
    import { fly, scale } from "svelte/transition";
    import { quintOut } from "svelte/easing";
    import type { ParsedEvent } from "$lib/types";
    import { markdownToHtml, formatDate } from "$lib/utils";
    import {
        Trash2,
        Calendar,
        Tag,
        Archive,
        Inbox,
        ChevronUp,
        ChevronDown,
    } from "lucide-svelte";
    import { Card, Button, Badge } from "$lib/components/ui";

    const dispatch = createEventDispatcher();

    export let event: ParsedEvent;
    export let isBeingDragged: boolean = false;
    export const draggable = true;

    let isDragging = false;
    let draggedOver = false;
    let startPos = { x: 0, y: 0 };
    let isDragStarted = false;

    function handleCardClick(e: any) {
        if (!isDragStarted) {
            dispatch("edit", event);
        }
    }

    function handleKeyDown(e: any) {
        if (e.key === "Enter") {
            handleCardClick(e);
        }
    }

    function handleMouseDown(e: MouseEvent) {
        startPos = { x: e.clientX, y: e.clientY };
        isDragStarted = false;

        function handleMouseMove(moveEvent: MouseEvent) {
            const distance = Math.sqrt(
                Math.pow(moveEvent.clientX - startPos.x, 2) +
                    Math.pow(moveEvent.clientY - startPos.y, 2),
            );

            if (distance > 5 && !isDragStarted) {
                isDragStarted = true;
                startDrag();
            }
        }

        function handleMouseUp() {
            document.removeEventListener("mousemove", handleMouseMove);
            document.removeEventListener("mouseup", handleMouseUp);

            setTimeout(() => {
                isDragStarted = false;
            }, 100);
        }

        document.addEventListener("mousemove", handleMouseMove);
        document.addEventListener("mouseup", handleMouseUp);
    }

    function startDrag() {
        // Store drag data in parent component
        dispatch("dragstart", {
            eventId: event.id,
            sourceStatus: event.status,
        });

        isDragging = true;
    }

    function handleDelete(e: Event) {
        e.stopPropagation();
        e.preventDefault();
        if (
            confirm(
                "Are you sure you want to delete this event? This action cannot be undone.",
            )
        ) {
            dispatch("delete", event.id);
        }
    }

    function handleMoveToArchived(e: Event) {
        e.stopPropagation();
        e.preventDefault();
        dispatch("statusChange", { eventId: event.id, newStatus: "Archived" });
    }

    function handleMoveToBacklog(e: Event) {
        e.stopPropagation();
        e.preventDefault();
        dispatch("statusChange", { eventId: event.id, newStatus: "Backlogs" });
    }

    function handleMoveUp(e: Event) {
        e.stopPropagation();
        e.preventDefault();
        dispatch("moveUp", event);
    }

    function handleMoveDown(e: Event) {
        e.stopPropagation();
        e.preventDefault();
        dispatch("moveDown", event);
    }

    function handleDragStart(e: DragEvent) {
        e.stopPropagation();

        if (e.dataTransfer) {
            e.dataTransfer.effectAllowed = "move";
            e.dataTransfer.dropEffect = "move";
            isDragging = true;

            // Store drag data in parent component instead of dataTransfer
            dispatch("dragstart", {
                eventId: event.id,
                sourceStatus: event.status,
            });
        }
    }

    function handleDragEnd() {
        isDragging = false;
        draggedOver = false;

        dispatch("dragend");

        // Reset after drag
        setTimeout(() => {
            // Reset any state if needed
        }, 100);
    }

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
</script>

<div transition:fly={{ y: 20, duration: 300, easing: quintOut }}>
    <div
        class="p-3 cursor-pointer transition-all duration-200 hover:shadow-md hover:scale-[1.01] transform {isDragging ||
        isBeingDragged
            ? 'opacity-50 rotate-2 scale-95 ring-2 ring-primary'
            : ''} group bg-card text-card-foreground rounded-lg border border-border shadow-sm w-full"
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
                class="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity"
            >
                <Button
                    variant="ghost"
                    size="icon"
                    on:click={handleMoveToBacklog}
                    class="h-6 w-6 hover:bg-primary hover:text-primary-foreground"
                    title="Move to Backlog"
                >
                    <Inbox class="h-3 w-3" />
                </Button>
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
        {#if event.tags.length > 0}
            <div class="flex flex-wrap gap-1 mb-2 min-w-0">
                {#each event.tags.slice(0, 2) as tag}
                    {@const tagColor = generateTagColor(tag)}
                    <Badge
                        variant="outline"
                        class="text-xs truncate max-w-20"
                        style="border-color: {tagColor}; background-color: {tagColor}20; color: {tagColor};"
                    >
                        {tag}
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

            <!-- Reorder buttons positioned absolutely -->
            <div
                class="absolute bottom-0 right-0 flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity"
            >
                <Button
                    variant="ghost"
                    size="sm"
                    on:click={handleMoveUp}
                    class="h-5 w-5 p-0 hover:bg-muted"
                    title="Move up"
                >
                    <ChevronUp class="h-3 w-3" />
                </Button>
                <Button
                    variant="ghost"
                    size="sm"
                    on:click={handleMoveDown}
                    class="h-5 w-5 p-0 hover:bg-muted"
                    title="Move down"
                >
                    <ChevronDown class="h-3 w-3" />
                </Button>
            </div>
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

    .line-clamp-3 {
        display: -webkit-box;
        -webkit-line-clamp: 3;
        line-clamp: 3;
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
