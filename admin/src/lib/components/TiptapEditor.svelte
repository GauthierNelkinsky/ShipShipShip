<script lang="ts">
    import { onMount, onDestroy, createEventDispatcher } from "svelte";
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
    } from "lucide-svelte";
    import ImageUploadModal from "./ImageUploadModal.svelte";

    export let content = "";
    export let placeholder = "Start writing...";
    export let showToolbar = true;

    const dispatch = createEventDispatcher();

    let editor: Editor;
    let element: HTMLElement;
    let imageModalOpen = false;

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
            content: content,
            onTransaction: () => {
                // Force re-render on transaction
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
                    "data-placeholder": placeholder,
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
        editor.commands.setContent(content);
    }

    function addLink() {
        const url = window.prompt("Enter URL:");
        if (url) {
            editor.chain().focus().setLink({ href: url }).run();
        }
    }

    function addImage() {
        imageModalOpen = true;
    }

    function handleImageSelected(event: CustomEvent) {
        const { url } = event.detail;
        if (url) {
            editor.chain().focus().setImage({ src: url }).run();
        }
    }

    function insertTable() {
        editor
            .chain()
            .focus()
            .insertTable({ rows: 3, cols: 3, withHeaderRow: true })
            .run();
    }
</script>

<div class="flex flex-col h-full">
    <!-- Editor -->
    <div bind:this={element} class="flex-1 min-h-0 bg-background"></div>

    <!-- Toolbar -->
    {#if showToolbar}
        <div
            class="flex flex-wrap items-center gap-1 p-2 bg-background/95 backdrop-blur-sm border-t border-border/50 shrink-0"
        >
            <!-- Text formatting -->
            <div class="flex items-center gap-1 pr-2">
                <button
                    type="button"
                    class="p-1.5 hover:bg-muted/50 transition-colors {editor?.isActive(
                        'bold',
                    )
                        ? 'bg-muted text-foreground'
                        : 'text-muted-foreground hover:text-foreground'}"
                    on:click={() => editor?.chain().focus().toggleBold().run()}
                    title="Bold"
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
                    title="Italic"
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
                    title="Strikethrough"
                >
                    <Strikethrough class="h-3.5 w-3.5" />
                </button>
                <button
                    type="button"
                    class="p-1.5 hover:bg-muted/50 transition-colors {editor?.isActive(
                        'code',
                    )
                        ? 'bg-muted text-foreground'
                        : 'text-muted-foreground hover:text-foreground'}"
                    on:click={() => editor?.chain().focus().toggleCode().run()}
                    title="Inline Code"
                >
                    <Code class="h-3.5 w-3.5" />
                </button>
            </div>

            <!-- Headings -->
            <div class="flex items-center gap-1 pr-2">
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
                    title="Heading 1"
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
                    title="Heading 2"
                >
                    <Heading2 class="h-3.5 w-3.5" />
                </button>
            </div>

            <!-- Lists -->
            <div class="flex items-center gap-1 pr-2">
                <button
                    type="button"
                    class="p-1.5 hover:bg-muted/50 transition-colors {editor?.isActive(
                        'bulletList',
                    )
                        ? 'bg-muted text-foreground'
                        : 'text-muted-foreground hover:text-foreground'}"
                    on:click={() =>
                        editor?.chain().focus().toggleBulletList().run()}
                    title="Bullet List"
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
                    title="Numbered List"
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
                    title="Quote"
                >
                    <Quote class="h-3.5 w-3.5" />
                </button>
            </div>

            <!-- Media & Links -->
            <div class="flex items-center gap-1 pr-2">
                <button
                    type="button"
                    class="p-1.5 hover:bg-muted/50 transition-colors {editor?.isActive(
                        'link',
                    )
                        ? 'bg-muted text-foreground'
                        : 'text-muted-foreground hover:text-foreground'}"
                    on:click={addLink}
                    title="Add Link"
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
                    title="Add Image"
                >
                    <ImageIcon class="h-3.5 w-3.5" />
                </button>
                <button
                    type="button"
                    class="p-1.5 hover:bg-muted/50 transition-colors text-muted-foreground hover:text-foreground"
                    on:click={insertTable}
                    title="Insert Table"
                >
                    <TableIcon class="h-3.5 w-3.5" />
                </button>
            </div>

            <!-- Formatting -->
            <div class="flex items-center gap-1 pr-2">
                <button
                    type="button"
                    class="p-1.5 hover:bg-muted/50 transition-colors {editor?.isActive(
                        'codeBlock',
                    )
                        ? 'bg-muted text-foreground'
                        : 'text-muted-foreground hover:text-foreground'}"
                    on:click={() =>
                        editor?.chain().focus().toggleCodeBlock().run()}
                    title="Code Block"
                >
                    <Code class="h-3.5 w-3.5" />
                </button>
                <button
                    type="button"
                    class="p-1.5 hover:bg-muted/50 transition-colors text-muted-foreground hover:text-foreground"
                    on:click={() =>
                        editor?.chain().focus().setHorizontalRule().run()}
                    title="Horizontal Rule"
                >
                    <Minus class="h-3.5 w-3.5" />
                </button>
            </div>

            <ImageUploadModal
                bind:isOpen={imageModalOpen}
                on:imageSelected={handleImageSelected}
            />

            <!-- Undo/Redo -->
            <div class="flex items-center gap-1">
                <button
                    type="button"
                    class="p-1.5 hover:bg-muted/50 transition-colors text-muted-foreground hover:text-foreground disabled:opacity-50 disabled:cursor-not-allowed"
                    on:click={() => editor?.chain().focus().undo().run()}
                    disabled={!editor?.can().undo()}
                    title="Undo"
                >
                    <Undo class="h-3.5 w-3.5" />
                </button>
                <button
                    type="button"
                    class="p-1.5 hover:bg-muted/50 transition-colors text-muted-foreground hover:text-foreground disabled:opacity-50 disabled:cursor-not-allowed"
                    on:click={() => editor?.chain().focus().redo().run()}
                    disabled={!editor?.can().redo()}
                    title="Redo"
                >
                    <Redo class="h-3.5 w-3.5" />
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
        padding-left: 1.25rem;
    }

    :global(.ProseMirror li) {
        margin: 0.25rem 0;
    }

    :global(.ProseMirror blockquote) {
        border-left: 4px solid hsl(var(--border));
        padding-left: 1rem;
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
