<script lang="ts">
    import { createEventDispatcher, onMount } from "svelte";
    import { api } from "$lib/api";
    import type {
        CreateEventRequest,
        UpdateEventRequest,
        EventStatus,
        ParsedEvent,
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
        Heading3,
        List,
        ListOrdered,
        Quote,
        Link as LinkIcon,
        Image as ImageIcon,
        Minus,
        Undo,
        Redo,
    } from "lucide-svelte";
    import { Button, Card, Input, Badge, DatePicker } from "$lib/components/ui";
    import TiptapEditor from "$lib/components/TiptapEditor.svelte";
    import ImageUploadModal from "$lib/components/ImageUploadModal.svelte";

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
    let tags: string[] = [];
    let tagColors: Record<string, string> = {};
    let media: string[] = [];

    // Tag management
    let existingTags: string[] = [];
    let existingTagColors: Record<string, string> = {};
    let showTagSelector = false;
    let newTagName = "";
    let newTagColor = "#3b82f6";

    const statusOptions = [
        { value: "Backlogs", label: "Backlogs" },
        { value: "Upcoming", label: "Upcoming" },
        { value: "Doing", label: "Doing" },
        { value: "Release", label: "Release" },
        { value: "Archived", label: "Archived" },
    ];

    onMount(async () => {
        await loadExistingTags();
    });

    async function loadExistingTags() {
        try {
            const allEvents = await api.getAllEvents();
            const allTags = new Set<string>();

            allEvents.forEach((event) => {
                try {
                    const eventTags = event.tags ? JSON.parse(event.tags) : [];
                    eventTags.forEach((tag: string) => {
                        allTags.add(tag);
                    });
                } catch (e) {
                    // Handle parsing errors
                }
            });

            existingTags = Array.from(allTags).sort();
            existingTags.forEach((tag) => {
                existingTagColors[tag] = generateTagColor(tag);
            });
        } catch (err) {
            console.error("Failed to load existing tags:", err);
        }
    }

    function generateTagColor(tag: string): string {
        const colors = [
            "#3b82f6",
            "#ef4444",
            "#10b981",
            "#f59e0b",
            "#8b5cf6",
            "#ec4899",
            "#06b6d4",
            "#84cc16",
            "#f97316",
            "#6366f1",
        ];

        let hash = 0;
        for (let i = 0; i < tag.length; i++) {
            hash = tag.charCodeAt(i) + ((hash << 5) - hash);
        }

        return colors[Math.abs(hash) % colors.length];
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
            tags = [...event.tags];
            media = [...event.media];
            date = event.date || "";

            // Set tag colors
            tagColors = {};
            tags.forEach((tag) => {
                tagColors[tag] =
                    existingTagColors[tag] || generateTagColor(tag);
            });
        } else {
            title = "";
            status = event?.status || "Backlogs";
            date = "";
            content = "";
            tags = [];
            tagColors = {};
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

    function addExistingTag(tag: string) {
        if (!tags.includes(tag)) {
            tags = [...tags, tag];
            tagColors = {
                ...tagColors,
                [tag]: existingTagColors[tag] || generateTagColor(tag),
            };
        }
        showTagSelector = false;
    }

    function handleTagSelectorToggle(e: Event) {
        e.stopPropagation();
        showTagSelector = !showTagSelector;
    }

    function createNewTag() {
        if (newTagName.trim() && !tags.includes(newTagName.trim())) {
            const tag = newTagName.trim();
            tags = [...tags, tag];
            tagColors = { ...tagColors, [tag]: newTagColor };

            // Add to existing tags
            if (!existingTags.includes(tag)) {
                existingTags = [...existingTags, tag].sort();
                existingTagColors = {
                    ...existingTagColors,
                    [tag]: newTagColor,
                };
            }

            newTagName = "";
            newTagColor = "#3b82f6";
            showTagSelector = false;
        }
    }

    function removeTag(tagToRemove: string) {
        tags = tags.filter((tag) => tag !== tagToRemove);
        const { [tagToRemove]: removed, ...remainingColors } = tagColors;
        tagColors = remainingColors;
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
                tags,
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
                        class="text-xs px-2 py-1 rounded border-none outline-none {statusOptions.find(
                            (s) => s.value === status,
                        )?.value === 'Backlogs'
                            ? 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-200'
                            : statusOptions.find((s) => s.value === status)
                                    ?.value === 'Upcoming'
                              ? 'bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-200'
                              : statusOptions.find((s) => s.value === status)
                                      ?.value === 'Doing'
                                ? 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200'
                                : statusOptions.find((s) => s.value === status)
                                        ?.value === 'Release'
                                  ? 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200'
                                  : 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200'}"
                    >
                        {#each statusOptions as option}
                            <option value={option.value}>{option.label}</option>
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
                                {#each tags as tag (tag)}
                                    <Badge
                                        variant="secondary"
                                        class="text-xs px-2 py-1 gap-1"
                                        style="background-color: {tagColors[
                                            tag
                                        ]}20; color: {tagColors[
                                            tag
                                        ]}; border-color: {tagColors[tag]}40;"
                                    >
                                        {tag}
                                        <button
                                            type="button"
                                            on:click={() => removeTag(tag)}
                                            class="ml-1 hover:text-destructive transition-colors"
                                        >
                                            <X class="h-3 w-3" />
                                        </button>
                                    </Badge>
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
                                            {#if existingTags.length > 0}
                                                <div class="mb-3">
                                                    <div
                                                        class="text-xs font-medium text-muted-foreground mb-2"
                                                    >
                                                        Existing Tags
                                                    </div>
                                                    <div
                                                        class="flex flex-wrap gap-1 max-h-20 overflow-y-auto"
                                                    >
                                                        {#each existingTags as tag}
                                                            <button
                                                                type="button"
                                                                on:click={() =>
                                                                    addExistingTag(
                                                                        tag,
                                                                    )}
                                                                class="text-xs px-2 py-1 rounded bg-muted hover:bg-muted-foreground/20 transition-colors"
                                                                disabled={tags.includes(
                                                                    tag,
                                                                )}
                                                                class:opacity-50={tags.includes(
                                                                    tag,
                                                                )}
                                                            >
                                                                {tag}
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
                                                        <input
                                                            type="color"
                                                            bind:value={
                                                                newTagColor
                                                            }
                                                            class="w-6 h-6 rounded border border-border cursor-pointer"
                                                        />
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

            <!-- Modal footer with toolbar and save button -->
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
                        <button
                            type="button"
                            class="p-1.5 hover:bg-muted/50 transition-colors {tiptapEditor?.isActive(
                                'heading',
                                { level: 3 },
                            )
                                ? 'bg-muted text-foreground'
                                : 'text-muted-foreground hover:text-foreground'}"
                            on:click={() =>
                                tiptapEditor
                                    ?.chain()
                                    .focus()
                                    .toggleHeading({ level: 3 })
                                    .run()}
                            title="Heading 3"
                        >
                            <Heading3 class="h-3.5 w-3.5" />
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

                <!-- Save button -->
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
{/if}

<!-- Image Upload Modal - Outside main modal to prevent z-index conflicts -->
<ImageUploadModal
    bind:isOpen={eventImageModalOpen}
    on:imageSelected={handleImageSelected}
/>
