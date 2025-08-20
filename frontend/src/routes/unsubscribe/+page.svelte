<script lang="ts">
    import { onMount } from "svelte";
    import { page } from "$app/stores";
    import { api } from "$lib/api";
    import { Button } from "$lib/components/ui";
    import { Mail, CheckCircle, AlertCircle, Loader2 } from "lucide-svelte";

    let email = "";
    let loading = false;
    let success = false;
    let error = "";
    let processing = false;

    onMount(() => {
        // Get email from URL query parameter
        email = $page.url.searchParams.get("email") || "";

        // If email is provided in URL, automatically attempt unsubscribe
        if (email) {
            handleUnsubscribe();
        }
    });

    async function handleUnsubscribe() {
        if (!email || !email.includes("@")) {
            error = "Please enter a valid email address";
            return;
        }

        processing = true;
        error = "";
        success = false;

        try {
            await api.unsubscribeFromNewsletter(email);
            success = true;
        } catch (err) {
            error = err instanceof Error ? err.message : "Failed to unsubscribe from newsletter";
        } finally {
            processing = false;
        }
    }

    function handleKeyPress(event: KeyboardEvent) {
        if (event.key === "Enter") {
            handleUnsubscribe();
        }
    }
</script>

<svelte:head>
    <title>Unsubscribe - Newsletter</title>
    <meta name="description" content="Unsubscribe from our newsletter" />
</svelte:head>

<div class="min-h-screen bg-gray-50 dark:bg-neutral-900 flex items-center justify-center py-12 px-4 sm:px-6 lg:px-8">
    <div class="max-w-md w-full space-y-8">
        <div class="text-center">
            <div class="mx-auto h-12 w-12 bg-blue-100 dark:bg-blue-900/20 rounded-full flex items-center justify-center">
                <Mail class="h-6 w-6 text-blue-600 dark:text-blue-400" />
            </div>
            <h2 class="mt-6 text-3xl font-bold text-gray-900 dark:text-neutral-100">
                Unsubscribe from Newsletter
            </h2>
            <p class="mt-2 text-sm text-gray-600 dark:text-neutral-400">
                We're sorry to see you go. Enter your email address to unsubscribe from our newsletter.
            </p>
        </div>

        <div class="bg-white dark:bg-neutral-800 shadow-sm rounded-lg p-6 space-y-6">
            {#if success}
                <div class="text-center space-y-4">
                    <div class="mx-auto h-12 w-12 bg-green-100 dark:bg-green-900/20 rounded-full flex items-center justify-center">
                        <CheckCircle class="h-6 w-6 text-green-600 dark:text-green-400" />
                    </div>
                    <div>
                        <h3 class="text-lg font-medium text-gray-900 dark:text-neutral-100">
                            Successfully Unsubscribed
                        </h3>
                        <p class="mt-1 text-sm text-gray-600 dark:text-neutral-400">
                            You have been successfully unsubscribed from our newsletter. You will no longer receive emails from us.
                        </p>
                        <p class="mt-2 text-sm text-gray-600 dark:text-neutral-400">
                            If you change your mind, you can always subscribe again on our website.
                        </p>
                    </div>
                    <Button
                        on:click={() => window.location.href = "/"}
                        class="w-full"
                    >
                        Back to Website
                    </Button>
                </div>
            {:else}
                <div class="space-y-4">
                    {#if error}
                        <div class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 text-red-800 dark:text-red-200 px-4 py-3 rounded-lg flex items-center gap-2">
                            <AlertCircle class="h-4 w-4" />
                            {error}
                        </div>
                    {/if}

                    <div>
                        <label for="email" class="block text-sm font-medium text-gray-700 dark:text-neutral-300 mb-2">
                            Email Address
                        </label>
                        <input
                            id="email"
                            type="email"
                            bind:value={email}
                            on:keypress={handleKeyPress}
                            placeholder="Enter your email address"
                            disabled={processing}
                            class="w-full px-3 py-2 border border-gray-300 dark:border-neutral-600 rounded-lg bg-white dark:bg-neutral-700 text-gray-900 dark:text-neutral-100 placeholder-gray-500 dark:placeholder-neutral-400 focus:outline-none focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-400 focus:border-transparent transition-colors disabled:opacity-50 disabled:cursor-not-allowed {error ? 'border-red-500 focus:ring-red-500' : ''}"
                        />
                    </div>

                    <Button
                        on:click={handleUnsubscribe}
                        disabled={processing || !email.trim()}
                        class="w-full"
                    >
                        {#if processing}
                            <Loader2 class="h-4 w-4 mr-2 animate-spin" />
                            Unsubscribing...
                        {:else}
                            Unsubscribe
                        {/if}
                    </Button>

                    <div class="text-center">
                        <a
                            href="/"
                            class="text-sm text-blue-600 dark:text-blue-400 hover:text-blue-700 dark:hover:text-blue-300 transition-colors"
                        >
                            Back to Website
                        </a>
                    </div>
                </div>
            {/if}
        </div>

        <div class="text-center">
            <p class="text-xs text-gray-500 dark:text-neutral-500">
                If you're having trouble unsubscribing, please contact our support team.
            </p>
        </div>
    </div>
</div>
