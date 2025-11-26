<script lang="ts">
    import { createEventDispatcher } from "svelte";
    import type { ParsedEvent } from "$lib/types";
    import { formatDate } from "$lib/utils";
    import { Card, Button, Badge } from "$lib/components/ui";
    import * as m from "$lib/paraglide/messages";

    import {
        Trash2,
        Edit,
        Calendar,
        Archive,
        GripVertical,
    } from "lucide-svelte";
    import { fly } from "svelte/transition";
    import { flip } from "svelte/animate";
    import { quintOut } from "svelte/easing";

    const dispatch = createEventDispatcher();

    export let events: ParsedEvent[] = [];
    export let loading = false;

    let draggingEventId: number | null = null;

    function handleEdit(event: ParsedEvent) {
        dispatch("edit", event);
    }

    function handleDelete(eventId: number) {
        dispatch("delete", eventId);
    }

    function handleMoveToArchived(event: ParsedEvent) {
        dispatch("statusChange", { eventId: event.id, newStatus: "Archived" });
    }

    function handleDragStart(e: DragEvent, event: ParsedEvent) {
        if (e.dataTransfer) {
            e.dataTransfer.effectAllowed = "move";
            e.dataTransfer.setData("text/plain", event.id.toString());
            e.dataTransfer.setData(
                "application/json",
                JSON.stringify({
                    eventId: event.id,
                    sourceStatus: event.status,
                    fromBacklog: true,
                }),
            );
        }
        draggingEventId = event.id;

        // Dispatch event for parent to track drag state
        dispatch("carddragstart", {
            eventId: event.id,
            sourceStatus: event.status,
        });
    }

    function handleDragEnd() {
        // Dispatch event for parent to clear drag state
        dispatch("carddragend");

        setTimeout(() => {
            draggingEventId = null;
        }, 100);
    }

    // Drag and drop reordering functionality removed
</script>

<Card class="overflow-hidden">
    <div class="overflow-x-auto max-h-[400px] overflow-y-auto pb-4">
        <table class="w-full">
            <thead class="border-b border-border">
                <tr class="bg-muted" style="opacity: 0.5;">
                    <th
                        class="py-2 px-3 w-8 text-center text-xs font-medium text-muted-foreground"
                        >#</th
                    >
                    <th
                        class="text-left py-2 px-3 font-medium text-sm text-muted-foreground"
                        >{m.backlog_table_name()}</th
                    >
                    <th
                        class="text-left py-2 px-3 font-medium text-sm text-muted-foreground"
                        >{m.backlog_table_tags()}</th
                    >
                    <th
                        class="text-left py-2 px-3 font-medium text-sm text-muted-foreground"
                        >{m.backlog_table_date()}</th
                    >
                    <th
                        class="text-left py-2 px-3 font-medium text-sm text-muted-foreground"
                        >{m.backlog_table_votes()}</th
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
                                {m.backlog_table_loading()}
                            </div>
                        </td>
                    </tr>
                {:else if events.length === 0}
                    <tr>
                        <td
                            colspan="6"
                            class="p-8 text-center text-muted-foreground"
                        >
                            {m.backlog_table_no_items()}
                        </td>
                    </tr>
                {:else}
                    {#each events as event, index (event.id)}
                        <tr
                            class="border-b border-border hover:bg-muted transition-colors group cursor-pointer {draggingEventId ===
                            event.id
                                ? 'opacity-50'
                                : ''}"
                            style="--hover-opacity: 0.2;"
                            draggable="true"
                            in:fly={{ y: -10, duration: 300, easing: quintOut }}
                            out:fly={{
                                x: -20,
                                duration: 200,
                                easing: quintOut,
                            }}
                            animate:flip={{ duration: 300, easing: quintOut }}
                            on:dragstart={(e) => handleDragStart(e, event)}
                            on:dragend={handleDragEnd}
                            on:click={() => handleEdit(event)}
                            on:mouseenter={(e) =>
                                (e.currentTarget.style.backgroundColor =
                                    "hsl(var(--muted) / 0.2)")}
                            on:mouseleave={(e) =>
                                (e.currentTarget.style.backgroundColor = "")}
                        >
                            <td class="py-2 px-3 w-8">
                                <!-- # column with drag handle -->
                                <div
                                    class="flex items-center justify-center text-muted-foreground gap-1"
                                >
                                    <GripVertical
                                        class="h-3 w-3 opacity-0 group-hover:opacity-100 transition-opacity"
                                    />
                                    <span class="text-xs">{index + 1}</span>
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
                                        title={m.backlog_table_edit()}
                                    >
                                        <Edit class="h-3 w-3" />
                                    </Button>
                                    <Button
                                        variant="ghost"
                                        size="icon"
                                        on:click={(e) => {
                                            e.stopPropagation();
                                            handleMoveToArchived(event);
                                        }}
                                        class="h-8 w-8 hover:bg-secondary hover:text-secondary-foreground"
                                        title={m.backlog_table_move_to_archived()}
                                    >
                                        <Archive class="h-3 w-3" />
                                    </Button>
                                    <Button
                                        variant="ghost"
                                        size="icon"
                                        on:click={(e) => {
                                            e.stopPropagation();
                                            handleDelete(event.id);
                                        }}
                                        class="h-8 w-8 hover:bg-destructive hover:text-destructive-foreground"
                                        title={m.backlog_table_delete()}
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

    /* Make header sticky while scrolling */
    thead {
        position: sticky;
        top: 0;
        z-index: 10;
        background-color: hsl(var(--background));
    }

    thead tr {
        box-shadow: 0 1px 0 0 hsl(var(--border));
    }

    /* Draggable row styling */
    tr[draggable="true"]:hover {
        background-color: hsl(var(--muted) / 0.3);
    }
</style>
