<script lang="ts">
    import { createEventDispatcher, onMount } from "svelte";
    import { api } from "$lib/api";
    import type { ParsedEvent } from "$lib/types";
    import {
        X,
        Share2,
        Send,
        Eye,
        Code2,
        Globe,
        Mail,
        CheckCircle,
        ChevronDown,
        Plus,
        Copy,
        Link,
    } from "lucide-svelte";
    import {
        Button,
        Card,
        Input,
        Switch,
        Textarea,
        Collapsible,
    } from "$lib/components/ui";

    const dispatch = createEventDispatcher();

    export let isOpen = false;
    export let event: ParsedEvent | null = null;
    export let newsletterEnabled = true;

    let loading = false;
    let error = "";

    // Publish state
    let isPublic = true;
    let hasPublicUrl = true;
    let emailSent = false;
    let emailSubject = "";
    let emailContent = "";
    let showEmailPreview = true;
    let publishLoading = false;
    let publishError = "";
    let urlCopied = false;
    let emailHistory: Array<{
        id: number;
        event_id: number;
        event_status: string;
        email_subject: string;
        email_template: string;
        subscriber_count: number;
        sent_at: string;
        created_at: string;
    }> = [];
    let historyLoading = false;

    // Load publish status when modal opens
    $: if (isOpen && event?.id) {
        loadPublishStatus();
    }

    async function loadPublishStatus() {
        if (!event?.id) return;

        try {
            loading = true;
            const response = await api.getEventPublishStatus(event.id);
            isPublic = response.is_public;
            hasPublicUrl = response.has_public_url;
            emailSent = response.email_sent;
            emailSubject = response.email_subject || "";

            // Always load newsletter preview to show content
            await loadNewsletterPreview();
            // Load email history only if newsletter is enabled
            if (newsletterEnabled) {
                await loadEmailHistory();
            }
        } catch (err) {
            error = "Failed to load publish status";
            console.error("Failed to load publish status:", err);
        } finally {
            loading = false;
        }
    }

    async function loadEmailHistory() {
        if (!event?.id || !newsletterEnabled) return;

        try {
            historyLoading = true;
            const response = await api.getEventEmailHistory(event.id);
            emailHistory = response.history || [];
        } catch (err) {
            console.error("Failed to load email history:", err);
            emailHistory = [];
        } finally {
            historyLoading = false;
        }
    }

    async function updatePublicStatus() {
        if (!event?.id) return;

        try {
            await api.updateEventPublicStatus(event.id, {
                is_public: isPublic,
            });
        } catch (err) {
            publishError = "Failed to update public status";
            console.error(err);
        }
    }

    async function updatePublicUrlStatus() {
        if (!event?.id) return;

        try {
            await api.updateEventPublicStatus(event.id, {
                has_public_url: hasPublicUrl,
            });
        } catch (err) {
            publishError = "Failed to update public URL status";
            console.error(err);
        }
    }

    async function loadNewsletterPreview() {
        if (!event?.id) return;

        try {
            const template =
                event.status === "Release"
                    ? "new_release"
                    : event.status === "Proposed"
                      ? "proposed_feature"
                      : "upcoming_feature";
            const response = await api.getEventNewsletterPreview(
                event.id,
                template,
            );
            if (!emailSubject) emailSubject = response.subject;
            emailContent = response.content;
            console.log("Newsletter preview loaded:", {
                emailSubject,
                emailContent: emailContent.length + " chars",
            });
        } catch (err) {
            publishError = "Failed to load newsletter preview";
            console.error("Failed to load newsletter preview:", err);
        }
    }

    async function sendNewsletter() {
        if (!event?.id || !emailSubject || !emailContent) return;

        try {
            publishLoading = true;
            publishError = "";
            const template =
                event.status === "Release"
                    ? "new_release"
                    : event.status === "Proposed"
                      ? "proposed_feature"
                      : "upcoming_feature";

            await api.sendEventNewsletter(event.id, {
                subject: emailSubject,
                content: emailContent,
                template: template,
            });

            // Reload the publish status and history to get updated information
            await loadPublishStatus();
            await loadEmailHistory();
            alert("Newsletter sent successfully!");
        } catch (err) {
            publishError =
                err instanceof Error
                    ? err.message
                    : "Failed to send newsletter";
        } finally {
            publishLoading = false;
        }
    }

    function closeModal() {
        isOpen = false;
        error = "";
        publishError = "";
        dispatch("close");
    }

    function handleKeydown(e: KeyboardEvent) {
        if (e.key === "Escape") {
            closeModal();
        }
    }

    function handleOutsideClick(e: MouseEvent) {
        if (e.target === e.currentTarget) {
            closeModal();
        }
    }

    async function copyPublicUrl() {
        if (!event) return;

        const publicUrl = `${window.location.origin}/${event.slug}`;
        try {
            await navigator.clipboard.writeText(publicUrl);
            urlCopied = true;
            setTimeout(() => {
                urlCopied = false;
            }, 2000);
        } catch (err) {
            console.error("Failed to copy URL:", err);
        }
    }
</script>

<svelte:window on:keydown={handleKeydown} />

{#if isOpen && event}
    <!-- Modal backdrop -->
    <div
        class="fixed inset-0 bg-black/50 z-50 flex items-center justify-center p-4"
        on:click={handleOutsideClick}
        role="dialog"
        aria-modal="true"
    >
        <!-- Modal content -->
        <div
            class="bg-background border border-border rounded-lg shadow-lg w-full max-w-6xl h-[90vh] flex flex-col"
            on:click={(e) => e.stopPropagation()}
        >
            <!-- Modal header -->
            <div
                class="flex items-center justify-between p-6 border-b border-border shrink-0"
            >
                <div class="flex items-center gap-3">
                    <Share2 class="h-5 w-5 text-primary" />
                    <div>
                        <h2 class="text-lg font-semibold">Share Event</h2>
                        <p class="text-sm text-muted-foreground">
                            {event.title}
                        </p>
                    </div>
                </div>
                <Button
                    variant="ghost"
                    size="icon"
                    on:click={closeModal}
                    class="text-muted-foreground hover:text-foreground"
                >
                    <X class="h-4 w-4" />
                </Button>
            </div>

            <!-- Modal body -->
            <div class="flex-1 overflow-y-auto p-8">
                {#if loading}
                    <div class="flex items-center justify-center min-h-32">
                        <div
                            class="h-8 w-8 animate-spin rounded-full border-2 border-primary border-t-transparent"
                        ></div>
                    </div>
                {:else}
                    {#if error}
                        <Card
                            class="p-4 mb-4 bg-destructive/10 border-destructive"
                        >
                            <p class="text-destructive text-sm">{error}</p>
                        </Card>
                    {/if}

                    {#if publishError}
                        <Card
                            class="p-4 mb-4 bg-destructive/10 border-destructive"
                        >
                            <p class="text-destructive text-sm">
                                {publishError}
                            </p>
                        </Card>
                    {/if}

                    <div class="space-y-4">
                        <!-- Public Page Channel -->
                        <Card class="p-3">
                            <div class="flex items-center justify-between">
                                <div class="flex items-center gap-3">
                                    <Globe class="h-5 w-5 text-primary" />
                                    <div>
                                        <h3 class="font-medium">
                                            Public Website
                                        </h3>
                                        <p
                                            class="text-xs text-muted-foreground"
                                        >
                                            Make this event visible on your
                                            public changelog page
                                        </p>
                                    </div>
                                </div>
                                <Switch
                                    bind:checked={isPublic}
                                    on:change={updatePublicStatus}
                                />
                            </div>
                        </Card>

                        <!-- Public URL Switch -->
                        <Card class="p-3">
                            <div class="flex items-center justify-between">
                                <div class="flex items-center gap-3">
                                    <Link class="h-5 w-5 text-primary" />
                                    <div>
                                        <h3 class="font-medium">Public URL</h3>
                                        {#if hasPublicUrl}
                                            <div
                                                class="flex items-center gap-2 mt-1"
                                            >
                                                <button
                                                    on:click={copyPublicUrl}
                                                    class="text-xs text-muted-foreground hover:text-foreground transition-colors flex items-center gap-1 group"
                                                >
                                                    <span class="font-mono"
                                                        >{window?.location
                                                            ?.origin ||
                                                            ""}/{event?.slug ||
                                                            ""}</span
                                                    >
                                                    {#if urlCopied}
                                                        <CheckCircle
                                                            class="h-3 w-3 text-green-500"
                                                        />
                                                    {:else}
                                                        <Copy
                                                            class="h-3 w-3 opacity-60 group-hover:opacity-100"
                                                        />
                                                    {/if}
                                                </button>
                                            </div>
                                        {/if}
                                    </div>
                                </div>
                                <Switch
                                    bind:checked={hasPublicUrl}
                                    on:change={updatePublicUrlStatus}
                                />
                            </div>
                        </Card>

                        <!-- Newsletter Channel -->
                        <Card class="p-3">
                            <div class="flex items-center gap-3 mb-3">
                                <Mail class="h-5 w-5 text-primary" />
                                <div>
                                    <h3 class="font-medium">
                                        Email Newsletter
                                    </h3>
                                    <p class="text-xs text-muted-foreground">
                                        Send to all newsletter subscribers
                                    </p>
                                </div>
                            </div>

                            <!-- Email History Table -->
                            <div class="space-y-3">
                                {#if historyLoading}
                                    <div
                                        class="flex items-center justify-center py-4"
                                    >
                                        <div
                                            class="animate-spin rounded-full h-4 w-4 border-b-2 border-primary"
                                        ></div>
                                    </div>
                                {:else if emailHistory.length === 0}
                                    <div
                                        class="text-center py-4 text-muted-foreground"
                                    >
                                        <Mail
                                            class="h-8 w-8 mx-auto mb-2 opacity-50"
                                        />
                                        <p class="text-xs">
                                            No newsletters sent for this event
                                        </p>
                                    </div>
                                {:else}
                                    <div class="max-h-32 overflow-y-auto">
                                        <div
                                            class="bg-muted/30 px-4 py-3 border-b border-border"
                                        >
                                            <div
                                                class="grid grid-cols-3 gap-4 text-xs font-medium text-muted-foreground uppercase tracking-wide"
                                            >
                                                <div>Date</div>
                                                <div>Subject</div>
                                                <div class="text-right">
                                                    Subscribers
                                                </div>
                                            </div>
                                        </div>
                                        <div class="divide-y divide-border">
                                            {#each emailHistory as entry (entry.id)}
                                                <div
                                                    class="px-4 py-3 hover:bg-muted/50"
                                                >
                                                    <div
                                                        class="grid grid-cols-3 gap-4 text-xs"
                                                    >
                                                        <div class="text-sm">
                                                            {new Date(
                                                                entry.sent_at,
                                                            ).toLocaleDateString(
                                                                "en-US",
                                                                {
                                                                    month: "short",
                                                                    day: "numeric",
                                                                    year: "numeric",
                                                                },
                                                            )}
                                                        </div>
                                                        <div
                                                            class="text-sm truncate"
                                                            title={entry.email_subject}
                                                        >
                                                            {entry.email_subject}
                                                        </div>
                                                        <div
                                                            class="text-sm text-right font-medium"
                                                        >
                                                            {entry.subscriber_count}
                                                        </div>
                                                    </div>
                                                </div>
                                            {/each}
                                        </div>
                                    </div>
                                {/if}
                            </div>

                            <!-- Send Newsletter Section -->
                            <div class="mt-4">
                                <Collapsible>
                                    <div slot="trigger" let:toggle let:open>
                                        <Button
                                            variant="outline"
                                            size="sm"
                                            class="w-full justify-between"
                                            on:click={toggle}
                                        >
                                            <div
                                                class="flex items-center gap-2"
                                            >
                                                <Plus class="h-4 w-4" />
                                                Send Newsletter
                                            </div>
                                            <div
                                                class="transition-transform duration-200"
                                                class:rotate-180={open}
                                            >
                                                <ChevronDown class="h-4 w-4" />
                                            </div>
                                        </Button>
                                    </div>
                                    <div class="mt-4 space-y-4">
                                        <div class="space-y-2">
                                            <label
                                                for="email-subject"
                                                class="text-sm font-medium"
                                            >
                                                Email Subject
                                            </label>
                                            <Input
                                                id="email-subject"
                                                bind:value={emailSubject}
                                                placeholder="Enter email subject line"
                                                class="h-10"
                                            />
                                        </div>

                                        <div class="space-y-4">
                                            <div
                                                class="flex items-center justify-between"
                                            >
                                                <label
                                                    for="email-content"
                                                    class="text-sm font-medium"
                                                >
                                                    Email Content
                                                </label>
                                                <div
                                                    class="flex items-center gap-1 bg-muted p-1 rounded-lg"
                                                >
                                                    <Button
                                                        variant={showEmailPreview
                                                            ? "default"
                                                            : "ghost"}
                                                        size="sm"
                                                        on:click={() =>
                                                            (showEmailPreview = true)}
                                                        class="h-8 px-3"
                                                    >
                                                        <Eye
                                                            class="h-3 w-3 mr-1"
                                                        />
                                                        Preview
                                                    </Button>
                                                    <Button
                                                        variant={!showEmailPreview
                                                            ? "default"
                                                            : "ghost"}
                                                        size="sm"
                                                        on:click={() =>
                                                            (showEmailPreview = false)}
                                                        class="h-8 px-3"
                                                    >
                                                        <Code2
                                                            class="h-3 w-3 mr-1"
                                                        />
                                                        HTML
                                                    </Button>
                                                </div>
                                            </div>

                                            {#if showEmailPreview}
                                                <div
                                                    class="border border-border rounded-lg p-4 bg-white dark:bg-card max-h-48 overflow-y-auto email-preview-content shadow-inner"
                                                >
                                                    {@html emailContent}
                                                </div>
                                            {:else}
                                                <Textarea
                                                    id="email-content"
                                                    bind:value={emailContent}
                                                    placeholder="Email HTML content will appear here..."
                                                    rows={10}
                                                    class="font-mono text-sm min-h-[240px] resize-none"
                                                />
                                            {/if}
                                        </div>

                                        <Button
                                            on:click={sendNewsletter}
                                            disabled={publishLoading ||
                                                !emailSubject ||
                                                !emailContent}
                                            class="w-full"
                                            size="sm"
                                            variant="default"
                                        >
                                            {#if publishLoading}
                                                <div
                                                    class="h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent mr-2"
                                                ></div>
                                                Sending Newsletter...
                                            {:else}
                                                <Send class="h-4 w-4 mr-2" />
                                                Send Newsletter Now
                                            {/if}
                                        </Button>
                                    </div>
                                </Collapsible>
                            </div>
                        </Card>
                    </div>
                {/if}
            </div>

            <!-- Modal footer -->
            <div
                class="flex justify-end gap-3 p-8 pt-4 border-t border-border shrink-0"
            >
                <Button variant="outline" on:click={closeModal}>Close</Button>
            </div>
        </div>
    </div>
{/if}

<style>
    :global(.email-preview-content) {
        /* Create an iframe-like environment for email content */
        all: initial;
        display: block !important;
        background: white;
        color: #333;
        font-family: Arial, sans-serif;
        line-height: 1.6;
    }

    :global(.email-preview-content *) {
        /* Reset all inherited styles from the parent page */
        all: unset;
        display: revert;
        box-sizing: border-box;
    }

    :global(.email-preview-content body) {
        font-family: Arial, sans-serif !important;
        line-height: 1.6 !important;
        color: #333 !important;
        max-width: 600px !important;
        margin: 0 auto !important;
        padding: 20px !important;
        display: block !important;
    }

    :global(.email-preview-content h1) {
        color: #000000 !important;
        text-align: center !important;
        font-size: 2em !important;
        margin: 0.67em 0 !important;
        font-weight: bold !important;
        display: block !important;
    }

    :global(.email-preview-content h2) {
        color: #000000 !important;
        margin-top: 0 !important;
        font-size: 1.5em !important;
        margin: 0.83em 0 !important;
        font-weight: bold !important;
        display: block !important;
    }

    :global(.email-preview-content h1 *) {
        color: #000000 !important;
    }

    :global(.email-preview-content h2 *) {
        color: #000000 !important;
    }

    :global(.email-preview-content p) {
        margin: 5px 0 !important;
        display: block !important;
        line-height: 1.6 !important;
    }

    :global(.email-preview-content div) {
        display: block !important;
    }

    :global(.email-preview-content a) {
        color: #2563eb !important;
        text-decoration: none !important;
        display: inline !important;
    }

    :global(.email-preview-content h1 a),
    :global(.email-preview-content h2 a),
    :global(.email-preview-content h3 a) {
        color: #000000 !important;
    }

    :global(.email-preview-content a[style*="background"]) {
        color: white !important;
    }

    :global(.email-preview-content hr) {
        border: none !important;
        border-top: 1px solid #eee !important;
        margin: 30px 0 !important;
        display: block !important;
    }
</style>
