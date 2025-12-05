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
    import { X, Plus, Save, Mail, Send } from "lucide-svelte";
    import {
        Button,
        Card,
        Input,
        Badge,
        DatePicker,
        Textarea,
    } from "$lib/components/ui";
    import TiptapEditor from "$lib/components/TiptapEditor.svelte";
    import ImageUploadModal from "$lib/components/ImageUploadModal.svelte";
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

    // Status management
    export let statuses: StatusDefinition[] = [];

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
    $: if (event && mode === "edit") {
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
    let eventImageModalOpen = false;

    function handleContentUpdate(event: CustomEvent) {
        content = event.detail.content;
    }

    function handleEditorReady(event: CustomEvent) {
        tiptapEditor = event.detail.editor;
    }

    function handleImageSelected(event: CustomEvent) {
        const { url } = event.detail;
        if (url && tiptapEditor) {
            tiptapEditor.chain().focus().setImage({ src: url }).run();
        }
    }

    function addExistingTag(tagId: number) {
        if (!tags.includes(tagId)) {
            tags = [...tags, tagId];
        }
        showTagSelector = false;
    }

    function handleTagSelectorToggle(e: Event) {
        e.stopPropagation();
        showTagSelector = !showTagSelector;
    }

    async function createNewTag() {
        if (newTagName.trim()) {
            try {
                const newTag = await api.createTag({
                    name: newTagName.trim(),
                    color: "#3B82F6", // Default blue color
                });

                // Add to available tags and select it
                availableTags = [...availableTags, newTag];
                tags = [...tags, newTag.id];

                newTagName = "";
                showTagSelector = false;
            } catch {
                error = m.event_modal_tag_create_error();
            }
        }
    }

    function removeTag(tagToRemove: number) {
        tags = tags.filter((tag) => tag !== tagToRemove);
    }

    function formatHistoryDate(dateString: string): string {
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
        if (showTagSelector && !target.closest(".tag-selector-container")) {
            showTagSelector = false;
        }
    }

    export { isOpen, event, mode };
</script>

<svelte:window on:keydown={handleKeydown} on:click={handleOutsideClick} />

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
                            showEmailPreview = false;
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
                        <div class="flex-1 overflow-hidden px-6 py-6">
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
                                    class="text-sm font-semibold mb-3 text-foreground"
                                >
                                    {m.event_modal_view()}
                                </h3>
                                <div
                                    class="flex rounded-md border border-border overflow-hidden"
                                >
                                    <button
                                        type="button"
                                        on:click={() =>
                                            (showEmailPreview = false)}
                                        class="flex-1 px-3 py-2 text-sm font-medium transition-colors {!showEmailPreview
                                            ? 'bg-primary text-primary-foreground'
                                            : 'bg-background hover:bg-muted'}"
                                    >
                                        {m.event_modal_edit()}
                                    </button>
                                    <button
                                        type="button"
                                        on:click={() =>
                                            (showEmailPreview = true)}
                                        class="flex-1 px-3 py-2 text-sm font-medium transition-colors {showEmailPreview
                                            ? 'bg-primary text-primary-foreground'
                                            : 'bg-background hover:bg-muted'}"
                                    >
                                        {m.event_modal_preview()}
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
                                                    class="text-sm font-medium text-foreground mb-1 truncate"
                                                    title={history.email_subject}
                                                >
                                                    {history.email_subject}
                                                </div>
                                                <div
                                                    class="flex items-center justify-between text-xs text-muted-foreground"
                                                >
                                                    <span
                                                        >{formatHistoryDate(
                                                            history.sent_at,
                                                        )}</span
                                                    >
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
                                    class="text-sm font-semibold mb-3 text-foreground"
                                >
                                    {m.event_modal_status()}
                                </h3>
                                <select
                                    bind:value={status}
                                    class="w-full h-9 text-sm rounded-md border border-input bg-background px-3 py-1 ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                                >
                                    {#each statuses as statusDef}
                                        <option value={statusDef.display_name}>
                                            {statusDef.display_name}
                                        </option>
                                    {/each}
                                </select>
                            </div>

                            <!-- Tags Section -->
                            <div class="mt-4">
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
                                            <Button
                                                variant="ghost"
                                                size="sm"
                                                on:click={handleTagSelectorToggle}
                                                class="text-xs h-6 px-2 text-muted-foreground hover:text-foreground"
                                            >
                                                <Plus class="h-3 w-3 mr-1" />
                                                {m.event_modal_add_tag()}
                                            </Button>

                                            {#if showTagSelector}
                                                <div
                                                    class="absolute top-full left-0 mt-1 bg-background border border-border rounded-md shadow-lg p-3 w-64 z-50"
                                                >
                                                    {#if availableTags.length > 0}
                                                        <div class="mb-3">
                                                            <div
                                                                class="text-xs font-medium text-muted-foreground mb-2"
                                                            >
                                                                {m.event_modal_available_tags()}
                                                            </div>
                                                            <div
                                                                class="flex flex-wrap gap-1 max-h-20 overflow-y-auto"
                                                            >
                                                                {#each availableTags as tag}
                                                                    <button
                                                                        type="button"
                                                                        on:click={() =>
                                                                            addExistingTag(
                                                                                tag.id,
                                                                            )}
                                                                        class="text-xs px-2 py-1 rounded transition-colors"
                                                                        style="background-color: {tag.color}20; color: {tag.color}; border: 1px solid {tag.color}40;"
                                                                        disabled={tags.includes(
                                                                            tag.id,
                                                                        )}
                                                                        class:opacity-50={tags.includes(
                                                                            tag.id,
                                                                        )}
                                                                    >
                                                                        {tag.name}
                                                                    </button>
                                                                {/each}
                                                            </div>
                                                        </div>
                                                    {/if}

                                                    <div
                                                        class="pt-3 border-t border-border"
                                                    >
                                                        <div
                                                            class="text-xs font-medium text-muted-foreground mb-2"
                                                        >
                                                            {m.event_modal_create_new_tag()}
                                                        </div>
                                                        <div class="space-y-2">
                                                            <Input
                                                                bind:value={
                                                                    newTagName
                                                                }
                                                                placeholder={m.event_modal_tag_name_placeholder()}
                                                                class="h-8 text-sm"
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
                                                            <div
                                                                class="flex items-center gap-2"
                                                            >
                                                                <Button
                                                                    size="sm"
                                                                    on:click={createNewTag}
                                                                    class="h-6 text-xs"
                                                                    disabled={!newTagName.trim()}
                                                                >
                                                                    {m.event_modal_create()}
                                                                </Button>
                                                            </div>
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
                    <div class="p-4">
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

<!-- Image Upload Modal - Outside main modal to prevent z-index conflicts -->
<ImageUploadModal
    bind:isOpen={eventImageModalOpen}
    on:imageSelected={handleImageSelected}
/>
