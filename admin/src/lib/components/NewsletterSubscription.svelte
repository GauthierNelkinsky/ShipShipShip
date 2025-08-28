<script lang="ts">
    import { Mail, Send, Check, X } from "lucide-svelte";
    import { Button, Input } from "$lib/components/ui";

    export let variant: "sidebar" | "inline" = "sidebar";

    let email = "";
    let loading = false;
    let success = false;
    let alreadySubscribed = false;
    let error = "";

    async function handleSubscribe() {
        if (!email || !email.includes("@")) {
            error = "Please enter a valid email address";
            return;
        }

        loading = true;
        error = "";
        alreadySubscribed = false;

        try {
            const response = await fetch("/api/newsletter/subscribe", {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                },
                body: JSON.stringify({ email }),
            });

            const data = await response.json();

            if (response.ok) {
                if (data.already_subscribed) {
                    alreadySubscribed = true;
                    email = "";
                    setTimeout(() => {
                        alreadySubscribed = false;
                    }, 3000);
                } else {
                    success = true;
                    email = "";
                    setTimeout(() => {
                        success = false;
                    }, 3000);
                }
            } else {
                error = data.error || "Failed to subscribe to newsletter";
            }
        } catch (err) {
            error = "Network error. Please try again.";
        } finally {
            loading = false;
        }
    }

    function handleKeyPress(event: KeyboardEvent) {
        if (event.key === "Enter") {
            handleSubscribe();
        }
    }
</script>

{#if variant === "sidebar"}
    <div class="bg-gray-50 dark:bg-neutral-800/50 rounded-lg p-4">
        <h3
            class="text-lg font-semibold mb-3 text-gray-900 dark:text-neutral-100 flex items-center gap-2"
        >
            <Mail class="h-5 w-5 text-primary" />
            Stay Updated
        </h3>

        <p class="text-muted-foreground text-sm mb-3">
            Get notified about new features and updates.
        </p>

        {#if success}
            <div
                class="bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800 text-green-800 dark:text-green-200 px-3 py-2 rounded-lg text-sm mb-3"
            >
                Successfully subscribed to newsletter!
            </div>
        {:else if alreadySubscribed}
            <div
                class="bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 text-blue-800 dark:text-blue-200 px-3 py-2 rounded-lg text-sm mb-3"
            >
                You are already subscribed to our newsletter!
            </div>
        {:else}
            <div class="space-y-2.5">
                <Input
                    type="email"
                    placeholder="Enter your email"
                    bind:value={email}
                    on:keypress={handleKeyPress}
                    disabled={loading}
                    class={error ? "border-red-500 focus:border-red-500" : ""}
                />

                {#if error}
                    <div class="text-sm text-red-600 dark:text-red-400">
                        {error}
                    </div>
                {/if}

                <Button
                    on:click={handleSubscribe}
                    disabled={loading || !email}
                    class="w-full"
                    size="sm"
                >
                    {#if loading}
                        <div
                            class="animate-spin rounded-full h-3.5 w-3.5 border-b-2 border-white mr-2"
                        ></div>
                        Subscribing...
                    {:else}
                        <Send class="h-3.5 w-3.5 mr-2" />
                        Subscribe
                    {/if}
                </Button>
            </div>
        {/if}
    </div>
{:else}
    <!-- Inline variant for mobile -->
    <div class="bg-gray-50 dark:bg-neutral-800/50 rounded-lg p-4">
        <h3
            class="text-lg font-semibold mb-3 text-gray-900 dark:text-neutral-100 flex items-center gap-2"
        >
            <Mail class="h-5 w-5 text-primary" />
            Stay Updated
        </h3>

        <p class="text-muted-foreground text-sm mb-3">
            Subscribe to our newsletter and be the first to know about new
            features and updates.
        </p>

        {#if success}
            <div
                class="bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800 text-green-800 dark:text-green-200 px-3 py-2 rounded-lg text-sm mb-3"
            >
                Successfully subscribed to newsletter!
            </div>
        {:else if alreadySubscribed}
            <div
                class="bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 text-blue-800 dark:text-blue-200 px-3 py-2 rounded-lg text-sm mb-3"
            >
                You are already subscribed to our newsletter!
            </div>
        {:else}
            <div class="space-y-2.5">
                <div class="flex gap-2">
                    <Input
                        type="email"
                        placeholder="Enter your email address"
                        bind:value={email}
                        on:keypress={handleKeyPress}
                        disabled={loading}
                        class={`flex-1 ${error ? "border-red-500 focus:border-red-500" : ""}`}
                    />
                    <Button
                        on:click={handleSubscribe}
                        disabled={loading || !email}
                        size="default"
                    >
                        {#if loading}
                            <div
                                class="animate-spin rounded-full h-3.5 w-3.5 border-b-2 border-white"
                            ></div>
                        {:else}
                            <Send class="h-3.5 w-3.5" />
                        {/if}
                    </Button>
                </div>

                {#if error}
                    <div class="text-sm text-red-600 dark:text-red-400">
                        {error}
                    </div>
                {/if}
            </div>
        {/if}
    </div>
{/if}
