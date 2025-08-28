<script lang="ts">
    import { createEventDispatcher } from "svelte";
    import { Button } from "$lib/components/ui";
    import { X } from "lucide-svelte";

    const dispatch = createEventDispatcher();

    export let open = false;
    export let title = "";
    export let description = "";
    export let cancelText = "Cancel";
    export let actionText = "Continue";
    export let actionVariant:
        | "default"
        | "destructive"
        | "outline"
        | "secondary"
        | "ghost"
        | "link" = "default";
    export let loading = false;

    function close() {
        open = false;
        dispatch("close");
    }

    function cancel() {
        dispatch("cancel");
        close();
    }

    function action() {
        dispatch("action");
    }

    function handleKeydown(event: KeyboardEvent) {
        if (event.key === "Escape" && !loading) {
            cancel();
        }
    }

    function handleBackdropClick(event: MouseEvent) {
        if (event.target === event.currentTarget && !loading) {
            cancel();
        }
    }
</script>

<svelte:window on:keydown={handleKeydown} />

{#if open}
    <!-- svelte-ignore a11y-click-events-have-key-events -->
    <!-- svelte-ignore a11y-no-static-element-interactions -->
    <!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
    <div
        class="fixed inset-0 z-50 bg-black/80 flex items-center justify-center p-4"
        on:click={handleBackdropClick}
        role="dialog"
        aria-modal="true"
        aria-labelledby="alert-dialog-title"
        aria-describedby="alert-dialog-description"
    >
        <!-- svelte-ignore a11y-click-events-have-key-events -->
        <!-- svelte-ignore a11y-no-static-element-interactions -->
        <!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
        <div
            class="bg-background border border-border rounded-lg shadow-lg w-full max-w-lg"
            on:click={(e) => e.stopPropagation()}
            role="alertdialog"
        >
            <!-- Header -->
            <div class="flex items-start justify-between p-6">
                <div class="flex-1">
                    {#if title}
                        <h2
                            id="alert-dialog-title"
                            class="text-lg font-semibold leading-none tracking-tight"
                        >
                            {title}
                        </h2>
                    {/if}
                    {#if description}
                        <p
                            id="alert-dialog-description"
                            class="text-sm text-muted-foreground mt-2"
                        >
                            {description}
                        </p>
                    {/if}
                </div>
                {#if !loading}
                    <Button
                        variant="ghost"
                        size="icon"
                        on:click={cancel}
                        class="h-6 w-6 p-0"
                    >
                        <X class="h-4 w-4" />
                        <span class="sr-only">Close</span>
                    </Button>
                {/if}
            </div>

            <!-- Content slot for custom content -->
            <div class="px-6">
                <slot />
            </div>

            <!-- Footer -->
            <div
                class="flex flex-col-reverse sm:flex-row sm:justify-end sm:space-x-2 p-6 pt-4"
            >
                <Button
                    variant="outline"
                    on:click={cancel}
                    disabled={loading}
                    class="mt-2 sm:mt-0"
                >
                    {cancelText}
                </Button>
                <Button
                    variant={actionVariant}
                    on:click={action}
                    disabled={loading}
                    class="min-w-20"
                >
                    {#if loading}
                        <div
                            class="h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent mr-2"
                        ></div>
                    {/if}
                    {actionText}
                </Button>
            </div>
        </div>
    </div>
{/if}

<style>
    @keyframes fadeIn {
        from {
            opacity: 0;
        }
        to {
            opacity: 1;
        }
    }

    @keyframes scaleIn {
        from {
            opacity: 0;
            transform: scale(0.95);
        }
        to {
            opacity: 1;
            transform: scale(1);
        }
    }

    [role="dialog"] {
        animation: fadeIn 0.2s ease-out;
    }

    [role="dialog"] > div {
        animation: scaleIn 0.2s ease-out;
    }
</style>
