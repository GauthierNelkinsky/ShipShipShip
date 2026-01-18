<script lang="ts">
    // Events management page - kanban board view
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
        Calendar,
        Columns,
        List,
        LayoutGrid,
        Eye,
        EyeOff,
    } from "lucide-svelte";
    import { Button, Input, Tooltip } from "$lib/components/ui";
    import { toast } from "svelte-sonner";
    import EventModal from "$lib/components/EventModal.svelte";
    import KanbanView from "$lib/components/KanbanView.svelte";
    import ListView from "$lib/components/ListView.svelte";
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

    // Status creation state
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

    // View mode state
    let viewMode: "kanban" | "list" = "kanban";
    let preferencesLoaded = false;

    // Load view preferences from localStorage
    function loadViewPreferences() {
        if (typeof window !== "undefined") {
            const savedViewMode = localStorage.getItem("eventsViewMode");
            if (savedViewMode === "kanban" || savedViewMode === "list") {
                viewMode = savedViewMode;
            }

            const savedHiddenStatuses = localStorage.getItem(
                "eventsHiddenStatuses",
            );
            if (savedHiddenStatuses) {
                try {
                    const parsed = JSON.parse(savedHiddenStatuses);
                    if (Array.isArray(parsed)) {
                        hiddenStatuses = new Set(parsed);
                    }
                } catch (e) {
                    console.error(
                        "Failed to parse hidden statuses from localStorage",
                        e,
                    );
                }
            }
        }
        preferencesLoaded = true;
    }

    // Save view mode to localStorage when it changes
    $: if (typeof window !== "undefined" && preferencesLoaded) {
        localStorage.setItem("eventsViewMode", viewMode);
    }

    // Save hidden statuses to localStorage when they change
    $: if (typeof window !== "undefined" && preferencesLoaded) {
        localStorage.setItem(
            "eventsHiddenStatuses",
            JSON.stringify(Array.from(hiddenStatuses)),
        );
    }

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

    // Hidden statuses state
    let hiddenStatuses: Set<string> = new Set();
    let showHiddenStatusesPopover = false;
    let showSortPopover = false;

    function toggleStatusVisibility(statusName: string) {
        const newSet = new Set(hiddenStatuses);
        if (newSet.has(statusName)) {
            newSet.delete(statusName);
        } else {
            newSet.add(statusName);
        }
        hiddenStatuses = newSet;
    }

    function unhideStatus(statusName: string) {
        const newSet = new Set(hiddenStatuses);
        newSet.delete(statusName);
        hiddenStatuses = newSet;
    }

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

    // Calculate hidden events count
    $: hiddenEventsCount = events.filter((e) =>
        hiddenStatuses.has(e.status),
    ).length;

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
        // Load view preferences from localStorage
        loadViewPreferences();

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

    async function handleColumnDrop(e: DragEvent, targetStatus: string) {
        e.preventDefault();
        if (!draggedColumnStatus || dropPosition === null) return;

        const sourceIndex = columns.findIndex(
            (col) => col.status === draggedColumnStatus,
        );
        if (sourceIndex === -1) {
            handleColumnDragEnd();
            return;
        }

        const targetIndex = columns.findIndex(
            (col) => col.status === targetStatus,
        );
        if (targetIndex === -1) {
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

    // Click outside handler for New popover and hidden statuses popover
    function handleClickOutside(event: MouseEvent) {
        const target = event.target as HTMLElement;

        if (showGlobalNew) {
            const popover = target.closest("[data-new-popover]");
            if (!popover) {
                showGlobalNew = false;
            }
        }

        if (showHiddenStatusesPopover) {
            const popover = target.closest("[data-hidden-statuses-popover]");
            if (!popover) {
                showHiddenStatusesPopover = false;
            }
        }

        if (showSortPopover) {
            const popover = target.closest("[data-sort-popover]");
            if (!popover) {
                showSortPopover = false;
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
                    class="h-8 text-sm pe-8"
                />
                <button
                    class="absolute end-2 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
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
            <div class="relative" data-sort-popover>
                <button
                    type="button"
                    class="w-8 h-8 flex items-center justify-center bg-background border rounded-md hover:bg-muted transition-colors"
                    on:click={() => (showSortPopover = !showSortPopover)}
                    aria-haspopup="true"
                    aria-expanded={showSortPopover}
                    title={getSortTooltip(globalSortOption)}
                >
                    <svelte:component
                        this={getSortIcon(globalSortOption)}
                        class="h-4 w-4"
                    />
                </button>

                {#if showSortPopover}
                    <div
                        class="absolute end-0 mt-1 w-48 rounded-md border bg-background shadow-md p-2 text-xs space-y-1 z-50"
                        role="menu"
                        transition:fly={{ y: -10, duration: 200 }}
                    >
                        <div
                            class="text-[10px] font-semibold text-muted-foreground mb-1 px-2 py-0.5 uppercase tracking-wider"
                        >
                            {m.events_page_sort_events()}
                        </div>
                        <button
                            type="button"
                            class="w-full text-start px-2 py-1.5 rounded hover:bg-muted transition-colors flex items-center gap-2"
                            on:click={() => {
                                globalSortOption = "DateAsc";
                                showSortPopover = false;
                            }}
                            role="menuitem"
                        >
                            <CalendarArrowUp class="h-4 w-4" />
                            {m.events_page_sort_option_date_asc()}
                        </button>
                        <button
                            type="button"
                            class="w-full text-start px-2 py-1.5 rounded hover:bg-muted transition-colors flex items-center gap-2"
                            on:click={() => {
                                globalSortOption = "DateDesc";
                                showSortPopover = false;
                            }}
                            role="menuitem"
                        >
                            <CalendarArrowDown class="h-4 w-4" />
                            {m.events_page_sort_option_date_desc()}
                        </button>
                        <button
                            type="button"
                            class="w-full text-start px-2 py-1.5 rounded hover:bg-muted transition-colors flex items-center gap-2"
                            on:click={() => {
                                globalSortOption = "TitleAsc";
                                showSortPopover = false;
                            }}
                            role="menuitem"
                        >
                            <ArrowDownAZ class="h-4 w-4" />
                            {m.events_page_sort_option_title_asc()}
                        </button>
                        <button
                            type="button"
                            class="w-full text-start px-2 py-1.5 rounded hover:bg-muted transition-colors flex items-center gap-2"
                            on:click={() => {
                                globalSortOption = "TitleDesc";
                                showSortPopover = false;
                            }}
                            role="menuitem"
                        >
                            <ArrowUpZA class="h-4 w-4" />
                            {m.events_page_sort_option_title_desc()}
                        </button>
                        <button
                            type="button"
                            class="w-full text-start px-2 py-1.5 rounded hover:bg-muted transition-colors flex items-center gap-2"
                            on:click={() => {
                                globalSortOption = "UpdatedAsc";
                                showSortPopover = false;
                            }}
                            role="menuitem"
                        >
                            <ClockArrowUp class="h-4 w-4" />
                            {m.events_page_sort_option_updated_asc()}
                        </button>
                        <button
                            type="button"
                            class="w-full text-start px-2 py-1.5 rounded hover:bg-muted transition-colors flex items-center gap-2"
                            on:click={() => {
                                globalSortOption = "UpdatedDesc";
                                showSortPopover = false;
                            }}
                            role="menuitem"
                        >
                            <ClockArrowDown class="h-4 w-4" />
                            {m.events_page_sort_option_updated_desc()}
                        </button>
                    </div>
                {/if}
            </div>

            <!-- Show hidden statuses button -->
            {#if hiddenStatuses.size > 0}
                <div class="relative" data-hidden-statuses-popover>
                    <button
                        type="button"
                        class="h-8 px-3 text-xs border rounded-md bg-background hover:bg-muted flex items-center gap-1.5 transition-colors"
                        on:click={() =>
                            (showHiddenStatusesPopover =
                                !showHiddenStatusesPopover)}
                        aria-haspopup="true"
                        aria-expanded={showHiddenStatusesPopover}
                        title="{hiddenStatuses.size} hidden status{hiddenStatuses.size >
                        1
                            ? 'es'
                            : ''}"
                    >
                        <EyeOff class="h-3.5 w-3.5" />
                        <span>{hiddenStatuses.size}</span>
                        <span>{m.events_page_hidden()}</span>
                    </button>

                    {#if showHiddenStatusesPopover}
                        <div
                            class="absolute start-0 mt-1 w-56 rounded-md border bg-background shadow-md p-2 text-xs space-y-1 z-50"
                            role="menu"
                            transition:fly={{ y: -10, duration: 200 }}
                        >
                            <div
                                class="text-[10px] font-semibold text-muted-foreground mb-1 px-2 py-0.5 uppercase tracking-wider"
                            >
                                {m.events_page_hidden_statuses()}
                            </div>
                            {#each Array.from(hiddenStatuses) as hiddenStatus}
                                {@const col = columns.find(
                                    (c) => c.status === hiddenStatus,
                                )}
                                {#if col}
                                    <button
                                        type="button"
                                        class="w-full text-start px-2 py-1.5 rounded hover:bg-muted transition-colors flex items-center justify-between gap-2"
                                        on:click={() => {
                                            unhideStatus(hiddenStatus);
                                            showHiddenStatusesPopover = false;
                                        }}
                                        role="menuitem"
                                    >
                                        <span class="truncate">{col.label}</span
                                        >
                                        <Eye
                                            class="h-3.5 w-3.5 shrink-0 opacity-70"
                                        />
                                    </button>
                                {/if}
                            {/each}
                        </div>
                    {/if}
                </div>
            {/if}
        </div>

        <!-- Settings and New buttons -->
        <div class="flex items-center gap-2">
            <!-- View mode toggle -->
            <div class="flex items-center gap-1 bg-muted rounded-md p-0.5">
                <button
                    type="button"
                    class="h-7 px-2 rounded flex items-center gap-1.5 text-xs transition-colors {viewMode ===
                    'kanban'
                        ? 'bg-background shadow-sm'
                        : 'hover:bg-background/50'}"
                    on:click={() => (viewMode = "kanban")}
                    title={m.events_page_view_kanban_title()}
                >
                    <LayoutGrid class="h-3.5 w-3.5" />
                    <span>{m.events_page_view_kanban()}</span>
                </button>
                <button
                    type="button"
                    class="h-7 px-2 rounded flex items-center gap-1.5 text-xs transition-colors {viewMode ===
                    'list'
                        ? 'bg-background shadow-sm'
                        : 'hover:bg-background/50'}"
                    on:click={() => (viewMode = "list")}
                    title={m.events_page_view_list_title()}
                >
                    <List class="h-3.5 w-3.5" />
                    <span>{m.events_page_view_list()}</span>
                </button>
            </div>

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
                        class="absolute end-0 mt-1 w-44 rounded-md border bg-background shadow p-2 text-xs space-y-1 z-30"
                        role="menu"
                        transition:fly={{ y: -10, duration: 200 }}
                    >
                        <div
                            class="text-[10px] font-semibold text-muted-foreground mb-1 px-2 py-0.5 uppercase tracking-wider"
                        >
                            {m.events_page_create_new()}
                        </div>
                        <button
                            type="button"
                            class="w-full text-start px-2 py-1 rounded hover:bg-muted flex items-center gap-2"
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
                            class="w-full text-start px-2 py-1 rounded hover:bg-muted flex items-center gap-2"
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
        {:else if viewMode === "kanban"}
            <!-- Kanban Board -->
            <KanbanView
                {columns}
                {statuses}
                {statusCategoryMap}
                {getEventsForStatus}
                {countForStatus}
                {hiddenStatuses}
                on:toggleStatusVisibility={(e) =>
                    toggleStatusVisibility(e.detail.statusName)}
                bind:draggedEventId
                bind:draggedEventStatus
                bind:dragOverColumn
                bind:draggedColumnStatus
                bind:dragOverColumnIndex
                bind:dropPosition
                bind:isCreatingNewStatus
                bind:newStatusName
                bind:newStatusInputEl
                bind:editingStatusId
                bind:editingStatusName
                bind:editingInputEl
                on:openCreateModal={(e) => openCreateModal(e.detail.status)}
                on:openEditModal={(e) => openEditModal(e.detail.event)}
                on:initiateDeleteEvent={(e) =>
                    initiateDeleteEvent(e.detail.event)}
                on:statusChange={(e) =>
                    handleStatusChange(e.detail.eventId, e.detail.newStatus)}
                on:initiateDeleteStatus={(e) =>
                    initiateDeleteStatus(e.detail.status)}
                on:startEditingStatus={(e) =>
                    startEditingStatus(e.detail.statusName)}
                on:commitEditingStatus={commitEditingStatus}
                on:cancelEditingStatus={cancelEditingStatus}
                on:createNewStatus={createNewStatus}
                on:cancelCreatingNewStatus={cancelCreatingNewStatus}
                on:drop={(e) => handleDrop(e.detail.event, e.detail.status)}
                on:dragover={(e) =>
                    handleDragOver(e.detail.event, e.detail.status)}
                on:dragenter={(e) => handleDragEnter(e.detail.event)}
                on:dragleave={handleDragLeave}
                on:columnDragStart={(e) =>
                    handleColumnDragStart(e.detail.event, e.detail.status)}
                on:columnDragOver={(e) =>
                    handleColumnDragOver(e.detail.event, e.detail.index)}
                on:columnDragEnter={(e) =>
                    handleColumnDragEnter(e.detail.index)}
                on:columnDragLeave={handleColumnDragLeave}
                on:columnDrop={(e) =>
                    handleColumnDrop(e.detail.event, e.detail.targetStatus)}
                on:columnDragEnd={handleColumnDragEnd}
            />
        {:else if viewMode === "list"}
            <!-- List View -->
            <ListView
                {columns}
                {statuses}
                {statusCategoryMap}
                {getEventsForStatus}
                {countForStatus}
                {hiddenStatuses}
                on:toggleStatusVisibility={(e) =>
                    toggleStatusVisibility(e.detail.statusName)}
                bind:editingStatusId
                bind:editingStatusName
                bind:editingInputEl
                bind:isCreatingNewStatus
                bind:newStatusName
                bind:newStatusInputEl
                on:openCreateModal={(e) => openCreateModal(e.detail.status)}
                on:openEditModal={(e) => openEditModal(e.detail.event)}
                on:initiateDeleteEvent={(e) =>
                    initiateDeleteEvent(e.detail.event)}
                on:statusChange={(e) =>
                    handleStatusChange(e.detail.eventId, e.detail.targetStatus)}
                on:reorderStatuses={(e) => {
                    const sourceIndex = columns.findIndex(
                        (col) => col.status === e.detail.sourceStatus,
                    );
                    const targetIndex = columns.findIndex(
                        (col) => col.status === e.detail.targetStatus,
                    );
                    if (sourceIndex !== -1 && targetIndex !== -1) {
                        draggedColumnStatus = e.detail.sourceStatus;
                        dropPosition = e.detail.dropPosition;
                        handleColumnDrop(
                            new DragEvent("drop"),
                            e.detail.targetStatus,
                        );
                    }
                }}
                on:startEditingStatus={(e) =>
                    startEditingStatus(e.detail.statusName)}
                on:initiateDeleteStatus={(e) =>
                    initiateDeleteStatus(e.detail.status)}
                on:commitEditingStatus={commitEditingStatus}
                on:cancelEditingStatus={cancelEditingStatus}
                on:createNewStatus={createNewStatus}
                on:cancelCreatingNewStatus={cancelCreatingNewStatus}
            />
        {/if}

        <!-- Summary -->
        <div class="mt-6 text-center text-xs text-muted-foreground">
            {#if hiddenEventsCount > 0}
                <Tooltip
                    content="{m.events_page_tooltip_visible({
                        visible: (events.length - hiddenEventsCount).toString(),
                    })} • {m.events_page_tooltip_hidden({
                        hidden: hiddenEventsCount.toString(),
                    })}"
                    side="top"
                >
                    {#snippet children()}
                        <span>
                            {m.events_page_total_events({
                                count: events.length.toString(),
                            })}
                        </span>
                    {/snippet}
                </Tooltip>
            {:else}
                {m.events_page_total_events({
                    count: events.length.toString(),
                })}
            {/if}
        </div>
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
