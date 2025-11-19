<script lang="ts">
    import { createEventDispatcher, onMount } from "svelte";
    import { api } from "$lib/api";

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
        Palette,
        Bold,
        Italic,
        Strikethrough,
        Code,
        Heading1,
        Heading2,
        List,
        ListOrdered,
        Quote,
        Link as LinkIcon,
        Image as ImageIcon,
        Minus,
        Undo,
        Redo,
        Send,
        Share2,
    } from "lucide-svelte";
    import { Button, Card, Input, Badge, DatePicker } from "$lib/components/ui";
    import TiptapEditor from "$lib/components/TiptapEditor.svelte";
    import ImageUploadModal from "$lib/components/ImageUploadModal.svelte";

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
    let status: EventStatus = "Backlogs";
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

    onMount(async () => {
        await loadAvailableTags();
    });

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
            status = event?.status || "Backlogs";
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
            } catch (err) {
                error = "Failed to create tag";
            }
        }
    }

    function removeTag(tagToRemove: number) {
        tags = tags.filter((tag) => tag !== tagToRemove);
    }

    async function handleSubmit() {
        if (!title.trim()) {
            error = "Title is required";
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
            error = err instanceof Error ? err.message : "Failed to save event";
        } finally {
            loading = false;
        }
    }

    function closeModal() {
        isOpen = false;
        dispatch("close");
    }

    async function handleShare() {
        // Check if there are unsaved changes
        const originalTags = event?.tags ? event.tags.map((tag) => tag.id) : [];
        const hasChanges =
            title !== (event?.title || "") ||
            status !== (event?.status || "Backlogs") ||
            content !== (event?.content || "") ||
            date !== (event?.date || "") ||
            JSON.stringify(tags.sort()) !==
                JSON.stringify(originalTags.sort()) ||
            JSON.stringify(media) !== JSON.stringify(event?.media || []);

        // If there are changes, offer to save them before sharing
        if (mode === "edit" && event && hasChanges) {
            const shouldSave = confirm(
                "You have unsaved changes. Would you like to save them before sharing?",
            );
            if (shouldSave) {
                try {
                    loading = true;
                    error = "";
                    const eventData = {
                        title: title.trim(),
                        status,
                        content: content.trim(),
                        date,
                        tag_ids: tags,
                        media,
                    };
                    const updated = await api.updateEvent(
                        event.id,
                        eventData as UpdateEventRequest,
                    );
                    // Keep local event in sync
                    event = updated;
                    // Dispatch publish BEFORE updated so parent still has the event when opening publish modal
                    dispatch("publish", updated);
                    dispatch("updated", updated);
                    closeModal();
                    return;
                } catch (err) {
                    error =
                        err instanceof Error
                            ? err.message
                            : "Failed to save event";
                    // Abort sharing if saving failed
                    return;
                } finally {
                    loading = false;
                }
            }
        }

        // No changes (or user chose not to save) -> publish first, then close
        if (event) {
            dispatch("publish", event);
        }
        closeModal();
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
            class="bg-background border border-border rounded-lg shadow-lg w-full max-w-6xl h-[90vh] flex flex-col"
        >
            <!-- Modal header -->
            <div
                class="flex items-center justify-between p-8 border-b border-border shrink-0"
            >
                <div class="flex items-center gap-6 flex-1">
                    <Input
                        bind:value={title}
                        placeholder="Enter event title..."
                        class="text-2xl font-bold border-none bg-transparent px-0 py-3 focus:ring-0 placeholder:text-muted-foreground flex-1"
                    />
                    <select
                        bind:value={status}
                        class="text-xs px-3 py-1.5 rounded-md border font-medium cursor-pointer"
                    >
                        {#each statuses as statusDef}
                            <option value={statusDef.display_name}>
                                {statusDef.display_name}
                            </option>
                        {/each}
                    </select>
                </div>
                <Button
                    variant="ghost"
                    size="icon"
                    on:click={closeModal}
                    class="text-muted-foreground hover:text-foreground ml-4"
                >
                    <X class="h-4 w-4" />
                </Button>
            </div>

            <!-- Modal body -->
            <div class="flex-1 overflow-y-auto">
                <div class="px-6 py-4 pb-20">
                    {#if error}
                        <Card
                            class="p-4 mb-4 bg-destructive/10 border-destructive"
                        >
                            <p class="text-destructive text-sm">{error}</p>
                        </Card>
                    {/if}

                    <!-- Tags and dates -->
                    <div class="mb-8">
                        <div class="flex items-center gap-4 flex-wrap">
                            <!-- Tags -->
                            <div
                                class="flex items-center gap-2 flex-wrap tag-selector-container"
                            >
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
                                        Add tag
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
                                                        Available Tags
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
                                                    Create New Tag
                                                </div>
                                                <div class="space-y-2">
                                                    <Input
                                                        bind:value={newTagName}
                                                        placeholder="Tag name"
                                                        class="h-8 text-sm"
                                                        on:keydown={(e) => {
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
                                                            Create
                                                        </Button>
                                                    </div>
                                                </div>
                                            </div>
                                        </div>
                                    {/if}
                                </div>
                            </div>

                            <!-- Dates -->
                            <div
                                class="flex items-center gap-4 text-sm text-muted-foreground"
                            >
                                <DatePicker
                                    bind:value={date}
                                    placeholder="Date"
                                />
                            </div>
                        </div>
                    </div>

                    <!-- Content Editor -->
                    <div>
                        <TiptapEditor
                            bind:content
                            on:update={handleContentUpdate}
                            on:ready={handleEditorReady}
                            placeholder="Describe your event, feature, or update in detail..."
                            showToolbar={false}
                        />
                    </div>
                </div>
            </div>

            <!-- Modal footer with toolbar and buttons -->
            <div
                class="p-6 shrink-0 flex justify-between items-center bg-background/95 backdrop-blur-sm"
            >
                <!-- Formatting toolbar -->
                <div class="flex flex-wrap items-center gap-1">
                    <!-- Text formatting -->
                    <div class="flex items-center gap-1 pr-2">
                        <button
                            type="button"
                            class="p-1.5 hover:bg-muted/50 transition-colors {tiptapEditor?.isActive(
                                'bold',
                            )
                                ? 'bg-muted text-foreground'
                                : 'text-muted-foreground hover:text-foreground'}"
                            on:click={() =>
                                tiptapEditor
                                    ?.chain()
                                    .focus()
                                    .toggleBold()
                                    .run()}
                            title="Bold"
                        >
                            <Bold class="h-3.5 w-3.5" />
                        </button>
                        <button
                            type="button"
                            class="p-1.5 hover:bg-muted/50 transition-colors {tiptapEditor?.isActive(
                                'italic',
                            )
                                ? 'bg-muted text-foreground'
                                : 'text-muted-foreground hover:text-foreground'}"
                            on:click={() =>
                                tiptapEditor
                                    ?.chain()
                                    .focus()
                                    .toggleItalic()
                                    .run()}
                            title="Italic"
                        >
                            <Italic class="h-3.5 w-3.5" />
                        </button>
                        <button
                            type="button"
                            class="p-1.5 hover:bg-muted/50 transition-colors {tiptapEditor?.isActive(
                                'strike',
                            )
                                ? 'bg-muted text-foreground'
                                : 'text-muted-foreground hover:text-foreground'}"
                            on:click={() =>
                                tiptapEditor
                                    ?.chain()
                                    .focus()
                                    .toggleStrike()
                                    .run()}
                            title="Strikethrough"
                        >
                            <Strikethrough class="h-3.5 w-3.5" />
                        </button>
                        <button
                            type="button"
                            class="p-1.5 hover:bg-muted/50 transition-colors {tiptapEditor?.isActive(
                                'code',
                            )
                                ? 'bg-muted text-foreground'
                                : 'text-muted-foreground hover:text-foreground'}"
                            on:click={() =>
                                tiptapEditor
                                    ?.chain()
                                    .focus()
                                    .toggleCode()
                                    .run()}
                            title="Inline Code"
                        >
                            <Code class="h-3.5 w-3.5" />
                        </button>
                    </div>

                    <!-- Headings -->
                    <div class="flex items-center gap-1 pr-2">
                        <button
                            type="button"
                            class="p-1.5 hover:bg-muted/50 transition-colors {tiptapEditor?.isActive(
                                'heading',
                                { level: 1 },
                            )
                                ? 'bg-muted text-foreground'
                                : 'text-muted-foreground hover:text-foreground'}"
                            on:click={() =>
                                tiptapEditor
                                    ?.chain()
                                    .focus()
                                    .toggleHeading({ level: 1 })
                                    .run()}
                            title="Heading 1"
                        >
                            <Heading1 class="h-3.5 w-3.5" />
                        </button>
                        <button
                            type="button"
                            class="p-1.5 hover:bg-muted/50 transition-colors {tiptapEditor?.isActive(
                                'heading',
                                { level: 2 },
                            )
                                ? 'bg-muted text-foreground'
                                : 'text-muted-foreground hover:text-foreground'}"
                            on:click={() =>
                                tiptapEditor
                                    ?.chain()
                                    .focus()
                                    .toggleHeading({ level: 2 })
                                    .run()}
                            title="Heading 2"
                        >
                            <Heading2 class="h-3.5 w-3.5" />
                        </button>
                    </div>

                    <!-- Lists -->
                    <div class="flex items-center gap-1 pr-2">
                        <button
                            type="button"
                            class="p-1.5 hover:bg-muted/50 transition-colors {tiptapEditor?.isActive(
                                'bulletList',
                            )
                                ? 'bg-muted text-foreground'
                                : 'text-muted-foreground hover:text-foreground'}"
                            on:click={() =>
                                tiptapEditor
                                    ?.chain()
                                    .focus()
                                    .toggleBulletList()
                                    .run()}
                            title="Bullet List"
                        >
                            <List class="h-3.5 w-3.5" />
                        </button>
                        <button
                            type="button"
                            class="p-1.5 hover:bg-muted/50 transition-colors {tiptapEditor?.isActive(
                                'orderedList',
                            )
                                ? 'bg-muted text-foreground'
                                : 'text-muted-foreground hover:text-foreground'}"
                            on:click={() =>
                                tiptapEditor
                                    ?.chain()
                                    .focus()
                                    .toggleOrderedList()
                                    .run()}
                            title="Numbered List"
                        >
                            <ListOrdered class="h-3.5 w-3.5" />
                        </button>
                        <button
                            type="button"
                            class="p-1.5 hover:bg-muted/50 transition-colors {tiptapEditor?.isActive(
                                'blockquote',
                            )
                                ? 'bg-muted text-foreground'
                                : 'text-muted-foreground hover:text-foreground'}"
                            on:click={() =>
                                tiptapEditor
                                    ?.chain()
                                    .focus()
                                    .toggleBlockquote()
                                    .run()}
                            title="Quote"
                        >
                            <Quote class="h-3.5 w-3.5" />
                        </button>
                    </div>

                    <!-- Media & Links -->
                    <div class="flex items-center gap-1 pr-2">
                        <button
                            type="button"
                            class="p-1.5 hover:bg-muted/50 transition-colors {tiptapEditor?.isActive(
                                'link',
                            )
                                ? 'bg-muted text-foreground'
                                : 'text-muted-foreground hover:text-foreground'}"
                            on:click={() => {
                                const url = window.prompt("Enter URL:");
                                if (url) {
                                    tiptapEditor
                                        ?.chain()
                                        .focus()
                                        .setLink({ href: url })
                                        .run();
                                }
                            }}
                            title="Add Link"
                        >
                            <LinkIcon class="h-3.5 w-3.5" />
                        </button>
                        <button
                            type="button"
                            class="p-1.5 hover:bg-muted/50 transition-colors text-muted-foreground hover:text-foreground"
                            on:click={() => {
                                eventImageModalOpen = true;
                            }}
                            title="Add Image"
                        >
                            <ImageIcon class="h-3.5 w-3.5" />
                        </button>
                    </div>

                    <!-- Formatting -->
                    <div class="flex items-center gap-1 pr-2">
                        <button
                            type="button"
                            class="p-1.5 hover:bg-muted/50 transition-colors {tiptapEditor?.isActive(
                                'codeBlock',
                            )
                                ? 'bg-muted text-foreground'
                                : 'text-muted-foreground hover:text-foreground'}"
                            on:click={() =>
                                tiptapEditor
                                    ?.chain()
                                    .focus()
                                    .toggleCodeBlock()
                                    .run()}
                            title="Code Block"
                        >
                            <Code class="h-3.5 w-3.5" />
                        </button>
                        <button
                            type="button"
                            class="p-1.5 hover:bg-muted/50 transition-colors text-muted-foreground hover:text-foreground"
                            on:click={() =>
                                tiptapEditor
                                    ?.chain()
                                    .focus()
                                    .setHorizontalRule()
                                    .run()}
                            title="Horizontal Rule"
                        >
                            <Minus class="h-3.5 w-3.5" />
                        </button>
                    </div>

                    <!-- Undo/Redo -->
                    <div class="flex items-center gap-1">
                        <button
                            type="button"
                            class="p-1.5 hover:bg-muted/50 transition-colors text-muted-foreground hover:text-foreground disabled:opacity-50 disabled:cursor-not-allowed"
                            on:click={() =>
                                tiptapEditor?.chain().focus().undo().run()}
                            disabled={!tiptapEditor?.can().undo()}
                            title="Undo"
                        >
                            <Undo class="h-3.5 w-3.5" />
                        </button>
                        <button
                            type="button"
                            class="p-1.5 hover:bg-muted/50 transition-colors text-muted-foreground hover:text-foreground disabled:opacity-50 disabled:cursor-not-allowed"
                            on:click={() =>
                                tiptapEditor?.chain().focus().redo().run()}
                            disabled={!tiptapEditor?.can().redo()}
                            title="Redo"
                        >
                            <Redo class="h-3.5 w-3.5" />
                        </button>
                    </div>
                </div>

                <!-- Buttons -->
                <div class="flex gap-2">
                    {#if mode === "edit" && event && status !== "Backlogs" && status !== "Archived"}
                        <Button
                            variant="outline"
                            size="sm"
                            on:click={handleShare}
                            class="min-w-24 gap-2"
                        >
                            <Share2 class="h-4 w-4" />
                            Share
                        </Button>
                    {/if}
                    <Button
                        variant="default"
                        size="sm"
                        on:click={handleSubmit}
                        disabled={loading || !title.trim()}
                        class="min-w-24 bg-primary hover:bg-primary/90 text-primary-foreground shadow-sm gap-2"
                    >
                        {#if loading}
                            <div
                                class="h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent"
                            ></div>
                        {:else}
                            <Save class="h-4 w-4" />
                        {/if}
                        {mode === "create" ? "Create" : "Save"}
                    </Button>
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
