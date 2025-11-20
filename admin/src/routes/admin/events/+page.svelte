<script lang="ts">
    import { onMount, tick } from "svelte";
    import { goto } from "$app/navigation";
    import { api } from "$lib/api";
    import { authStore } from "$lib/stores/auth";
    import { parseEvent, groupEventsByStatus } from "$lib/utils";
    import type { ParsedEvent, EventStatus } from "$lib/types";
    import {
        Plus,
        ArrowDownAZ,
        ArrowUpZA,
        CalendarArrowUp,
        CalendarArrowDown,
        ClockArrowUp,
        ClockArrowDown,
        Search,
        X,
        ChevronDown,
        Trash,
        Calendar,
        Tag,
        Pencil,
        Check,
        Settings,
        Workflow,
    } from "lucide-svelte";
    import { Button, Card, ScrollArea, Input } from "$lib/components/ui";
    import { toast } from "svelte-sonner";
    import EventModal from "$lib/components/EventModal.svelte";
    import PublishModal from "$lib/components/PublishModal.svelte";
    import StatusModal from "$lib/components/StatusModal.svelte";
    import StatusMappingModal from "$lib/components/StatusMappingModal.svelte";
    import KanbanCard from "$lib/components/KanbanCard.svelte";
    import BacklogTable from "$lib/components/BacklogTable.svelte";
    import ArchivedTable from "$lib/components/ArchivedTable.svelte";

    import {
        Tabs,
        TabsList,
        TabsTrigger,
        TabsContent,
    } from "$lib/components/ui";
    import { fly } from "svelte/transition";

    let events: ParsedEvent[] = [];
    let loading = true;
    let error = "";
    let activeTab = "backlogs";
    let showGlobalNew = false;

    // Newsletter settings
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    let newsletterEnabled = false;

    // Modal state
    let isModalOpen = false;
    let modalMode: "create" | "edit" = "create";
    let editingEvent: ParsedEvent | null = null;

    // Publish modal state
    let isPublishModalOpen = false;
    let publishingEvent: ParsedEvent | null = null;

    // Status modal state
    let isStatusModalOpen = false;
    let isStatusMappingModalOpen = false;
    let unmappedStatusesCount = 0;
    let dragOverColumn: string | null = null;
    let draggedEventId: number | null = null;
    let draggedEventStatus: string | null = null;

    // Store status to category mappings
    let statusCategoryMap: Map<
        string,
        {
            categoryId: string | null;
            categoryLabel: string;
            categoryDescription: string;
        }
    > = new Map();

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
    // Track sorting state for dynamic columns; will be filled when statuses load
    let sortState: SortState = {};

    // Dynamic status definitions (fetched from backend)
    type StatusDefinition = {
        id: number;
        display_name: string;
        order: number;
        is_reserved: boolean;
    };

    let statuses: StatusDefinition[] = [];
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    let statusLoading = false;
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    let statusError = "";

    // Delete confirmation modal state (for status)
    let showDeleteModal = false;
    let pendingDeleteStatus: StatusDefinition | null = null;
    let pendingDeleteEventsCount = 0;
    let deleting = false;

    // Delete confirmation modal state (for event)
    let showDeleteEventModal = false;
    let pendingDeleteEvent: ParsedEvent | null = null;
    let deletingEvent = false;

    // Inline status title editing
    let editingStatusId: number | null = null;
    let editingStatusName: string = "";
    let editingInputEl: HTMLInputElement | null = null;

    function initiateDeleteStatus(def: StatusDefinition) {
        if (def.is_reserved) return;
        pendingDeleteStatus = def;
        pendingDeleteEventsCount = events.filter(
            (e) => e.status === def.display_name,
        ).length;
        showDeleteModal = true;
    }

    function cancelDeleteStatus() {
        if (deleting) return;
        showDeleteModal = false;
        pendingDeleteStatus = null;
        pendingDeleteEventsCount = 0;
    }

    async function confirmDeleteStatus() {
        if (!pendingDeleteStatus || deleting) return;
        deleting = true;
        const targetName = pendingDeleteStatus.display_name;

        // Move affected events to Backlogs first (backend requires status not in use to delete)
        const affected = events.filter((e) => e.status === targetName);
        try {
            await Promise.all(
                affected.map((ev) =>
                    api.updateEvent(ev.id, { status: "Backlogs" }),
                ),
            );
            // Optimistically update local events
            events = events.map((e) =>
                e.status === targetName ? { ...e, status: "Backlogs" } : e,
            );

            // Now delete the status definition
            await api.deleteStatus(pendingDeleteStatus.id);
            statuses = statuses.filter((s) => s.id !== pendingDeleteStatus!.id);
            rebuildColumns();

            // Force a fresh reload of all events so backlog table reflects moved items
            try {
                const refreshed = await api.getAllEvents();
                events = refreshed.map(parseEvent);
            } catch {
                // swallow refresh errors; optimistic state already updated
            }

            toast("Status deleted", {
                description: `${targetName} removed, ${affected.length} ${
                    affected.length === 1 ? "event" : "events"
                } moved to Backlogs`,
            });
        } catch (e) {
            toast("Delete failed", {
                description: e instanceof Error ? e.message : "Unknown error",
            });
        } finally {
            deleting = false;
            showDeleteModal = false;
            pendingDeleteStatus = null;
            pendingDeleteEventsCount = 0;
        }
    }

    async function startEditingStatus(displayName: string) {
        const def = statuses.find((s) => s.display_name === displayName);
        if (!def) return;
        editingStatusId = def.id;
        editingStatusName = def.display_name;
        // Wait for DOM to update then focus/select input
        await tick();
        editingInputEl?.focus();
        editingInputEl?.select();
    }

    async function commitEditingStatus() {
        if (editingStatusId === null) return;
        const def = statuses.find((s) => s.id === editingStatusId);
        if (!def) {
            editingStatusId = null;
            return;
        }
        const newName = editingStatusName.trim();

        // Check if anything changed
        if (!newName || newName === def.display_name) {
            editingStatusId = null;
            return;
        }

        try {
            await api.updateStatus(def.id, {
                display_name: newName,
            });
            // Update local statuses (reload or mutate)
            statuses = statuses.map((s) =>
                s.id === def.id ? { ...s, display_name: newName } : s,
            );
            // Update events that had old status name
            events = events.map((e) =>
                e.status === def.display_name ? { ...e, status: newName } : e,
            );
            // Rebuild columns with new names
            rebuildColumns();
            toast("Status updated", {
                description: `${def.display_name} updated successfully`,
            });
        } catch (e) {
            toast("Update failed", {
                description: e instanceof Error ? e.message : "Unknown error",
            });
        } finally {
            editingStatusId = null;
        }
    }

    function cancelEditingStatus() {
        editingStatusId = null;
        editingStatusName = "";
    }

    // Status management UI removed

    // Kanban columns derived from non-reserved statuses
    let columns: { status: EventStatus; label: string }[] = [];

    function rebuildColumns() {
        columns = statuses
            .filter((s) => !s.is_reserved)
            .sort(
                (a, b) =>
                    a.order - b.order ||
                    a.display_name.localeCompare(b.display_name),
            )
            .map((s) => {
                return {
                    status: s.display_name as EventStatus,
                    label: s.display_name,
                };
            });

        // Initialize sort state for any new columns
        for (const c of columns) {
            if (!sortState[c.status]) {
                sortState[c.status] = "DateAsc";
            }
        }
    }

    async function loadStatuses() {
        try {
            statusLoading = true;
            statusError = "";
            const data = await api.getStatuses();
            statuses = data;
            rebuildColumns();
        } catch (e) {
            statusError =
                e instanceof Error ? e.message : "Failed to load statuses";
        } finally {
            statusLoading = false;
        }
    }

    // Status CRUD functions removed

    function countForStatus(status: string): number {
        return filterEvents(
            events.filter((e) => e.status === status),
            searchQuery,
        ).length;
    }

    // (Deprecated static columns replaced by dynamic status definitions)
    // columns now rebuilt from statuses in rebuildColumns()

    // Group events by status - reactive variable that updates when events change
    $: groupedEvents = groupEventsByStatus(events);

    // Reactive function that updates when events or search/sort inputs change.
    // For reserved statuses (Backlogs / Archived) use grouped buckets.
    // For all dynamic statuses, filter directly so renamed statuses continue to show.
    $: getEventsForStatus = (status: string): ParsedEvent[] => {
        const key = status.toLowerCase();

        if (key === "backlogs" || key === "archived") {
            const bucket = groupedEvents[key as "backlogs" | "archived"] || [];
            return filterEvents(bucket, searchQuery);
        }

        // Dynamic status column: filter by exact status string
        let list = events.filter((e) => e.status === status);
        list = filterEvents(list, searchQuery);
        return sortEvents(list, globalSortOption);
    };

    // Track filtered counts for all statuses
    // Reference `events` directly so Svelte tracks reactivity when events load/update
    $: filteredBacklogCount = filterEvents(
        groupEventsByStatus(events).backlogs || [],
        searchQuery,
    ).length;
    $: filteredArchivedCount = filterEvents(
        groupEventsByStatus(events).archived || [],
        searchQuery,
    ).length;
    $: filteredProposedCount = filterEvents(
        groupedEvents.proposed || [],
        searchQuery,
    ).length;
    $: filteredUpcomingCount = filterEvents(
        groupedEvents.upcoming || [],
        searchQuery,
    ).length;
    $: filteredReleaseCount = filterEvents(
        groupedEvents.release || [],
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
    function _cycleSortOption(status: string): void {
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
        // Wait for authentication to be initialized before loading events & statuses
        const unsubscribe = authStore.subscribe(async (auth) => {
            if (auth.initialized && auth.isAuthenticated) {
                await Promise.all([
                    loadEvents(),
                    loadStatuses(),
                    checkUnmappedStatuses(),
                ]);
                unsubscribe();
            } else if (auth.initialized && !auth.isAuthenticated) {
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

    async function checkUnmappedStatuses() {
        try {
            const manifestData = await api.getThemeManifest();
            const mappingsData = await api.getStatusMappings();

            if (!manifestData.manifest) {
                unmappedStatusesCount = 0;
                return;
            }

            // Build status to category mapping
            statusCategoryMap.clear();
            if (mappingsData.mappings) {
                mappingsData.mappings.forEach((m: any) => {
                    // Find the category in the manifest to get description
                    const category = manifestData.manifest.categories.find(
                        (c: any) => c.id === m.category_id,
                    );
                    statusCategoryMap.set(m.status_name, {
                        categoryId: m.category_id,
                        categoryLabel: m.category_label,
                        categoryDescription: category?.description || "",
                    });
                });
            }
            // Add unmapped statuses
            if (mappingsData.unmapped_statuses) {
                mappingsData.unmapped_statuses.forEach((u: any) => {
                    statusCategoryMap.set(u.status_name, {
                        categoryId: null,
                        categoryLabel: "Unmapped",
                        categoryDescription: "",
                    });
                });
            }
            statusCategoryMap = new Map(statusCategoryMap); // Trigger reactivity

            // Count categories that have no statuses mapped to them
            const mappedCategoryIds = new Set(
                mappingsData.mappings?.map((m: any) => m.category_id) || [],
            );

            const emptyCategoriesCount =
                manifestData.manifest.categories.filter(
                    (cat: any) => !mappedCategoryIds.has(cat.id),
                ).length;

            unmappedStatusesCount = emptyCategoriesCount;
        } catch (err) {
            console.error("Failed to check unmapped statuses:", err);
            unmappedStatusesCount = 0;
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

    function initiateDeleteEvent(eventId: number) {
        const event = events.find((e) => e.id === eventId);
        if (!event) return;
        pendingDeleteEvent = event;
        showDeleteEventModal = true;
    }

    function cancelDeleteEvent() {
        showDeleteEventModal = false;
        pendingDeleteEvent = null;
    }

    async function confirmDeleteEvent() {
        if (!pendingDeleteEvent) return;

        const eventId = pendingDeleteEvent.id;

        try {
            deletingEvent = true;
            await api.deleteEvent(eventId);
            // Force array recreation for reactivity
            events = events.filter((e) => e.id !== eventId);
            await tick();
            toast.success("Event deleted successfully");
        } catch (err) {
            const errorMessage =
                err instanceof Error ? err.message : "Failed to delete event";
            error = errorMessage;
            toast.error(errorMessage);
        } finally {
            deletingEvent = false;
            showDeleteEventModal = false;
            pendingDeleteEvent = null;
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
                const label = newStatus;

                toast("Event moved successfully!", {
                    description: `"${event.title}" has been moved to ${label}.`,
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

    function _getEventStatus(eventId: number): string | null {
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
                const _sourceIndex = dragData.sourceIndex;
                // This will be handled by BacklogTable's own reorder logic
                return;
            }
            // No longer allow kanban to backlog drops
        } catch {
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
    <!-- Page Title -->
    <div class="mb-8">
        <h1 class="text-xl font-semibold mb-1">Events</h1>
        <p class="text-muted-foreground text-sm">
            Search, sort, and organize your events
        </p>
    </div>

    <!-- Controls Row -->
    <div class="flex items-center justify-between mb-4 pb-4 border-b">
        <!-- Search bar and Sort button -->
        <div class="flex items-center gap-2">
            <div class="relative w-[360px]">
                <Input
                    type="text"
                    placeholder="Search events..."
                    bind:value={searchQuery}
                    class="h-8 text-sm pr-8"
                />
                <button
                    class="absolute right-2 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
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
            <div class="relative w-8 h-8 flex-shrink-0">
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
                    class="flex items-center justify-center w-full h-full bg-background border rounded-md hover:bg-muted cursor-pointer"
                >
                    <svelte:component
                        this={getSortIcon(globalSortOption)}
                        class="h-4 w-4"
                    />
                </div>
            </div>
        </div>

        <!-- Settings and New buttons -->
        <div class="flex items-center gap-2">
            <!-- Settings button -->
            <div class="relative">
                <button
                    type="button"
                    class="h-8 w-8 border rounded-md bg-background hover:bg-muted flex items-center justify-center"
                    on:click={() => (isStatusMappingModalOpen = true)}
                    title="Configure status to category mappings for the public page"
                >
                    <Settings class="h-4 w-4" />
                </button>
                {#if unmappedStatusesCount > 0}
                    <span
                        class="absolute -top-1 -right-1"
                        title="{unmappedStatusesCount} unmapped status{unmappedStatusesCount >
                        1
                            ? 'es'
                            : ''}"
                    >
                        <span class="relative flex h-2.5 w-2.5">
                            <span
                                class="absolute inline-flex h-full w-full animate-ping rounded-full bg-amber-400 opacity-75"
                            ></span>
                            <span
                                class="relative inline-flex h-2.5 w-2.5 rounded-full bg-amber-500"
                            ></span>
                        </span>
                    </span>
                {/if}
            </div>

            <!-- New button -->
            <div class="relative inline-block">
                <button
                    type="button"
                    class="h-8 px-3 text-xs border rounded-md bg-background hover:bg-muted flex items-center gap-1"
                    on:click={() => (showGlobalNew = !showGlobalNew)}
                    aria-haspopup="true"
                    aria-expanded={showGlobalNew}
                >
                    <Plus class="h-3 w-3" />
                    <span>New</span>
                    <ChevronDown
                        class="h-3 w-3 transition-transform {showGlobalNew
                            ? 'rotate-180'
                            : ''}"
                    />
                </button>
                {#if showGlobalNew}
                    <div
                        class="absolute right-0 mt-1 w-44 rounded-md border bg-background shadow p-2 text-xs space-y-1 z-30"
                        role="menu"
                    >
                        <button
                            type="button"
                            class="w-full text-left px-2 py-1 rounded hover:bg-muted flex items-center gap-2"
                            on:click={() => {
                                showGlobalNew = false;
                                openCreateModal();
                            }}
                            role="menuitem"
                        >
                            <Calendar class="h-4 w-4" />
                            New Event
                        </button>
                        <button
                            type="button"
                            class="w-full text-left px-2 py-1 rounded hover:bg-muted flex items-center gap-2"
                            on:click={() => {
                                showGlobalNew = false;
                                isStatusModalOpen = true;
                            }}
                            role="menuitem"
                        >
                            <Tag class="h-4 w-4" />
                            New Status
                        </button>
                    </div>
                {/if}
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
            <div class="w-full overflow-x-auto">
                <div class="flex gap-2 lg:gap-4 min-h-0 pb-3 w-max">
                    {#each columns as column (column.status)}
                        <div class="flex-shrink-0 w-[320px] group">
                            <!-- Column Header -->
                            <div class="mb-3">
                                <div class="flex items-center justify-between">
                                    <div class="flex items-center gap-1.5">
                                        {#if editingStatusId !== null && statuses.find((s) => s.id === editingStatusId)?.display_name === column.status}
                                            <!-- Name input -->
                                            <input
                                                class="text-sm font-medium bg-transparent border border-border rounded px-2 py-0.5 focus:outline-none w-32"
                                                bind:this={editingInputEl}
                                                bind:value={editingStatusName}
                                                on:keydown={(e) => {
                                                    if (e.key === "Enter") {
                                                        commitEditingStatus();
                                                    } else if (
                                                        e.key === "Escape"
                                                    ) {
                                                        cancelEditingStatus();
                                                    }
                                                }}
                                                aria-label="Edit status name"
                                            />
                                            <!-- Validate button -->
                                            <button
                                                type="button"
                                                class="h-6 w-6 flex items-center justify-center text-green-600 hover:text-green-700 hover:bg-green-50 rounded transition-colors -ml-1"
                                                on:click={commitEditingStatus}
                                                title="Save changes"
                                            >
                                                <Check class="h-4 w-4" />
                                            </button>
                                            <!-- Cancel button -->
                                            <button
                                                type="button"
                                                class="h-6 w-6 flex items-center justify-center text-muted-foreground hover:text-foreground hover:bg-muted rounded transition-colors -ml-0.5"
                                                on:click={cancelEditingStatus}
                                                title="Cancel"
                                            >
                                                <X class="h-4 w-4" />
                                            </button>
                                        {:else}
                                            <!-- Status name -->
                                            <span class="text-sm font-medium">
                                                {column.label}
                                            </span>
                                            <!-- Edit pencil icon (visible on hover) -->
                                            {#if statuses.find((s) => s.display_name === column.status)}
                                                <button
                                                    type="button"
                                                    class="h-5 w-5 flex items-center justify-center text-muted-foreground hover:text-foreground opacity-0 group-hover:opacity-100 transition-opacity"
                                                    on:click={() =>
                                                        startEditingStatus(
                                                            column.status,
                                                        )}
                                                    title="Edit status"
                                                >
                                                    <Pencil class="h-3 w-3" />
                                                </button>
                                            {/if}
                                        {/if}
                                    </div>
                                    <div class="flex items-center gap-2">
                                        <div
                                            class="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity"
                                        >
                                            <!-- Per-column Add button removed -->
                                            {#if statuses.find((s) => s.display_name === column.status)}
                                                <button
                                                    type="button"
                                                    class="h-6 w-6 flex items-center justify-center text-red-500 hover:text-red-600 hover:bg-red-50 dark:hover:bg-red-950/50 rounded transition-colors"
                                                    aria-label="Delete status"
                                                    on:click={() =>
                                                        initiateDeleteStatus(
                                                            statuses.find(
                                                                (s) =>
                                                                    s.display_name ===
                                                                    column.status,
                                                            )!,
                                                        )}
                                                >
                                                    <Trash class="h-4 w-4" />
                                                </button>
                                            {/if}
                                        </div>
                                        {#if statusCategoryMap.has(column.status)}
                                            {@const mapping =
                                                statusCategoryMap.get(
                                                    column.status,
                                                )}
                                            {#if mapping?.categoryId}
                                                <div
                                                    class="relative group/mapping"
                                                >
                                                    <span
                                                        class="text-muted-foreground"
                                                    >
                                                        <Workflow
                                                            class="h-3.5 w-3.5"
                                                        />
                                                    </span>
                                                    <!-- Popover -->
                                                    <div
                                                        class="absolute right-0 top-full mt-2 w-64 p-3 rounded-lg border bg-popover text-popover-foreground shadow-lg opacity-0 invisible group-hover/mapping:opacity-100 group-hover/mapping:visible transition-all z-50"
                                                    >
                                                        <div
                                                            class="text-xs font-semibold mb-1"
                                                        >
                                                            {mapping.categoryLabel}
                                                        </div>
                                                        <div
                                                            class="text-xs text-muted-foreground"
                                                        >
                                                            {mapping.categoryDescription}
                                                        </div>
                                                    </div>
                                                </div>
                                            {/if}
                                        {/if}
                                        <span
                                            class="text-xs text-muted-foreground bg-muted rounded px-1.5 py-0.5"
                                        >
                                            {countForStatus(column.status)}
                                        </span>
                                    </div>
                                </div>
                            </div>

                            <!-- Column Content -->
                            <div
                                class="h-[550px] rounded-lg border-2 border-dashed transition-colors bg-muted/30 {dragOverColumn ===
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

                                        {#each getEventsForStatus(column.status) as event (event.id)}
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
                                                        initiateDeleteEvent(
                                                            e.detail,
                                                        )}
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
                                                <!-- Status modal removed -->
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
                            <!-- Mapping indicator for active tab -->
                            {#if activeTab === "backlogs" && statusCategoryMap.has("Backlogs")}
                                {@const mapping =
                                    statusCategoryMap.get("Backlogs")}
                                {#if mapping?.categoryId}
                                    <div class="relative group/mapping">
                                        <span class="text-muted-foreground">
                                            <Workflow class="h-4 w-4" />
                                        </span>
                                        <!-- Popover -->
                                        <div
                                            class="absolute right-0 top-full mt-2 w-64 p-3 rounded-lg border bg-popover text-popover-foreground shadow-lg opacity-0 invisible group-hover/mapping:opacity-100 group-hover/mapping:visible transition-all z-50"
                                        >
                                            <div
                                                class="text-xs font-semibold mb-1"
                                            >
                                                {mapping.categoryLabel}
                                            </div>
                                            <div
                                                class="text-xs text-muted-foreground"
                                            >
                                                {mapping.categoryDescription}
                                            </div>
                                        </div>
                                    </div>
                                {/if}
                            {:else if activeTab === "archived" && statusCategoryMap.has("Archived")}
                                {@const mapping =
                                    statusCategoryMap.get("Archived")}
                                {#if mapping?.categoryId}
                                    <div class="relative group/mapping">
                                        <span class="text-muted-foreground">
                                            <Workflow class="h-4 w-4" />
                                        </span>
                                        <!-- Popover -->
                                        <div
                                            class="absolute right-0 top-full mt-2 w-64 p-3 rounded-lg border bg-popover text-popover-foreground shadow-lg opacity-0 invisible group-hover/mapping:opacity-100 group-hover/mapping:visible transition-all z-50"
                                        >
                                            <div
                                                class="text-xs font-semibold mb-1"
                                            >
                                                {mapping.categoryLabel}
                                            </div>
                                            <div
                                                class="text-xs text-muted-foreground"
                                            >
                                                {mapping.categoryDescription}
                                            </div>
                                        </div>
                                    </div>
                                {/if}
                            {/if}

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
                                                groupedEvents.backlogs || [],
                                                searchQuery,
                                            ),
                                            globalSortOption,
                                        )}
                                        {loading}
                                        {statuses}
                                        on:edit={(e) => openEditModal(e.detail)}
                                        on:delete={(e) =>
                                            initiateDeleteEvent(e.detail)}
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
                                            groupedEvents.archived || [],
                                            searchQuery,
                                        ),
                                        globalSortOption,
                                    )}
                                    {loading}
                                    on:edit={(e) => openEditModal(e.detail)}
                                    on:delete={(e) =>
                                        initiateDeleteEvent(e.detail)}
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

{#if showDeleteModal}
    <div
        class="fixed inset-0 z-50 flex items-center justify-center bg-black/40"
    >
        <div class="bg-background rounded-lg p-5 w-full max-w-sm space-y-4">
            <h2 class="text-sm font-semibold">Delete status?</h2>
            <p class="text-xs text-muted-foreground">
                {pendingDeleteEventsCount}
                {pendingDeleteEventsCount === 1 ? "event" : "events"} will be moved
                to Backlogs before deletion.
            </p>
            <div class="flex justify-end gap-2 text-xs">
                <Button
                    variant="outline"
                    size="sm"
                    on:click={cancelDeleteStatus}
                    disabled={deleting}
                >
                    Cancel
                </Button>
                <Button
                    size="sm"
                    on:click={confirmDeleteStatus}
                    disabled={deleting}
                >
                    {deleting ? "Deleting..." : "Confirm"}
                </Button>
            </div>
        </div>
    </div>
{/if}

{#if showDeleteEventModal && pendingDeleteEvent}
    <div
        class="fixed inset-0 z-50 flex items-center justify-center bg-black/40"
    >
        <div class="bg-background rounded-lg p-5 w-full max-w-sm space-y-4">
            <h2 class="text-sm font-semibold">Delete event?</h2>
            <p class="text-xs text-muted-foreground">
                Are you sure you want to delete "<strong
                    >{pendingDeleteEvent.title}</strong
                >"? This action cannot be undone.
            </p>
            <div class="flex justify-end gap-2 text-xs">
                <Button
                    variant="outline"
                    size="sm"
                    on:click={cancelDeleteEvent}
                    disabled={deletingEvent}
                >
                    Cancel
                </Button>
                <Button
                    size="sm"
                    on:click={confirmDeleteEvent}
                    disabled={deletingEvent}
                >
                    {deletingEvent ? "Deleting..." : "Confirm"}
                </Button>
            </div>
        </div>
    </div>
{/if}

<!-- Event Modal -->
<EventModal
    bind:isOpen={isModalOpen}
    mode={modalMode}
    event={editingEvent}
    {statuses}
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
    on:close={() => {
        isPublishModalOpen = false;
        publishingEvent = null;
    }}
/>

<StatusModal
    bind:isOpen={isStatusModalOpen}
    on:created={async (e) => {
        await loadStatuses();
        rebuildColumns();
        toast.success(`Status "${e.detail.display_name}" created successfully`);
    }}
    on:close={() => {
        isStatusModalOpen = false;
    }}
/>

<StatusMappingModal
    bind:isOpen={isStatusMappingModalOpen}
    onClose={async () => {
        isStatusMappingModalOpen = false;
        await checkUnmappedStatuses();
    }}
/>

<style>
    /* Kanban column scrollbar styling to match dark mode */
    .overflow-y-auto {
        scrollbar-width: thin;
        scrollbar-color: hsl(var(--border)) transparent;
    }

    .overflow-y-auto::-webkit-scrollbar {
        width: 8px;
        height: 8px;
    }

    .overflow-y-auto::-webkit-scrollbar-track {
        background: transparent;
    }

    .overflow-y-auto::-webkit-scrollbar-thumb {
        background-color: hsl(var(--border));
        border-radius: 4px;
        border: 2px solid transparent;
        background-clip: content-box;
    }

    .overflow-y-auto::-webkit-scrollbar-thumb:hover {
        background-color: hsl(var(--border) / 0.8);
    }

    .overflow-y-auto::-webkit-scrollbar-corner {
        background: transparent;
    }
</style>
