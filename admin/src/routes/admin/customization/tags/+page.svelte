<script lang="ts">
    import { onMount } from "svelte";
    import { Button, Input, Badge } from "$lib/components/ui";
    import { api } from "$lib/api";
    import type { Tag, TagUsage } from "$lib/types";
    import {
        Plus,
        Edit2,
        Trash2,
        Save,
        X,
        Tag as TagIcon,
    } from "lucide-svelte";
    import { toast } from "svelte-sonner";
    import * as m from "$lib/paraglide/messages";

    let loading = true;
    let saving = false;
    let tags: Tag[] = [];
    let tagUsage: TagUsage[] = [];

    // New tag form
    let showNewTagForm = false;
    let newTagName = "";
    let newTagColor = "#3b82f6";

    // Editing
    let editingTag: Tag | null = null;
    let editTagName = "";
    let editTagColor = "";

    // Delete modal
    let showDeleteModal = false;
    let pendingDeleteTag: Tag | null = null;
    let pendingDeleteUsageCount = 0;

    // Color presets
    const colorPresets = [
        "#EF4444", // Red
        "#F97316", // Orange
        "#F59E0B", // Amber
        "#EAB308", // Yellow
        "#84CC16", // Lime
        "#22C55E", // Green
        "#10B981", // Emerald
        "#06B6D4", // Cyan
        "#0EA5E9", // Sky
        "#3B82F6", // Blue
        "#6366F1", // Indigo
        "#8B5CF6", // Violet
        "#A855F7", // Purple
        "#D946EF", // Fuchsia
        "#EC4899", // Pink
        "#F43F5E", // Rose
    ];

    onMount(async () => {
        await loadTags();
    });

    async function loadTags() {
        try {
            loading = true;

            // Load tags and their usage statistics
            const [tagsData, usageData] = await Promise.all([
                api.getTags(),
                api.getTagUsage(),
            ]);

            // Sort tags to put Feedback first, then alphabetically
            tags = tagsData.sort((a, b) => {
                if (a.name.toLowerCase() === "feedback") return -1;
                if (b.name.toLowerCase() === "feedback") return 1;
                return a.name.localeCompare(b.name);
            });
            tagUsage = usageData;
        } catch (err) {
            const errorMessage =
                err instanceof Error ? err.message : m.tags_load_failed();
            toast.error(m.tags_load_failed(), {
                description: errorMessage,
            });
        } finally {
            loading = false;
        }
    }

    async function createTag() {
        if (!newTagName.trim()) {
            toast.error(m.tags_name_required());
            return;
        }

        if (!isValidHexColor(newTagColor)) {
            toast.error(m.tags_invalid_color());
            return;
        }

        try {
            saving = true;

            await api.createTag({
                name: newTagName.trim(),
                color: newTagColor,
            });

            toast.success(m.tags_created(), {
                description: m.tags_created_description({
                    tagName: newTagName.trim(),
                }),
            });
            newTagName = "";
            newTagColor = "#3B82F6";
            showNewTagForm = false;
            await loadTags();
        } catch (err) {
            const errorMessage =
                err instanceof Error ? err.message : m.tags_create_failed();
            toast.error(m.tags_create_failed(), {
                description: errorMessage,
            });
        } finally {
            saving = false;
        }
    }

    function initiateDeleteTag(tag: Tag) {
        // Protect Feedback tag
        if (tag.name.toLowerCase() === "feedback") {
            toast.error(m.tags_feedback_cannot_delete());
            return;
        }

        const usage = tagUsage.find((u) => u.id === tag.id);
        pendingDeleteTag = tag;
        pendingDeleteUsageCount = usage?.count || 0;
        showDeleteModal = true;
    }

    function cancelDelete() {
        showDeleteModal = false;
        pendingDeleteTag = null;
        pendingDeleteUsageCount = 0;
    }

    async function confirmDelete() {
        if (!pendingDeleteTag) return;

        try {
            saving = true;

            const tagName = pendingDeleteTag.name;
            await api.deleteTag(pendingDeleteTag.id);
            toast.success(m.tags_deleted(), {
                description: m.tags_deleted_description({ tagName }),
            });
            await loadTags();
            showDeleteModal = false;
            pendingDeleteTag = null;
            pendingDeleteUsageCount = 0;
        } catch (err) {
            const errorMessage =
                err instanceof Error ? err.message : m.tags_delete_failed();
            toast.error(m.tags_delete_failed(), {
                description: errorMessage,
            });
        } finally {
            saving = false;
        }
    }

    function startEdit(tag: Tag) {
        editingTag = tag;
        editTagName = tag.name;
        editTagColor = tag.color;
    }

    function cancelEdit() {
        editingTag = null;
        editTagName = "";
        editTagColor = "";
    }

    async function saveEdit() {
        if (!editingTag) return;

        if (!editTagName.trim()) {
            toast.error(m.tags_name_required());
            return;
        }

        if (!isValidHexColor(editTagColor)) {
            toast.error(m.tags_invalid_color());
            return;
        }

        // For Feedback tag, only allow color changes
        const isUpdatingName =
            editingTag.name.toLowerCase() === "feedback" &&
            editTagName !== editingTag.name;
        if (isUpdatingName) {
            toast.error(m.tags_feedback_name_cannot_change());
            return;
        }

        try {
            saving = true;

            await api.updateTag(editingTag.id, {
                name: editTagName.trim(),
                color: editTagColor,
            });

            toast.success(m.tags_updated(), {
                description: m.tags_updated_description({
                    tagName: editTagName.trim(),
                }),
            });
            editingTag = null;
            editTagName = "";
            editTagColor = "";
            await loadTags();
        } catch (err) {
            const errorMessage =
                err instanceof Error ? err.message : m.tags_update_failed();
            toast.error(m.tags_update_failed(), {
                description: errorMessage,
            });
        } finally {
            saving = false;
        }
    }

    function isValidHexColor(color: string): boolean {
        return /^#[0-9A-F]{6}$/i.test(color);
    }

    function getUsageCount(tagId: number): number {
        const usage = tagUsage.find((u) => u.id === tagId);
        return usage?.count || 0;
    }
</script>

<svelte:head>
    <title>{m.tags_page_title()}</title>
</svelte:head>

<div class="max-w-4xl mx-auto">
    <div class="mb-8">
        <h1 class="text-xl font-semibold mb-1">{m.tags_heading()}</h1>
        <p class="text-muted-foreground text-sm">
            {m.tags_subheading()}
        </p>
    </div>

    {#if !showNewTagForm}
        <div class="mb-6 flex justify-end">
            <Button
                variant="outline"
                size="sm"
                on:click={() => {
                    showNewTagForm = true;
                }}
            >
                <Plus class="h-4 w-4 mr-2" />
                {m.tags_new_tag()}
            </Button>
        </div>
    {/if}

    {#if loading}
        <div class="flex items-center justify-center min-h-64">
            <div
                class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"
            ></div>
        </div>
    {:else}
        <div class="mb-6">
            {#if showNewTagForm}
                <div class="border border-border rounded-lg p-4 bg-card">
                    <h3 class="text-sm font-medium mb-3">
                        {m.tags_create_new_tag()}
                    </h3>
                    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                        <div>
                            <label
                                class="block text-sm font-medium mb-1"
                                for="new-tag-name"
                            >
                                {m.tags_tag_name()}
                            </label>
                            <Input
                                id="new-tag-name"
                                bind:value={newTagName}
                                placeholder={m.tags_name_placeholder()}
                                disabled={saving}
                            />
                        </div>
                        <div>
                            <label
                                class="block text-sm font-medium mb-1"
                                for="new-tag-color"
                            >
                                {m.tags_color()}
                            </label>
                            <div class="flex gap-2">
                                <input
                                    type="color"
                                    bind:value={newTagColor}
                                    class="w-10 h-10 rounded border border-input"
                                    disabled={saving}
                                />
                                <Input
                                    bind:value={newTagColor}
                                    placeholder="#3B82F6"
                                    disabled={saving}
                                />
                            </div>
                        </div>
                    </div>

                    <div class="mt-4">
                        <p class="text-sm font-medium mb-2">
                            {m.tags_color_presets()}
                        </p>
                        <div class="flex flex-wrap gap-2">
                            {#each colorPresets as preset}
                                <button
                                    type="button"
                                    class="w-6 h-6 rounded border border-input cursor-pointer hover:scale-110 transition-transform"
                                    style="background-color: {preset}"
                                    on:click={() => (newTagColor = preset)}
                                    disabled={saving}
                                    aria-label={m.tags_select_color({
                                        color: preset,
                                    })}
                                ></button>
                            {/each}
                        </div>
                    </div>

                    <div class="flex justify-end gap-2 mt-4">
                        <Button
                            variant="ghost"
                            size="sm"
                            on:click={() => {
                                showNewTagForm = false;
                                newTagName = "";
                                newTagColor = "#3B82F6";
                            }}
                            disabled={saving}
                        >
                            <X class="h-4 w-4 mr-2" />
                            {m.tags_cancel()}
                        </Button>
                        <Button
                            size="sm"
                            on:click={createTag}
                            disabled={saving}
                        >
                            {#if saving}
                                <div
                                    class="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"
                                ></div>
                            {:else}
                                <Save class="h-4 w-4 mr-2" />
                            {/if}
                            {m.tags_create_tag()}
                        </Button>
                    </div>
                </div>
            {/if}

            {#if tags.length === 0}
                <div
                    class="text-center py-12 border border-border rounded-lg bg-card"
                >
                    <TagIcon class="h-12 w-12 mx-auto mb-3 opacity-50" />
                    <p class="text-lg font-medium mb-1">{m.tags_no_tags()}</p>
                    <p class="text-sm text-muted-foreground">
                        {m.tags_create_first()}
                    </p>
                </div>
            {:else}
                <div class="border border-border rounded-lg overflow-hidden">
                    <table class="w-full">
                        <thead class="bg-muted/50 border-b border-border">
                            <tr>
                                <th
                                    class="text-left py-3 px-3 font-medium text-sm"
                                >
                                    {m.tags_table_tag()}
                                </th>
                                <th
                                    class="text-left py-3 px-3 font-medium text-sm"
                                >
                                    {m.tags_table_preview()}
                                </th>
                                <th
                                    class="text-left py-3 px-3 font-medium text-sm"
                                >
                                    {m.tags_table_color()}
                                </th>
                                <th
                                    class="text-left py-3 px-3 font-medium text-sm"
                                >
                                    {m.tags_table_usage()}
                                </th>
                                <th
                                    class="text-right py-3 px-3 font-medium text-sm"
                                >
                                    {m.tags_table_actions()}
                                </th>
                            </tr>
                        </thead>
                        <tbody>
                            {#each tags as tag (tag.id)}
                                <tr
                                    class="border-b border-border last:border-b-0 hover:bg-muted/25"
                                >
                                    {#if editingTag?.id === tag.id}
                                        <td class="py-2 px-3">
                                            <div
                                                class="flex items-center gap-2"
                                            >
                                                <div
                                                    class="w-3 h-3 rounded-full border border-border"
                                                    style="background-color: {editTagColor}"
                                                ></div>
                                                <Input
                                                    bind:value={editTagName}
                                                    class="text-sm"
                                                    disabled={saving ||
                                                        tag.name.toLowerCase() ===
                                                            "feedback"}
                                                    placeholder={tag.name.toLowerCase() ===
                                                    "feedback"
                                                        ? m.tags_name_cannot_change()
                                                        : ""}
                                                />
                                            </div>
                                        </td>
                                        <td class="py-2 px-3">
                                            <Badge
                                                style="background-color: {editTagColor}20; color: {editTagColor}; border-color: {editTagColor}"
                                            >
                                                {editTagName}
                                            </Badge>
                                        </td>
                                        <td class="py-2 px-3">
                                            <div
                                                class="flex gap-2 items-center"
                                            >
                                                <input
                                                    type="color"
                                                    bind:value={editTagColor}
                                                    class="w-6 h-6 rounded border border-input"
                                                    disabled={saving}
                                                />
                                                <Input
                                                    bind:value={editTagColor}
                                                    class="text-xs font-mono"
                                                    disabled={saving}
                                                />
                                            </div>
                                        </td>
                                        <td class="py-2 px-3">
                                            <span
                                                class="text-sm text-muted-foreground"
                                            >
                                                {m.tags_usage_count({
                                                    count: getUsageCount(
                                                        tag.id,
                                                    ).toString(),
                                                })}
                                            </span>
                                        </td>
                                        <td class="py-2 px-3 text-right">
                                            <div class="flex justify-end gap-1">
                                                <Button
                                                    variant="ghost"
                                                    size="sm"
                                                    on:click={cancelEdit}
                                                    disabled={saving}
                                                >
                                                    <X class="h-4 w-4" />
                                                </Button>
                                                <Button
                                                    variant="ghost"
                                                    size="sm"
                                                    on:click={saveEdit}
                                                    disabled={saving}
                                                >
                                                    {#if saving}
                                                        <div
                                                            class="animate-spin rounded-full h-4 w-4 border-b-2 border-current"
                                                        ></div>
                                                    {:else}
                                                        <Save class="h-4 w-4" />
                                                    {/if}
                                                </Button>
                                            </div>
                                        </td>
                                    {:else}
                                        <td class="py-2 px-3">
                                            <div
                                                class="flex items-center gap-2"
                                            >
                                                <div
                                                    class="w-3 h-3 rounded-full border border-border"
                                                    style="background-color: {tag.color}"
                                                ></div>
                                                <span
                                                    class="font-medium text-sm"
                                                    >{tag.name}</span
                                                >
                                                {#if tag.name.toLowerCase() === "feedback"}
                                                    <span
                                                        class="text-xs text-muted-foreground"
                                                        title={m.tags_system_tag_tooltip()}
                                                        >🔒</span
                                                    >
                                                {/if}
                                            </div>
                                        </td>
                                        <td class="py-2 px-3">
                                            <Badge
                                                style="background-color: {tag.color}20; color: {tag.color}; border-color: {tag.color}"
                                            >
                                                {tag.name}
                                                {#if tag.name.toLowerCase() === "feedback"}
                                                    <span
                                                        class="ml-1 text-xs opacity-60"
                                                        >🔒</span
                                                    >
                                                {/if}
                                            </Badge>
                                        </td>
                                        <td class="py-2 px-3">
                                            <code
                                                class="text-xs bg-muted px-2 py-1 rounded"
                                                >{tag.color}</code
                                            >
                                        </td>
                                        <td class="py-2 px-3">
                                            <span
                                                class="text-sm text-muted-foreground"
                                            >
                                                {m.tags_usage_count({
                                                    count: getUsageCount(
                                                        tag.id,
                                                    ).toString(),
                                                })}
                                            </span>
                                        </td>
                                        <td class="py-2 px-3 text-right">
                                            <div class="flex justify-end gap-1">
                                                <Button
                                                    variant="ghost"
                                                    size="sm"
                                                    on:click={() =>
                                                        startEdit(tag)}
                                                    disabled={saving}
                                                    title={tag.name.toLowerCase() ===
                                                    "feedback"
                                                        ? m.tags_edit_color_only()
                                                        : m.tags_edit_tag()}
                                                >
                                                    <Edit2 class="h-4 w-4" />
                                                </Button>
                                                <Button
                                                    variant="ghost"
                                                    size="sm"
                                                    on:click={() =>
                                                        initiateDeleteTag(tag)}
                                                    disabled={saving ||
                                                        tag.name.toLowerCase() ===
                                                            "feedback"}
                                                    class={tag.name.toLowerCase() ===
                                                    "feedback"
                                                        ? "opacity-50 cursor-not-allowed"
                                                        : getUsageCount(
                                                                tag.id,
                                                            ) > 0
                                                          ? "text-orange-600 hover:text-orange-700"
                                                          : ""}
                                                    title={tag.name.toLowerCase() ===
                                                    "feedback"
                                                        ? m.tags_feedback_cannot_delete_tooltip()
                                                        : getUsageCount(
                                                                tag.id,
                                                            ) > 0
                                                          ? m.tags_delete_warning(
                                                                {
                                                                    count: getUsageCount(
                                                                        tag.id,
                                                                    ).toString(),
                                                                },
                                                            )
                                                          : m.tags_delete_tag()}
                                                >
                                                    <Trash2 class="h-4 w-4" />
                                                </Button>
                                            </div>
                                        </td>
                                    {/if}
                                </tr>
                            {/each}
                        </tbody>
                    </table>
                </div>
            {/if}
        </div>
    {/if}
</div>

{#if showDeleteModal && pendingDeleteTag}
    <div
        class="fixed inset-0 bg-black/50 flex items-center justify-center z-50"
        on:click={cancelDelete}
        on:keydown={(e) => e.key === "Escape" && cancelDelete()}
        role="dialog"
        aria-modal="true"
        tabindex="-1"
    >
        <div
            class="bg-background rounded-lg p-5 w-full max-w-sm space-y-4"
            on:click={(e) => e.stopPropagation()}
            role="none"
        >
            <h2 class="text-sm font-semibold">
                {m.tags_delete_modal_title({ tagName: pendingDeleteTag.name })}
            </h2>
            <p class="text-xs text-muted-foreground">
                {#if pendingDeleteUsageCount > 0}
                    <span
                        class="text-orange-600 dark:text-orange-400 font-medium"
                    >
                        {m.tags_warning()}
                    </span>
                    {m.tags_delete_modal_warning({
                        count: pendingDeleteUsageCount.toString(),
                    })}
                {:else}
                    {m.tags_delete_modal_cannot_undo()}
                {/if}
            </p>
            <div class="flex justify-end gap-2 text-xs">
                <Button
                    variant="outline"
                    size="sm"
                    on:click={cancelDelete}
                    disabled={saving}
                >
                    {m.tags_cancel()}
                </Button>
                <Button
                    variant="destructive"
                    size="sm"
                    on:click={confirmDelete}
                    disabled={saving}
                >
                    {#if saving}
                        {m.tags_deleting()}
                    {:else}
                        {m.tags_delete_tag_button()}
                    {/if}
                </Button>
            </div>
        </div>
    </div>
{/if}
