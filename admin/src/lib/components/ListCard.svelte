<script lang="ts">
    import { createEventDispatcher } from "svelte";
    import type { ParsedEvent } from "$lib/types";
    import { Calendar, Trash2 } from "lucide-svelte";
    import { Badge } from "$lib/components/ui";
    import { formatDate } from "$lib/utils";
    import Icon from "@iconify/svelte";

    const dispatch = createEventDispatcher();

    export let event: ParsedEvent;
    export let sourceStatus: string = "";

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

    function handleDelete() {
        dispatch("delete", event);
    }

    function handleClick() {
        dispatch("edit", event);
    }

    function handleKeyDown(e: KeyboardEvent) {
        if (e.key === "Enter" || e.key === " ") {
            e.preventDefault();
            handleClick();
        }
    }

    function handleDragStart(e: DragEvent) {
        e.stopPropagation();
        dispatch("dragstart", { event, sourceStatus });
    }

    function handleDragEnd(e: DragEvent) {
        e.stopPropagation();
        dispatch("dragend");
    }
</script>

<div
    class="group w-full px-4 py-2.5 relative cursor-pointer"
    role="button"
    tabindex="0"
    draggable="true"
    on:click={handleClick}
    on:keydown={handleKeyDown}
    on:dragstart={handleDragStart}
    on:dragend={handleDragEnd}
>
    <div class="flex items-center justify-between gap-4">
        <!-- Left: Title and Metadata (first line) -->
        <div class="flex-1 min-w-0 space-y-0.5">
            <!-- First line: Title + Date + Tags + Reactions -->
            <div class="flex items-center gap-2 flex-wrap">
                <span class="text-sm font-medium truncate max-w-xs">
                    {event.title}
                </span>

                {#if event.date}
                    <span class="text-muted-foreground/40">•</span>
                    <span
                        class="inline-flex items-center gap-1 text-xs text-muted-foreground flex-shrink-0"
                    >
                        <Calendar class="h-3 w-3" />
                        {formatDate(event.date)}
                    </span>
                {/if}

                {#if event.tags && event.tags.length > 0}
                    <span class="text-muted-foreground/40">•</span>
                    <div class="flex flex-wrap gap-1">
                        {#each event.tags.slice(0, 2) as tag}
                            <Badge
                                variant="outline"
                                class="text-xs px-1.5 py-0 h-auto"
                                style="background-color: {tag.color}20; color: {tag.color}; border-color: {tag.color}40"
                            >
                                {tag.name}
                            </Badge>
                        {/each}
                        {#if event.tags.length > 2}
                            <Badge
                                variant="secondary"
                                class="text-xs px-1.5 py-0 h-auto"
                            >
                                +{event.tags.length - 2}
                            </Badge>
                        {/if}
                    </div>
                {/if}

                {#if visibleReactions.length > 0}
                    <span class="text-muted-foreground/40">•</span>
                    <div class="flex items-center gap-1">
                        {#each visibleReactions as reaction}
                            <div
                                class="inline-flex items-center gap-1 px-1.5 py-0.5 rounded-full border border-border bg-background text-xs"
                                title="{reaction.reaction_type}: {reaction.count}"
                            >
                                <Icon
                                    icon={reactionIcons[
                                        reaction.reaction_type
                                    ] || "fluent-emoji-flat:thumbs-up"}
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

            <!-- Second line: Content Preview -->
            {#if event.content}
                <div class="text-xs text-muted-foreground truncate">
                    {event.content
                        .replace(/[#*`_~[\]]/g, "")
                        .trim()
                        .substring(0, 120)}{event.content.length > 120
                        ? "..."
                        : ""}
                </div>
            {/if}
        </div>

        <!-- Right: Actions (visible on hover) -->
        <div
            class="flex items-center gap-0.5 opacity-0 group-hover:opacity-100 transition-opacity flex-shrink-0"
        >
            <button
                type="button"
                class="h-6 w-6 flex items-center justify-center rounded hover:bg-destructive/10 text-muted-foreground hover:text-destructive transition-colors"
                on:click|stopPropagation={handleDelete}
                title="Delete event"
            >
                <Trash2 class="h-3 w-3" />
            </button>
        </div>
    </div>
</div>
