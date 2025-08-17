<script lang="ts">
    import { onMount } from "svelte";
    import { Button, Input, Badge } from "$lib/components/ui";
    import { api } from "$lib/api";
    import type { Event } from "$lib/types";
    import { tagColorStore } from "$lib/stores/tagColors";
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
    let allEvents: Event[] = [];
    let existingTags: string[] = [];
    let tagUsageCount: Record<string, number> = {};

    // New tag form
    let showNewTagForm = false;
    let newTagName = "";
    let newTagColor = "#3b82f6";

    // Edit tag
    let editingTag: string | null = null;
    let editTagName = "";
    let editTagColor = "";

    const colorPresets = [
        { name: "Blue", value: "#3b82f6" },
        { name: "Red", value: "#ef4444" },
        { name: "Green", value: "#10b981" },
        { name: "Yellow", value: "#f59e0b" },
        { name: "Purple", value: "#8b5cf6" },
        { name: "Pink", value: "#ec4899" },
        { name: "Cyan", value: "#06b6d4" },
        { name: "Lime", value: "#84cc16" },
        { name: "Orange", value: "#f97316" },
        { name: "Indigo", value: "#6366f1" },
    ];

    onMount(async () => {
        tagColorStore.init();
        await loadTags();
    });

    async function loadTags() {
        try {
            loading = true;
            error = "";

            // Get all events to extract tags
            allEvents = await api.getAllEvents();
            const allTagsSet = new Set<string>();
            const usageCount: Record<string, number> = {};

            allEvents.forEach((event) => {
                try {
                    const eventTags = event.tags ? JSON.parse(event.tags) : [];
                    eventTags.forEach((tag: string) => {
                        allTagsSet.add(tag);
                        usageCount[tag] = (usageCount[tag] || 0) + 1;
                    });
                } catch (e) {
                    console.error("Error parsing tags for event:", event.id, e);
                }
            });

            existingTags = Array.from(allTagsSet).sort();
            tagUsageCount = usageCount;
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

        if (
            existingTags.some(
                (tag) => tag.toLowerCase() === newTagName.trim().toLowerCase(),
            )
        ) {
            error = "A tag with this name already exists";
            return;
        }

        try {
            saving = true;
            error = "";
            success = "";

            // Add to local state
            const tagName = newTagName.trim();
            existingTags = [...existingTags, tagName].sort();
            tagColorStore.setColor(tagName, newTagColor);
            tagUsageCount[tagName] = 0;

            success = "Tag created successfully";
            newTagName = "";
            newTagColor = "#3b82f6";
            showNewTagForm = false;
        } catch (err) {
            error = err instanceof Error ? err.message : "Failed to create tag";
        } finally {
            saving = false;
        }
    }

    async function deleteTag(tagName: string) {
        // Protect Feedback tag
        if (tagName.toLowerCase() === "feedback") {
            error =
                "The 'Feedback' tag cannot be deleted as it's used by the system.";
            return;
        }

        const usageCount = tagUsageCount[tagName] || 0;

        if (usageCount > 0) {
            if (
                !confirm(
                    `This tag is used by ${usageCount} event(s). Deleting it will remove it from all events. Are you sure?`,
                )
            ) {
                return;
            }
        } else {
            if (
                !confirm(
                    `Are you sure you want to delete the tag "${tagName}"?`,
                )
            ) {
                return;
            }
        }

        try {
            saving = true;
            error = "";
            success = "";

            // Remove tag from all events that use it
            const eventsToUpdate = allEvents.filter((event) => {
                try {
                    const eventTags = event.tags ? JSON.parse(event.tags) : [];
                    return eventTags.includes(tagName);
                } catch {
                    return false;
                }
            });

            // Update each event
            for (const event of eventsToUpdate) {
                try {
                    const eventTags = JSON.parse(event.tags);
                    const updatedTags = eventTags.filter(
                        (tag: string) => tag !== tagName,
                    );

                    await api.updateEvent(event.id, {
                        tags: updatedTags,
                    });
                } catch (e) {
                    console.error("Error updating event:", event.id, e);
                }
            }

            // Remove from tag colors
            tagColorStore.removeColor(tagName);

            success = "Tag deleted successfully";
            await loadTags(); // Reload to reflect changes
        } catch (err) {
            error = err instanceof Error ? err.message : "Failed to delete tag";
        } finally {
            saving = false;
        }
    }

    function startEdit(tag: string) {
        // Protect Feedback tag
        if (tag.toLowerCase() === "feedback") {
            error =
                "The 'Feedback' tag cannot be edited as it's used by the system.";
            return;
        }

        editingTag = tag;
        editTagName = tag;
        editTagColor = tagColorStore.getColor(tag);
        error = "";
        success = "";
    }

    function cancelEdit() {
        editingTag = null;
        editTagName = "";
        editTagColor = "";
        error = "";
    }

    async function saveEdit() {
        if (!editTagName.trim()) {
            error = "Tag name is required";
            return;
        }

        if (
            editTagName !== editingTag &&
            existingTags.some(
                (tag) => tag.toLowerCase() === editTagName.trim().toLowerCase(),
            )
        ) {
            error = "A tag with this name already exists";
            return;
        }

        try {
            saving = true;
            error = "";
            success = "";

            const oldTag = editingTag!;
            const newTag = editTagName.trim();

            // Get events that use this tag
            const eventsToUpdate = allEvents.filter((event) => {
                try {
                    const eventTags = event.tags ? JSON.parse(event.tags) : [];
                    return eventTags.includes(oldTag);
                } catch {
                    return false;
                }
            });

            // Update all events that use this tag
            for (const event of eventsToUpdate) {
                try {
                    const eventTags = JSON.parse(event.tags);
                    const updatedTags = eventTags.map((tag: string) =>
                        tag === oldTag ? newTag : tag,
                    );

                    await api.updateEvent(event.id, {
                        tags: updatedTags,
                    });
                } catch (e) {
                    console.error("Error updating event:", event.id, e);
                }
            }

            // Update local state
            const tagIndex = existingTags.indexOf(oldTag);
            if (tagIndex !== -1) {
                if (oldTag !== newTag) {
                    // Name changed - update the tag name in the array
                    existingTags[tagIndex] = newTag;
                    existingTags = [...existingTags].sort();

                    // Transfer color and usage count to new name
                    tagColorStore.renameTag(oldTag, newTag);
                    tagColorStore.setColor(newTag, editTagColor);
                    tagUsageCount[newTag] = tagUsageCount[oldTag] || 0;
                    delete tagUsageCount[oldTag];
                } else {
                    // Only color changed - just update the color
                    tagColorStore.setColor(oldTag, editTagColor);
                }
            }

            // Reload events to ensure consistency
            await loadTags();

            success = "Tag updated successfully";
            editingTag = null;
            editTagName = "";
            editTagColor = "";
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

    $: if (success) {
        setTimeout(() => {
            success = "";
        }, 3000);
    }
</script>

<svelte:head>
    <title>Tags - Admin</title>
</svelte:head>

<div class="max-w-4xl mx-auto">
    <div class="mb-6">
        <div class="flex items-center justify-between">
            <div>
                <h1 class="text-xl font-semibold mb-1">Tags</h1>
                <p class="text-muted-foreground text-sm">
                    Manage tags used in your changelog entries. Create new tags
                    or edit existing ones.
                </p>
            </div>
            {#if !showNewTagForm}
                <Button
                    on:click={() => {
                        showNewTagForm = true;
                        clearMessages();
                    }}
                    size="sm"
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
                class="mb-4 p-3 bg-green-50 border border-green-200 rounded-lg text-green-800 text-sm"
            >
                {success}
            </div>
        {/if}

        {#if error}
            <div
                class="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg text-red-800 text-sm"
            >
                {error}
            </div>
        {/if}

        <!-- Create New Tag -->
        <div class="mb-6">
            {#if showNewTagForm}
                <div
                    class="border border-border rounded-lg p-4 mb-4 bg-muted/50"
                >
                    <h3 class="text-sm font-medium mb-3">Create New Tag</h3>
                    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                        <div>
                            <label
                                for="newTagName"
                                class="block text-sm font-medium mb-2"
                            >
                                Tag Name
                            </label>
                            <Input
                                id="newTagName"
                                bind:value={newTagName}
                                placeholder="Enter tag name"
                                disabled={saving}
                            />
                        </div>
                        <div>
                            <label
                                for="newTagColor"
                                class="block text-sm font-medium mb-2"
                            >
                                Color
                            </label>
                            <div class="flex gap-2">
                                <input
                                    id="newTagColor"
                                    type="color"
                                    bind:value={newTagColor}
                                    class="w-12 h-10 rounded border border-border cursor-pointer"
                                    disabled={saving}
                                />
                                <Input
                                    bind:value={newTagColor}
                                    placeholder="#3b82f6"
                                    class="flex-1"
                                    disabled={saving}
                                />
                            </div>
                        </div>
                    </div>

                    <!-- Color Presets -->
                    <div class="mt-4">
                        <p class="text-sm font-medium mb-2">Quick Colors</p>
                        <div class="flex flex-wrap gap-2">
                            {#each colorPresets as preset}
                                <button
                                    type="button"
                                    class="w-6 h-6 rounded border border-border cursor-pointer hover:scale-110 transition-transform"
                                    style="background-color: {preset.value}"
                                    title={preset.name}
                                    on:click={() =>
                                        (newTagColor = preset.value)}
                                    disabled={saving}
                                ></button>
                            {/each}
                        </div>
                    </div>

                    <div class="flex justify-end gap-2 mt-4">
                        <Button
                            variant="outline"
                            size="sm"
                            on:click={() => {
                                showNewTagForm = false;
                                newTagName = "";
                                newTagColor = "#3b82f6";
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
                            disabled={saving || !newTagName.trim()}
                        >
                            {#if saving}
                                <div
                                    class="animate-spin rounded-full h-4 w-4 border-b-2 border-primary-foreground mr-2"
                                ></div>
                            {:else}
                                <Save class="h-4 w-4 mr-2" />
                            {/if}
                            Create Tag
                        </Button>
                    </div>
                </div>
            {/if}

            <!-- Tags Table -->
            {#if existingTags.length === 0}
                <div
                    class="text-center py-12 text-muted-foreground border border-dashed border-border rounded-lg"
                >
                    <TagIcon class="h-12 w-12 mx-auto mb-3 opacity-50" />
                    <p class="text-lg font-medium mb-1">No tags yet</p>
                    <p class="text-sm">
                        Create your first tag or add some events with tags to
                        get started
                    </p>
                </div>
            {:else}
                <div class="border border-border rounded-lg overflow-hidden">
                    <table class="w-full">
                        <thead class="bg-muted/50 border-b border-border">
                            <tr>
                                <th
                                    class="text-left py-2 px-3 text-sm font-medium text-muted-foreground"
                                    >Color</th
                                >
                                <th
                                    class="text-left py-2 px-3 text-sm font-medium text-muted-foreground"
                                    >Tag Name</th
                                >
                                <th
                                    class="text-left py-2 px-3 text-sm font-medium text-muted-foreground"
                                    >Usage</th
                                >
                                <th
                                    class="text-left py-2 px-3 text-sm font-medium text-muted-foreground"
                                    >Type</th
                                >
                                <th
                                    class="text-right py-2 px-3 text-sm font-medium text-muted-foreground"
                                    >Actions</th
                                >
                            </tr>
                        </thead>
                        <tbody>
                            {#each existingTags as tag (tag)}
                                <tr
                                    class="border-b border-border last:border-b-0 hover:bg-muted/25 transition-colors"
                                >
                                    {#if editingTag === tag}
                                        <!-- Edit Mode -->
                                        <td class="py-2 px-3">
                                            <div
                                                class="flex items-center gap-2"
                                            >
                                                <input
                                                    type="color"
                                                    bind:value={editTagColor}
                                                    class="w-8 h-8 rounded border border-border cursor-pointer flex-shrink-0"
                                                    disabled={saving}
                                                />
                                                <Input
                                                    bind:value={editTagColor}
                                                    placeholder="#3b82f6"
                                                    class="w-24 text-sm h-8"
                                                    disabled={saving}
                                                />
                                            </div>
                                        </td>
                                        <td class="py-2 px-3">
                                            <Input
                                                bind:value={editTagName}
                                                placeholder="Tag name"
                                                class="max-w-xs h-8"
                                                disabled={saving}
                                            />
                                        </td>
                                        <td
                                            class="py-2 px-3 text-sm text-muted-foreground"
                                        >
                                            {tagUsageCount[tag] || 0} event{(tagUsageCount[
                                                tag
                                            ] || 0) !== 1
                                                ? "s"
                                                : ""}
                                        </td>
                                        <td class="py-2 px-3">
                                            {#if tag.toLowerCase() === "feedback"}
                                                <span
                                                    class="text-xs text-muted-foreground"
                                                    >System tag</span
                                                >
                                            {:else}
                                                <span
                                                    class="text-xs text-muted-foreground"
                                                    >Custom</span
                                                >
                                            {/if}
                                        </td>
                                        <td class="py-2 px-3 text-right">
                                            <div
                                                class="flex items-center justify-end gap-1"
                                            >
                                                <Button
                                                    variant="ghost"
                                                    size="sm"
                                                    on:click={cancelEdit}
                                                    disabled={saving}
                                                >
                                                    <X class="h-4 w-4" />
                                                </Button>
                                                <Button
                                                    size="sm"
                                                    on:click={saveEdit}
                                                    disabled={saving ||
                                                        !editTagName.trim()}
                                                >
                                                    {#if saving}
                                                        <div
                                                            class="animate-spin rounded-full h-4 w-4 border-b-2 border-primary-foreground"
                                                        ></div>
                                                    {:else}
                                                        <Save class="h-4 w-4" />
                                                    {/if}
                                                </Button>
                                            </div>
                                        </td>
                                    {:else}
                                        <!-- View Mode -->
                                        <td class="py-2 px-3">
                                            <div
                                                class="flex items-center gap-2"
                                            >
                                                <div
                                                    class="w-4 h-4 rounded"
                                                    style="background-color: {tagColorStore.getColor(
                                                        tag,
                                                    )}"
                                                ></div>
                                            </div>
                                        </td>
                                        <td class="py-2 px-3">
                                            <Badge
                                                variant="outline"
                                                style={tagColorStore.getColor(
                                                    tag,
                                                )
                                                    ? `border-color: ${tagColorStore.getColor(tag)}; background-color: ${tagColorStore.getColor(tag)}20; color: ${tagColorStore.getColor(tag)};`
                                                    : ""}
                                            >
                                                {tag}
                                                {#if tag.toLowerCase() === "feedback"}
                                                    <span
                                                        class="ml-1 text-xs opacity-60"
                                                        >🔒</span
                                                    >
                                                {/if}
                                            </Badge>
                                        </td>
                                        <td
                                            class="py-2 px-3 text-sm text-muted-foreground"
                                        >
                                            {tagUsageCount[tag] || 0} event{(tagUsageCount[
                                                tag
                                            ] || 0) !== 1
                                                ? "s"
                                                : ""}
                                        </td>
                                        <td class="py-2 px-3">
                                            {#if tag.toLowerCase() === "feedback"}
                                                <Badge
                                                    variant="secondary"
                                                    class="text-xs"
                                                    >System</Badge
                                                >
                                            {:else}
                                                <span
                                                    class="text-xs text-muted-foreground"
                                                    >Custom</span
                                                >
                                            {/if}
                                        </td>
                                        <td class="py-2 px-3 text-right">
                                            <div
                                                class="flex items-center justify-end gap-1"
                                            >
                                                <Button
                                                    variant="ghost"
                                                    size="sm"
                                                    on:click={() =>
                                                        startEdit(tag)}
                                                    disabled={saving ||
                                                        tag.toLowerCase() ===
                                                            "feedback"}
                                                    class={tag.toLowerCase() ===
                                                    "feedback"
                                                        ? "opacity-50"
                                                        : ""}
                                                >
                                                    <Edit2 class="h-4 w-4" />
                                                </Button>
                                                <Button
                                                    variant="ghost"
                                                    size="sm"
                                                    class="text-destructive hover:text-destructive {tag.toLowerCase() ===
                                                    'feedback'
                                                        ? 'opacity-50'
                                                        : ''}"
                                                    on:click={() =>
                                                        deleteTag(tag)}
                                                    disabled={saving ||
                                                        tag.toLowerCase() ===
                                                            "feedback"}
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
