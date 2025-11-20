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

    function resetForm() {
        name = "";
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
        class="fixed inset-0 bg-black/50 z-[60] flex items-center justify-center p-4"
        on:click={handleBackdropClick}
        on:keydown={(e) => {
            if (e.key === "Escape") close();
        }}
        role="dialog"
        aria-modal="true"
        aria-labelledby="modal-title"
        tabindex="-1"
    >
        <!-- Modal content -->
        <div
            class="bg-background rounded-lg shadow-xl w-full max-w-sm"
            on:click|stopPropagation
            role="none"
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
