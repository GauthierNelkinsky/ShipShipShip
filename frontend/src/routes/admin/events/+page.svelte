<script lang="ts">
    import { onMount, tick } from "svelte";
    import { goto } from "$app/navigation";
    import { api } from "$lib/api";
    import { authStore } from "$lib/stores/auth";
    import { parseEvent, groupEventsByStatus } from "$lib/utils";
    import type { ParsedEvent, EventStatus } from "$lib/types";
    import { Plus, ArrowLeft } from "lucide-svelte";
    import { Button, Card, Badge, ScrollArea } from "$lib/components/ui";
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

    // Make groupedEvents reactive to events changes
    let groupedEvents: ReturnType<typeof groupEventsByStatus>;

    // Reactive statement that updates when events change
    $: groupedEvents = groupEventsByStatus(events);

    // Reactive function that updates when groupedEvents changes
    $: getEventsForStatus = (status: string): ParsedEvent[] => {
        if (!groupedEvents) return [];
        const key = status.toLowerCase();
        switch (key) {
            case "backlogs":
                return groupedEvents.backlogs || [];
            case "proposed":
                return groupedEvents.proposed || [];
            case "upcoming":
                return groupedEvents.upcoming || [];
            case "release":
                return groupedEvents.release || [];
            case "archived":
                return groupedEvents.archived || [];
            default:
                return [];
        }
    };

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

            events = events.map((e) =>
                e.id === eventId ? parseEvent(updatedEvent) : e,
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

        // Use stored drag data instead of dataTransfer
        if (!draggedEventId) {
            return;
        }

        // Find the event to get its current status
        const sourceEvent = events.find((e) => e.id === draggedEventId);
        if (!sourceEvent) {
            return;
        }

        // Only change status if it's different
        if (sourceEvent.status !== newStatus) {
            handleStatusChange(draggedEventId, newStatus);
        }

        // Don't clear drag data here - let dragend handler do it with delay
        // This prevents race conditions between drop and dragend events
    }

    function getEventStatus(eventId: number): string | null {
        const event = events.find((e) => e.id === eventId);
        return event ? event.status : null;
    }

    async function handleMoveUp(event: ParsedEvent) {
        try {
            const statusEvents = getEventsForStatus(event.status);
            const currentIndex = statusEvents.findIndex(
                (e) => e.id === event.id,
            );

            if (currentIndex <= 0) {
                return;
            }

            // Add loading state to show fading animation
            loading = true;

            // Move up means moving to previous position (currentIndex - 1)
            const newOrder = currentIndex - 1;

            await api.reorderEvent(event.id, newOrder, event.status);

            // Reload events to get updated order from backend
            await loadEvents();
        } catch (err) {
            error = `Failed to move event up: ${(err as any)?.message || err}`;
        } finally {
            loading = false;
        }
    }

    async function handleMoveDown(event: ParsedEvent) {
        try {
            const statusEvents = getEventsForStatus(event.status);
            const currentIndex = statusEvents.findIndex(
                (e) => e.id === event.id,
            );

            if (currentIndex >= statusEvents.length - 1) {
                return;
            }

            // Add loading state to show fading animation
            loading = true;

            // Move down means moving to next position (currentIndex + 1)
            const newOrder = currentIndex + 1;

            await api.reorderEvent(event.id, newOrder, event.status);

            // Reload events to get updated order from backend
            await loadEvents();
        } catch (err) {
            error = `Failed to move event down: ${(err as any)?.message || err}`;
        } finally {
            loading = false;
        }
    }

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

    async function handleBacklogReorder(reorderedEvents: ParsedEvent[]) {
        try {
            // Update order values for each event based on their new position
            for (let i = 0; i < reorderedEvents.length; i++) {
                const event = reorderedEvents[i];
                try {
                    await api.reorderEvent(event.id, i, "Backlogs");
                } catch (apiErr) {
                    console.error(
                        `Failed to reorder event ${event.id}:`,
                        apiErr,
                    );
                    throw new Error("Failed to save new order to backend");
                }
            }

            // Reload events to get updated order from backend
            await loadEvents();
        } catch (err) {
            error =
                err instanceof Error ? err.message : "Failed to reorder events";
            // Revert on error by reloading events
            await loadEvents();
        }
    }

    async function handleArchivedReorder(reorderedEvents: ParsedEvent[]) {
        try {
            // Update order values for each event based on their new position
            for (let i = 0; i < reorderedEvents.length; i++) {
                const event = reorderedEvents[i];
                try {
                    await api.reorderEvent(event.id, i, "Archived");
                } catch (apiErr) {
                    console.error(
                        `Failed to reorder event ${event.id}:`,
                        apiErr,
                    );
                    throw new Error("Failed to save new order to backend");
                }
            }

            // Reload events to get updated order from backend
            await loadEvents();
        } catch (err) {
            error =
                err instanceof Error ? err.message : "Failed to reorder events";
            // Revert on error by reloading events
            await loadEvents();
        }
    }
</script>

<svelte:head>
    <title>Manage Events - Admin</title>
</svelte:head>

<div class="w-full">
    <!-- Header -->
    <div class="flex items-center justify-between mb-4">
        <div>
            <h1 class="text-xl font-semibold mb-1">Events</h1>
            <p class="text-muted-foreground text-sm">
                Drag and drop to organize
            </p>
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
                                            {getEventsForStatus(column.status)
                                                .length}
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
                                            on:dragover={(e) => {
                                                e.preventDefault();
                                                if (
                                                    draggedEventId &&
                                                    draggedEventStatus ===
                                                        column.status
                                                ) {
                                                    e.dataTransfer.dropEffect =
                                                        "move";
                                                }
                                            }}
                                            on:drop={(e) => {
                                                e.preventDefault();
                                                if (
                                                    draggedEventId &&
                                                    draggedEventStatus ===
                                                        column.status
                                                ) {
                                                    const firstEvent =
                                                        getEventsForStatus(
                                                            column.status,
                                                        )[0];
                                                    if (firstEvent) {
                                                        handleDrop(
                                                            e,
                                                            column.status,
                                                        );
                                                    }
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
                                                    on:moveUp={(e) =>
                                                        handleMoveUp(e.detail)}
                                                    on:moveDown={(e) =>
                                                        handleMoveDown(
                                                            e.detail,
                                                        )}
                                                    on:dragstart={(e) => {
                                                        draggedEventId =
                                                            e.detail.eventId;
                                                        draggedEventStatus =
                                                            e.detail
                                                                .sourceStatus;
                                                    }}
                                                    isBeingDragged={draggedEventId ===
                                                        event.id}
                                                    on:dragend={() => {
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
                                    {(groupedEvents["backlogs"] || []).length}
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
                                    {(groupedEvents["archived"] || []).length}
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
                                        events={groupedEvents["backlogs"] || []}
                                        {loading}
                                        on:edit={(e) => openEditModal(e.detail)}
                                        on:delete={(e) =>
                                            handleDelete(e.detail)}
                                        on:reorder={(e) =>
                                            handleBacklogReorder(e.detail)}
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
                                    events={groupedEvents["archived"] || []}
                                    {loading}
                                    on:edit={(e) => openEditModal(e.detail)}
                                    on:delete={(e) => handleDelete(e.detail)}
                                    on:statusChange={(e) =>
                                        handleStatusChange(
                                            e.detail.eventId,
                                            e.detail.newStatus,
                                        )}
                                    on:reorder={(e) =>
                                        handleArchivedReorder(e.detail)}
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
