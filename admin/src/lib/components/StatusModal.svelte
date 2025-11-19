<script lang="ts">
    import { createEventDispatcher } from "svelte";
    import { api } from "$lib/api";
    import { X, Save, Tag } from "lucide-svelte";
    import { Button, Input } from "$lib/components/ui";

    const dispatch = createEventDispatcher();

    export let isOpen = false;

    let loading = false;
    let error = "";

    // Form fields
    let name = "";
    let color = "#3b82f6"; // Default blue color

    // Predefined color palette
    const colorPalette = [
        "#3b82f6", // blue
        "#8b5cf6", // purple
        "#ec4899", // pink
        "#f43f5e", // rose
        "#ef4444", // red
        "#f97316", // orange
        "#f59e0b", // amber
        "#eab308", // yellow
        "#84cc16", // lime
        "#22c55e", // green
        "#10b981", // emerald
        "#14b8a6", // teal
        "#06b6d4", // cyan
        "#0ea5e9", // sky
        "#6366f1", // indigo
        "#a855f7", // violet
    ];

    function resetForm() {
        name = "";
        color = "#3b82f6";
        error = "";
    }

    function closeModal() {
        resetForm();
        isOpen = false;
        dispatch("close");
    }

    async function handleSubmit() {
        if (!name.trim()) {
            error = "Status name is required";
            return;
        }

        try {
            loading = true;
            error = "";

            const statusData = {
                display_name: name.trim(),
                color: color,
            };

            const newStatus = await api.createStatus(statusData);
            dispatch("created", newStatus);
            closeModal();
        } catch (err) {
            error =
                err instanceof Error ? err.message : "Failed to create status";
        } finally {
            loading = false;
        }
    }

    function handleKeydown(e: KeyboardEvent) {
        if (e.key === "Escape") {
            closeModal();
        }
        if ((e.metaKey || e.ctrlKey) && e.key === "Enter") {
            handleSubmit();
        }
    }
</script>

<svelte:window on:keydown={handleKeydown} />

{#if isOpen}
    <!-- Modal backdrop -->
    <div
        class="fixed inset-0 bg-black/50 z-40 flex items-center justify-center p-4"
    >
        <!-- Modal content -->
        <div
            class="bg-background border border-border rounded-lg shadow-lg w-full max-w-md flex flex-col"
        >
            <!-- Modal header -->
            <div
                class="flex items-center justify-between p-6 border-b border-border"
            >
                <div class="flex items-center gap-2">
                    <Tag class="h-5 w-5" />
                    <h2 class="text-lg font-semibold">Create New Status</h2>
                </div>
                <button
                    on:click={closeModal}
                    class="text-muted-foreground hover:text-foreground"
                    disabled={loading}
                >
                    <X class="h-5 w-5" />
                </button>
            </div>

            <!-- Modal body -->
            <div class="flex-1 overflow-y-auto p-6 space-y-6">
                {#if error}
                    <div
                        class="p-3 text-sm bg-destructive/10 text-destructive rounded-md border border-destructive/20"
                    >
                        {error}
                    </div>
                {/if}

                <!-- Name input -->
                <div class="space-y-2">
                    <label for="status-name" class="text-sm font-medium">
                        Status Name <span class="text-destructive">*</span>
                    </label>
                    <Input
                        id="status-name"
                        type="text"
                        placeholder="e.g., In Progress, Done, Testing..."
                        bind:value={name}
                        disabled={loading}
                        class="w-full"
                    />
                </div>

                <!-- Color picker -->
                <div class="space-y-2">
                    <label for="status-color" class="text-sm font-medium">
                        Color
                    </label>
                    <div class="space-y-3">
                        <!-- Color palette -->
                        <div class="grid grid-cols-8 gap-2">
                            {#each colorPalette as paletteColor}
                                <button
                                    type="button"
                                    class="w-8 h-8 rounded-md border-2 transition-all hover:scale-110"
                                    class:border-foreground={color ===
                                        paletteColor}
                                    class:border-transparent={color !==
                                        paletteColor}
                                    style="background-color: {paletteColor}"
                                    on:click={() => (color = paletteColor)}
                                    disabled={loading}
                                    title={paletteColor}
                                ></button>
                            {/each}
                        </div>

                        <!-- Custom color input -->
                        <div class="flex items-center gap-2">
                            <input
                                id="status-color"
                                type="color"
                                bind:value={color}
                                disabled={loading}
                                class="w-12 h-8 rounded-md border border-border cursor-pointer"
                            />
                            <Input
                                type="text"
                                bind:value={color}
                                disabled={loading}
                                placeholder="#3b82f6"
                                class="flex-1 font-mono text-sm"
                            />
                        </div>

                        <!-- Preview -->
                        <div
                            class="flex items-center gap-2 p-3 rounded-md border"
                        >
                            <span class="text-sm text-muted-foreground"
                                >Preview:</span
                            >
                            <div
                                class="px-3 py-1 rounded-md text-xs font-medium"
                                style="background-color: {color}20; color: {color}; border: 1px solid {color}40"
                            >
                                {name || "Status Name"}
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            <!-- Modal footer -->
            <div
                class="flex items-center justify-end gap-2 p-6 border-t border-border"
            >
                <Button
                    variant="outline"
                    on:click={closeModal}
                    disabled={loading}
                >
                    Cancel
                </Button>
                <Button on:click={handleSubmit} disabled={loading}>
                    {#if loading}
                        Creating...
                    {:else}
                        <Save class="h-4 w-4 mr-2" />
                        Create Status
                    {/if}
                </Button>
            </div>
        </div>
    </div>
{/if}
