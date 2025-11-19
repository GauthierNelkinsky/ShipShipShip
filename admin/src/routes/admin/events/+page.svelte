<script lang="ts">
    import { onMount, tick } from "svelte";
    import { goto } from "$app/navigation";
    import { api } from "$lib/api";
    import { authStore } from "$lib/stores/auth";
    import { parseEvent, groupEventsByStatus } from "$lib/utils";
    import type { ParsedEvent, EventStatus } from "$lib/types";
    import {
        Plus,
        ArrowLeft,
        ArrowDownAZ,
        ArrowUpZA,
        CalendarArrowUp,
        CalendarArrowDown,
        ClockArrowUp,
        ClockArrowDown,
        Search,
        X,
        ChevronDown,
    } from "lucide-svelte";
    import { Button, Card, Badge, ScrollArea, Input } from "$lib/components/ui";
    import { toast } from "svelte-sonner";
    import EventModal from "$lib/components/EventModal.svelte";
    import PublishModal from "$lib/components/PublishModal.svelte";
    import KanbanCard from "$lib/components/KanbanCard.svelte";
    import BacklogTable from "$lib/components/BacklogTable.svelte";
    import ArchivedTable from "$lib/components/ArchivedTable.svelte";

    import {
        Tabs,
        TabsList,
        TabsTrigger,
        TabsContent,
    } from "$lib/components/ui";
    import { flip } from "svelte/animate";
    import { fly } from "svelte/transition";

    let events: ParsedEvent[] = [];
    let loading = true;
    let error = "";
    let activeTab = "backlogs";

    // Newsletter settings
    let newsletterEnabled = false;

    // Modal state
    let isModalOpen = false;
    let modalMode: "create" | "edit" = "create";
    let editingEvent: ParsedEvent | null = null;

    // Publish modal state
    let isPublishModalOpen = false;
    let publishingEvent: ParsedEvent | null = null;
    let dragOverColumn: string | null = null;
    let draggedEventId: number | null = null;
    let draggedEventStatus: string | null = null;

    // Search and global sort state
    let searchQuery = "";
    let globalSortOption: SortOption = "DateAsc";

    // Sort options for column events
    type SortOption =
        | "DateAsc"
        | "DateDesc"
        | "TitleAsc"
        | "TitleDesc"
        | "UpdatedAsc"
        | "UpdatedDesc";
    type SortState = {
        [status: string]: SortOption;
    };

    // Track sorting state for each column
    let sortState: SortState = {
        Proposed: "DateAsc",
        Upcoming: "DateAsc",
        Release: "DateAsc",
    };

    // Kanban columns (excluding Backlogs and Archived)
    const columns: { status: EventStatus; label: string; color: string }[] = [
        {
            status: "Proposed",
            label: "Proposed",
            color: "bg-purple-50 border-purple-200 dark:bg-purple-900/30 dark:border-purple-700",
        },
        {
            status: "Upcoming",
            label: "Upcoming",
            color: "bg-yellow-50 border-yellow-200 dark:bg-yellow-900/30 dark:border-yellow-700",
        },
        {
            status: "Release",
            label: "Release",
            color: "bg-green-50 border-green-200 dark:bg-green-900/30 dark:border-green-700",
        },
    ];

    // Group events by status (computed lazily to avoid reactive cycle)
    function groupedEvents() {
        return groupEventsByStatus(events);
    }

    // Reactive function that updates when events or search/sort inputs change.
    // Explicitly reference `events` so Svelte tracks it as a dependency.
    $: getEventsForStatus = (status: string): ParsedEvent[] => {
        // Touch events for reactivity
        const all = events;

        // Build grouped structure inline to avoid missing dependency tracking.
        const grouped = groupEventsByStatus(all);

        const key = status.toLowerCase();
        let list: ParsedEvent[] = [];

        switch (key) {
            case "backlogs":
                list = grouped.backlogs || [];
                break;
            case "proposed":
                list = grouped.proposed || [];
                break;
            case "upcoming":
                list = grouped.upcoming || [];
                break;
            case "release":
                list = grouped.release || [];
                break;
            case "archived":
                list = grouped.archived || [];
                break;
            default:
                return [];
        }

        // Always filter by current search query
        list = filterEvents(list, searchQuery);

        // Apply sorting only for kanban columns
        if (
            status === "Proposed" ||
            status === "Upcoming" ||
            status === "Release"
        ) {
            return sortEvents(list, globalSortOption);
        }

        return list;
    };

    // Track filtered counts for all statuses
    $: filteredBacklogCount = filterEvents(
        groupedEvents().backlogs || [],
        searchQuery,
    ).length;
    $: filteredArchivedCount = filterEvents(
        groupedEvents().archived || [],
        searchQuery,
    ).length;
    $: filteredProposedCount = filterEvents(
        groupedEvents().proposed || [],
        searchQuery,
    ).length;
    $: filteredUpcomingCount = filterEvents(
        groupedEvents().upcoming || [],
        searchQuery,
    ).length;
    $: filteredReleaseCount = filterEvents(
        groupedEvents().release || [],
        searchQuery,
    ).length;

    // Track if we have any search results
    $: hasSearchResults =
        !searchQuery.trim() ||
        filteredProposedCount > 0 ||
        filteredUpcomingCount > 0 ||
        filteredReleaseCount > 0 ||
        filteredBacklogCount > 0 ||
        filteredArchivedCount > 0;

    // Function to sort events based on sort option
    function sortEvents(
        events: ParsedEvent[],
        sortOption: SortOption,
    ): ParsedEvent[] {
        const sortedEvents = [...events];

        switch (sortOption) {
            case "DateAsc":
                sortedEvents.sort(
                    (a, b) =>
                        new Date(b.date).getTime() - new Date(a.date).getTime(),
                );
                break;
            case "DateDesc":
                sortedEvents.sort(
                    (a, b) =>
                        new Date(a.date).getTime() - new Date(b.date).getTime(),
                );
                break;
            case "TitleAsc":
                sortedEvents.sort((a, b) => a.title.localeCompare(b.title));
                break;
            case "TitleDesc":
                sortedEvents.sort((a, b) => b.title.localeCompare(a.title));
                break;
            case "UpdatedAsc":
                sortedEvents.sort(
                    (a, b) =>
                        new Date(b.updated_at || b.created_at).getTime() -
                        new Date(a.updated_at || a.created_at).getTime(),
                );
                break;
            case "UpdatedDesc":
                sortedEvents.sort(
                    (a, b) =>
                        new Date(a.updated_at || a.created_at).getTime() -
                        new Date(b.updated_at || b.created_at).getTime(),
                );
                break;
        }

        return sortedEvents;
    }

    // Function to cycle through sort options
    function cycleSortOption(status: string): void {
        const currentSort = sortState[status];
        let newSort: SortOption;

        switch (currentSort) {
            case "DateAsc":
                newSort = "DateDesc";
                break;
            case "DateDesc":
                newSort = "TitleAsc";
                break;
            case "TitleAsc":
                newSort = "TitleDesc";
                break;
            case "TitleDesc":
                newSort = "UpdatedAsc";
                break;
            case "UpdatedAsc":
                newSort = "UpdatedDesc";
                break;
            case "UpdatedDesc":
            default:
                newSort = "DateAsc";
                break;
        }

        sortState[status] = newSort;
        // Force reactivity update
        sortState = { ...sortState };
    }

    // Function to get the appropriate sort icon
    function getSortIcon(sortOption: SortOption): typeof Plus {
        switch (sortOption) {
            case "DateAsc":
                return CalendarArrowUp;
            case "DateDesc":
                return CalendarArrowDown;
            case "TitleAsc":
                return ArrowDownAZ;
            case "TitleDesc":
                return ArrowUpZA;
            case "UpdatedAsc":
                return ClockArrowUp;
            case "UpdatedDesc":
                return ClockArrowDown;
            default:
                return CalendarArrowUp;
        }
    }

    // Removed cycle function as we now use a dropdown
    $: if (globalSortOption) {
        // Force update for tables when sort option changes
        events = [...events];
    }

    // Function to filter events based on search query
    function filterEvents(events: ParsedEvent[], query: string): ParsedEvent[] {
        if (!query.trim()) return events;

        const lowercaseQuery = query.toLowerCase();
        return events.filter(
            (event) =>
                event.title.toLowerCase().includes(lowercaseQuery) ||
                (event.content?.toLowerCase().includes(lowercaseQuery) ??
                    false) ||
                (event.tags?.some((tag) =>
                    tag.name.toLowerCase().includes(lowercaseQuery),
                ) ??
                    false) ||
                (event.date?.toLowerCase().includes(lowercaseQuery) ?? false),
        );
    }

    // Function to get tooltip text for sort options
    function getSortTooltip(sortOption: SortOption): string {
        switch (sortOption) {
            case "DateAsc":
                return "Sorted by date (newest first)";
            case "DateDesc":
                return "Sorted by date (oldest first)";
            case "TitleAsc":
                return "Sorted by title (A-Z)";
            case "TitleDesc":
                return "Sorted by title (Z-A)";
            case "UpdatedAsc":
                return "Sorted by last update (newest first)";
            case "UpdatedDesc":
                return "Sorted by last update (oldest first)";
            default:
                return "Change sort order";
        }
    }

    onMount(async () => {
        // Wait for authentication to be initialized before loading events
        const unsubscribe = authStore.subscribe(async (auth) => {
            if (auth.initialized && auth.isAuthenticated) {
                await loadEvents();
                unsubscribe();
            } else if (auth.initialized && !auth.isAuthenticated) {
                // User is not authenticated, redirect to login
                goto("/admin/login");
                unsubscribe();
            }
        });
    });

    async function loadEvents() {
        try {
            loading = true;
            error = "";
            const data = await api.getAllEvents();
            events = data.map(parseEvent);

            // Load newsletter settings
            await loadNewsletterSettings();
        } catch (err) {
            error =
                err instanceof Error ? err.message : "Failed to load events";
            console.error("Failed to load events:", err);
        } finally {
            loading = false;
        }
    }

    async function loadNewsletterSettings() {
        try {
            const settings = await api.getSettings();
            newsletterEnabled = !!settings?.newsletter_enabled;
        } catch (err) {
            console.error("Failed to load newsletter settings:", err);
            newsletterEnabled = false;
        }
    }

    // Remove manual updateGroupedEvents function since it's now reactive
    // function updateGroupedEvents() {
    //     groupedEvents = groupEventsByStatus(events);
    // }

    // Helper function to safely access grouped events

    // No additional reactive statements needed since groupedEvents is now reactive

    function openCreateModal(status?: EventStatus) {
        modalMode = "create";
        editingEvent = null;
        if (status) {
            // Pre-set the status for the new event
            editingEvent = { status } as ParsedEvent;
        }
        isModalOpen = true;
    }

    function openEditModal(event: ParsedEvent) {
        modalMode = "edit";
        editingEvent = event;
        isModalOpen = true;
    }

    async function handleDelete(eventId: number) {
        try {
            await api.deleteEvent(eventId);
            events = events.filter((e) => e.id !== eventId);
        } catch (err) {
            error =
                err instanceof Error ? err.message : "Failed to delete event";
        }
    }

    function handlePublish(event: ParsedEvent) {
        publishingEvent = event;
        isPublishModalOpen = true;
    }

    async function handleEventCreated(event: CustomEvent) {
        const newEvent = parseEvent(event.detail);
        events = [...events, newEvent];
        await tick();
    }

    async function handleEventUpdated(event: CustomEvent) {
        const updatedEvent = parseEvent(event.detail);

        // Force complete array recreation to ensure reactivity
        events = events.map((e) =>
            e.id === updatedEvent.id ? updatedEvent : e,
        );

        // Force DOM update
        await tick();
    }

    async function handleStatusChange(eventId: number, newStatus: EventStatus) {
        try {
            const event = events.find((e) => e.id === eventId);
            if (!event) {
                return;
            }

            const oldStatus = event.status;

            const updatedEvent = await api.updateEvent(eventId, {
                status: newStatus,
            });

            // Force status to newStatus regardless of what backend returns (prevents UI reverting)
            events = events.map((e) =>
                e.id === eventId
                    ? { ...parseEvent(updatedEvent), status: newStatus }
                    : e,
            );

            // Force DOM update to ensure animations work
            await tick();

            // Show toast notification for successful move only if status actually changed
            if (oldStatus !== newStatus) {
                const statusLabels = {
                    Backlogs: "Backlogs",
                    Proposed: "Proposed",
                    Upcoming: "Upcoming",
                    Release: "Release",
                    Archived: "Archived",
                };

                toast("Event moved successfully!", {
                    description: `"${event.title}" has been moved to ${statusLabels[newStatus]}.`,
                    action: {
                        label: "Share Update",
                        onClick: () => {
                            publishingEvent = event;
                            isPublishModalOpen = true;
                        },
                    },
                });
            }
        } catch (err) {
            console.error("Error updating event status:", err);
            error =
                err instanceof Error
                    ? err.message
                    : "Failed to update event status";
        }
    }

    function handleDrop(e: DragEvent, newStatus: EventStatus) {
        e.preventDefault();
        e.stopPropagation();
        dragOverColumn = null;

        if (!draggedEventId) return;

        const sourceIndex = events.findIndex((ev) => ev.id === draggedEventId);
        if (sourceIndex === -1) return;

        const sourceEvent = events[sourceIndex];
        if (sourceEvent.status === newStatus) return;

        // Optimistic update
        const prevStatus = sourceEvent.status;
        const prevEventsSnapshot = events;
        events = events.map((ev) =>
            ev.id === draggedEventId ? { ...ev, status: newStatus } : ev,
        );

        // Persist change (will remap with fresh data); revert if fails
        handleStatusChange(draggedEventId, newStatus).catch((err) => {
            console.error("Failed to persist status change, reverting:", err);
            events = prevEventsSnapshot.map((ev) =>
                ev.id === draggedEventId ? { ...ev, status: prevStatus } : ev,
            );
        });
        // Drag data cleared later by dragend
    }

    function getEventStatus(eventId: number): string | null {
        const event = events.find((e) => e.id === eventId);
        return event ? event.status : null;
    }

    // Move up/down functionality removed

    function handleDragOver(e: DragEvent, columnStatus?: string) {
        e.preventDefault();
        e.dataTransfer!.dropEffect = "move";
        if (columnStatus) {
            dragOverColumn = columnStatus;
        }
    }

    function handleDragEnter(e: DragEvent) {
        e.preventDefault();
    }

    function handleDragLeave() {
        dragOverColumn = null;
    }

    function handleBacklogDrop(e: DragEvent) {
        e.preventDefault();
        const eventId = parseInt(e.dataTransfer?.getData("text/plain") || "0");

        try {
            const dragData = JSON.parse(
                e.dataTransfer?.getData("application/json") || "{}",
            );

            if (eventId && dragData.sourceType === "backlog") {
                // Handle reordering within backlog only
                const sourceIndex = dragData.sourceIndex;
                // This will be handled by BacklogTable's own reorder logic
                return;
            }
            // No longer allow kanban to backlog drops
        } catch (err) {
            // No fallback for failed parsing
        }
    }

    function handleBacklogDragOver(e: DragEvent) {
        e.preventDefault();
    }

    // Reordering functionality removed
</script>

<svelte:head>
    <title>Manage Events - Admin</title>
</svelte:head>

<div class="w-full">
    <!-- Header with integrated search and sort -->
    <div
        class="flex flex-col md:flex-row items-start md:items-center justify-between gap-3 mb-6 sticky top-0 z-10 bg-background py-2"
    >
        <div>
            <h1 class="text-xl font-semibold mb-1">Events</h1>
            <p class="text-muted-foreground text-sm">
                Search, sort, and organize your events
            </p>
        </div>

        <div class="flex items-center gap-2 w-full md:w-auto">
            <!-- Search bar -->
            <div class="relative flex-1 min-w-[220px]">
                <Input
                    type="text"
                    placeholder="Search events..."
                    bind:value={searchQuery}
                    class="h-8 text-sm"
                />
                <button
                    class="absolute right-2 top-1/2 transform -translate-y-1/2 text-muted-foreground"
                    on:click={() => (searchQuery = "")}
                    title={searchQuery ? "Clear search" : "Search"}
                >
                    {#if searchQuery}
                        <X class="h-4 w-4" />
                    {:else}
                        <Search class="h-4 w-4" />
                    {/if}
                </button>
            </div>

            <!-- Sort dropdown -->
            <div class="relative w-8 h-8">
                <select
                    bind:value={globalSortOption}
                    class="appearance-none absolute inset-0 opacity-0 w-full h-full cursor-pointer z-10"
                    title={getSortTooltip(globalSortOption)}
                >
                    <option value="DateAsc">Date (newest first)</option>
                    <option value="DateDesc">Date (oldest first)</option>
                    <option value="TitleAsc">Title (A-Z)</option>
                    <option value="TitleDesc">Title (Z-A)</option>
                    <option value="UpdatedAsc">Updated (newest first)</option>
                    <option value="UpdatedDesc">Updated (oldest first)</option>
                </select>
                <div
                    class="flex items-center justify-center w-full h-full bg-background border rounded-md hover:bg-muted"
                >
                    <svelte:component
                        this={getSortIcon(globalSortOption)}
                        class="h-4 w-4"
                    />
                </div>
            </div>
        </div>
    </div>

    <main class="w-full">
        {#if error}
            <div
                class="bg-destructive border border-destructive text-destructive px-3 py-2 rounded text-sm mb-3"
                style="background-color: hsl(var(--destructive) / 0.1); border-color: hsl(var(--destructive) / 0.2);"
            >
                {error}
            </div>
        {/if}

        <!-- Search -->

        {#if loading}
            <div class="flex items-center justify-center min-h-32">
                <div
                    class="animate-spin rounded-full h-6 w-6 border-b-2 border-primary"
                ></div>
            </div>
        {:else if searchQuery && !hasSearchResults}
            <div
                class="flex flex-col items-center justify-center min-h-32 gap-2"
            >
                <div class="text-muted-foreground text-lg">
                    No events found matching "{searchQuery}"
                </div>
                <Button variant="outline" on:click={() => (searchQuery = "")}
                    >Clear Search</Button
                >
            </div>
        {:else}
            <!-- Kanban Board -->
            <div class="w-full">
                <div class="flex gap-2 lg:gap-4 min-h-0 pb-3">
                    {#each columns as column (column.status)}
                        <div class="flex-1 min-w-0 max-w-sm">
                            <!-- Column Header -->
                            <div class="mb-3">
                                <div class="flex items-center justify-between">
                                    <h2
                                        class="font-medium text-sm text-foreground"
                                    >
                                        {column.label}
                                    </h2>
                                    <div class="flex items-center gap-2">
                                        <Button
                                            variant="outline"
                                            size="sm"
                                            on:click={() =>
                                                openCreateModal(column.status)}
                                            class="h-6 px-2 text-xs"
                                        >
                                            <Plus class="h-3 w-3 mr-1" />
                                            Add
                                        </Button>

                                        <span
                                            class="text-xs text-muted-foreground bg-muted rounded px-1.5 py-0.5"
                                        >
                                            {column.status === "Proposed"
                                                ? filteredProposedCount
                                                : column.status === "Upcoming"
                                                  ? filteredUpcomingCount
                                                  : filteredReleaseCount}
                                        </span>
                                    </div>
                                </div>
                            </div>

                            <!-- Column Content -->
                            <div
                                class="h-[550px] rounded-lg border-2 border-dashed transition-colors {column.color} {dragOverColumn ===
                                column.status
                                    ? 'ring-2 ring-primary border-primary'
                                    : ''} overflow-hidden"
                                on:drop={(e) => handleDrop(e, column.status)}
                                on:dragover={(e) =>
                                    handleDragOver(e, column.status)}
                                on:dragenter={handleDragEnter}
                                on:dragleave={handleDragLeave}
                                role="region"
                                aria-label="Drop zone for {column.label} events"
                            >
                                <div class="h-full overflow-y-auto">
                                    <div class="space-y-2 p-3 min-w-0">
                                        <!-- Drop zone at top of column -->
                                        <div
                                            class="h-2 transition-all duration-200 {draggedEventId &&
                                            draggedEventStatus ===
                                                column.status &&
                                            getEventsForStatus(column.status)
                                                .length > 0
                                                ? 'bg-primary/20 rounded border border-primary border-dashed'
                                                : ''}"
                                            role="region"
                                            aria-label="Column drop zone"
                                            on:dragover={(e) => {
                                                e.preventDefault();
                                                if (
                                                    draggedEventId &&
                                                    draggedEventStatus ===
                                                        column.status &&
                                                    e.dataTransfer
                                                ) {
                                                    e.dataTransfer.dropEffect =
                                                        "move";
                                                }
                                            }}
                                            on:drop={(e) => {
                                                e.preventDefault();
                                                if (
                                                    draggedEventId &&
                                                    draggedEventStatus !==
                                                        column.status
                                                ) {
                                                    handleDrop(
                                                        e,
                                                        column.status,
                                                    );
                                                }
                                            }}
                                        ></div>

                                        {#each getEventsForStatus(column.status) as event, index (event.id)}
                                            <div
                                                class="group relative"
                                                in:fly={{
                                                    y: 20,
                                                    duration: 200,
                                                }}
                                            >
                                                <KanbanCard
                                                    {event}
                                                    on:edit={(e) =>
                                                        openEditModal(e.detail)}
                                                    on:delete={(e) =>
                                                        handleDelete(e.detail)}
                                                    on:publish={(e) =>
                                                        handlePublish(e.detail)}
                                                    on:statusChange={(e) =>
                                                        handleStatusChange(
                                                            e.detail.eventId,
                                                            e.detail.newStatus,
                                                        )}
                                                    on:carddragstart={(e) => {
                                                        draggedEventId =
                                                            e.detail.eventId;
                                                        draggedEventStatus =
                                                            e.detail
                                                                .sourceStatus;
                                                    }}
                                                    isBeingDragged={draggedEventId ===
                                                        event.id}
                                                    on:carddragend={() => {
                                                        // Delay clearing drag data to prevent race condition with drop event
                                                        setTimeout(() => {
                                                            draggedEventId =
                                                                null;
                                                            draggedEventStatus =
                                                                null;
                                                        }, 100);
                                                    }}
                                                />
                                            </div>
                                        {/each}

                                        {#if getEventsForStatus(column.status).length === 0}
                                            <div
                                                class="text-center py-6 text-muted-foreground text-xs"
                                            >
                                                No events
                                            </div>
                                        {/if}
                                    </div>
                                </div>
                            </div>
                        </div>
                    {/each}
                </div>
            </div>

            <!-- Backlogs and Archived Events Tabs -->
            <div class="mt-8">
                <Tabs bind:value={activeTab} className="w-full">
                    <div class="flex items-center justify-between mb-6">
                        <div
                            class="inline-flex h-10 items-center justify-center rounded-md bg-muted p-1 text-muted-foreground"
                            use:TabsList
                        >
                            <button
                                class="inline-flex items-center justify-center whitespace-nowrap rounded-sm px-3 py-1.5 text-sm font-medium ring-offset-background transition-all focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50"
                                use:TabsTrigger={{
                                    value: "backlogs",
                                    activeValue: activeTab,
                                }}
                                on:click={() => (activeTab = "backlogs")}
                            >
                                Backlogs
                                <span
                                    class="ml-2 text-xs text-muted-foreground bg-background rounded px-1.5 py-0.5"
                                >
                                    {filteredBacklogCount}
                                </span>
                            </button>
                            <button
                                class="inline-flex items-center justify-center whitespace-nowrap rounded-sm px-3 py-1.5 text-sm font-medium ring-offset-background transition-all focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50"
                                use:TabsTrigger={{
                                    value: "archived",
                                    activeValue: activeTab,
                                }}
                                on:click={() => (activeTab = "archived")}
                            >
                                Archived
                                <span
                                    class="ml-2 text-xs text-muted-foreground bg-background rounded px-1.5 py-0.5"
                                >
                                    {filteredArchivedCount}
                                </span>
                            </button>
                        </div>

                        <div class="flex items-center gap-2">
                            <Button
                                variant="outline"
                                size="sm"
                                on:click={() =>
                                    openCreateModal(
                                        activeTab === "backlogs"
                                            ? "Backlogs"
                                            : "Archived",
                                    )}
                                class="flex items-center gap-1.5 text-xs py-1.5 px-3"
                            >
                                <Plus class="h-3 w-3" />
                                Add {activeTab === "backlogs"
                                    ? "Backlog"
                                    : "Archived"}
                            </Button>
                        </div>
                    </div>

                    <div
                        class="mt-2 ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
                        use:TabsContent={{
                            value: "backlogs",
                            activeValue: activeTab,
                        }}
                    >
                        <div class="mb-3">
                            <p class="text-muted-foreground text-xs">
                                Events waiting to be scheduled
                            </p>
                        </div>
                        <Card
                            class="border-2 border-dashed bg-gray-50 border-gray-200 dark:bg-gray-800 dark:border-gray-700 p-3"
                            role="region"
                            aria-label="Backlogs drop zone"
                        >
                            <ScrollArea class="max-h-96">
                                <div
                                    role="region"
                                    aria-label="Backlog drop zone"
                                    on:drop={handleBacklogDrop}
                                    on:dragover={handleBacklogDragOver}
                                >
                                    <BacklogTable
                                        events={sortEvents(
                                            filterEvents(
                                                groupedEvents().backlogs || [],
                                                searchQuery,
                                            ),
                                            globalSortOption,
                                        )}
                                        {loading}
                                        on:edit={(e) => openEditModal(e.detail)}
                                        on:delete={(e) =>
                                            handleDelete(e.detail)}
                                        on:statusChange={(e) =>
                                            handleStatusChange(
                                                e.detail.eventId,
                                                e.detail.newStatus,
                                            )}
                                    />
                                </div>
                            </ScrollArea>
                        </Card>
                    </div>

                    <div
                        class="mt-2 ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
                        use:TabsContent={{
                            value: "archived",
                            activeValue: activeTab,
                        }}
                    >
                        <div class="mb-3">
                            <p class="text-muted-foreground text-xs">
                                Events archived from public view
                            </p>
                        </div>
                        <Card
                            class="border-2 border-dashed bg-red-50 border-red-200 dark:bg-red-900 dark:border-red-800 p-3"
                            style="background-color: rgb(239 68 68 / 0.1);"
                            role="region"
                            aria-label="Archived events"
                        >
                            <ScrollArea class="max-h-96">
                                <ArchivedTable
                                    events={sortEvents(
                                        filterEvents(
                                            groupedEvents().archived || [],
                                            searchQuery,
                                        ),
                                        globalSortOption,
                                    )}
                                    {loading}
                                    on:edit={(e) => openEditModal(e.detail)}
                                    on:delete={(e) => handleDelete(e.detail)}
                                    on:statusChange={(e) =>
                                        handleStatusChange(
                                            e.detail.eventId,
                                            e.detail.newStatus,
                                        )}
                                />
                            </ScrollArea>
                        </Card>
                    </div>
                </Tabs>
            </div>

            <!-- Summary -->
            <div class="mt-6 text-center text-xs text-muted-foreground">
                {events.length} total
            </div>
        {/if}
    </main>
</div>

<!-- Event Modal -->
<EventModal
    bind:isOpen={isModalOpen}
    mode={modalMode}
    event={editingEvent}
    on:created={handleEventCreated}
    on:updated={async (e) => {
        await handleEventUpdated(e);
        isModalOpen = false;
        editingEvent = null;
    }}
    on:publish={(e) => {
        publishingEvent = e.detail;
        isPublishModalOpen = true;
    }}
    on:close={() => {
        isModalOpen = false;
        editingEvent = null;
    }}
/>

<!-- Publish Modal -->
<PublishModal
    bind:isOpen={isPublishModalOpen}
    event={publishingEvent}
    {newsletterEnabled}
    on:close={() => {
        isPublishModalOpen = false;
        publishingEvent = null;
    }}
/>
