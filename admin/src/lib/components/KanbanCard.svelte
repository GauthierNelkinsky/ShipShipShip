<script lang="ts">
    import { createEventDispatcher } from "svelte";
    import { fly } from "svelte/transition";
    import { quintOut } from "svelte/easing";
    import type { ParsedEvent } from "$lib/types";
    import { markdownToHtml, formatDate } from "$lib/utils";
    import * as m from "$lib/paraglide/messages";

    import { Trash2, Calendar } from "lucide-svelte";
    import { Button, Badge } from "$lib/components/ui";
    import Icon from "@iconify/svelte";

    const dispatch = createEventDispatcher();

    export let event: ParsedEvent;
    export let isBeingDragged: boolean = false;
    export const draggable = true;

    let isDragging = false;
    let isDragStarted = false;

    const reactionIcons: Record<string, string> = {
        thumbs_up: "fluent-emoji-flat:thumbs-up",
        heart: "fluent-emoji-flat:red-heart",
        fire: "fluent-emoji-flat:fire",
        party: "fluent-emoji-flat:party-popper",
        eyes: "fluent-emoji-flat:eyes",
        lightbulb: "fluent-emoji-flat:light-bulb",
        thinking: "fluent-emoji-flat:thinking-face",
        thumbs_down: "fluent-emoji-flat:thumbs-down",
    };

    $: visibleReactions = event.reaction_summary?.reactions
        ? event.reaction_summary.reactions.filter((r: any) => r.count > 0)
        : [];

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

    function handleDragStart(e: DragEvent) {
        e.stopPropagation();

        if (e.dataTransfer) {
            e.dataTransfer.effectAllowed = "move";
            e.dataTransfer.dropEffect = "move";
            isDragging = true;
            isDragStarted = true;

            // Store drag data in parent component instead of dataTransfer
            dispatch("carddragstart", {
                eventId: event.id,
                sourceStatus: event.status,
            });
        }
    }

    function handleDragEnd() {
        isDragging = false;
        isDragStarted = false;

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
                class="font-semibold text-xs sm:text-sm leading-tight flex-1 pe-1 break-words overflow-wrap-anywhere"
            >
                {event.title}
            </h3>
            <div
                class="flex items-center gap-1 opacity-0 group-hover/card:opacity-100 transition-opacity"
            >
                <Button
                    variant="ghost"
                    size="icon"
                    on:click={handleDelete}
                    class="h-6 w-6 hover:bg-destructive hover:text-destructive-foreground"
                    title={m.kanban_card_delete()}
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

            <!-- Reactions Display -->
            {#if visibleReactions.length > 0}
                <div class="flex flex-wrap items-center gap-1 min-w-0">
                    {#each visibleReactions as reaction}
                        <div
                            class="inline-flex items-center gap-1 px-1.5 py-0.5 rounded-full border border-border bg-background text-xs"
                            title="{reaction.reaction_type}: {reaction.count}"
                        >
                            <Icon
                                icon={reactionIcons[reaction.reaction_type] ||
                                    "fluent-emoji-flat:thumbs-up"}
                                class="h-3.5 w-3.5"
                            />
                            <span class="text-muted-foreground font-medium">
                                {reaction.count}
                            </span>
                        </div>
                    {/each}
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
