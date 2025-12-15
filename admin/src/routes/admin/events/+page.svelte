<script lang="ts">
    import { onMount, tick } from "svelte";
    import { goto } from "$app/navigation";
    import { api } from "$lib/api";
    import { authStore } from "$lib/stores/auth";
    import { parseEvent } from "$lib/utils";
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
        Pencil,
        Check,
        Link2,
        Unlink,
        Columns,
    } from "lucide-svelte";
    import { Button, Input } from "$lib/components/ui";
    import { toast } from "svelte-sonner";
    import EventModal from "$lib/components/EventModal.svelte";
    import StatusModal from "$lib/components/StatusModal.svelte";
    import KanbanCard from "$lib/components/KanbanCard.svelte";
    import * as m from "$lib/paraglide/messages";

    import { fly } from "svelte/transition";

    let events: ParsedEvent[] = [];
    let loading = true;
    let error = "";
    let showGlobalNew = false;

    // Modal state
    let isModalOpen = false;
    let modalMode: "create" | "edit" = "create";
    let editingEvent: ParsedEvent | null = null;

    // Status modal state
    let isStatusModalOpen = false;
    let isCreatingNewStatus = false;
    let newStatusName = "";
    let newStatusInputEl: HTMLInputElement;
    let dragOverColumn: string | null = null;
    let draggedEventId: number | null = null;
    let draggedEventStatus: string | null = null;

    // Column drag state
    let draggedColumnStatus: string | null = null;
    let dragOverColumnIndex: number | null = null;
    let dropPosition: "before" | "after" | null = null;

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
        // Prevent deleting the last status
        if (statuses.length <= 1) {
            toast.error(m.events_page_cannot_delete_last_status_message());
            return;
        }

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

        // Move affected events to first status by order (excluding the one being deleted)
        const affected = events.filter((e) => e.status === targetName);
        const sortedStatuses = [...statuses]
            .filter((s) => s.display_name !== targetName)
            .sort((a, b) => a.order - b.order);
        const firstStatus = sortedStatuses[0];

        if (!firstStatus) {
            toast.error(m.events_page_cannot_delete_last_status());
            deleting = false;
            showDeleteModal = false;
            pendingDeleteStatus = null;
            pendingDeleteEventsCount = 0;
            return;
        }

        const fallbackStatus = firstStatus.display_name;
        try {
            await Promise.all(
                affected.map((ev) =>
                    api.updateEvent(ev.id, { status: fallbackStatus }),
                ),
            );

            events = events.map((e) =>
                e.status === targetName ? { ...e, status: fallbackStatus } : e,
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

            toast(m.events_page_status_deleted(), {
                description: m.events_page_status_deleted_description({
                    statusName: targetName,
                    count: affected.length.toString(),
                    eventWord:
                        affected.length === 1
                            ? m.events_page_event()
                            : m.events_page_events(),
                }),
            });
        } catch (e) {
            toast(m.events_page_delete_failed(), {
                description:
                    e instanceof Error
                        ? e.message
                        : m.events_page_unknown_error(),
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
            toast(m.events_page_status_updated(), {
                description: m.events_page_status_updated_description({
                    statusName: def.display_name,
                }),
            });
        } catch (e) {
            toast(m.events_page_update_failed(), {
                description:
                    e instanceof Error
                        ? e.message
                        : m.events_page_unknown_error(),
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

    // Reactive function that updates when events or search/sort inputs change.
    // Filter events in real-time by status name
    $: getEventsForStatus = (status: string): ParsedEvent[] => {
        // Filter events by status name
        let list = events.filter((e) => e.status === status);
        list = filterEvents(list, searchQuery);
        return sortEvents(list, globalSortOption);
    };

    // Track filtered event count for search results
    $: filteredEventCount = filterEvents(events, searchQuery).length;

    $: hasSearchResults = !searchQuery.trim() || filteredEventCount > 0;

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
                return m.events_page_sort_date_newest();
            case "DateDesc":
                return m.events_page_sort_date_oldest();
            case "TitleAsc":
                return m.events_page_sort_title_az();
            case "TitleDesc":
                return m.events_page_sort_title_za();
            case "UpdatedAsc":
                return m.events_page_sort_updated_newest();
            case "UpdatedDesc":
                return m.events_page_sort_updated_oldest();
            default:
                return m.events_page_sort_change_order();
        }
    }

    onMount(async () => {
        // Wait for authentication to be initialized before loading events & statuses
        const unsubscribe = authStore.subscribe(async (auth) => {
            if (auth.initialized && auth.isAuthenticated) {
                await Promise.all([
                    loadEvents(),
                    loadStatuses(),
                    loadCategoryMappings(),
                ]);
                unsubscribe();
            } else if (auth.initialized && !auth.isAuthenticated) {
                goto("/login");
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
        } catch (err) {
            error =
                err instanceof Error ? err.message : "Failed to load events";
            console.error("Failed to load events:", err);
        } finally {
            loading = false;
        }
    }

    async function loadCategoryMappings() {
        try {
            const manifestData = await api.getThemeManifest();
            const mappingsData = await api.getStatusMappings();

            if (!manifestData.manifest) {
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
                        categoryLabel: m.events_page_not_linked_to_theme(),
                        categoryDescription: "",
                    });
                });
            }
            statusCategoryMap = new Map(statusCategoryMap); // Trigger reactivity
        } catch (err) {
            console.error("Failed to load category mappings:", err);
        }
    }

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

    async function startCreatingNewStatus() {
        showGlobalNew = false;
        isCreatingNewStatus = true;
        newStatusName = "";
        await tick();
        newStatusInputEl?.focus();
    }

    function cancelCreatingNewStatus() {
        isCreatingNewStatus = false;
        newStatusName = "";
    }

    async function createNewStatus() {
        const trimmedName = newStatusName.trim();

        if (!trimmedName) {
            toast.error(m.events_page_status_name_empty());
            return;
        }

        // Check if a status with this display name already exists
        const existingStatus = statuses.find(
            (s) => s.display_name.toLowerCase() === trimmedName.toLowerCase(),
        );

        if (existingStatus) {
            toast.error(
                m.events_page_status_already_exists({
                    statusName: trimmedName,
                }),
            );
            return;
        }

        try {
            await api.createStatus({
                display_name: trimmedName,
            });

            toast.success(
                m.events_page_status_created_success({
                    statusName: trimmedName,
                }),
            );
            isCreatingNewStatus = false;
            newStatusName = "";
            await loadStatuses();
            await loadCategoryMappings();
        } catch (err: any) {
            console.error("Failed to create status:", err);
            const errorMessage = err.message || "Failed to create status";

            if (errorMessage.includes("already exists")) {
                toast.error(m.events_page_status_similar_exists());
            } else {
                toast.error(errorMessage);
            }
        }
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
            toast.success(m.events_page_event_deleted_success());
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

                toast(m.events_page_event_moved_success(), {
                    description: m.events_page_event_moved_description({
                        eventTitle: event.title,
                        statusLabel: label,
                    }),
                    action: {
                        label: m.events_page_share_update(),
                        onClick: () => {
                            openEditModal(event);
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

    // Column drag and drop handlers
    function handleColumnDragStart(e: DragEvent, status: string) {
        draggedColumnStatus = status;
        if (e.dataTransfer) {
            e.dataTransfer.effectAllowed = "move";
            e.dataTransfer.setData("text/plain", status);
        }
    }

    function handleColumnDragOver(e: DragEvent, index: number) {
        e.preventDefault();
        if (draggedColumnStatus) {
            const rect = (
                e.currentTarget as HTMLElement
            ).getBoundingClientRect();
            const midpoint = rect.left + rect.width / 2;
            const mouseX = e.clientX;

            dragOverColumnIndex = index;
            dropPosition = mouseX < midpoint ? "before" : "after";

            if (e.dataTransfer) {
                e.dataTransfer.dropEffect = "move";
            }
        }
    }

    function handleColumnDragEnter(index: number) {
        if (draggedColumnStatus) {
            dragOverColumnIndex = index;
        }
    }

    function handleColumnDragLeave() {
        dragOverColumnIndex = null;
        dropPosition = null;
    }

    function handleColumnDragEnd() {
        draggedColumnStatus = null;
        dragOverColumnIndex = null;
        dropPosition = null;
    }

    async function handleColumnDrop(e: DragEvent, targetIndex: number) {
        e.preventDefault();
        if (!draggedColumnStatus || dropPosition === null) return;

        const sourceIndex = columns.findIndex(
            (col) => col.status === draggedColumnStatus,
        );
        if (sourceIndex === -1) {
            handleColumnDragEnd();
            return;
        }

        // Calculate actual insertion index based on drop position
        let insertIndex = targetIndex;
        if (dropPosition === "after") {
            insertIndex = targetIndex + 1;
        }

        // Adjust if dragging from before to after
        if (sourceIndex < insertIndex) {
            insertIndex--;
        }

        if (sourceIndex === insertIndex) {
            handleColumnDragEnd();
            return;
        }

        // Reorder columns array
        const newColumns = [...columns];
        const [movedColumn] = newColumns.splice(sourceIndex, 1);
        newColumns.splice(insertIndex, 0, movedColumn);
        columns = newColumns;

        // Update statuses array to match new order
        const reorderedStatuses = newColumns
            .map((col, idx) => {
                const status = statuses.find(
                    (s) => s.display_name === col.status,
                );
                if (status) {
                    return { ...status, order: idx };
                }
                return null;
            })
            .filter(Boolean) as StatusDefinition[];

        // Save to backend
        try {
            const orderData = {
                order: reorderedStatuses.map((s) => ({
                    id: s.id,
                    order: s.order,
                })),
            };
            await api.reorderStatuses(orderData);

            // Update local statuses array
            statuses = statuses.map((s) => {
                const updated = reorderedStatuses.find((rs) => rs.id === s.id);
                return updated || s;
            });

            toast.success(m.events_page_column_order_updated());
        } catch (error) {
            console.error("Failed to reorder columns:", error);
            toast.error(m.events_page_column_order_failed());
            // Revert on error
            await loadStatuses();
            rebuildColumns();
        }

        handleColumnDragEnd();
    }

    // Click outside handler for New popover
    function handleClickOutside(event: MouseEvent) {
        if (showGlobalNew) {
            const target = event.target as HTMLElement;
            const popover = target.closest("[data-new-popover]");
            if (!popover) {
                showGlobalNew = false;
            }
        }
    }

    // Reordering functionality removed
</script>

<svelte:window on:click={handleClickOutside} />

<svelte:head>
    <title>{m.events_page_title()}</title>
</svelte:head>

<div class="w-full">
    <!-- Page Title -->
    <div class="mb-8">
        <h1 class="text-xl font-semibold mb-1">{m.events_page_heading()}</h1>
        <p class="text-muted-foreground text-sm">
            {m.events_page_subheading()}
        </p>
    </div>

    <!-- Controls Row -->
    <div class="flex items-center justify-between mb-4 pb-4 border-b">
        <!-- Search bar and Sort button -->
        <div class="flex items-center gap-2">
            <div class="relative w-[360px]">
                <Input
                    type="text"
                    placeholder={m.events_page_search_placeholder()}
                    bind:value={searchQuery}
                    class="h-8 text-sm pr-8"
                />
                <button
                    class="absolute right-2 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                    on:click={() => (searchQuery = "")}
                    title={searchQuery
                        ? m.events_page_clear_search()
                        : m.events_page_search()}
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
                    <option value="DateAsc"
                        >{m.events_page_sort_option_date_asc()}</option
                    >
                    <option value="DateDesc"
                        >{m.events_page_sort_option_date_desc()}</option
                    >
                    <option value="TitleAsc"
                        >{m.events_page_sort_option_title_asc()}</option
                    >
                    <option value="TitleDesc"
                        >{m.events_page_sort_option_title_desc()}</option
                    >
                    <option value="UpdatedAsc"
                        >{m.events_page_sort_option_updated_asc()}</option
                    >
                    <option value="UpdatedDesc"
                        >{m.events_page_sort_option_updated_desc()}</option
                    >
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
            <!-- New Button with Popover -->
            <!-- New button -->
            <div class="relative inline-block" data-new-popover>
                <button
                    type="button"
                    class="h-8 px-3 text-xs border rounded-md bg-background hover:bg-muted flex items-center gap-1"
                    on:click={() => (showGlobalNew = !showGlobalNew)}
                    aria-haspopup="true"
                    aria-expanded={showGlobalNew}
                >
                    <Plus class="h-3 w-3" />
                    <span>{m.events_page_new()}</span>
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
                        transition:fly={{ y: -10, duration: 200 }}
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
                            {m.events_page_new_event()}
                        </button>
                        <button
                            type="button"
                            class="w-full text-left px-2 py-1 rounded hover:bg-muted flex items-center gap-2"
                            on:click={startCreatingNewStatus}
                            role="menuitem"
                        >
                            <Columns class="h-4 w-4" />
                            {m.events_page_new_status()}
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
                    {m.events_page_no_events_found({ query: searchQuery })}
                </div>
                <Button variant="outline" on:click={() => (searchQuery = "")}
                    >{m.events_page_clear_search_button()}</Button
                >
            </div>
        {:else}
            <!-- Kanban Board -->
            <div class="w-full overflow-x-auto">
                {#if columns.length === 0}
                    <div
                        class="flex flex-col items-center justify-center py-16 px-4 text-center"
                    >
                        <div class="text-muted-foreground mb-4">
                            <svg
                                xmlns="http://www.w3.org/2000/svg"
                                class="h-16 w-16 mx-auto mb-3 opacity-50"
                                fill="none"
                                viewBox="0 0 24 24"
                                stroke="currentColor"
                            >
                                <path
                                    stroke-linecap="round"
                                    stroke-linejoin="round"
                                    stroke-width="1.5"
                                    d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"
                                />
                            </svg>
                            <p class="text-lg font-medium mb-2">
                                {m.events_page_no_statuses_configured()}
                            </p>
                            <p class="text-sm">
                                {m.events_page_create_first_status()}
                            </p>
                        </div>
                        <Button
                            variant="outline"
                            on:click={() => (isStatusModalOpen = true)}
                            class="mt-2"
                        >
                            <Plus class="h-4 w-4 mr-2" />
                            {m.events_page_create_status()}
                        </Button>
                    </div>
                {:else}
                    <div class="flex gap-2 lg:gap-4 min-h-0 pb-3 w-max">
                        <!-- Drop indicator for before first column -->
                        {#if draggedColumnStatus && dragOverColumnIndex === 0 && dropPosition === "before"}
                            <div class="w-1 bg-primary/50 rounded-full"></div>
                        {/if}

                        <!-- New Status Creation Column -->
                        {#if isCreatingNewStatus}
                            <div
                                class="flex-shrink-0 w-64 lg:w-72"
                                style="height: calc(100vh - 280px);"
                            >
                                <div class="mb-3">
                                    <div
                                        class="flex items-center justify-between px-3 py-2 rounded-md border-2 border-primary bg-primary/5"
                                    >
                                        <div
                                            class="flex items-center gap-2 flex-1"
                                        >
                                            <input
                                                bind:this={newStatusInputEl}
                                                bind:value={newStatusName}
                                                type="text"
                                                placeholder={m.events_page_status_name_placeholder()}
                                                class="flex-1 bg-transparent border-none outline-none text-sm font-medium"
                                                on:keydown={(e) => {
                                                    if (e.key === "Enter") {
                                                        createNewStatus();
                                                    } else if (
                                                        e.key === "Escape"
                                                    ) {
                                                        cancelCreatingNewStatus();
                                                    }
                                                }}
                                            />
                                            <button
                                                type="button"
                                                class="text-green-600 hover:text-green-700 transition-colors"
                                                on:click={createNewStatus}
                                                title={m.events_page_save()}
                                            >
                                                <Check class="h-4 w-4" />
                                            </button>
                                            <button
                                                type="button"
                                                class="text-muted-foreground hover:text-foreground transition-colors"
                                                on:click={cancelCreatingNewStatus}
                                                title={m.events_page_cancel()}
                                            >
                                                <X class="h-4 w-4" />
                                            </button>
                                        </div>
                                    </div>
                                </div>
                                <div
                                    class="rounded-lg border-2 border-dashed border-primary/30 bg-muted/20 overflow-hidden"
                                    style="height: calc(100vh - 280px);"
                                >
                                    <div
                                        class="flex items-center justify-center h-full text-xs text-muted-foreground"
                                    >
                                        {m.events_page_press_enter_to_create()}
                                    </div>
                                </div>
                            </div>
                        {/if}

                        {#each columns as column, index (column.status)}
                            <!-- Drop indicator before (not first) -->
                            {#if draggedColumnStatus && dragOverColumnIndex === index && dropPosition === "before" && index > 0}
                                <div
                                    class="w-1 bg-primary rounded-full self-stretch -mx-1.5 z-10"
                                ></div>
                            {/if}

                            <div
                                class="flex-shrink-0 w-[320px] group transition-opacity {draggedColumnStatus ===
                                column.status
                                    ? 'opacity-50'
                                    : ''}"
                                draggable={!draggedEventId}
                                on:dragstart={(e) => {
                                    // Only allow column drag if not dragging an event
                                    if (!draggedEventId) {
                                        handleColumnDragStart(e, column.status);
                                    }
                                }}
                                on:dragover={(e) => {
                                    // Only handle column drag over if dragging a column
                                    if (draggedColumnStatus) {
                                        handleColumnDragOver(e, index);
                                    }
                                }}
                                on:dragenter={() => {
                                    if (draggedColumnStatus) {
                                        handleColumnDragEnter(index);
                                    }
                                }}
                                on:dragleave={() => {
                                    if (draggedColumnStatus) {
                                        handleColumnDragLeave();
                                    }
                                }}
                                on:drop={(e) => {
                                    if (draggedColumnStatus) {
                                        handleColumnDrop(e, index);
                                    }
                                }}
                                on:dragend={() => {
                                    if (draggedColumnStatus) {
                                        handleColumnDragEnd();
                                    }
                                }}
                                role="none"
                            >
                                <!-- Column Header -->
                                <div class="mb-3 cursor-move">
                                    <div
                                        class="flex items-center justify-between"
                                    >
                                        <div
                                            class="flex items-center gap-1.5 flex-1"
                                        >
                                            {#if editingStatusId !== null && statuses.find((s) => s.id === editingStatusId)?.display_name === column.status}
                                                <div
                                                    class="flex items-center justify-between px-3 py-2 rounded-md border-2 border-primary bg-primary/5 flex-1"
                                                >
                                                    <div
                                                        class="flex items-center gap-2 flex-1"
                                                    >
                                                        <!-- Name input -->
                                                        <input
                                                            class="flex-1 bg-transparent border-none outline-none text-sm font-medium"
                                                            bind:this={
                                                                editingInputEl
                                                            }
                                                            bind:value={
                                                                editingStatusName
                                                            }
                                                            on:keydown={(e) => {
                                                                if (
                                                                    e.key ===
                                                                    "Enter"
                                                                ) {
                                                                    commitEditingStatus();
                                                                } else if (
                                                                    e.key ===
                                                                    "Escape"
                                                                ) {
                                                                    cancelEditingStatus();
                                                                }
                                                            }}
                                                            placeholder={m.events_page_status_name_placeholder()}
                                                            maxlength="50"
                                                        />
                                                        <!-- Check button -->
                                                        <button
                                                            type="button"
                                                            class="text-green-600 hover:text-green-700 transition-colors"
                                                            on:click={commitEditingStatus}
                                                            title={m.events_page_save()}
                                                        >
                                                            <Check
                                                                class="h-4 w-4"
                                                            />
                                                        </button>
                                                        <!-- Cancel button -->
                                                        <button
                                                            type="button"
                                                            class="text-muted-foreground hover:text-foreground transition-colors"
                                                            on:click={cancelEditingStatus}
                                                            title={m.events_page_cancel()}
                                                        >
                                                            <X
                                                                class="h-4 w-4"
                                                            />
                                                        </button>
                                                    </div>
                                                </div>
                                            {:else}
                                                <!-- Status name -->
                                                <span
                                                    class="text-sm font-medium"
                                                >
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
                                                        title={m.events_page_edit_status()}
                                                    >
                                                        <Pencil
                                                            class="h-3 w-3"
                                                        />
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
                                                        class="h-6 w-6 flex items-center justify-center rounded transition-colors {statuses.length <=
                                                        1
                                                            ? 'text-muted-foreground/30 cursor-not-allowed'
                                                            : 'text-red-500 hover:text-red-600 hover:bg-red-50 dark:hover:bg-red-950/50'}"
                                                        aria-label={m.events_page_delete_status_aria()}
                                                        disabled={statuses.length <=
                                                            1}
                                                        title={statuses.length <=
                                                        1
                                                            ? m.events_page_cannot_delete_last_status()
                                                            : m.events_page_delete_status_tooltip()}
                                                        on:click={() =>
                                                            initiateDeleteStatus(
                                                                statuses.find(
                                                                    (s) =>
                                                                        s.display_name ===
                                                                        column.status,
                                                                )!,
                                                            )}
                                                    >
                                                        <Trash
                                                            class="h-4 w-4"
                                                        />
                                                    </button>
                                                {/if}
                                            </div>
                                            {#if statusCategoryMap.has(column.status)}
                                                {@const mapping =
                                                    statusCategoryMap.get(
                                                        column.status,
                                                    )}
                                                <div
                                                    class="relative group/mapping"
                                                    title={mapping?.categoryId
                                                        ? m.events_page_linked_to_theme_title(
                                                              {
                                                                  categoryLabel:
                                                                      mapping.categoryLabel,
                                                              },
                                                          )
                                                        : m.events_page_not_linked_to_theme()}
                                                >
                                                    {#if mapping?.categoryId}
                                                        <Link2
                                                            class="h-3.5 w-3.5 text-blue-600 dark:text-blue-500"
                                                        />
                                                    {:else}
                                                        <Unlink
                                                            class="h-3.5 w-3.5 text-muted-foreground/50"
                                                        />
                                                    {/if}
                                                    <!-- Tooltip on hover -->
                                                    <div
                                                        class="absolute right-0 top-full mt-2 w-64 p-3 rounded-lg border bg-popover text-popover-foreground shadow-lg opacity-0 invisible group-hover/mapping:opacity-100 group-hover/mapping:visible transition-all z-50 pointer-events-none"
                                                    >
                                                        {#if mapping?.categoryId}
                                                            <div
                                                                class="flex items-center gap-2 mb-2"
                                                            >
                                                                <Link2
                                                                    class="h-3.5 w-3.5 text-blue-600 dark:text-blue-500 flex-shrink-0"
                                                                />
                                                                <span
                                                                    class="text-xs font-semibold text-blue-600 dark:text-blue-500"
                                                                >
                                                                    {m.events_page_linked_to_theme()}
                                                                </span>
                                                            </div>
                                                            <div
                                                                class="text-xs mb-1"
                                                            >
                                                                <strong
                                                                    >{mapping.categoryLabel}</strong
                                                                >
                                                            </div>
                                                            {#if mapping.categoryDescription}
                                                                <div
                                                                    class="text-xs text-muted-foreground"
                                                                >
                                                                    {mapping.categoryDescription}
                                                                </div>
                                                            {/if}
                                                        {:else}
                                                            <div
                                                                class="flex items-center gap-2 mb-2"
                                                            >
                                                                <Unlink
                                                                    class="h-3.5 w-3.5 text-muted-foreground flex-shrink-0"
                                                                />
                                                                <span
                                                                    class="text-xs font-semibold"
                                                                >
                                                                    {m.events_page_not_linked_to_theme()}
                                                                </span>
                                                            </div>
                                                            <div
                                                                class="text-xs text-muted-foreground"
                                                            >
                                                                {m.events_page_not_linked_description()}
                                                            </div>
                                                        {/if}
                                                    </div>
                                                </div>
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
                                    class="rounded-lg border-2 border-dashed transition-colors bg-muted/30 {dragOverColumn ===
                                        column.status && draggedEventId
                                        ? 'ring-2 ring-primary border-primary'
                                        : ''} overflow-hidden relative flex flex-col"
                                    style="height: calc(100vh - 280px);"
                                    on:drop={(e) => {
                                        if (
                                            draggedEventId &&
                                            !draggedColumnStatus
                                        ) {
                                            handleDrop(e, column.status);
                                        }
                                    }}
                                    on:dragover={(e) => {
                                        if (
                                            draggedEventId &&
                                            !draggedColumnStatus
                                        ) {
                                            handleDragOver(e, column.status);
                                        }
                                    }}
                                    on:dragenter={(e) => {
                                        if (
                                            draggedEventId &&
                                            !draggedColumnStatus
                                        ) {
                                            handleDragEnter(e);
                                        }
                                    }}
                                    on:dragleave={() => {
                                        if (
                                            draggedEventId &&
                                            !draggedColumnStatus
                                        ) {
                                            handleDragLeave();
                                        }
                                    }}
                                    role="region"
                                    aria-label="Drop zone for {column.label} events"
                                >
                                    <div class="flex-1 overflow-y-auto">
                                        <div class="space-y-2 p-3 min-w-0">
                                            <!-- Drop zone at top of column -->
                                            <div
                                                class="h-2 transition-all duration-200 {draggedEventId &&
                                                draggedEventStatus ===
                                                    column.status &&
                                                getEventsForStatus(
                                                    column.status,
                                                ).length > 0
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
                                                            openEditModal(
                                                                e.detail,
                                                            )}
                                                        on:delete={(e) =>
                                                            initiateDeleteEvent(
                                                                e.detail,
                                                            )}
                                                        on:statusChange={(e) =>
                                                            handleStatusChange(
                                                                e.detail
                                                                    .eventId,
                                                                e.detail
                                                                    .newStatus,
                                                            )}
                                                        on:carddragstart={(
                                                            e,
                                                        ) => {
                                                            draggedEventId =
                                                                e.detail
                                                                    .eventId;
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
                                                    {m.events_page_no_events()}
                                                </div>
                                            {/if}
                                        </div>
                                    </div>

                                    <!-- New Event Button -->
                                    <div
                                        class="absolute bottom-0 left-0 right-0 p-3 opacity-0 group-hover:opacity-100 translate-y-2 group-hover:translate-y-0 transition-all duration-200 pointer-events-none group-hover:pointer-events-auto"
                                    >
                                        <button
                                            type="button"
                                            class="w-full py-2 px-3 text-xs font-medium rounded-md bg-background border border-border hover:bg-muted text-foreground transition-all flex items-center justify-center gap-2"
                                            on:click={() =>
                                                openCreateModal(column.status)}
                                        >
                                            <Plus class="h-3.5 w-3.5" />
                                            <span
                                                >{m.events_page_new_event()}</span
                                            >
                                        </button>
                                    </div>
                                </div>
                            </div>

                            <!-- Drop indicator after -->
                            {#if draggedColumnStatus && dragOverColumnIndex === index && dropPosition === "after"}
                                <div
                                    class="w-1 bg-primary rounded-full self-stretch -mx-1.5 z-10"
                                ></div>
                            {/if}
                        {/each}
                    </div>
                {/if}
            </div>

            <!-- Summary -->
            <div class="mt-6 text-center text-xs text-muted-foreground">
                {m.events_page_total_events({
                    count: events.length.toString(),
                })}
            </div>
        {/if}
    </main>
</div>

{#if showDeleteModal}
    <div
        class="fixed inset-0 z-50 flex items-center justify-center bg-black/40"
    >
        <div class="bg-background rounded-lg p-5 w-full max-w-sm space-y-4">
            <h2 class="text-sm font-semibold">
                {m.events_page_delete_status_title()}
            </h2>
            <p class="text-xs text-muted-foreground">
                {m.events_page_delete_status_message({
                    count: pendingDeleteEventsCount.toString(),
                    eventWord:
                        pendingDeleteEventsCount === 1
                            ? m.events_page_event()
                            : m.events_page_events(),
                })}
            </p>
            <div class="flex justify-end gap-2 text-xs">
                <Button
                    variant="outline"
                    size="sm"
                    on:click={cancelDeleteStatus}
                    disabled={deleting}
                >
                    {m.events_page_cancel()}
                </Button>
                <Button
                    size="sm"
                    on:click={confirmDeleteStatus}
                    disabled={deleting}
                >
                    {deleting
                        ? m.events_page_deleting()
                        : m.events_page_confirm()}
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
            <h2 class="text-sm font-semibold">
                {m.events_page_delete_event_title()}
            </h2>
            <p class="text-xs text-muted-foreground">
                {m.events_page_delete_event_message({
                    eventTitle: pendingDeleteEvent.title,
                })}
            </p>
            <div class="flex justify-end gap-2 text-xs">
                <Button
                    variant="outline"
                    size="sm"
                    on:click={cancelDeleteEvent}
                    disabled={deletingEvent}
                >
                    {m.events_page_cancel()}
                </Button>
                <Button
                    size="sm"
                    on:click={confirmDeleteEvent}
                    disabled={deletingEvent}
                >
                    {deletingEvent
                        ? m.events_page_deleting()
                        : m.events_page_confirm()}
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
    on:close={() => {
        isModalOpen = false;
        editingEvent = null;
    }}
/>

<StatusModal
    bind:isOpen={isStatusModalOpen}
    on:created={async (e) => {
        await loadStatuses();
        rebuildColumns();
        toast.success(
            m.events_page_status_created({ statusName: e.detail.display_name }),
        );
    }}
    on:close={() => {
        isStatusModalOpen = false;
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
