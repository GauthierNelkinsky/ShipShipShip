<script lang="ts">
    import { createEventDispatcher } from "svelte";
    import { api } from "$lib/api";
    import { Button } from "$lib/components/ui";
    import * as m from "$lib/paraglide/messages";
    import { X, Upload, Image as ImageIcon } from "lucide-svelte";

    const dispatch = createEventDispatcher();

    export let isOpen = false;

    let selectedFile: File | null = null;
    let uploading = false;
    let error = "";
    let dragOver = false;

    // File validation
    const MAX_FILE_SIZE = 10 * 1024 * 1024; // 10MB
    const ALLOWED_TYPES = [
        "image/jpeg",
        "image/jpg",
        "image/png",
        "image/gif",
        "image/webp",
    ];

    function closeModal() {
        isOpen = false;
        resetForm();
        dispatch("close");
    }

    function resetForm() {
        selectedFile = null;
        uploading = false;
        error = "";
        dragOver = false;
    }

    function handleFileSelect(event: Event) {
        const target = event.target as HTMLInputElement;
        if (target.files && target.files[0]) {
            validateAndSetFile(target.files[0]);
        }
    }

    function validateAndSetFile(file: File) {
        error = "";

        // Check file type
        if (!ALLOWED_TYPES.includes(file.type)) {
            error = m.image_upload_modal_invalid_file_type();
            return;
        }

        // Check file size
        if (file.size > MAX_FILE_SIZE) {
            error = m.image_upload_modal_file_size_error();
            return;
        }

        selectedFile = file;
    }

    function handleDrop(event: DragEvent) {
        event.preventDefault();
        dragOver = false;

        const files = event.dataTransfer?.files;
        if (files && files[0]) {
            validateAndSetFile(files[0]);
        }
    }

    function handleDragOver(event: DragEvent) {
        event.preventDefault();
        dragOver = true;
    }

    function handleDragLeave(event: DragEvent) {
        event.preventDefault();
        dragOver = false;
    }

    async function handleUpload() {
        if (!selectedFile) return;

        uploading = true;
        error = "";

        try {
            const result = await api.uploadImage(selectedFile);
            dispatch("imageSelected", { url: result.url, type: "upload" });
            closeModal();
        } catch (err) {
            error =
                err instanceof Error
                    ? err.message
                    : m.image_upload_modal_upload_failed();
        } finally {
            uploading = false;
        }
    }

    function handleKeydown(event: KeyboardEvent) {
        if (event.key === "Escape") {
            closeModal();
        }
    }

    function handleBackdropClick(event: MouseEvent) {
        if (event.target === event.currentTarget) {
            closeModal();
        }
    }

    function formatFileSize(bytes: number): string {
        if (bytes === 0) return "0 Bytes";
        const k = 1024;
        const sizes = ["Bytes", "KB", "MB"];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
    }
</script>

<svelte:window on:keydown={handleKeydown} />

{#if isOpen}
    <div
        class="fixed inset-0 z-50 bg-black bg-opacity-50 flex items-center justify-center p-4"
        on:click={handleBackdropClick}
        on:keydown={(e) => {
            if (e.key === "Escape") close();
        }}
        role="dialog"
        aria-modal="true"
        tabindex="0"
    >
        <div
            class="bg-background border border-border rounded-lg shadow-lg w-full max-w-md"
            on:click={(e) => e.stopPropagation()}
            role="none"
        >
            <!-- Header -->
            <div class="flex items-center justify-between p-6 pb-4">
                <h2 class="text-lg font-semibold">
                    {m.image_upload_modal_add_image()}
                </h2>
                <Button variant="ghost" size="icon" on:click={closeModal}>
                    <X class="h-4 w-4" />
                </Button>
            </div>

            <!-- Content -->
            <div class="px-6">
                <!-- File Upload Section -->
                <div class="space-y-4">
                    <!-- Drop Zone -->
                    <div
                        class="border-2 border-dashed rounded-lg p-6 text-center transition-colors {dragOver
                            ? 'border-primary bg-primary/5'
                            : 'border-border'} {selectedFile
                            ? 'border-green-500 bg-green-50 dark:bg-green-950'
                            : ''}"
                        on:drop={handleDrop}
                        on:dragover={handleDragOver}
                        on:dragleave={handleDragLeave}
                        role="button"
                        tabindex="0"
                    >
                        {#if selectedFile}
                            <div class="space-y-2">
                                <ImageIcon
                                    class="h-8 w-8 mx-auto text-green-600"
                                />
                                <p class="text-sm font-medium text-foreground">
                                    {selectedFile.name}
                                </p>
                                <p class="text-xs text-muted-foreground">
                                    {formatFileSize(selectedFile.size)}
                                </p>
                                <Button
                                    variant="outline"
                                    size="sm"
                                    on:click={() => (selectedFile = null)}
                                >
                                    {m.image_upload_modal_choose_different_file()}
                                </Button>
                            </div>
                        {:else}
                            <div class="space-y-2">
                                <Upload
                                    class="h-8 w-8 mx-auto text-muted-foreground"
                                />
                                <p class="text-sm font-medium text-foreground">
                                    {m.image_upload_modal_drop_image()}
                                    <label
                                        class="text-primary cursor-pointer hover:underline"
                                    >
                                        {m.image_upload_modal_browse()}
                                        <input
                                            type="file"
                                            accept="image/*"
                                            on:change={handleFileSelect}
                                            class="hidden"
                                        />
                                    </label>
                                </p>
                                <p class="text-xs text-muted-foreground">
                                    {m.image_upload_modal_supported_formats()}
                                </p>
                            </div>
                        {/if}
                    </div>
                </div>

                <!-- Error Message -->
                {#if error}
                    <div
                        class="mt-4 p-3 rounded-lg bg-destructive/10 border border-destructive/20"
                    >
                        <p class="text-sm text-destructive">{error}</p>
                    </div>
                {/if}
            </div>

            <!-- Footer -->
            <div class="flex justify-end gap-3 p-6 pt-4">
                <Button
                    variant="outline"
                    on:click={closeModal}
                    disabled={uploading}
                >
                    {m.image_upload_modal_cancel()}
                </Button>

                <Button
                    on:click={handleUpload}
                    disabled={!selectedFile || uploading}
                    class="min-w-20"
                >
                    {#if uploading}
                        <div
                            class="h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent"
                        ></div>
                    {:else}
                        {m.image_upload_modal_upload()}
                    {/if}
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
</style>
