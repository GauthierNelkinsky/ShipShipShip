<script lang="ts">
    import { createEventDispatcher } from "svelte";
    import type { ParsedEvent, Tag as TagType } from "$lib/types";
    import { formatDate } from "$lib/utils";

    import {
        Trash2,
        Edit,
        Tag,
        Calendar,
        GripVertical,
        ArrowUp,
        Archive,
        Inbox,
    } from "lucide-svelte";
    import { Card, Button, Badge } from "$lib/components/ui";
    import { flip } from "svelte/animate";

    const dispatch = createEventDispatcher();

    export let events: ParsedEvent[] = [];
    export let loading = false;

    let draggedIndex: number | null = null;
    let dropTargetIndex: number | null = null;

    function handleEdit(event: ParsedEvent) {
        dispatch("edit", event);
    }

    function handleDelete(eventId: number) {
        if (
            confirm(
                "Are you sure you want to delete this event? This action cannot be undone.",
            )
        ) {
            dispatch("delete", eventId);
        }
    }

    function handleMoveToBacklog(event: ParsedEvent) {
        dispatch("statusChange", { eventId: event.id, newStatus: "Backlogs" });
    }

    function handleDragStart(e: DragEvent, index: number) {
        if (e.dataTransfer) {
            e.dataTransfer.effectAllowed = "move";
            e.dataTransfer.setData("text/plain", events[index].id.toString());
            e.dataTransfer.setData(
                "application/json",
                JSON.stringify({
                    id: events[index].id,
                    sourceType: "archived",
                    sourceIndex: index,
                }),
            );
            draggedIndex = index;
        }
    }

    function handleDragEnd() {
        draggedIndex = null;
        dropTargetIndex = null;
    }

    function handleDragOver(e: DragEvent, index: number) {
        e.preventDefault();
        if (draggedIndex !== null && draggedIndex !== index) {
            dropTargetIndex = index;
        }
    }

    function handleDragLeave() {
        dropTargetIndex = null;
    }

    function handleDrop(e: DragEvent, targetIndex: number) {
        e.preventDefault();

        if (draggedIndex === null || draggedIndex === targetIndex) {
            draggedIndex = null;
            dropTargetIndex = null;
            return;
        }

        // Reorder the events array
        const newEvents = [...events];
        const draggedEvent = newEvents[draggedIndex];
        newEvents.splice(draggedIndex, 1);
        newEvents.splice(targetIndex, 0, draggedEvent);

        // Dispatch reorder event
        dispatch("reorder", newEvents);

        draggedIndex = null;
        dropTargetIndex = null;
    }

    function truncateText(text: string, maxLength: number = 100): string {
        if (text.length <= maxLength) return text;
        return text.slice(0, maxLength) + "...";
    }
</script>

<Card class="overflow-hidden">
    <div class="overflow-x-auto">
        <table class="w-full">
            <thead class="border-b border-border">
                <tr class="bg-muted" style="opacity: 0.5;">
                    <th
                        class="text-left py-2 px-3 font-medium text-sm text-muted-foreground w-8"
                    ></th>
                    <th
                        class="text-left py-2 px-3 font-medium text-sm text-muted-foreground"
                        >Name</th
                    >
                    <th
                        class="text-left py-2 px-3 font-medium text-sm text-muted-foreground"
                        >Tags</th
                    >
                    <th
                        class="text-left py-2 px-3 font-medium text-sm text-muted-foreground"
                        >Date</th
                    >
                    <th
                        class="text-left py-2 px-3 font-medium text-sm text-muted-foreground"
                        >Votes</th
                    >
                    <th
                        class="text-right py-2 px-3 font-medium text-sm text-muted-foreground"
                    ></th>
                </tr>
            </thead>
            <tbody>
                {#if loading}
                    <tr>
                        <td
                            colspan="6"
                            class="p-8 text-center text-muted-foreground"
                        >
                            <div class="flex items-center justify-center gap-2">
                                <div
                                    class="animate-spin rounded-full h-4 w-4 border-b-2 border-primary"
                                ></div>
                                Loading...
                            </div>
                        </td>
                    </tr>
                {:else if events.length === 0}
                    <tr>
                        <td
                            colspan="6"
                            class="p-8 text-center text-muted-foreground"
                        >
                            No archived items found
                        </td>
                    </tr>
                {:else}
                    {#each events as event, index (event.id)}
                        <tr
                            class="border-b border-border hover:bg-muted transition-colors group cursor-pointer {draggedIndex ===
                            index
                                ? 'opacity-50'
                                : ''} {dropTargetIndex === index
                                ? 'bg-primary/10'
                                : ''}"
                            style="--hover-opacity: 0.2;"
                            on:click={() => handleEdit(event)}
                            on:mouseenter={(e) =>
                                draggedIndex === null &&
                                (e.currentTarget.style.backgroundColor =
                                    "hsl(var(--muted) / 0.2)")}
                            on:mouseleave={(e) =>
                                (e.currentTarget.style.backgroundColor = "")}
                            on:dragover={(e) => handleDragOver(e, index)}
                            on:dragleave={handleDragLeave}
                            on:drop={(e) => handleDrop(e, index)}
                            animate:flip={{ duration: 300 }}
                        >
                            <!-- Drag Handle -->
                            <td class="py-2 px-3 w-8">
                                <div
                                    class="cursor-grab active:cursor-grabbing flex items-center justify-center text-muted-foreground hover:text-foreground transition-colors"
                                    role="button"
                                    tabindex="0"
                                    draggable="true"
                                    on:dragstart={(e) =>
                                        handleDragStart(e, index)}
                                    on:dragend={handleDragEnd}
                                    on:click={(e) => e.stopPropagation()}
                                    on:keydown={(e) => {
                                        if (
                                            e.key === "Enter" ||
                                            e.key === " "
                                        ) {
                                            e.preventDefault();
                                            e.stopPropagation();
                                        }
                                    }}
                                    title="Drag to reorder"
                                    aria-label="Drag to reorder event"
                                >
                                    <GripVertical class="h-4 w-4" />
                                </div>
                            </td>

                            <!-- Name -->
                            <td class="py-2 px-3">
                                <div
                                    class="font-medium text-sm text-foreground"
                                >
                                    {event.title}
                                </div>
                            </td>

                            <!-- Tags -->
                            <td class="py-2 px-3">
                                {#if event.tags && Array.isArray(event.tags) && event.tags.length > 0}
                                    <div class="flex flex-wrap gap-1">
                                        {#each event.tags.slice(0, 3) as tag}
                                            <Badge
                                                variant="outline"
                                                class="text-xs"
                                                style="background-color: {tag.color}20; color: {tag.color}; border-color: {tag.color}"
                                            >
                                                {tag.name}
                                            </Badge>
                                        {/each}
                                        {#if event.tags.length > 3}
                                            <Badge
                                                variant="secondary"
                                                class="text-xs"
                                            >
                                                +{event.tags.length - 3}
                                            </Badge>
                                        {/if}
                                    </div>
                                {:else}
                                    <div class="text-sm text-muted-foreground">
                                        -
                                    </div>
                                {/if}
                            </td>

                            <!-- Date -->
                            <td class="py-2 px-3">
                                {#if event.date}
                                    <div
                                        class="flex items-center gap-1 text-sm text-muted-foreground"
                                    >
                                        <Calendar class="h-3 w-3" />
                                        {formatDate(event.date)}
                                    </div>
                                {:else}
                                    <div class="text-sm text-muted-foreground">
                                        -
                                    </div>
                                {/if}
                            </td>

                            <!-- Votes -->
                            <td class="py-2 px-3">
                                <div class="text-sm text-muted-foreground">
                                    {event.votes}
                                </div>
                            </td>

                            <!-- Actions -->
                            <td class="py-2 px-3">
                                <div
                                    class="flex items-center justify-end gap-1 opacity-0 group-hover:opacity-100 transition-opacity"
                                >
                                    <Button
                                        variant="ghost"
                                        size="icon"
                                        on:click={(e) => {
                                            e.stopPropagation();
                                            handleEdit(event);
                                        }}
                                        class="h-8 w-8"
                                        title="Edit event"
                                    >
                                        <Edit class="h-3 w-3" />
                                    </Button>
                                    <Button
                                        variant="ghost"
                                        size="icon"
                                        on:click={(e) => {
                                            e.stopPropagation();
                                            handleMoveToBacklog(event);
                                        }}
                                        class="h-8 w-8 hover:bg-primary hover:text-primary-foreground"
                                        title="Move to Backlog"
                                    >
                                        <Inbox class="h-3 w-3" />
                                    </Button>
                                    <Button
                                        variant="ghost"
                                        size="icon"
                                        on:click={(e) => {
                                            e.stopPropagation();
                                            handleDelete(event.id);
                                        }}
                                        class="h-8 w-8 hover:bg-destructive hover:text-destructive-foreground"
                                        title="Delete event"
                                    >
                                        <Trash2 class="h-3 w-3" />
                                    </Button>
                                </div>
                            </td>
                        </tr>
                    {/each}
                {/if}
            </tbody>
        </table>
    </div>
</Card>

<style>
    /* Ensure table doesn't break on mobile */
    @media (max-width: 768px) {
        table {
            font-size: 0.875rem;
        }

        th,
        td {
            padding: 0.75rem 0.5rem;
        }
    }
</style>
