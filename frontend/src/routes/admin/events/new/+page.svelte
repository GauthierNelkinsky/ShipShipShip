<script lang="ts">
    import { onMount } from "svelte";
    import { goto } from "$app/navigation";
    import { api } from "$lib/api";
    import type { CreateEventRequest, EventStatus } from "$lib/types";
    import { ArrowLeft, Save, Plus, X } from "lucide-svelte";
    import { Button, Card, Input, Textarea, Badge } from "$lib/components/ui";

    let loading = false;
    let error = "";
    let success = false;

    // Form data
    let title = "";
    let status: EventStatus = "Backlogs";
    let date = "";
    let content = "";
    let tags: string[] = [];
    let media: string[] = [];
    let newTag = "";
    let newMedia = "";

    const statusOptions: { value: EventStatus; label: string }[] = [
        { value: "Backlogs", label: "Backlogs" },
        { value: "Upcoming", label: "Upcoming" },
        { value: "Doing", label: "Doing" },
        { value: "Release", label: "Release" },
        { value: "Archived", label: "Archived" },
    ];

    onMount(async () => {
        // Check authentication
        if (!api.isAuthenticated()) {
            goto("/admin");
            return;
        }

        try {
            await api.validateToken();
        } catch (err) {
            api.clearToken();
            goto("/admin");
        }
    });

    async function handleSubmit() {
        if (!title.trim()) {
            error = "Title is required";
            return;
        }

        loading = true;
        error = "";
        success = false;

        try {
            const eventData: CreateEventRequest = {
                title: title.trim(),
                tags,
                media,
                status,
                date: date.trim(),
                content: content.trim(),
            };

            await api.createEvent(eventData);
            success = true;

            // Redirect after a short delay
            setTimeout(() => {
                goto("/admin/events");
            }, 1500);
        } catch (err) {
            error =
                err instanceof Error ? err.message : "Failed to create event";
        } finally {
            loading = false;
        }
    }

    function addTag() {
        if (newTag.trim() && !tags.includes(newTag.trim())) {
            tags = [...tags, newTag.trim()];
            newTag = "";
        }
    }

    function removeTag(index: number) {
        tags = tags.filter((_, i) => i !== index);
    }

    function addMedia() {
        if (newMedia.trim() && !media.includes(newMedia.trim())) {
            media = [...media, newMedia.trim()];
            newMedia = "";
        }
    }

    function removeMedia(index: number) {
        media = media.filter((_, i) => i !== index);
    }

    function handleTagKeyDown(event: KeyboardEvent) {
        if (event.key === "Enter") {
            event.preventDefault();
            addTag();
        }
    }

    function handleMediaKeyDown(event: KeyboardEvent) {
        if (event.key === "Enter") {
            event.preventDefault();
            addMedia();
        }
    }
</script>

<svelte:head>
    <title>New Event - Admin</title>
</svelte:head>

<div class="min-h-screen bg-background">
    <!-- Header -->
    <header class="border-b border-border/40">
        <div class="container max-w-4xl mx-auto px-4">
            <div class="flex h-14 items-center justify-between">
                <div class="flex items-center space-x-2">
                    <a
                        href="/admin/events"
                        class="inline-flex items-center justify-center whitespace-nowrap rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 hover:bg-accent hover:text-accent-foreground h-10 px-4 py-2 gap-2"
                    >
                        <ArrowLeft class="h-4 w-4" />
                        Events
                    </a>
                    <span class="text-muted-foreground">/</span>
                    <span class="font-medium">New Event</span>
                </div>
            </div>
        </div>
    </header>

    <main class="container max-w-4xl mx-auto px-4 py-8">
        <div class="mb-8">
            <h1 class="text-3xl font-bold mb-2">Create New Event</h1>
            <p class="text-muted-foreground">
                Add a new event to your changelog
            </p>
        </div>

        {#if success}
            <div
                class="bg-green-50 border border-green-200 text-green-800 px-4 py-3 rounded-md mb-6 dark:bg-green-900/20 dark:border-green-800 dark:text-green-200"
            >
                Event created successfully! Redirecting...
            </div>
        {/if}

        {#if error}
            <div
                class="bg-destructive/10 border border-destructive/20 text-destructive px-4 py-3 rounded-md mb-6"
            >
                {error}
            </div>
        {/if}

        <form on:submit|preventDefault={handleSubmit} class="space-y-8">
            <!-- Basic Information -->
            <Card class="p-6">
                <h2 class="text-lg font-semibold mb-4">Basic Information</h2>
                <div class="grid gap-6">
                    <div>
                        <label for="title" class="label block mb-2"
                            >Title *</label
                        >
                        <Input
                            id="title"
                            type="text"
                            bind:value={title}
                            placeholder="Enter event title"
                            disabled={loading}
                            required
                        />
                    </div>

                    <div class="grid gap-6 md:grid-cols-2">
                        <div>
                            <label for="status" class="label block mb-2"
                                >Status *</label
                            >
                            <select
                                id="status"
                                bind:value={status}
                                class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                                disabled={loading}
                                required
                            >
                                {#each statusOptions as option}
                                    <option value={option.value}
                                        >{option.label}</option
                                    >
                                {/each}
                            </select>
                        </div>

                        <div>
                            <label for="date" class="label block mb-2"
                                >Date</label
                            >
                            <Input
                                id="date"
                                type="date"
                                bind:value={date}
                                placeholder="YYYY-MM-DD"
                                disabled={loading}
                            />
                        </div>
                    </div>
                </div>
            </Card>

            <!-- Tags -->
            <Card class="p-6">
                <h2 class="text-lg font-semibold mb-4">Tags</h2>
                <div class="space-y-4">
                    <div class="flex gap-2">
                        <Input
                            type="text"
                            bind:value={newTag}
                            on:keydown={handleTagKeyDown}
                            placeholder="Add a tag"
                            class="flex-1"
                            disabled={loading}
                        />
                        <Button
                            variant="outline"
                            type="button"
                            on:click={addTag}
                            class="flex items-center gap-2"
                            disabled={loading || !newTag.trim()}
                        >
                            <Plus class="h-4 w-4" />
                            Add
                        </Button>
                    </div>

                    {#if tags.length > 0}
                        <div class="flex flex-wrap gap-2">
                            {#each tags as tag, index (tag)}
                                <Badge
                                    variant="secondary"
                                    class="inline-flex items-center gap-1"
                                >
                                    {tag}
                                    <Button
                                        variant="ghost"
                                        size="icon"
                                        type="button"
                                        on:click={() => removeTag(index)}
                                        class="h-4 w-4 hover:text-destructive p-0"
                                        disabled={loading}
                                    >
                                        <X class="h-3 w-3" />
                                    </Button>
                                </Badge>
                            {/each}
                        </div>
                    {/if}
                </div>
            </Card>

            <!-- Media -->
            <Card class="p-6">
                <h2 class="text-lg font-semibold mb-4">Media</h2>
                <div class="space-y-4">
                    <div class="flex gap-2">
                        <Input
                            type="url"
                            bind:value={newMedia}
                            on:keydown={handleMediaKeyDown}
                            placeholder="Add media URL (image, video, etc.)"
                            class="flex-1"
                            disabled={loading}
                        />
                        <Button
                            variant="outline"
                            type="button"
                            on:click={addMedia}
                            class="flex items-center gap-2"
                            disabled={loading || !newMedia.trim()}
                        >
                            <Plus class="h-4 w-4" />
                            Add
                        </Button>
                    </div>

                    {#if media.length > 0}
                        <div class="space-y-2">
                            {#each media as mediaUrl, index (mediaUrl)}
                                <div
                                    class="flex items-center gap-2 p-2 bg-muted rounded-md"
                                >
                                    <span class="flex-1 text-sm truncate"
                                        >{mediaUrl}</span
                                    >
                                    <Button
                                        variant="ghost"
                                        size="icon"
                                        type="button"
                                        on:click={() => removeMedia(index)}
                                        class="h-6 w-6 text-destructive hover:text-destructive/80"
                                        disabled={loading}
                                    >
                                        <X class="h-4 w-4" />
                                    </Button>
                                </div>
                            {/each}
                        </div>
                    {/if}
                </div>
            </Card>

            <!-- Content -->
            <Card class="p-6">
                <h2 class="text-lg font-semibold mb-4">Content</h2>
                <div>
                    <label for="content" class="label block mb-2"
                        >Markdown Content</label
                    >
                    <Textarea
                        id="content"
                        bind:value={content}
                        placeholder="Write your content in Markdown format..."
                        rows={10}
                        disabled={loading}
                    />
                    <p class="text-xs text-muted-foreground mt-1">
                        Supports Markdown formatting: **bold**, *italic*,
                        `code`, etc.
                    </p>
                </div>
            </Card>

            <!-- Actions -->
            <div class="flex gap-4">
                <Button
                    type="submit"
                    class="flex items-center gap-2"
                    disabled={loading || !title.trim()}
                >
                    {#if loading}
                        <div
                            class="animate-spin rounded-full h-4 w-4 border-b-2 border-white"
                        ></div>
                        Creating...
                    {:else}
                        <Save class="h-4 w-4" />
                        Create Event
                    {/if}
                </Button>

                <a
                    href="/admin/events"
                    class="inline-flex items-center justify-center whitespace-nowrap rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 border border-input bg-background hover:bg-accent hover:text-accent-foreground h-10 px-4 py-2"
                >
                    Cancel
                </a>
            </div>
        </form>
    </main>
</div>
