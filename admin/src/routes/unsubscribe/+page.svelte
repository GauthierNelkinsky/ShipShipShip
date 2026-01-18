<script lang="ts">
    import { onMount } from "svelte";
    import { api } from "$lib/api";
    import * as m from "$lib/paraglide/messages";
    import { Button, Input, Card } from "$lib/components/ui";
    import { Mail, CheckCircle, AlertCircle } from "lucide-svelte";

    let email = "";
    let loading = false;
    let success = false;
    let error = "";

    // Get email from URL query parameter
    onMount(() => {
        const params = new URLSearchParams(window.location.search);
        const emailParam = params.get("email");
        if (emailParam) {
            email = emailParam;
        }
    });

    async function handleUnsubscribe() {
        if (!email.trim()) {
            error = "Please enter your email address";
            return;
        }

        // Basic email validation
        const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        if (!emailRegex.test(email)) {
            error = "Please enter a valid email address";
            return;
        }

        loading = true;
        error = "";

        try {
            await api.unsubscribeFromNewsletter(email);
            success = true;
        } catch (err) {
            error = err instanceof Error ? err.message : m.unsubscribe_error();
        } finally {
            loading = false;
        }
    }

    function handleKeydown(e: KeyboardEvent) {
        if (e.key === "Enter" && !loading && !success) {
            handleUnsubscribe();
        }
    }
</script>

<svelte:head>
    <title>{m.unsubscribe_page_title()}</title>
</svelte:head>

<div class="min-h-screen bg-background flex items-center justify-center p-4">
    <Card class="w-full max-w-md p-8">
        {#if success}
            <!-- Success State -->
            <div class="text-center space-y-6">
                <div class="flex justify-center">
                    <div
                        class="w-16 h-16 bg-green-100 dark:bg-green-900/20 rounded-full flex items-center justify-center"
                    >
                        <CheckCircle
                            class="w-8 h-8 text-green-600 dark:text-green-400"
                        />
                    </div>
                </div>

                <div class="space-y-2">
                    <h1 class="text-2xl font-bold text-foreground">
                        {m.unsubscribe_success()}
                    </h1>
                    <p class="text-muted-foreground">
                        {m.unsubscribe_success_description()}
                    </p>
                </div>
            </div>
        {:else}
            <!-- Unsubscribe Form -->
            <div class="space-y-6">
                <div class="text-center space-y-2">
                    <div class="flex justify-center mb-4">
                        <div
                            class="w-16 h-16 bg-primary/10 rounded-full flex items-center justify-center"
                        >
                            <Mail class="w-8 h-8 text-primary" />
                        </div>
                    </div>

                    <h1 class="text-2xl font-bold text-foreground">
                        {m.unsubscribe_heading()}
                    </h1>
                    <p class="text-muted-foreground">
                        {m.unsubscribe_description()}
                    </p>
                </div>

                <div class="space-y-4">
                    <div>
                        <Input
                            type="email"
                            bind:value={email}
                            placeholder={m.unsubscribe_email_placeholder()}
                            disabled={loading}
                            on:keydown={handleKeydown}
                            class="w-full"
                        />
                    </div>

                    {#if error}
                        <div
                            class="flex items-center gap-2 p-3 bg-destructive/10 border border-destructive/20 rounded-md"
                        >
                            <AlertCircle
                                class="w-4 h-4 text-destructive flex-shrink-0"
                            />
                            <p class="text-sm text-destructive">{error}</p>
                        </div>
                    {/if}

                    <Button
                        on:click={handleUnsubscribe}
                        disabled={loading || !email.trim()}
                        class="w-full"
                    >
                        {#if loading}
                            <div
                                class="w-4 h-4 border-2 border-current border-t-transparent rounded-full animate-spin me-2"
                            ></div>
                            {m.unsubscribe_processing()}
                        {:else}
                            {m.unsubscribe_button()}
                        {/if}
                    </Button>
                </div>
            </div>
        {/if}
    </Card>
</div>

<style>
    :global(body) {
        margin: 0;
        padding: 0;
    }
</style>
