<script lang="ts">
    import { createEventDispatcher, onMount } from "svelte";
    import { api } from "$lib/api";
    import * as m from "$lib/paraglide/messages";

    import type {
        CreateEventRequest,
        UpdateEventRequest,
        EventStatus,
        ParsedEvent,
        Tag,
    } from "$lib/types";
    import {
        X,
        Plus,
        Save,
        Mail,
        Send,
        Eye,
        Pencil,
        Calendar,
        ChevronDown,
        Check,
        Loader2,
    } from "lucide-svelte";
    import { fly } from "svelte/transition";
    import { cn } from "$lib/utils";
    import {
        Button,
        Card,
        Input,
        Badge,
        DatePicker,
        Textarea,
    } from "$lib/components/ui";
    import TiptapEditor from "$lib/components/TiptapEditor.svelte";

    import Icon from "@iconify/svelte";

    interface StatusDefinition {
        id: number;
        display_name: string;
        order: number;
        is_reserved: boolean;
    }

    const dispatch = createEventDispatcher();

    let isOpen = false;
    let event: ParsedEvent | null = null;
    let mode: "create" | "edit" = "create";

    let loading = false;
    let error = "";

    // Form fields
    let title = "";
    let status: EventStatus = "";
    let date = "";
    let content = "";
    let tags: number[] = []; // Array of tag IDs
    let media: string[] = [];

    // Tag management
    let availableTags: Tag[] = [];
    let showTagSelector = false;
    let newTagName = "";
    let newTagColor = "#3B82F6"; // Default blue color
    let tagToDelete: number | null = null;
    let deletingTag = false;
    let tagButtonElement: HTMLElement | null = null;
    let tagPopoverTop = 0;
    let tagPopoverLeft = 0;
    let tagSearchTerm = "";

    // Status management
    export let statuses: StatusDefinition[] = [];

    // Status selector
    let statusSelectOpen = false;
    let statusSearchTerm = "";
    let statusButtonElement: HTMLButtonElement;
    let statusDropdownElement: HTMLDivElement;

    // Newsletter management
    let emailSubject = "";
    let emailContent = "";
    let showEmailPreview = false;
    let newsletterLoading = false;
    let newsletterError = "";
    let showNewsletterMode = false;
    let newsletterHistory: Array<{
        id: number;
        event_id: number;
        event_status: string;
        email_subject: string;
        email_template: string;
        subscriber_count: number;
        sent_at: string;
        created_at: string;
    }> = [];
    let historyLoading = false;

    // Computed filtered statuses
    $: filteredStatuses = statuses.filter((s) =>
        s.display_name.toLowerCase().includes(statusSearchTerm.toLowerCase()),
    );

    // Computed filtered tags
    $: filteredTags = availableTags.filter((tag) =>
        tag.name.toLowerCase().includes(tagSearchTerm.toLowerCase()),
    );

    const reactionIcons: Record<string, string> = {
        thumbs_up: "fluent-emoji-flat:thumbs-up",
        heart: "fluent-emoji-flat:red-heart",
        fire: "fluent-emoji-flat:fire",
        party: "fluent-emoji-flat:party-popper",
        eyes: "fluent-emoji-flat:eyes",
        lightbulb: "fluent-emoji-flat:light-bulb",
        thinking: "fluent-emoji-flat:thinking-face",
        thumbs_down: "fluent-emoji-flat:thumbs-down",
    };

    $: visibleReactions = event?.reaction_summary?.reactions
        ? event.reaction_summary.reactions.filter((r: any) => r.count > 0)
        : [];

    // Auto-load newsletter template when event changes in edit mode
    $: if (event && mode === "edit" && showNewsletterMode) {
        loadNewsletterPreview();
    }

    onMount(async () => {
        await loadAvailableTags();
    });

    async function loadNewsletterPreview() {
        if (!event?.id) return;

        try {
            const response = await api.getEventNewsletterPreview(
                event.id,
                "event",
            );
            if (!emailSubject) emailSubject = response.subject;
            emailContent = response.content;
        } catch (err) {
            newsletterError = m.event_modal_newsletter_preview_error();
            console.error("Failed to load newsletter preview:", err);
        }
    }

    async function loadNewsletterHistory() {
        if (!event?.id) return;

        historyLoading = true;
        try {
            const response = await api.getEventEmailHistory(event.id);
            newsletterHistory = response.history || [];
        } catch (err) {
            console.error("Failed to load newsletter history:", err);
            newsletterHistory = [];
        } finally {
            historyLoading = false;
        }
    }

    async function sendNewsletter() {
        if (!event?.id || !emailSubject || !emailContent) return;

        try {
            newsletterLoading = true;
            newsletterError = "";

            await api.sendEventNewsletter(event.id, {
                subject: emailSubject,
                content: emailContent,
                template: "event",
            });

            alert(m.event_modal_newsletter_sent_success());

            // Reload newsletter history to show the newly sent email
            await loadNewsletterHistory();

            emailSubject = "";
            emailContent = "";
            showEmailPreview = false;
            showNewsletterMode = false;
        } catch (err) {
            newsletterError =
                err instanceof Error
                    ? err.message
                    : m.event_modal_newsletter_send_error();
        } finally {
            newsletterLoading = false;
        }
    }

    async function loadAvailableTags() {
        try {
            availableTags = await api.getTags();
        } catch (err) {
            console.error("Failed to load available tags:", err);
        }
    }

    // Reset form when modal opens
    $: if (isOpen && (mode === "create" || (mode === "edit" && event))) {
        resetForm();
    }

    function resetForm() {
        if (mode === "edit" && event) {
            title = event.title;
            status = event.status;
            content = event.content || "";
            tags = event.tags.map((tag) => tag.id);
            media = [...event.media];
            date = event.date || "";
        } else {
            title = "";
            status =
                event?.status ||
                (statuses.length > 0 ? statuses[0].display_name : "");
            date = "";
            content = "";
            tags = [];
            media = [];
        }
        error = "";
        showTagSelector = false;
    }

    let tiptapEditor: any = null;

    function handleContentUpdate(event: CustomEvent) {
        content = event.detail.content;
    }

    function handleEditorReady(event: CustomEvent) {
        tiptapEditor = event.detail.editor;
    }

    function addExistingTag(tagId: number) {
        if (!tags.includes(tagId)) {
            tags = [...tags, tagId];
        }
        showTagSelector = false;
    }

    function handleTagSelectorToggle(e: Event) {
        e.stopPropagation();
        if (!showTagSelector && tagButtonElement) {
            const rect = tagButtonElement.getBoundingClientRect();
            tagPopoverTop = rect.bottom + 4; // 4px margin
            tagPopoverLeft = rect.right - 288 - 8; // 288px (w-72) + 8px margin
        } else if (!showTagSelector) {
            tagSearchTerm = ""; // Reset search when closing
        }
        showTagSelector = !showTagSelector;
    }

    async function createNewTag() {
        if (newTagName.trim()) {
            try {
                const newTag = await api.createTag({
                    name: newTagName.trim(),
                    color: newTagColor,
                });

                // Add to available tags and select it
                availableTags = [...availableTags, newTag];
                tags = [...tags, newTag.id];

                newTagName = "";
                newTagColor = "#3B82F6"; // Reset to default
                showTagSelector = false;
            } catch {
                error = m.event_modal_tag_create_error();
            }
        }
    }

    async function deleteTag(tagId: number) {
        if (deletingTag) return;

        try {
            deletingTag = true;
            await api.deleteTag(tagId);

            // Remove from available tags
            availableTags = availableTags.filter((t) => t.id !== tagId);

            // Remove from selected tags if present
            tags = tags.filter((t) => t !== tagId);

            tagToDelete = null;
        } catch (err) {
            error = err instanceof Error ? err.message : "Failed to delete tag";
        } finally {
            deletingTag = false;
        }
    }

    function confirmDeleteTag(tagId: number) {
        tagToDelete = tagId;
        showTagSelector = false;
    }

    function cancelDeleteTag() {
        tagToDelete = null;
    }

    function removeTag(tagToRemove: number) {
        tags = tags.filter((t) => t !== tagToRemove);
    }

    function toggleStatusSelect() {
        statusSelectOpen = !statusSelectOpen;
    }

    function handleStatusClickOutside(event: MouseEvent) {
        const target = event.target as HTMLElement;
        if (
            statusSelectOpen &&
            statusButtonElement &&
            statusDropdownElement &&
            !statusButtonElement.contains(target) &&
            !statusDropdownElement.contains(target)
        ) {
            statusSelectOpen = false;
        }
    }

    function formatHistoryDate(dateString: string) {
        try {
            const date = new Date(dateString);
            const now = new Date();
            const diffInSeconds = Math.floor(
                (now.getTime() - date.getTime()) / 1000,
            );

            if (diffInSeconds < 60) return m.event_modal_time_just_now();
            if (diffInSeconds < 3600)
                return m.event_modal_time_minutes_ago({
                    minutes: Math.floor(diffInSeconds / 60),
                });
            if (diffInSeconds < 86400)
                return m.event_modal_time_hours_ago({
                    hours: Math.floor(diffInSeconds / 3600),
                });
            if (diffInSeconds < 604800)
                return m.event_modal_time_days_ago({
                    days: Math.floor(diffInSeconds / 86400),
                });

            return date.toLocaleDateString("en-US", {
                month: "short",
                day: "numeric",
                year:
                    date.getFullYear() !== now.getFullYear()
                        ? "numeric"
                        : undefined,
            });
        } catch {
            return dateString;
        }
    }

    async function handleSubmit() {
        if (!title.trim()) {
            error = m.event_modal_title_required();
            return;
        }

        loading = true;
        error = "";

        try {
            const eventData = {
                title: title.trim(),
                status,
                content: content.trim(),
                date: date,
                tag_ids: tags,
                media,
            };

            let result;
            if (mode === "create") {
                result = await api.createEvent(eventData as CreateEventRequest);
                dispatch("created", result);
            } else if (mode === "edit" && event) {
                result = await api.updateEvent(
                    event.id,
                    eventData as UpdateEventRequest,
                );
                dispatch("updated", result);
            }

            closeModal();
        } catch (err) {
            error =
                err instanceof Error ? err.message : m.event_modal_save_error();
        } finally {
            loading = false;
        }
    }

    function closeModal() {
        isOpen = false;
        dispatch("close");
    }

    function handleKeydown(e: KeyboardEvent) {
        if (e.key === "Escape") {
            if (showTagSelector) {
                showTagSelector = false;
            } else {
                closeModal();
            }
        }
        if ((e.metaKey || e.ctrlKey) && e.key === "Enter") {
            handleSubmit();
        }
    }

    function handleOutsideClick(e: Event) {
        const target = e.target as Element;
        if (showTagSelector) {
            const tagPopover = document.querySelector(".tag-popover");
            const tagButton = tagButtonElement;
            if (
                tagPopover &&
                !tagPopover.contains(target) &&
                tagButton &&
                !tagButton.contains(target)
            ) {
                showTagSelector = false;
            }
        }
    }

    export { isOpen, event, mode };
</script>

<svelte:window
    on:keydown={handleKeydown}
    on:click={handleOutsideClick}
    on:click={handleStatusClickOutside}
/>

{#if isOpen}
    <!-- Modal backdrop -->
    <div
        class="fixed inset-0 bg-black/50 z-40 flex items-center justify-center p-4"
    >
        <!-- Modal content -->
        <div
            class="bg-background border border-border rounded-lg shadow-lg w-full max-w-7xl h-[90vh] flex flex-col"
        >
            <!-- Modal header -->
            <div
                class="flex items-center justify-between p-8 border-b border-border shrink-0"
            >
                {#if showNewsletterMode}
                    <h2 class="text-2xl font-bold">
                        {m.event_modal_send_newsletter()}
                    </h2>
                    <Button
                        variant="ghost"
                        on:click={() => {
                            showNewsletterMode = false;
                            emailSubject = "";
                            emailContent = "";
                            newsletterError = "";
                        }}
                        class="ml-4"
                    >
                        {m.event_modal_cancel()}
                    </Button>
                {:else}
                    <Input
                        bind:value={title}
                        placeholder={m.event_modal_title_placeholder()}
                        class="text-2xl font-bold border-none bg-transparent px-0 py-3 focus:ring-0 placeholder:text-muted-foreground flex-1"
                    />
                    <Button
                        variant="ghost"
                        size="icon"
                        on:click={closeModal}
                        class="text-muted-foreground hover:text-foreground ml-4"
                    >
                        <X class="h-4 w-4" />
                    </Button>
                {/if}
            </div>

            <!-- Modal body with sidebar -->
            <div class="flex flex-1 overflow-hidden">
                <!-- Main content area -->
                <div class="flex-1 flex flex-col overflow-hidden">
                    {#if error}
                        <Card
                            class="m-6 mb-0 p-4 bg-destructive/10 border-destructive"
                        >
                            <p class="text-destructive text-sm">{error}</p>
                        </Card>
                    {/if}

                    {#if showNewsletterMode}
                        <!-- Newsletter Content -->
                        <div class="flex-1 overflow-hidden px-6 py-6">
                            {#if newsletterError}
                                <Card
                                    class="p-4 mb-4 bg-destructive/10 border-destructive"
                                >
                                    <p class="text-destructive text-sm">
                                        {newsletterError}
                                    </p>
                                </Card>
                            {/if}

                            {#if showEmailPreview}
                                <div
                                    class="p-6 bg-background border border-border rounded-md h-full overflow-y-auto"
                                >
                                    {@html emailContent}
                                </div>
                            {:else}
                                <Textarea
                                    id="newsletter-content"
                                    bind:value={emailContent}
                                    placeholder={m.event_modal_newsletter_content_placeholder()}
                                    class="h-full font-mono text-sm resize-none"
                                />
                            {/if}
                        </div>
                    {:else}
                        <!-- Content Editor -->
                        <div class="flex-1 overflow-hidden px-6 pt-6 pb-2">
                            <TiptapEditor
                                bind:this={tiptapEditor}
                                bind:content
                                on:update={handleContentUpdate}
                                on:ready={handleEditorReady}
                            />
                        </div>
                    {/if}
                </div>

                <!-- Right Sidebar -->
                <div
                    class="w-80 border-l border-border flex flex-col bg-muted/30"
                >
                    {#if showNewsletterMode}
                        <!-- Newsletter Sidebar -->
                        <div class="flex-1 overflow-y-auto p-6 space-y-4">
                            <!-- Subject Section -->
                            <div>
                                <h3
                                    class="text-sm font-semibold mb-3 text-foreground"
                                >
                                    {m.event_modal_subject()}
                                </h3>
                                <Input
                                    id="newsletter-subject"
                                    bind:value={emailSubject}
                                    placeholder={m.event_modal_subject_placeholder()}
                                    class="text-sm"
                                />
                            </div>

                            <!-- View Toggle -->
                            <div>
                                <h3
                                    class="text-sm font-semibold mb-2 text-foreground"
                                >
                                    {m.event_modal_view()}
                                </h3>
                                <div
                                    class="flex items-center gap-1 bg-muted rounded-md p-0.5 w-fit"
                                >
                                    <button
                                        type="button"
                                        class="h-7 px-2 rounded flex items-center gap-1.5 text-xs transition-colors {!showEmailPreview
                                            ? 'bg-background shadow-sm'
                                            : 'hover:bg-background/50'}"
                                        on:click={() =>
                                            (showEmailPreview = false)}
                                    >
                                        <Pencil class="h-3.5 w-3.5" />
                                        <span>{m.event_modal_edit()}</span>
                                    </button>
                                    <button
                                        type="button"
                                        class="h-7 px-2 rounded flex items-center gap-1.5 text-xs transition-colors {showEmailPreview
                                            ? 'bg-background shadow-sm'
                                            : 'hover:bg-background/50'}"
                                        on:click={() =>
                                            (showEmailPreview = true)}
                                    >
                                        <Eye class="h-3.5 w-3.5" />
                                        <span>{m.event_modal_preview()}</span>
                                    </button>
                                </div>
                            </div>

                            <!-- History Section -->
                            <div>
                                <h3
                                    class="text-sm font-semibold mb-3 text-foreground"
                                >
                                    {m.event_modal_history()}
                                </h3>
                                {#if historyLoading}
                                    <div
                                        class="text-xs text-muted-foreground text-center py-4"
                                    >
                                        {m.event_modal_loading()}
                                    </div>
                                {:else if newsletterHistory.length === 0}
                                    <div
                                        class="text-xs text-muted-foreground text-center py-4"
                                    >
                                        {m.event_modal_no_newsletters_sent()}
                                    </div>
                                {:else}
                                    <div
                                        class="space-y-2 max-h-[300px] overflow-y-auto"
                                    >
                                        {#each newsletterHistory as history}
                                            <div
                                                class="p-3 bg-muted/50 rounded-lg border border-border"
                                            >
                                                <div
                                                    class="text-xs font-medium text-foreground mb-1.5 truncate"
                                                    title={history.email_subject}
                                                >
                                                    {history.email_subject}
                                                </div>
                                                <div
                                                    class="flex items-center justify-between text-xs text-muted-foreground"
                                                >
                                                    <div
                                                        class="flex items-center gap-1.5"
                                                    >
                                                        <Calendar
                                                            class="h-3 w-3 flex-shrink-0"
                                                        />
                                                        <span
                                                            >{formatHistoryDate(
                                                                history.sent_at,
                                                            )}</span
                                                        >
                                                    </div>
                                                    <span
                                                        >{m.event_modal_recipients(
                                                            {
                                                                count: history.subscriber_count,
                                                            },
                                                        )}</span
                                                    >
                                                </div>
                                            </div>
                                        {/each}
                                    </div>
                                {/if}
                            </div>
                        </div>
                    {:else}
                        <div class="flex-1 overflow-y-auto p-6 space-y-4">
                            <!-- Status Section -->
                            <div>
                                <h3
                                    class="text-sm font-semibold mb-2 text-foreground"
                                >
                                    {m.event_modal_status()}
                                </h3>

                                <div class="relative">
                                    <button
                                        bind:this={statusButtonElement}
                                        type="button"
                                        on:click={toggleStatusSelect}
                                        class="h-9 px-3 text-sm border rounded-md bg-background hover:bg-muted flex items-center gap-2 transition-colors w-full justify-between"
                                        aria-haspopup="true"
                                        aria-expanded={statusSelectOpen}
                                    >
                                        <span class="truncate text-left flex-1">
                                            {status || "Select status..."}
                                        </span>
                                        <ChevronDown
                                            class={cn(
                                                "h-4 w-4 shrink-0 opacity-50 transition-transform duration-200",
                                                statusSelectOpen &&
                                                    "rotate-180",
                                            )}
                                        />
                                    </button>

                                    {#if statusSelectOpen}
                                        <div
                                            bind:this={statusDropdownElement}
                                            transition:fly={{
                                                duration: 200,
                                                y: -10,
                                            }}
                                            class="absolute left-0 mt-1 w-full rounded-md border bg-background shadow-md p-2 text-sm space-y-1 z-50 max-h-48 overflow-y-auto"
                                            role="menu"
                                        >
                                            {#if filteredStatuses.length === 0}
                                                <div
                                                    class="py-6 text-center text-sm text-muted-foreground"
                                                >
                                                    No statuses found
                                                </div>
                                            {:else}
                                                {#each filteredStatuses as statusDef}
                                                    <button
                                                        type="button"
                                                        class="w-full text-left px-2 py-1.5 rounded hover:bg-muted transition-colors flex items-center justify-between gap-2"
                                                        on:click={() => {
                                                            status =
                                                                statusDef.display_name;
                                                            statusSelectOpen = false;
                                                        }}
                                                        role="menuitem"
                                                    >
                                                        <span class="truncate"
                                                            >{statusDef.display_name}</span
                                                        >
                                                        <span
                                                            class={cn(
                                                                "flex h-4 w-4 items-center justify-center shrink-0",
                                                                status ===
                                                                    statusDef.display_name
                                                                    ? "opacity-100"
                                                                    : "opacity-0",
                                                            )}
                                                        >
                                                            <Check
                                                                class="h-4 w-4"
                                                            />
                                                        </span>
                                                    </button>
                                                {/each}
                                            {/if}
                                        </div>
                                    {/if}
                                </div>
                            </div>

                            <!-- Tags Section -->
                            <div>
                                <h3
                                    class="text-sm font-semibold mb-3 text-foreground"
                                >
                                    {m.event_modal_tags()}
                                </h3>
                                <div class="space-y-2">
                                    <div class="flex flex-wrap gap-2">
                                        {#each tags as tagId (tagId)}
                                            {@const tag = availableTags.find(
                                                (t) => t.id === tagId,
                                            )}
                                            {#if tag}
                                                <Badge
                                                    variant="outline"
                                                    class="text-xs mr-1 mb-1 flex items-center gap-1"
                                                    style="background-color: {tag.color}20; color: {tag.color}; border-color: {tag.color}"
                                                >
                                                    {tag.name}
                                                    <button
                                                        type="button"
                                                        on:click={() =>
                                                            removeTag(tagId)}
                                                        class="hover:text-destructive"
                                                    >
                                                        <X class="h-3 w-3" />
                                                    </button>
                                                </Badge>
                                            {/if}
                                        {/each}

                                        <div class="relative">
                                            <div bind:this={tagButtonElement}>
                                                <Button
                                                    variant="ghost"
                                                    size="sm"
                                                    on:click={handleTagSelectorToggle}
                                                    class="text-xs h-6 px-2 text-muted-foreground hover:text-foreground"
                                                >
                                                    <Plus
                                                        class="h-3 w-3 mr-1"
                                                    />
                                                    {m.event_modal_add_tag()}
                                                </Button>
                                            </div>

                                            {#if showTagSelector}
                                                <div
                                                    class="tag-popover fixed bg-background border border-border rounded-md shadow-lg p-2 w-72 z-[9999]"
                                                    style="top: {tagPopoverTop}px; left: {tagPopoverLeft}px;"
                                                    on:click|stopPropagation
                                                >
                                                    {#if availableTags.length > 0}
                                                        <div class="mb-2">
                                                            <div
                                                                class="text-xs font-medium text-muted-foreground mb-1.5 px-1"
                                                            >
                                                                {m.event_modal_available_tags()}
                                                            </div>
                                                            <Input
                                                                bind:value={
                                                                    tagSearchTerm
                                                                }
                                                                placeholder="Search tags..."
                                                                class="h-7 text-xs mb-1.5"
                                                            />
                                                            <div
                                                                class="space-y-0.5 max-h-48 overflow-y-auto"
                                                            >
                                                                {#each filteredTags as tag}
                                                                    <div
                                                                        class="flex items-center justify-between gap-2 px-1.5 py-1 rounded hover:bg-accent/50 transition-colors group"
                                                                    >
                                                                        <button
                                                                            type="button"
                                                                            on:click={() =>
                                                                                addExistingTag(
                                                                                    tag.id,
                                                                                )}
                                                                            class="flex items-center gap-2 text-xs flex-1 text-left"
                                                                            disabled={tags.includes(
                                                                                tag.id,
                                                                            )}
                                                                            class:opacity-50={tags.includes(
                                                                                tag.id,
                                                                            )}
                                                                        >
                                                                            <span
                                                                                class="w-2.5 h-2.5 rounded-full flex-shrink-0"
                                                                                style="background-color: {tag.color};"

                                                                            ></span>
                                                                            <span
                                                                                class="truncate"
                                                                            >
                                                                                {tag.name}
                                                                            </span>
                                                                        </button>
                                                                        <button
                                                                            type="button"
                                                                            on:click|stopPropagation={() =>
                                                                                confirmDeleteTag(
                                                                                    tag.id,
                                                                                )}
                                                                            class="opacity-0 group-hover:opacity-100 transition-opacity p-1 rounded hover:bg-destructive/20 flex-shrink-0"
                                                                            title="Delete tag"
                                                                        >
                                                                            <X
                                                                                class="h-3 w-3 text-destructive"
                                                                            />
                                                                        </button>
                                                                    </div>
                                                                {/each}
                                                                {#if filteredTags.length === 0}
                                                                    <div
                                                                        class="text-xs text-muted-foreground text-center py-4"
                                                                    >
                                                                        No tags
                                                                        found
                                                                    </div>
                                                                {/if}
                                                            </div>
                                                        </div>
                                                    {/if}

                                                    <div
                                                        class="pt-2 border-t border-border"
                                                    >
                                                        <div
                                                            class="text-xs font-medium text-muted-foreground mb-1.5 px-1"
                                                        >
                                                            {m.event_modal_create_new_tag()}
                                                        </div>
                                                        <div
                                                            class="space-y-1.5"
                                                        >
                                                            <div
                                                                class="flex items-center gap-1.5"
                                                            >
                                                                <input
                                                                    type="color"
                                                                    bind:value={
                                                                        newTagColor
                                                                    }
                                                                    class="h-7 w-10 rounded border border-border cursor-pointer flex-shrink-0"
                                                                    title="Choose tag color"
                                                                />
                                                                <Input
                                                                    bind:value={
                                                                        newTagName
                                                                    }
                                                                    placeholder={m.event_modal_tag_name_placeholder()}
                                                                    class="h-7 text-xs flex-1"
                                                                    on:keydown={(
                                                                        e,
                                                                    ) => {
                                                                        if (
                                                                            e.key ===
                                                                            "Enter"
                                                                        ) {
                                                                            createNewTag();
                                                                        }
                                                                    }}
                                                                />
                                                            </div>
                                                            <Button
                                                                size="sm"
                                                                on:click={createNewTag}
                                                                class="h-7 text-xs w-full"
                                                                disabled={!newTagName.trim()}
                                                            >
                                                                {m.event_modal_create()}
                                                            </Button>
                                                        </div>
                                                    </div>
                                                </div>
                                            {/if}

                                            <!-- Delete Confirmation Dialog -->
                                            {#if tagToDelete !== null}
                                                {@const tagName =
                                                    availableTags.find(
                                                        (t) =>
                                                            t.id ===
                                                            tagToDelete,
                                                    )?.name}
                                                <div
                                                    class="fixed inset-0 z-50 flex items-center justify-center bg-black/40"
                                                >
                                                    <div
                                                        class="bg-background rounded-lg p-5 w-full max-w-sm space-y-4"
                                                    >
                                                        <h2
                                                            class="text-sm font-semibold"
                                                        >
                                                            Delete Tag
                                                        </h2>
                                                        <p
                                                            class="text-xs text-muted-foreground"
                                                        >
                                                            Are you sure you
                                                            want to delete the
                                                            tag "{tagName}"?
                                                            This action cannot
                                                            be undone.
                                                        </p>
                                                        <div
                                                            class="flex justify-end gap-2 text-xs"
                                                        >
                                                            <Button
                                                                variant="outline"
                                                                size="sm"
                                                                on:click={cancelDeleteTag}
                                                                disabled={deletingTag}
                                                            >
                                                                Cancel
                                                            </Button>
                                                            <Button
                                                                size="sm"
                                                                on:click={() =>
                                                                    deleteTag(
                                                                        tagToDelete,
                                                                    )}
                                                                disabled={deletingTag}
                                                            >
                                                                {#if deletingTag}
                                                                    Deleting...
                                                                {:else}
                                                                    Confirm
                                                                {/if}
                                                            </Button>
                                                        </div>
                                                    </div>
                                                </div>
                                            {/if}
                                        </div>
                                    </div>
                                </div>

                                <!-- Date Section -->
                                <div class="mt-4">
                                    <h3
                                        class="text-sm font-semibold mb-3 text-foreground"
                                    >
                                        {m.event_modal_date()}
                                    </h3>
                                    <DatePicker
                                        bind:value={date}
                                        placeholder={m.event_modal_date_placeholder()}
                                        includeTime={true}
                                    />
                                </div>

                                <!-- Reactions Section -->
                                {#if mode === "edit" && visibleReactions.length > 0}
                                    <div class="mt-4">
                                        <h3
                                            class="text-sm font-semibold mb-3 text-foreground"
                                        >
                                            {m.event_modal_reactions()}
                                        </h3>
                                        <div class="flex flex-wrap gap-1.5">
                                            {#each visibleReactions as reaction}
                                                <div
                                                    class="inline-flex items-center gap-1 px-2 py-1 rounded-full border border-border bg-background text-xs"
                                                    title="{reaction.reaction_type}: {reaction.count}"
                                                >
                                                    <Icon
                                                        icon={reactionIcons[
                                                            reaction
                                                                .reaction_type
                                                        ] ||
                                                            "fluent-emoji-flat:thumbs-up"}
                                                        class="h-3.5 w-3.5"
                                                    />
                                                    <span
                                                        class="text-muted-foreground font-medium"
                                                    >
                                                        {reaction.count}
                                                    </span>
                                                </div>
                                            {/each}
                                        </div>
                                    </div>
                                {/if}

                                <!-- Newsletter Section -->
                                {#if event}
                                    <div class="mt-4">
                                        <h3
                                            class="text-sm font-semibold mb-3 text-foreground"
                                        >
                                            {m.event_modal_newsletter()}
                                        </h3>
                                        <Button
                                            variant="outline"
                                            size="sm"
                                            on:click={() => {
                                                showNewsletterMode = true;
                                                showEmailPreview = true;
                                                newsletterError = "";
                                                loadNewsletterPreview();
                                                loadNewsletterHistory();
                                            }}
                                            class="w-full justify-start gap-2"
                                        >
                                            <Mail class="h-3.5 w-3.5" />
                                            <span class="text-xs"
                                                >{m.event_modal_send_newsletter()}</span
                                            >
                                        </Button>
                                    </div>
                                {/if}
                            </div>
                        </div>
                    {/if}

                    <!-- Action Buttons at Bottom of Sidebar -->
                    <div class="p-2">
                        {#if showNewsletterMode}
                            <Button
                                variant="default"
                                size="default"
                                on:click={sendNewsletter}
                                disabled={newsletterLoading ||
                                    !emailSubject ||
                                    !emailContent}
                                class="w-full gap-2"
                            >
                                {#if newsletterLoading}
                                    <div
                                        class="h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent"
                                    ></div>
                                {:else}
                                    <Send class="h-4 w-4" />
                                {/if}
                                {m.event_modal_send_newsletter()}
                            </Button>
                        {:else}
                            <Button
                                variant="default"
                                size="default"
                                on:click={handleSubmit}
                                disabled={loading || !title.trim()}
                                class="w-full gap-2"
                            >
                                {#if loading}
                                    <div
                                        class="h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent"
                                    ></div>
                                {:else}
                                    <Save class="h-4 w-4" />
                                {/if}
                                {mode === "create"
                                    ? m.event_modal_create_button()
                                    : m.event_modal_save_button()}
                            </Button>
                        {/if}
                    </div>
                </div>
            </div>
        </div>
    </div>
{/if}
