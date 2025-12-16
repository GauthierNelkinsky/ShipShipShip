<script lang="ts">
    import { createEventDispatcher } from "svelte";
    import type { ParsedEvent } from "$lib/types";
    import { Plus, Link2, Unlink, Settings, Eye } from "lucide-svelte";
    import { Button } from "$lib/components/ui";
    import * as m from "$lib/paraglide/messages";
    import ListCard from "$lib/components/ListCard.svelte";
    import { flip } from "svelte/animate";

    const dispatch = createEventDispatcher();

    type StatusCategoryMapping = {
        categoryId: string | null;
        categoryLabel: string;
        categoryDescription: string;
    };

    type StatusDefinition = {
        id: number;
        display_name: string;
        order: number;
        is_reserved: boolean;
    };

    // Type definitions
    type Column = {
        status: string;
        label: string;
    };

    // Props
    export let columns: Column[] = [];
    export let statuses: StatusDefinition[] = [];
    export let hiddenStatuses: Set<string> = new Set();
    export let getEventsForStatus: (status: string) => ParsedEvent[];
    export let countForStatus: (status: string) => number;
    export let statusCategoryMap: Map<string, StatusCategoryMapping> =
        new Map();

    // Drag and drop state
    let draggedEventId: number | null = null;
    let draggedEventStatus: string | null = null;
    let dragOverStatus: string | null = null;
    let draggedStatusIndex: number | null = null;
    let dragOverStatusIndex: number | null = null;
    let dropPosition: "before" | "after" | null = null;

    // Auto-scroll state
    let scrollInterval: number | null = null;

    // Status editing state
    export let editingStatusId: number | null = null;
    export let editingStatusName: string = "";
    export let editingInputEl: HTMLInputElement | null = null;

    // Status creation state
    export let isCreatingNewStatus: boolean = false;
    export let newStatusName: string = "";
    export let newStatusInputEl: HTMLInputElement | null = null;

    // Click outside handler for new status creation
    let isCreationJustStarted = false;

    function handleClickOutside(event: MouseEvent) {
        if (isCreatingNewStatus && !isCreationJustStarted) {
            const target = event.target as HTMLElement;
            const newStatusSection = target.closest(
                "[data-new-status-section]",
            );
            if (!newStatusSection) {
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

    function handleOpenEditModal(event: ParsedEvent) {
        dispatch("openEditModal", { event });
    }

    function handleInitiateDeleteEvent(event: ParsedEvent) {
        dispatch("initiateDeleteEvent", { event });
    }

    function handleCreateNewStatus() {
        dispatch("createNewStatus");
    }

    function handleCancelCreatingNewStatus() {
        dispatch("cancelCreatingNewStatus");
    }

    function handleStartEditingStatus(statusName: string) {
        dispatch("startEditingStatus", { statusName });
    }

    function handleInitiateDeleteStatus(status: StatusDefinition) {
        dispatch("initiateDeleteStatus", { status });
    }

    function handleCommitEditingStatus() {
        dispatch("commitEditingStatus");
    }

    function handleCancelEditingStatus() {
        dispatch("cancelEditingStatus");
    }

    // Helper function to find nearest status section
    function findNearestStatus(y: number): string | null {
        const statusElements = document.querySelectorAll("[data-list-status]");
        let nearestStatus: string | null = null;
        let nearestDistance = Infinity;

        statusElements.forEach((el) => {
            const rect = el.getBoundingClientRect();
            const statusCenter = rect.top + rect.height / 2;
            const distance = Math.abs(y - statusCenter);

            if (distance < nearestDistance) {
                nearestDistance = distance;
                nearestStatus = el.getAttribute("data-list-status");
            }
        });

        return nearestStatus;
    }

    function startAutoScroll(direction: "up" | "down") {
        if (scrollInterval !== null) return;

        scrollInterval = window.setInterval(() => {
            const scrollAmount = direction === "down" ? 15 : -15;
            window.scrollBy(0, scrollAmount);
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
        if (draggedEventId || draggedStatusIndex !== null) {
            const scrollThreshold = 150; // pixels from viewport edge
            const viewportHeight = window.innerHeight;
            const distanceFromBottom = viewportHeight - e.clientY;
            const distanceFromTop = e.clientY;

            if (distanceFromBottom < scrollThreshold) {
                startAutoScroll("down");
            } else if (distanceFromTop < scrollThreshold) {
                startAutoScroll("up");
            } else {
                stopAutoScroll();
            }
        }

        if (draggedEventId && !draggedStatusIndex) {
            const nearestStatus = findNearestStatus(e.clientY);
            if (nearestStatus && nearestStatus !== dragOverStatus) {
                dragOverStatus = nearestStatus;
            }
        } else if (draggedStatusIndex !== null) {
            // Clear event drop indicator when dragging statuses
            dragOverStatus = null;

            // Calculate which status section we're over for status reordering
            const nearestStatus = findNearestStatus(e.clientY);
            if (nearestStatus) {
                const targetIndex = visibleColumns.findIndex(
                    (col) => col.status === nearestStatus,
                );
                if (targetIndex !== -1) {
                    handleStatusDragOver(e, targetIndex);
                }
            }
        }
    }

    function handleContainerDrop(e: DragEvent) {
        e.preventDefault();
        stopAutoScroll();

        if (draggedEventId && !draggedStatusIndex) {
            const nearestStatus = findNearestStatus(e.clientY);
            if (nearestStatus) {
                handleCardDrop(e, nearestStatus);
            }
        } else if (draggedStatusIndex !== null) {
            // Handle status drop in dead zones
            const nearestStatus = findNearestStatus(e.clientY);
            if (nearestStatus) {
                const targetIndex = visibleColumns.findIndex(
                    (col) => col.status === nearestStatus,
                );
                if (targetIndex !== -1) {
                    handleStatusDrop(e, targetIndex);
                }
            }
        }
    }

    // Drag and drop handlers - Event cards
    function handleCardDragStart(event: ParsedEvent, sourceStatus: string) {
        draggedEventId = event.id;
        draggedEventStatus = sourceStatus;
    }

    function handleCardDragEnd() {
        stopAutoScroll();
        setTimeout(() => {
            draggedEventId = null;
            draggedEventStatus = null;
            dragOverStatus = null;
        }, 100);
    }

    function handleCardDrop(e: DragEvent, targetStatus: string) {
        e.preventDefault();
        if (draggedEventId && draggedEventStatus) {
            dispatch("statusChange", {
                eventId: draggedEventId,
                sourceStatus: draggedEventStatus,
                targetStatus,
            });
        }
        dragOverStatus = null;
    }

    function handleCardDragOver(e: DragEvent) {
        e.preventDefault();
    }

    function handleCardDragEnter(status: string) {
        dragOverStatus = status;
    }

    function handleCardDragLeave() {
        dragOverStatus = null;
    }

    // Drag and drop handlers - Status sections
    function handleStatusDragStart(e: DragEvent, index: number) {
        draggedStatusIndex = index;
        dragOverStatus = null; // Clear event drop indicator
        if (e.dataTransfer) {
            e.dataTransfer.effectAllowed = "move";
        }
    }

    function handleStatusDragOver(e: DragEvent, index: number) {
        e.preventDefault();
        if (draggedStatusIndex === null || draggedStatusIndex === index) {
            dragOverStatusIndex = null;
            dropPosition = null;
            return;
        }

        const rect = (e.currentTarget as HTMLElement).getBoundingClientRect();
        const midpoint = rect.top + rect.height / 2;

        if (e.clientY < midpoint) {
            dropPosition = "before";
        } else {
            dropPosition = "after";
        }
        dragOverStatusIndex = index;
    }

    function handleStatusDrop(e: DragEvent, targetIndex: number) {
        e.preventDefault();
        if (draggedStatusIndex === null || draggedStatusIndex === targetIndex)
            return;

        const sourceStatus = visibleColumns[draggedStatusIndex].status;
        const targetStatus = visibleColumns[targetIndex].status;

        dispatch("reorderStatuses", {
            sourceStatus,
            targetStatus,
            dropPosition,
        });

        draggedStatusIndex = null;
        dragOverStatusIndex = null;
        dropPosition = null;
    }

    function handleStatusDragEnd() {
        stopAutoScroll();
        draggedStatusIndex = null;
        dragOverStatusIndex = null;
        dropPosition = null;
    }

    // Filter out hidden statuses
    $: visibleColumns = columns.filter(
        (col) => !hiddenStatuses.has(col.status),
    );
</script>

<svelte:window on:click={handleClickOutside} />

<div class="w-full">
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
                <Plus class="h-4 w-4 mr-2" />
                {m.events_page_create_first_status()}
            </Button>
        </div>
    {:else}
        <div
            class="pb-4"
            on:dragover={handleContainerDragOver}
            on:drop={handleContainerDrop}
            on:dragleave={() => stopAutoScroll()}
            role="list"
        >
            <!-- New Status Creation Section -->
            {#if isCreatingNewStatus}
                <div class="mb-4" data-new-status-section>
                    <div
                        class="group/section space-y-3 relative"
                        role="listitem"
                    >
                        <!-- Status Header -->
                        <div
                            class="flex items-center justify-between pb-2 border-b border-primary"
                        >
                            <div class="flex items-center gap-2 flex-1 min-w-0">
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

                        <!-- Events List -->
                        <div
                            class="rounded-lg border-2 border-dashed border-primary bg-primary/5 py-16"
                        >
                            <div
                                class="flex items-center justify-center text-xs text-primary/70"
                            >
                                {m.events_page_press_enter_to_create()}
                            </div>
                        </div>
                    </div>
                </div>
            {/if}

            {#each visibleColumns as column, index (column.status)}
                {@const events = getEventsForStatus(column.status)}
                {@const categoryMapping = statusCategoryMap.get(column.status)}

                <div
                    class="mb-4"
                    style="will-change: transform;"
                    animate:flip={{ duration: 300 }}
                >
                    <!-- Drop indicator before first status -->
                    {#if draggedStatusIndex !== null && dragOverStatusIndex === index && dropPosition === "before" && index === 0}
                        <div
                            class="h-0.5 bg-primary rounded-full pointer-events-none mb-2"
                        ></div>
                    {/if}

                    <!-- Drop indicator between statuses -->
                    {#if draggedStatusIndex !== null && dragOverStatusIndex === index && dropPosition === "before" && index > 0}
                        <div
                            class="h-0.5 bg-primary rounded-full pointer-events-none my-2"
                        ></div>
                    {/if}

                    <div
                        class="group/section space-y-3 relative {draggedStatusIndex ===
                        index
                            ? 'opacity-50'
                            : ''}"
                        role="listitem"
                        data-list-status={column.status}
                        draggable={!draggedEventId &&
                            !(
                                editingStatusId !== null &&
                                statuses.find((s) => s.id === editingStatusId)
                                    ?.display_name === column.status
                            )}
                        on:dragstart={(e) => handleStatusDragStart(e, index)}
                        on:dragover={(e) => handleStatusDragOver(e, index)}
                        on:drop={(e) => handleStatusDrop(e, index)}
                        on:dragend={handleStatusDragEnd}
                    >
                        <!-- Status Header -->
                        <div
                            class="flex items-center justify-between pb-2 border-b {editingStatusId !==
                                null &&
                            statuses.find((s) => s.id === editingStatusId)
                                ?.display_name === column.status
                                ? 'cursor-auto'
                                : 'cursor-move'}"
                        >
                            <div class="flex items-center gap-2 flex-1 min-w-0">
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
                                        placeholder={m.events_page_status_name_placeholder()}
                                        maxlength="50"
                                    />
                                {:else}
                                    <span class="text-sm font-medium">
                                        {column.label}
                                    </span>
                                    <button
                                        type="button"
                                        class="h-5 w-5 flex items-center justify-center text-muted-foreground hover:text-foreground opacity-0 group-hover/section:opacity-100 transition-opacity"
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
                                        on:mousedown={(e) => e.preventDefault()}
                                        on:click={() =>
                                            handleInitiateDeleteStatus(
                                                currentStatus,
                                            )}
                                    >
                                        {m.events_page_delete()}
                                    </button>
                                {:else}
                                    <!-- Show "Add event" button, Settings icon, Link icon and count when not editing -->
                                    <Button
                                        variant="ghost"
                                        size="sm"
                                        class="h-7 px-2 text-xs opacity-0 group-hover/section:opacity-100 transition-opacity"
                                        on:click={() =>
                                            handleOpenCreateModal(
                                                column.status,
                                            )}
                                    >
                                        <Plus class="h-3.5 w-3.5 mr-1" />
                                        {m.events_page_new_event()}
                                    </Button>
                                    {#if statuses.find((s) => s.display_name === column.status)}
                                        <button
                                            type="button"
                                            class="h-6 w-6 flex items-center justify-center text-muted-foreground hover:text-foreground opacity-0 group-hover/section:opacity-100 transition-opacity rounded hover:bg-muted"
                                            on:click={() =>
                                                handleStartEditingStatus(
                                                    column.status,
                                                )}
                                            title={m.events_page_settings()}
                                        >
                                            <Settings class="h-4 w-4" />
                                        </button>
                                    {/if}
                                    {#if categoryMapping}
                                        <div
                                            class="relative group/mapping"
                                            title={categoryMapping.categoryId
                                                ? m.events_page_linked_to_theme_title(
                                                      {
                                                          categoryLabel:
                                                              categoryMapping.categoryLabel,
                                                      },
                                                  )
                                                : m.events_page_not_linked_to_theme()}
                                        >
                                            {#if categoryMapping.categoryId}
                                                <Link2
                                                    class="h-3.5 w-3.5 text-blue-600 dark:text-blue-500"
                                                />
                                            {:else}
                                                <Unlink
                                                    class="h-3.5 w-3.5 text-muted-foreground/50"
                                                />
                                            {/if}
                                            <div
                                                class="absolute right-0 top-full mt-2 w-64 p-3 rounded-lg border bg-popover text-popover-foreground shadow-lg opacity-0 invisible group-hover/mapping:opacity-100 group-hover/mapping:visible transition-all z-50 pointer-events-none"
                                            >
                                                {#if categoryMapping.categoryId}
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
                                                    <div class="text-xs mb-1">
                                                        <strong
                                                            >{categoryMapping.categoryLabel}</strong
                                                        >
                                                    </div>
                                                    {#if categoryMapping.categoryDescription}
                                                        <div
                                                            class="text-xs text-muted-foreground"
                                                        >
                                                            {categoryMapping.categoryDescription}
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

                        <!-- Events List -->
                        <div
                            class={dragOverStatus === column.status &&
                            draggedEventStatus !== column.status
                                ? "bg-primary/10 rounded-lg border-2 border-primary border-dashed"
                                : ""}
                            role="region"
                            on:drop={(e) => handleCardDrop(e, column.status)}
                            on:dragover={handleCardDragOver}
                            on:dragenter={() =>
                                handleCardDragEnter(column.status)}
                            on:dragleave={handleCardDragLeave}
                        >
                            {#if events.length === 0}
                                <div
                                    class="py-6 text-center text-xs text-muted-foreground"
                                >
                                    {m.events_page_no_events()}
                                </div>
                            {:else}
                                <div class="space-y-1.5">
                                    {#each events as event (event.id)}
                                        <div
                                            class="border rounded-lg bg-background hover:shadow-md hover:scale-[1.01] transition-all duration-200 transform {draggedEventId ===
                                            event.id
                                                ? 'opacity-50'
                                                : ''}"
                                            animate:flip={{ duration: 300 }}
                                        >
                                            <ListCard
                                                {event}
                                                sourceStatus={column.status}
                                                on:edit={(e) =>
                                                    handleOpenEditModal(
                                                        e.detail,
                                                    )}
                                                on:delete={(e) =>
                                                    handleInitiateDeleteEvent(
                                                        e.detail,
                                                    )}
                                                on:dragstart={() =>
                                                    handleCardDragStart(
                                                        event,
                                                        column.status,
                                                    )}
                                                on:dragend={handleCardDragEnd}
                                            />
                                        </div>
                                    {/each}
                                </div>
                            {/if}
                        </div>
                    </div>

                    <!-- Drop indicator after last status -->
                    {#if draggedStatusIndex !== null && dragOverStatusIndex === index && dropPosition === "after"}
                        <div
                            class="h-0.5 bg-primary rounded-full pointer-events-none my-2"
                        ></div>
                    {/if}
                </div>
            {/each}
        </div>
    {/if}
</div>
