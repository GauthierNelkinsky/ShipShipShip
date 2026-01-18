<script lang="ts">
    // KanbanView component - displays events in a kanban board layout
    import { createEventDispatcher } from "svelte";
    import { fly } from "svelte/transition";
    import { flip } from "svelte/animate";
    import type { ParsedEvent } from "$lib/types";
    import { Plus, Settings, Link2, Unlink, Eye } from "lucide-svelte";
    import { Button } from "$lib/components/ui";
    import KanbanCard from "$lib/components/KanbanCard.svelte";
    import * as m from "$lib/paraglide/messages";

    const dispatch = createEventDispatcher();

    // Type definitions
    type StatusDefinition = {
        id: number;
        display_name: string;
        order: number;
        is_reserved: boolean;
    };

    type Column = {
        status: string;
        label: string;
    };

    type StatusCategoryMapping = {
        categoryId: string | null;
        categoryLabel: string;
        categoryDescription: string;
    };

    // Props
    export let columns: Column[] = [];
    export let statuses: StatusDefinition[] = [];
    export let hiddenStatuses: Set<string> = new Set();
    export let statusCategoryMap: Map<string, StatusCategoryMapping> =
        new Map();
    export let getEventsForStatus: (status: string) => ParsedEvent[];
    export let countForStatus: (status: string) => number;

    // Drag and drop state
    export let draggedEventId: number | null = null;
    export let draggedEventStatus: string | null = null;
    export let dragOverColumn: string | null = null;
    export let draggedColumnStatus: string | null = null;
    export let dragOverColumnIndex: number | null = null;
    export let dropPosition: "before" | "after" | null = null;

    // Auto-scroll state
    let scrollInterval: number | null = null;
    let kanbanScrollContainer: HTMLElement | null = null;

    // Status editing state
    export let isCreatingNewStatus: boolean = false;
    export let newStatusName: string = "";
    export let newStatusInputEl: HTMLInputElement | null = null;
    export let editingStatusId: number | null = null;
    export let editingStatusName: string = "";
    export let editingInputEl: HTMLInputElement | null = null;

    // Click outside handler for new status creation
    let isCreationJustStarted = false;

    function handleClickOutside(event: MouseEvent) {
        if (isCreatingNewStatus && !isCreationJustStarted) {
            const target = event.target as HTMLElement;
            const newStatusColumn = target.closest("[data-new-status-column]");
            if (!newStatusColumn) {
                handleCancelCreatingNewStatus();
            }
        }
        isCreationJustStarted = false;
    }

    // Watch for when creation starts
    $: if (isCreatingNewStatus) {
        isCreationJustStarted = true;
    }

    // Event handlers
    function handleOpenCreateModal(status: string) {
        dispatch("openCreateModal", { status });
    }

    function handleToggleStatusVisibility(statusName: string) {
        dispatch("toggleStatusVisibility", { statusName });
    }

    function handleOpenEditModal(event: CustomEvent) {
        dispatch("openEditModal", { event: event.detail });
    }

    function handleInitiateDeleteEvent(event: CustomEvent) {
        dispatch("initiateDeleteEvent", { event: event.detail });
    }

    function handleStatusChange(event: CustomEvent) {
        dispatch("statusChange", event.detail);
    }

    function handleCardDragStart(event: CustomEvent) {
        draggedEventId = event.detail.eventId;
        draggedEventStatus = event.detail.sourceStatus;
    }

    function handleCardDragEnd() {
        stopAutoScroll();
        setTimeout(() => {
            draggedEventId = null;
            draggedEventStatus = null;
        }, 100);
    }

    function handleInitiateDeleteStatus(status: StatusDefinition) {
        dispatch("initiateDeleteStatus", { status });
    }

    function handleStartEditingStatus(statusName: string) {
        dispatch("startEditingStatus", { statusName });
    }

    function handleCommitEditingStatus() {
        dispatch("commitEditingStatus");
    }

    function handleCancelEditingStatus() {
        dispatch("cancelEditingStatus");
    }

    function handleCreateNewStatus() {
        dispatch("createNewStatus");
    }

    function handleCancelCreatingNewStatus() {
        dispatch("cancelCreatingNewStatus");
    }

    function handleDrop(e: DragEvent, status: string) {
        dispatch("drop", { event: e, status });
    }

    function findNearestColumn(x: number): string | null {
        const columnElements = document.querySelectorAll(
            "[data-kanban-column]",
        );
        let nearestColumn: string | null = null;
        let nearestDistance = Infinity;

        columnElements.forEach((el) => {
            const rect = el.getBoundingClientRect();
            const columnCenter = rect.left + rect.width / 2;
            const distance = Math.abs(x - columnCenter);

            if (distance < nearestDistance) {
                nearestDistance = distance;
                nearestColumn = el.getAttribute("data-kanban-column");
            }
        });

        return nearestColumn;
    }

    function startAutoScroll(direction: "left" | "right") {
        if (scrollInterval !== null) return;

        scrollInterval = window.setInterval(() => {
            if (!kanbanScrollContainer) return;

            const scrollAmount = direction === "right" ? 15 : -15;
            kanbanScrollContainer.scrollLeft += scrollAmount;
        }, 16) as unknown as number; // ~60fps
    }

    function stopAutoScroll() {
        if (scrollInterval !== null) {
            clearInterval(scrollInterval);
            scrollInterval = null;
        }
    }

    function handleContainerDragOver(e: DragEvent) {
        e.preventDefault();

        // Auto-scroll detection using viewport edges
        if (draggedEventId || draggedColumnStatus) {
            const scrollThreshold = 150; // pixels from viewport edge
            const viewportWidth = window.innerWidth;
            const distanceFromRight = viewportWidth - e.clientX;
            const distanceFromLeft = e.clientX;

            if (distanceFromRight < scrollThreshold) {
                startAutoScroll("right");
            } else if (distanceFromLeft < scrollThreshold) {
                startAutoScroll("left");
            } else {
                stopAutoScroll();
            }
        }

        if (draggedEventId && !draggedColumnStatus) {
            const nearestStatus = findNearestColumn(e.clientX);
            if (nearestStatus) {
                dragOverColumn = nearestStatus;
            }
        } else if (draggedColumnStatus) {
            const nearestStatus = findNearestColumn(e.clientX);
            if (nearestStatus && nearestStatus !== draggedColumnStatus) {
                const targetIndex = visibleColumns.findIndex(
                    (col) => col.status === nearestStatus,
                );
                if (targetIndex !== -1) {
                    handleColumnDragOver(e, targetIndex);
                }
            }
        }
    }

    function handleContainerDrop(e: DragEvent) {
        e.preventDefault();
        if (draggedEventId && !draggedColumnStatus) {
            const nearestStatus = findNearestColumn(e.clientX);
            if (nearestStatus) {
                handleDrop(e, nearestStatus);
            }
        } else if (draggedColumnStatus) {
            const nearestStatus = findNearestColumn(e.clientX);
            const targetIndex = visibleColumns.findIndex(
                (col) => col.status === nearestStatus,
            );
            if (targetIndex !== -1) {
                handleColumnDrop(e, targetIndex);
            }
        }
    }

    function handleDragOver(e: DragEvent, status: string) {
        dispatch("dragover", { event: e, status });
    }

    function handleDragEnter(e: DragEvent) {
        dispatch("dragenter", { event: e });
    }

    function handleDragLeave() {
        dispatch("dragleave");
    }

    function handleColumnDragStart(e: DragEvent, status: string) {
        dispatch("columnDragStart", { event: e, status });
    }

    function handleColumnDragOver(e: DragEvent, index: number) {
        dispatch("columnDragOver", { event: e, index });
    }

    function handleColumnDragEnter(index: number) {
        dispatch("columnDragEnter", { index });
    }

    function handleColumnDragLeave() {
        dispatch("columnDragLeave");
    }

    function handleColumnDrop(e: DragEvent, index: number) {
        const targetStatus = visibleColumns[index]?.status;
        dispatch("columnDrop", { event: e, targetStatus });
    }

    function handleColumnDragEnd() {
        stopAutoScroll();
        dispatch("columnDragEnd");
    }

    // Filter out hidden statuses
    $: visibleColumns = columns.filter(
        (col) => !hiddenStatuses.has(col.status),
    );
</script>

<svelte:window on:click={handleClickOutside} />

<div class="w-full overflow-x-auto" bind:this={kanbanScrollContainer}>
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
                on:click={() => {
                    isCreatingNewStatus = true;
                }}
                class="mt-2"
            >
                <Plus class="h-4 w-4 me-2" />
                {m.events_page_create_first_status()}
            </Button>
        </div>
    {:else}
        <div
            class="flex gap-2 lg:gap-4 min-h-0 pb-3 w-max"
            role="region"
            aria-label="Kanban columns"
            on:dragover={handleContainerDragOver}
            on:drop={handleContainerDrop}
            on:dragleave={() => stopAutoScroll()}
        >
            <!-- Drop indicator for before first column -->
            {#if draggedColumnStatus && dragOverColumnIndex === 0 && dropPosition === "before"}
                <div class="w-1 bg-primary/50 rounded-full"></div>
            {/if}

            <!-- New Status Creation Column -->
            {#if isCreatingNewStatus}
                <div class="flex-shrink-0 w-[320px]" data-new-status-column>
                    <!-- Column Header -->
                    <div class="mb-3">
                        <div class="flex items-center justify-between">
                            <div class="flex items-center gap-1.5 flex-1">
                                <input
                                    bind:this={newStatusInputEl}
                                    bind:value={newStatusName}
                                    type="text"
                                    placeholder={m.events_page_status_name_placeholder()}
                                    class="flex-1 bg-transparent border-none outline-none text-sm font-medium border-b border-primary/50 pb-0.5 rounded-none focus:border-primary transition-colors"
                                    on:keydown={(e) => {
                                        if (e.key === "Enter") {
                                            handleCreateNewStatus();
                                        } else if (e.key === "Escape") {
                                            handleCancelCreatingNewStatus();
                                        }
                                    }}
                                    on:blur={handleCreateNewStatus}
                                />
                            </div>
                        </div>
                    </div>

                    <!-- Column Content -->
                    <div
                        class="rounded-lg border-2 border-dashed border-primary bg-primary/5 overflow-hidden flex flex-col"
                        style="height: calc(100vh - 280px);"
                    >
                        <div class="flex-1 flex items-center justify-center">
                            <div class="text-xs text-primary/70">
                                {m.events_page_press_enter_to_create()}
                            </div>
                        </div>
                    </div>
                </div>
            {/if}

            {#each visibleColumns as column, index (column.status)}
                <div
                    class="flex gap-2 lg:gap-4"
                    animate:flip={{ duration: 300 }}
                >
                    <!-- Drop indicator before (not first) -->
                    {#if draggedColumnStatus && dragOverColumnIndex === index && dropPosition === "before" && index > 0}
                        <div
                            class="w-1 bg-primary rounded-full self-stretch z-10"
                            style="height: calc(100vh - 280px);"
                        ></div>
                    {/if}

                    <div
                        class="flex-shrink-0 w-[320px] group transition-opacity {draggedColumnStatus ===
                        column.status
                            ? 'opacity-50'
                            : ''}"
                        data-kanban-column={column.status}
                        draggable={!draggedEventId &&
                            !(
                                editingStatusId !== null &&
                                statuses.find((s) => s.id === editingStatusId)
                                    ?.display_name === column.status
                            )}
                        on:dragstart={(e) => {
                            if (!draggedEventId) {
                                handleColumnDragStart(e, column.status);
                            }
                        }}
                        on:dragover={(e) => {
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
                        <div
                            class="mb-3 {editingStatusId !== null &&
                            statuses.find((s) => s.id === editingStatusId)
                                ?.display_name === column.status
                                ? 'cursor-auto'
                                : 'cursor-move'}"
                        >
                            <div class="flex items-center justify-between">
                                <div class="flex items-center gap-1.5 flex-1">
                                    {#if editingStatusId !== null && statuses.find((s) => s.id === editingStatusId)?.display_name === column.status}
                                        <input
                                            class="flex-1 bg-transparent border-none outline-none text-sm font-medium border-b border-primary/50 pb-0.5 rounded-none focus:border-primary transition-colors"
                                            bind:this={editingInputEl}
                                            bind:value={editingStatusName}
                                            on:keydown={(e) => {
                                                if (e.key === "Enter") {
                                                    handleCommitEditingStatus();
                                                } else if (e.key === "Escape") {
                                                    handleCancelEditingStatus();
                                                }
                                            }}
                                            on:blur={handleCommitEditingStatus}
                                            on:mousedown={(e) =>
                                                e.stopPropagation()}
                                            on:click={(e) =>
                                                e.stopPropagation()}
                                            placeholder={m.events_page_status_name_placeholder()}
                                            maxlength="50"
                                        />
                                    {:else}
                                        <span class="text-sm font-medium">
                                            {column.label}
                                        </span>
                                        <button
                                            type="button"
                                            class="h-5 w-5 flex items-center justify-center text-muted-foreground hover:text-foreground opacity-0 group-hover:opacity-100 transition-opacity"
                                            on:click={() =>
                                                handleToggleStatusVisibility(
                                                    column.status,
                                                )}
                                            title={m.events_page_hide_status()}
                                        >
                                            <Eye class="h-3.5 w-3.5" />
                                        </button>
                                    {/if}
                                </div>
                                <div class="flex items-center gap-2">
                                    {#if editingStatusId !== null && statuses.find((s) => s.id === editingStatusId)?.display_name === column.status}
                                        <!-- Show Delete button when editing -->
                                        {@const currentStatus = statuses.find(
                                            (s) => s.id === editingStatusId,
                                        )!}
                                        <button
                                            type="button"
                                            class="text-xs font-medium px-2 py-0.5 rounded transition-colors {statuses.length <=
                                            1
                                                ? 'text-muted-foreground/30 cursor-not-allowed'
                                                : 'text-red-500 hover:text-red-600 hover:bg-red-50 dark:hover:bg-red-950/50'}"
                                            disabled={statuses.length <= 1}
                                            title={statuses.length <= 1
                                                ? m.events_page_cannot_delete_last_status()
                                                : m.events_page_delete_status_tooltip()}
                                            on:mousedown={(e) =>
                                                e.preventDefault()}
                                            on:click={() =>
                                                handleInitiateDeleteStatus(
                                                    currentStatus,
                                                )}
                                        >
                                            {m.events_page_delete()}
                                        </button>
                                    {:else}
                                        <!-- Show Settings icon, Link icon and count when not editing -->
                                        {#if statuses.find((s) => s.display_name === column.status)}
                                            <button
                                                type="button"
                                                class="h-6 w-6 flex items-center justify-center text-muted-foreground hover:text-foreground opacity-0 group-hover:opacity-100 transition-opacity rounded hover:bg-muted"
                                                on:click={() =>
                                                    handleStartEditingStatus(
                                                        column.status,
                                                    )}
                                                title={m.events_page_settings()}
                                            >
                                                <Settings class="h-4 w-4" />
                                            </button>
                                        {/if}
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
                                                <div
                                                    class="absolute end-0 top-full mt-2 w-64 p-3 rounded-lg border bg-popover text-popover-foreground shadow-lg opacity-0 invisible group-hover/mapping:opacity-100 group-hover/mapping:visible transition-all z-50 pointer-events-none"
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
                                    {/if}
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
                                if (draggedEventId && !draggedColumnStatus) {
                                    e.preventDefault();
                                    handleDrop(e, column.status);
                                }
                            }}
                            on:dragover={(e) => {
                                if (draggedEventId && !draggedColumnStatus) {
                                    e.preventDefault();
                                    handleDragOver(e, column.status);
                                    dragOverColumn = column.status;
                                }
                            }}
                            on:dragenter={(e) => {
                                if (draggedEventId && !draggedColumnStatus) {
                                    e.preventDefault();
                                    handleDragEnter(e);
                                    dragOverColumn = column.status;
                                }
                            }}
                            on:dragleave={() => {
                                if (draggedEventId && !draggedColumnStatus) {
                                    handleDragLeave();
                                }
                            }}
                            role="region"
                            aria-label="Drop zone for {column.label} events"
                        >
                            <div class="flex-1 overflow-y-auto">
                                <div class="space-y-2 p-3 pb-16 min-w-0">
                                    <!-- Drop zone at top of column -->
                                    <div
                                        class="h-2 transition-all duration-200 {draggedEventId &&
                                        draggedEventStatus === column.status &&
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
                                                handleDrop(e, column.status);
                                            }
                                        }}
                                    ></div>

                                    {#each getEventsForStatus(column.status) as event (event.id)}
                                        <div
                                            class="group relative"
                                            in:fly={{ y: 20, duration: 200 }}
                                        >
                                            <KanbanCard
                                                {event}
                                                on:edit={handleOpenEditModal}
                                                on:delete={handleInitiateDeleteEvent}
                                                on:statusChange={handleStatusChange}
                                                on:carddragstart={handleCardDragStart}
                                                isBeingDragged={draggedEventId ===
                                                    event.id}
                                                on:carddragend={handleCardDragEnd}
                                            />
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
                                class="absolute bottom-0 start-0 end-0 p-3 opacity-0 group-hover:opacity-100 translate-y-2 group-hover:translate-y-0 transition-all duration-200 pointer-events-none group-hover:pointer-events-auto"
                            >
                                <button
                                    type="button"
                                    class="w-full py-2 px-3 text-xs font-medium rounded-md bg-background border border-border hover:bg-muted text-foreground transition-all flex items-center justify-center gap-2"
                                    on:click={() =>
                                        handleOpenCreateModal(column.status)}
                                >
                                    <Plus class="h-3.5 w-3.5" />
                                    <span>{m.events_page_new_event()}</span>
                                </button>
                            </div>
                        </div>
                    </div>

                    <!-- Drop indicator after -->
                    {#if draggedColumnStatus && dragOverColumnIndex === index && dropPosition === "after"}
                        <div
                            class="w-1 bg-primary rounded-full self-stretch z-10"
                            style="height: calc(100vh - 280px);"
                        ></div>
                    {/if}
                </div>
            {/each}
        </div>
    {/if}
</div>

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
