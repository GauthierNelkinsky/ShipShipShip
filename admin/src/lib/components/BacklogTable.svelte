<script lang="ts">
    import { createEventDispatcher } from "svelte";
    import type { ParsedEvent } from "$lib/types";
    import { formatDate } from "$lib/utils";
    import { Card, Button, Badge } from "$lib/components/ui";

    import { Trash2, Edit, Calendar, ArrowUp, Archive } from "lucide-svelte";
    import { fly } from "svelte/transition";
    import { flip } from "svelte/animate";
    import { onMount } from "svelte";
    import { quintOut } from "svelte/easing";

    interface StatusDefinition {
        id: number;
        display_name: string;
        order: number;
        is_reserved: boolean;
    }

    const dispatch = createEventDispatcher();

    export let events: ParsedEvent[] = [];
    export let loading = false;
    export let statuses: StatusDefinition[] = [];

    let dropdownOpenIndex: number | null = null;
    let dropdownPosition = { top: 0, right: 0 };

    function toggleDropdown(index: number, e: MouseEvent) {
        e.stopPropagation();

        // Get button position for dropdown placement
        const button = e.currentTarget as HTMLElement;
        const rect = button.getBoundingClientRect();

        // Store button position for dropdown positioning
        dropdownPosition = {
            top: rect.bottom + 5,
            right: window.innerWidth - rect.right,
        };

        // Toggle dropdown
        dropdownOpenIndex = dropdownOpenIndex === index ? null : index;
    }

    function handleClickOutside(event: MouseEvent) {
        const target = event.target as HTMLElement;
        if (!target.closest(".status-dropdown")) {
            dropdownOpenIndex = null;
        }
    }

    // When a click happens outside the dropdown, close it

    onMount(() => {
        const handleDocumentClick = (event: MouseEvent) =>
            handleClickOutside(event);
        document.addEventListener("click", handleDocumentClick);
        return () => document.removeEventListener("click", handleDocumentClick);
    });

    function handleEdit(event: ParsedEvent) {
        dispatch("edit", event);
    }

    function handleDelete(eventId: number) {
        dispatch("delete", eventId);
    }

    function handleStatusChange(event: ParsedEvent, newStatus: string) {
        dispatch("statusChange", { eventId: event.id, newStatus });
        dropdownOpenIndex = null;
    }

    function handleMoveToArchived(event: ParsedEvent) {
        dispatch("statusChange", { eventId: event.id, newStatus: "Archived" });
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
                            No backlog items found
                        </td>
                    </tr>
                {:else}
                    {#each events as event, index (event.id)}
                        <tr
                            class="border-b border-border hover:bg-muted transition-colors group cursor-pointer"
                            style="--hover-opacity: 0.2;"
                            in:fly={{ y: -10, duration: 300, easing: quintOut }}
                            out:fly={{
                                x: -20,
                                duration: 200,
                                easing: quintOut,
                            }}
                            animate:flip={{ duration: 300, easing: quintOut }}
                            on:click={() => handleEdit(event)}
                            on:mouseenter={(e) =>
                                (e.currentTarget.style.backgroundColor =
                                    "hsl(var(--muted) / 0.2)")}
                            on:mouseleave={(e) =>
                                (e.currentTarget.style.backgroundColor = "")}
                        >
                            <td class="py-2 px-3 w-8">
                                <!-- # column -->
                                <div
                                    class="flex items-center justify-center text-muted-foreground"
                                >
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
                                        title="Edit event"
                                    >
                                        <Edit class="h-3 w-3" />
                                    </Button>
                                    <div class="relative status-dropdown">
                                        <Button
                                            variant="ghost"
                                            size="icon"
                                            on:click={(e) =>
                                                toggleDropdown(index, e)}
                                            class="h-8 w-8 hover:bg-accent hover:text-accent-foreground"
                                            title="Move to another status"
                                        >
                                            <ArrowUp class="h-3 w-3" />
                                        </Button>

                                        {#if dropdownOpenIndex === index}
                                            <div
                                                transition:fly={{
                                                    duration: 150,
                                                    y: 5,
                                                }}
                                                class="fixed w-32 rounded-md border border-border bg-popover shadow-md z-[100]"
                                                style="top: {dropdownPosition.top}px; right: {dropdownPosition.right}px;"
                                                role="menu"
                                                aria-orientation="vertical"
                                                tabindex="0"
                                                on:click|stopPropagation
                                                on:keydown={(e) => {
                                                    if (e.key === "Escape")
                                                        closeDropdown();
                                                }}
                                            >
                                                <div class="p-1">
                                                    <div
                                                        class="px-2 py-1 text-[11px] font-medium text-muted-foreground"
                                                    >
                                                        Move to
                                                    </div>
                                                    <div
                                                        class="h-px bg-border mb-1"
                                                    ></div>
                                                    {#each statuses.filter((s) => !s.is_reserved) as status}
                                                        <Button
                                                            variant="ghost"
                                                            on:click={(e) => {
                                                                e.stopPropagation();
                                                                handleStatusChange(
                                                                    event,
                                                                    status.display_name,
                                                                );
                                                            }}
                                                            class="flex items-center w-full px-2 py-1.5 text-xs rounded-sm justify-start h-auto"
                                                            role="menuitem"
                                                        >
                                                            <span
                                                                class="text-xs font-medium max-w-[200px] truncate block"
                                                                title={status.display_name}
                                                            >
                                                                {status.display_name}
                                                            </span>
                                                        </Button>
                                                    {/each}
                                                </div>
                                            </div>
                                        {/if}
                                    </div>
                                    <Button
                                        variant="ghost"
                                        size="icon"
                                        on:click={(e) => {
                                            e.stopPropagation();
                                            handleMoveToArchived(event);
                                        }}
                                        class="h-8 w-8 hover:bg-secondary hover:text-secondary-foreground"
                                        title="Move to Archived"
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

    /* Ensure dropdowns appear above other elements */
    :global(.status-dropdown .fixed) {
        position: fixed !important;
        z-index: 9999 !important;
    }
</style>
