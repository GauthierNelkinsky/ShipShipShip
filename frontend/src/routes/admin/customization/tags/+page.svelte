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
        AlertTriangle,
    } from "lucide-svelte";

    let loading = true;
    let saving = false;
    let error = "";
    let success = "";
    let tags: Tag[] = [];
    let tagUsage: TagUsage[] = [];

    // New tag form
    let showNewTagForm = false;
    let newTagName = "";
    let newTagColor = "#3B82F6";

    // Edit tag
    let editingTag: Tag | null = null;
    let editTagName = "";
    let editTagColor = "";

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
            error = "";

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
            error = err instanceof Error ? err.message : "Failed to load tags";
        } finally {
            loading = false;
        }
    }

    async function createTag() {
        if (!newTagName.trim()) {
            error = "Tag name is required";
            return;
        }

        if (!isValidHexColor(newTagColor)) {
            error = "Please enter a valid hex color (e.g., #FF0000)";
            return;
        }

        try {
            saving = true;
            error = "";

            await api.createTag({
                name: newTagName.trim(),
                color: newTagColor,
            });

            success = "Tag created successfully";
            newTagName = "";
            newTagColor = "#3B82F6";
            showNewTagForm = false;
            await loadTags();
        } catch (err) {
            error = err instanceof Error ? err.message : "Failed to create tag";
        } finally {
            saving = false;
        }
    }

    async function deleteTag(tag: Tag) {
        // Protect Feedback tag
        if (tag.name.toLowerCase() === "feedback") {
            error =
                "The 'Feedback' tag cannot be deleted as it's used by the system.";
            return;
        }

        const usage = tagUsage.find((u) => u.id === tag.id);
        let confirmMessage;

        if (usage && usage.count > 0) {
            confirmMessage = `⚠️ WARNING: The tag "${tag.name}" is currently used by ${usage.count} event(s).\n\nDeleting this tag will remove it from all events that use it.\n\nAre you sure you want to continue?`;
        } else {
            confirmMessage = `Are you sure you want to delete the tag "${tag.name}"?`;
        }

        if (!confirm(confirmMessage)) {
            return;
        }

        try {
            saving = true;
            error = "";

            await api.deleteTag(tag.id);
            success = "Tag deleted successfully";
            await loadTags();
        } catch (err) {
            error = err instanceof Error ? err.message : "Failed to delete tag";
        } finally {
            saving = false;
        }
    }

    function startEdit(tag: Tag) {
        editingTag = tag;
        editTagName = tag.name;
        editTagColor = tag.color;
        clearMessages();
    }

    function cancelEdit() {
        editingTag = null;
        editTagName = "";
        editTagColor = "";
    }

    async function saveEdit() {
        if (!editingTag) return;

        if (!editTagName.trim()) {
            error = "Tag name is required";
            return;
        }

        if (!isValidHexColor(editTagColor)) {
            error = "Please enter a valid hex color (e.g., #FF0000)";
            return;
        }

        // For Feedback tag, only allow color changes
        const isUpdatingName =
            editingTag.name.toLowerCase() === "feedback" &&
            editTagName !== editingTag.name;
        if (isUpdatingName) {
            error = "The name of the 'Feedback' tag cannot be changed.";
            return;
        }

        try {
            saving = true;
            error = "";

            await api.updateTag(editingTag.id, {
                name: editTagName.trim(),
                color: editTagColor,
            });

            success = "Tag updated successfully";
            editingTag = null;
            editTagName = "";
            editTagColor = "";
            await loadTags();
        } catch (err) {
            error = err instanceof Error ? err.message : "Failed to update tag";
        } finally {
            saving = false;
        }
    }

    function clearMessages() {
        error = "";
        success = "";
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
    <title>Tag Management - Admin</title>
</svelte:head>

<div class="max-w-4xl mx-auto">
    <div class="mb-6">
        <div class="flex items-center justify-between">
            <div>
                <h1 class="text-xl font-semibold mb-1">Tag Management</h1>
                <p class="text-muted-foreground text-sm">
                    Create and manage tags for organizing your events. Each tag
                    can have a custom color for visual distinction.
                </p>
            </div>
            {#if !showNewTagForm}
                <Button
                    variant="outline"
                    size="sm"
                    on:click={() => {
                        showNewTagForm = true;
                        clearMessages();
                    }}
                >
                    <Plus class="h-4 w-4 mr-2" />
                    New Tag
                </Button>
            {/if}
        </div>
    </div>

    {#if loading}
        <div class="flex items-center justify-center min-h-64">
            <div
                class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"
            ></div>
        </div>
    {:else}
        {#if success}
            <div
                class="mb-4 p-3 bg-green-50 border border-green-200 rounded-md text-green-800 text-sm"
            >
                {success}
            </div>
        {/if}

        {#if error}
            <div
                class="mb-4 p-3 bg-red-50 border border-red-200 rounded-md text-red-800 text-sm flex items-center"
            >
                <AlertTriangle class="h-4 w-4 mr-2" />
                {error}
            </div>
        {/if}

        <div class="mb-6">
            {#if showNewTagForm}
                <div class="border border-border rounded-lg p-4 bg-card">
                    <h3 class="text-sm font-medium mb-3">Create New Tag</h3>
                    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                        <div>
                            <label
                                class="block text-sm font-medium mb-1"
                                for="new-tag-name"
                            >
                                Tag Name
                            </label>
                            <Input
                                id="new-tag-name"
                                bind:value={newTagName}
                                placeholder="Enter tag name"
                                disabled={saving}
                            />
                        </div>
                        <div>
                            <label
                                class="block text-sm font-medium mb-1"
                                for="new-tag-color"
                            >
                                Color
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
                        <p class="text-sm font-medium mb-2">Color Presets</p>
                        <div class="flex flex-wrap gap-2">
                            {#each colorPresets as preset}
                                <button
                                    type="button"
                                    class="w-6 h-6 rounded border border-input cursor-pointer hover:scale-110 transition-transform"
                                    style="background-color: {preset}"
                                    on:click={() => (newTagColor = preset)}
                                    disabled={saving}
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
                                clearMessages();
                            }}
                            disabled={saving}
                        >
                            <X class="h-4 w-4 mr-2" />
                            Cancel
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
                            Create Tag
                        </Button>
                    </div>
                </div>
            {/if}

            {#if tags.length === 0}
                <div
                    class="text-center py-12 border border-border rounded-lg bg-card"
                >
                    <TagIcon class="h-12 w-12 mx-auto mb-3 opacity-50" />
                    <p class="text-lg font-medium mb-1">No tags created yet</p>
                    <p class="text-sm text-muted-foreground">
                        Create your first tag to start organizing your events.
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
                                    Tag
                                </th>
                                <th
                                    class="text-left py-3 px-3 font-medium text-sm"
                                >
                                    Preview
                                </th>
                                <th
                                    class="text-left py-3 px-3 font-medium text-sm"
                                >
                                    Color
                                </th>
                                <th
                                    class="text-left py-3 px-3 font-medium text-sm"
                                >
                                    Usage
                                </th>
                                <th
                                    class="text-right py-3 px-3 font-medium text-sm"
                                >
                                    Actions
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
                                                        ? "Name cannot be changed"
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
                                                {getUsageCount(tag.id)} events
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
                                                        title="System tag - name cannot be changed"
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
                                                {getUsageCount(tag.id)} events
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
                                                        ? "Edit color only"
                                                        : "Edit tag"}
                                                >
                                                    <Edit2 class="h-4 w-4" />
                                                </Button>
                                                <Button
                                                    variant="ghost"
                                                    size="sm"
                                                    on:click={() =>
                                                        deleteTag(tag)}
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
                                                        ? "Feedback tag cannot be deleted"
                                                        : getUsageCount(
                                                                tag.id,
                                                            ) > 0
                                                          ? `⚠️ Will remove from ${getUsageCount(tag.id)} events`
                                                          : "Delete tag"}
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
