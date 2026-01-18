<script lang="ts">
    import { onMount, onDestroy, createEventDispatcher } from "svelte";
    import { scale } from "svelte/transition";
    import { Editor } from "@tiptap/core";
    import StarterKit from "@tiptap/starter-kit";
    import Link from "@tiptap/extension-link";
    import Image from "@tiptap/extension-image";
    import { Table } from "@tiptap/extension-table";
    import { TableRow } from "@tiptap/extension-table-row";
    import { TableHeader } from "@tiptap/extension-table-header";
    import { TableCell } from "@tiptap/extension-table-cell";
    import {
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
        Table as TableIcon,
        Minus,
        Undo,
        Redo,
        ArrowUpToLine,
        ArrowDownToLine,
        ArrowLeftToLine,
        ArrowRightToLine,
        Trash2,
        ListMinus,
        Grid2x2X,
    } from "lucide-svelte";
    import * as m from "$lib/paraglide/messages";
    import { api } from "$lib/api";
    import { toast } from "svelte-sonner";
    import { isRTL } from "$lib/utils";

    export let content = "";
    export let placeholder = "";
    export let showToolbar = true;

    $: effectivePlaceholder = placeholder || m.tiptap_editor_placeholder();
    $: isRtlLocale = isRTL();

    const dispatch = createEventDispatcher();

    let editor: Editor;
    let element: HTMLElement;
    let fileInputElement: HTMLInputElement;
    let isInTable = false;
    let tableElement: HTMLElement | null = null;
    let popoverPosition = { top: 0, left: 0 };

    // Helper function to convert relative image URLs to absolute URLs
    function convertImageUrlsToAbsolute(html: string): string {
        if (!html) return html;

        // Replace relative image URLs with absolute URLs
        return html.replace(
            /<img([^>]+)src=["']([^"']+)["']/g,
            (match, attrs, src) => {
                // If already absolute, keep as-is
                if (src.startsWith("http://") || src.startsWith("https://")) {
                    return match;
                }
                // Convert relative to absolute
                const absoluteUrl = api.getImageUrl(src);
                return `<img${attrs}src="${absoluteUrl}"`;
            },
        );
    }

    onMount(() => {
        editor = new Editor({
            element: element,
            extensions: [
                StarterKit.configure({
                    link: false, // Exclude link from StarterKit to avoid duplicate
                    codeBlock: {
                        HTMLAttributes: {
                            class: "code-block-custom",
                        },
                    },
                    code: {
                        HTMLAttributes: {
                            class: "inline-code-custom",
                        },
                    },
                }),
                Link.configure({
                    openOnClick: false,
                    HTMLAttributes: {
                        class: "text-blue-600 hover:underline cursor-pointer",
                    },
                }),
                Image.configure({
                    HTMLAttributes: {
                        class: "max-w-full h-auto rounded-md my-2",
                    },
                }),
                Table.configure({
                    resizable: true,
                }),
                TableRow,
                TableHeader.configure({
                    HTMLAttributes: {
                        class: "bg-muted font-semibold",
                    },
                }),
                TableCell,
            ],
            content: convertImageUrlsToAbsolute(content),
            onTransaction: () => {
                // Force re-render on transaction
                editor = editor;
                isInTable = editor?.isActive("table") || false;
                if (isInTable) {
                    updatePopoverPosition();
                }
            },
            onUpdate: ({ editor }) => {
                const html = editor.getHTML();
                content = html;
                dispatch("update", { content: html });
            },
            onCreate: ({ editor }) => {
                dispatch("ready", { editor });
            },
            editorProps: {
                attributes: {
                    class: "prose prose-sm max-w-none focus:outline-none min-h-[300px] p-4",
                    "data-placeholder": effectivePlaceholder,
                },
            },
        });
    });

    onDestroy(() => {
        if (editor) {
            editor.destroy();
        }
    });

    // Update content when prop changes
    $: if (editor && content !== editor.getHTML()) {
        const contentWithAbsoluteUrls = convertImageUrlsToAbsolute(content);
        editor.commands.setContent(contentWithAbsoluteUrls);
    }

    function addLink() {
        const url = window.prompt(m.tiptap_editor_enter_url());
        if (url) {
            editor.chain().focus().setLink({ href: url }).run();
        }
    }

    function addImage() {
        fileInputElement?.click();
    }

    async function handleImageUpload(event: Event) {
        const target = event.target as HTMLInputElement;
        const file = target.files?.[0];
        if (!file) return;

        try {
            const result = await api.uploadImage(file);
            const absoluteUrl = api.getImageUrl(result.url);
            editor.chain().focus().setImage({ src: absoluteUrl }).run();
            toast.success("Image uploaded successfully");
        } catch (err) {
            toast.error(
                err instanceof Error ? err.message : "Failed to upload image",
            );
        }

        // Reset the input
        target.value = "";
    }

    function insertTable() {
        editor
            .chain()
            .focus()
            .insertTable({ rows: 3, cols: 3, withHeaderRow: true })
            .run();
    }

    function addColumnBefore() {
        editor?.chain().focus().addColumnBefore().run();
    }

    function addColumnAfter() {
        editor?.chain().focus().addColumnAfter().run();
    }

    function deleteColumn() {
        editor?.chain().focus().deleteColumn().run();
    }

    function addRowBefore() {
        editor?.chain().focus().addRowBefore().run();
    }

    function addRowAfter() {
        editor?.chain().focus().addRowAfter().run();
    }

    function deleteRow() {
        editor?.chain().focus().deleteRow().run();
    }

    function deleteTable() {
        editor?.chain().focus().deleteTable().run();
    }

    function mergeCells() {
        editor?.chain().focus().mergeCells().run();
    }

    function splitCell() {
        editor?.chain().focus().splitCell().run();
    }

    function updatePopoverPosition() {
        if (!editor || !element) return;

        // Find the table element that contains the current selection
        const { view, state } = editor;
        const { from } = state.selection;

        // Get the DOM node at the current cursor position
        const domAtPos = view.domAtPos(from);
        const node = domAtPos.node;

        // Find the closest table element from the cursor position
        let tableNode: HTMLElement | null = null;
        if (node instanceof HTMLElement) {
            tableNode = node.closest("table");
        } else if (node.parentElement) {
            tableNode = node.parentElement.closest("table");
        }

        if (tableNode) {
            tableElement = tableNode as HTMLElement;
            const rect = tableNode.getBoundingClientRect();
            const containerRect =
                element.parentElement?.getBoundingClientRect() || editorRect;

            // Position at bottom right of table, absolute to viewport
            popoverPosition = {
                top: rect.bottom + 8,
                left: rect.right - 10,
            };
        }
    }
</script>

<div class="flex flex-col h-full">
    <!-- Editor -->
    <div
        bind:this={element}
        class="flex-1 min-h-0 overflow-y-auto bg-background"
    ></div>

    <!-- Toolbar -->
    {#if showToolbar}
        <div
            class="flex flex-wrap items-center gap-1 p-2 bg-background/95 backdrop-blur-sm border-t border-border/50 shrink-0"
        >
            <!-- Text formatting -->
            <div class="flex items-center gap-1 pe-2">
                <button
                    type="button"
                    class="p-1.5 hover:bg-muted/50 transition-colors {editor?.isActive(
                        'bold',
                    )
                        ? 'bg-muted text-foreground'
                        : 'text-muted-foreground hover:text-foreground'}"
                    on:click={() => editor?.chain().focus().toggleBold().run()}
                    title={m.tiptap_editor_bold()}
                >
                    <Bold class="h-3.5 w-3.5" />
                </button>
                <button
                    type="button"
                    class="p-1.5 hover:bg-muted/50 transition-colors {editor?.isActive(
                        'italic',
                    )
                        ? 'bg-muted text-foreground'
                        : 'text-muted-foreground hover:text-foreground'}"
                    on:click={() =>
                        editor?.chain().focus().toggleItalic().run()}
                    title={m.tiptap_editor_italic()}
                >
                    <Italic class="h-3.5 w-3.5" />
                </button>
                <button
                    type="button"
                    class="p-1.5 hover:bg-muted/50 transition-colors {editor?.isActive(
                        'strike',
                    )
                        ? 'bg-muted text-foreground'
                        : 'text-muted-foreground hover:text-foreground'}"
                    on:click={() =>
                        editor?.chain().focus().toggleStrike().run()}
                    title={m.tiptap_editor_strikethrough()}
                >
                    <Strikethrough class="h-3.5 w-3.5" />
                </button>
                <button
                    type="button"
                    class="p-1.5 hover:bg-muted/50 transition-colors {editor?.isActive(
                        'code',
                    ) || editor?.isActive('codeBlock')
                        ? 'bg-muted text-foreground'
                        : 'text-muted-foreground hover:text-foreground'}"
                    on:click={() => {
                        // If text is selected, toggle inline code
                        // If no selection or whole line, toggle code block
                        const { from, to } = editor.state.selection;
                        const hasSelection = from !== to;

                        if (hasSelection) {
                            editor?.chain().focus().toggleCode().run();
                        } else {
                            editor?.chain().focus().toggleCodeBlock().run();
                        }
                    }}
                    title={editor?.state.selection.from !==
                    editor?.state.selection.to
                        ? m.tiptap_editor_inline_code()
                        : m.tiptap_editor_code_block()}
                >
                    <Code class="h-3.5 w-3.5" />
                </button>
            </div>

            <!-- Headings -->
            <div class="flex items-center gap-1 pe-2">
                <button
                    type="button"
                    class="p-1.5 hover:bg-muted/50 transition-colors {editor?.isActive(
                        'heading',
                        { level: 1 },
                    )
                        ? 'bg-muted text-foreground'
                        : 'text-muted-foreground hover:text-foreground'}"
                    on:click={() =>
                        editor
                            ?.chain()
                            .focus()
                            .toggleHeading({ level: 1 })
                            .run()}
                    title={m.tiptap_editor_heading_1()}
                >
                    <Heading1 class="h-3.5 w-3.5" />
                </button>
                <button
                    type="button"
                    class="p-1.5 hover:bg-muted/50 transition-colors {editor?.isActive(
                        'heading',
                        { level: 2 },
                    )
                        ? 'bg-muted text-foreground'
                        : 'text-muted-foreground hover:text-foreground'}"
                    on:click={() =>
                        editor
                            ?.chain()
                            .focus()
                            .toggleHeading({ level: 2 })
                            .run()}
                    title={m.tiptap_editor_heading_2()}
                >
                    <Heading2 class="h-3.5 w-3.5" />
                </button>
            </div>

            <!-- Lists -->
            <div class="flex items-center gap-1 pe-2">
                <button
                    type="button"
                    class="p-1.5 hover:bg-muted/50 transition-colors {editor?.isActive(
                        'bulletList',
                    )
                        ? 'bg-muted text-foreground'
                        : 'text-muted-foreground hover:text-foreground'}"
                    on:click={() =>
                        editor?.chain().focus().toggleBulletList().run()}
                    title={m.tiptap_editor_bullet_list()}
                >
                    <List class="h-3.5 w-3.5" />
                </button>
                <button
                    type="button"
                    class="p-1.5 hover:bg-muted/50 transition-colors {editor?.isActive(
                        'orderedList',
                    )
                        ? 'bg-muted text-foreground'
                        : 'text-muted-foreground hover:text-foreground'}"
                    on:click={() =>
                        editor?.chain().focus().toggleOrderedList().run()}
                    title={m.tiptap_editor_numbered_list()}
                >
                    <ListOrdered class="h-3.5 w-3.5" />
                </button>
                <button
                    type="button"
                    class="p-1.5 hover:bg-muted/50 transition-colors {editor?.isActive(
                        'blockquote',
                    )
                        ? 'bg-muted text-foreground'
                        : 'text-muted-foreground hover:text-foreground'}"
                    on:click={() =>
                        editor?.chain().focus().toggleBlockquote().run()}
                    title={m.tiptap_editor_quote()}
                >
                    <Quote class="h-3.5 w-3.5" />
                </button>
            </div>

            <!-- Media & Links -->
            <div class="flex items-center gap-1 pe-2">
                <button
                    type="button"
                    class="p-1.5 hover:bg-muted/50 transition-colors {editor?.isActive(
                        'link',
                    )
                        ? 'bg-muted text-foreground'
                        : 'text-muted-foreground hover:text-foreground'}"
                    on:click={addLink}
                    title={m.tiptap_editor_add_link()}
                >
                    <LinkIcon class="h-3.5 w-3.5" />
                </button>
                <button
                    type="button"
                    class="p-1.5 hover:bg-muted/50 transition-colors text-muted-foreground hover:text-foreground"
                    on:click={(e) => {
                        e.preventDefault();
                        e.stopPropagation();
                        addImage();
                    }}
                    title={m.tiptap_editor_add_image()}
                >
                    <ImageIcon class="h-3.5 w-3.5" />
                </button>
                <button
                    type="button"
                    class="p-1.5 hover:bg-muted/50 transition-colors text-muted-foreground hover:text-foreground"
                    on:click={insertTable}
                    title={m.tiptap_editor_insert_table()}
                >
                    <TableIcon class="h-3.5 w-3.5" />
                </button>
            </div>

            <!-- Formatting -->
            <div class="flex items-center gap-1 pe-2">
                <button
                    type="button"
                    class="p-1.5 hover:bg-muted/50 transition-colors text-muted-foreground hover:text-foreground"
                    on:click={() =>
                        editor?.chain().focus().setHorizontalRule().run()}
                    title={m.tiptap_editor_horizontal_rule()}
                >
                    <Minus class="h-3.5 w-3.5" />
                </button>
            </div>

            <!-- Hidden file input -->
            <input
                type="file"
                accept="image/*,.ico"
                bind:this={fileInputElement}
                on:change={handleImageUpload}
                class="hidden"
            />

            <!-- Undo/Redo -->
            <div class="flex items-center gap-1">
                <button
                    type="button"
                    class="p-1.5 hover:bg-muted/50 transition-colors text-muted-foreground hover:text-foreground disabled:opacity-50 disabled:cursor-not-allowed"
                    on:click={() => editor?.chain().focus().undo().run()}
                    disabled={!editor?.can().undo()}
                    title={m.tiptap_editor_undo()}
                >
                    {#if isRtlLocale}
                        <Redo class="h-3.5 w-3.5" />
                    {:else}
                        <Undo class="h-3.5 w-3.5" />
                    {/if}
                </button>
                <button
                    type="button"
                    class="p-1.5 hover:bg-muted/50 transition-colors text-muted-foreground hover:text-foreground disabled:opacity-50 disabled:cursor-not-allowed"
                    on:click={() => editor?.chain().focus().redo().run()}
                    disabled={!editor?.can().redo()}
                    title={m.tiptap_editor_redo()}
                >
                    {#if isRtlLocale}
                        <Undo class="h-3.5 w-3.5" />
                    {:else}
                        <Redo class="h-3.5 w-3.5" />
                    {/if}
                </button>
            </div>
        </div>
    {/if}

    <!-- Floating Table Controls Popover -->
    {#if isInTable && tableElement}
        <div
            transition:scale={{ duration: 150, start: 0.95 }}
            class="fixed z-50 bg-popover border border-border rounded-md shadow-lg p-1"
            style="top: {popoverPosition.top}px; left: {popoverPosition.left}px; transform: translateX(-100%);"
        >
            <div class="flex items-center gap-0.5">
                <button
                    type="button"
                    class="p-1 hover:bg-muted/50 transition-colors text-muted-foreground hover:text-foreground rounded"
                    on:click={addRowBefore}
                    title="Add row before"
                >
                    <ArrowUpToLine class="h-3 w-3" />
                </button>
                <button
                    type="button"
                    class="p-1 hover:bg-muted/50 transition-colors text-muted-foreground hover:text-foreground rounded"
                    on:click={addRowAfter}
                    title="Add row after"
                >
                    <ArrowDownToLine class="h-3 w-3" />
                </button>
                <button
                    type="button"
                    class="p-1 hover:bg-muted/50 transition-colors text-muted-foreground hover:text-foreground rounded"
                    on:click={deleteRow}
                    title="Delete row"
                >
                    <ListMinus class="h-3 w-3" />
                </button>
                <div class="w-px h-3 bg-border mx-0.5"></div>
                <button
                    type="button"
                    class="p-1 hover:bg-muted/50 transition-colors text-muted-foreground hover:text-foreground rounded"
                    on:click={addColumnBefore}
                    title="Add column before"
                >
                    <ArrowLeftToLine class="h-3 w-3" />
                </button>
                <button
                    type="button"
                    class="p-1 hover:bg-muted/50 transition-colors text-muted-foreground hover:text-foreground rounded"
                    on:click={addColumnAfter}
                    title="Add column after"
                >
                    <ArrowRightToLine class="h-3 w-3" />
                </button>
                <button
                    type="button"
                    class="p-1 hover:bg-muted/50 transition-colors text-muted-foreground hover:text-foreground rounded"
                    on:click={deleteColumn}
                    title="Delete column"
                >
                    <Grid2x2X class="h-3 w-3" />
                </button>
                <div class="w-px h-3 bg-border mx-0.5"></div>
                <button
                    type="button"
                    class="p-1 hover:bg-destructive/10 transition-colors text-destructive hover:text-destructive rounded"
                    on:click={deleteTable}
                    title="Delete table"
                >
                    <Trash2 class="h-3 w-3" />
                </button>
            </div>
        </div>
    {/if}
</div>

<style>
    :global(.ProseMirror) {
        outline: none;
    }

    :global(.ProseMirror[data-placeholder]:empty::before) {
        content: attr(data-placeholder);
        color: hsl(var(--muted-foreground));
        pointer-events: none;
        position: absolute;
        top: 1rem;
        left: 1rem;
        font-size: 0.875rem;
        opacity: 0.7;
        z-index: 1;
    }

    :global(.ProseMirror:focus[data-placeholder]:empty::before) {
        opacity: 0.5;
    }

    /* Override any global prose styles for code elements */
    :global(.ProseMirror code) {
        background: hsl(var(--muted)) !important;
        color: hsl(var(--foreground)) !important;
        position: relative;
        border-radius: calc(var(--radius) - 2px);
        padding: 0.15rem 0.3rem !important;
        font-family:
            ui-monospace, SFMono-Regular, "Menlo", "Monaco", "Cascadia Code",
            "Segoe UI Mono", "Roboto Mono", "Oxygen Mono", "Ubuntu Monospace",
            "Source Code Pro", "Fira Code", "Droid Sans Mono", "Courier New",
            monospace !important;
        font-size: 0.875rem;
        font-weight: 600;
    }

    :global(.ProseMirror code::before),
    :global(.ProseMirror code::after) {
        content: none !important;
    }

    :global(.ProseMirror pre) {
        background: hsl(var(--muted)) !important;
        color: hsl(var(--foreground)) !important;
        padding: 1rem !important;
        border-radius: calc(var(--radius) - 2px);
        overflow-x: auto;
        margin: 1rem 0;
        font-family:
            ui-monospace, SFMono-Regular, "Menlo", "Monaco", "Cascadia Code",
            "Segoe UI Mono", "Roboto Mono", "Oxygen Mono", "Ubuntu Monospace",
            "Source Code Pro", "Fira Code", "Droid Sans Mono", "Courier New",
            monospace !important;
    }

    :global(.ProseMirror pre code) {
        background: none !important;
        padding: 0 !important;
        font-size: 0.875rem;
        font-weight: normal;
        border-radius: 0;
    }

    :global(.ProseMirror pre::before),
    :global(.ProseMirror pre::after),
    :global(.ProseMirror pre code::before),
    :global(.ProseMirror pre code::after) {
        content: none !important;
    }

    :global(.ProseMirror h1) {
        font-size: 1.25rem;
        font-weight: bold;
        margin: 0.75rem 0 0.25rem;
        line-height: 1.2;
    }

    :global(.ProseMirror h2) {
        font-size: 1.125rem;
        font-weight: bold;
        margin: 0.5rem 0 0.25rem;
        line-height: 1.3;
    }

    :global(.ProseMirror p) {
        margin: 0.25rem 0;
        line-height: 1.5;
    }

    :global(.ProseMirror ul),
    :global(.ProseMirror ol) {
        margin: 0.25rem 0;
        padding-inline-start: 1.25rem;
    }

    :global(.ProseMirror li) {
        margin: 0.25rem 0;
    }

    :global(.ProseMirror blockquote) {
        border-left: 4px solid hsl(var(--border));
        padding-inline-start: 1rem;
        margin: 1rem 0;
        font-style: italic;
        color: hsl(var(--muted-foreground));
    }

    :global(.ProseMirror table) {
        border-collapse: collapse;
        width: 100%;
        margin: 1rem 0;
        border: 1px solid hsl(var(--border));
        border-radius: 0.375rem;
        overflow: hidden;
    }

    :global(.ProseMirror th),
    :global(.ProseMirror td) {
        border: 1px solid hsl(var(--border));
        padding: 0.5rem 1rem;
        text-align: left;
    }

    :global(.ProseMirror th) {
        background-color: hsl(var(--muted));
        font-weight: 600;
    }

    :global(.ProseMirror hr) {
        border: none;
        border-top: 1px solid hsl(var(--border));
        margin: 2rem 0;
    }

    :global(.ProseMirror img) {
        max-width: 100%;
        height: auto;
        border-radius: 0.375rem;
        margin: 1rem 0;
    }

    :global(.ProseMirror a) {
        color: hsl(var(--primary));
        text-decoration: underline;
        cursor: pointer;
    }

    :global(.ProseMirror a:hover) {
        text-decoration: none;
    }
</style>
