<script lang="ts">
    import { createEventDispatcher } from "svelte";
    import { api } from "$lib/api";
    import { X } from "lucide-svelte";
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
        "#ef4444", // red
        "#f97316", // orange
        "#22c55e", // green
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
        class="fixed inset-0 bg-black/40 z-50 flex items-center justify-center p-4"
        on:click={closeModal}
        role="presentation"
    >
        <!-- Modal content -->
        <div
            class="bg-background rounded-lg shadow-xl w-full max-w-sm"
            on:click|stopPropagation
            role="dialog"
            aria-labelledby="modal-title"
        >
            <!-- Modal header -->
            <div class="flex items-center justify-between p-5 pb-3">
                <h2 id="modal-title" class="text-lg font-semibold">
                    New Status
                </h2>
                <button
                    on:click={closeModal}
                    class="text-muted-foreground hover:text-foreground transition-colors rounded-full p-1 hover:bg-muted"
                    disabled={loading}
                    aria-label="Close"
                >
                    <X class="h-5 w-5" />
                </button>
            </div>

            <!-- Modal body -->
            <div class="px-5 pb-4 space-y-4">
                {#if error}
                    <div
                        class="p-3 text-sm bg-destructive/10 text-destructive rounded-lg border border-destructive/20"
                    >
                        {error}
                    </div>
                {/if}

                <!-- Name input -->
                <div class="space-y-2">
                    <label for="status-name" class="text-sm font-medium block">
                        Name
                    </label>
                    <Input
                        id="status-name"
                        type="text"
                        placeholder="In Progress"
                        bind:value={name}
                        disabled={loading}
                        class="w-full"
                        autofocus
                    />
                </div>

                <!-- Color picker -->
                <div class="space-y-2">
                    <label class="text-sm font-medium block">Color</label>

                    <!-- Color palette grid -->
                    <div class="grid grid-cols-6 gap-2">
                        {#each colorPalette as paletteColor}
                            <button
                                type="button"
                                class="w-full aspect-square rounded-md transition-all hover:scale-105 relative"
                                class:ring-2={color === paletteColor}
                                class:ring-foreground={color === paletteColor}
                                class:ring-offset-1={color === paletteColor}
                                class:ring-offset-background={color ===
                                    paletteColor}
                                style="background-color: {paletteColor}"
                                on:click={() => (color = paletteColor)}
                                disabled={loading}
                                aria-label="Select color {paletteColor}"
                            >
                                {#if color === paletteColor}
                                    <div
                                        class="absolute inset-0 flex items-center justify-center"
                                    >
                                        <div
                                            class="w-1.5 h-1.5 rounded-full bg-white shadow"
                                        ></div>
                                    </div>
                                {/if}
                            </button>
                        {/each}
                    </div>

                    <!-- Custom color input -->
                    <div class="flex items-center gap-2">
                        <input
                            type="color"
                            bind:value={color}
                            disabled={loading}
                            class="w-9 h-9 rounded-md border border-border cursor-pointer"
                            aria-label="Custom color picker"
                        />
                        <Input
                            type="text"
                            bind:value={color}
                            disabled={loading}
                            placeholder="#3b82f6"
                            class="flex-1 font-mono text-xs"
                        />
                    </div>
                </div>
            </div>

            <!-- Modal footer -->
            <div
                class="flex items-center justify-end gap-2 px-5 py-3 bg-muted/20 rounded-b-lg border-t border-border"
            >
                <Button
                    variant="ghost"
                    on:click={closeModal}
                    disabled={loading}
                >
                    Cancel
                </Button>
                <Button
                    on:click={handleSubmit}
                    disabled={loading || !name.trim()}
                >
                    {#if loading}
                        <div
                            class="animate-spin rounded-full h-4 w-4 border-2 border-background border-t-transparent mr-2"
                        ></div>
                        Creating...
                    {:else}
                        Create
                    {/if}
                </Button>
            </div>
        </div>
    </div>
{/if}
