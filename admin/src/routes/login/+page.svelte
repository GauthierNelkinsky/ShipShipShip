<script lang="ts">
    import { onMount } from "svelte";
    import { goto } from "$app/navigation";
    import { api } from "$lib/api";
    import { authStore } from "$lib/stores/auth";
    import { Eye, EyeOff, Lock, User } from "lucide-svelte";
    import { Button, Card, Input } from "$lib/components/ui";

    let username = "";
    let password = "";
    let showPassword = false;
    let loading = false;
    let error = "";

    onMount(async () => {
        // Check if user is already authenticated
        const isAuthenticated = await authStore.init();
        if (isAuthenticated) {
            goto("/admin/events");
        }
    });

    async function handleLogin() {
        if (!username || !password) {
            error = "Please enter both username and password";
            return;
        }

        loading = true;
        error = "";

        try {
            const loginResult = await api.login(username, password);
            // Set authenticated state in the store
            authStore.setAuthenticated();
            goto("/admin/events");
        } catch (err) {
            error = err instanceof Error ? err.message : "Login failed";
        } finally {
            loading = false;
        }
    }

    function handleKeyDown(event: KeyboardEvent) {
        if (event.key === "Enter") {
            handleLogin();
        }
    }
</script>

<svelte:head>
    <title>Admin Login - Changelog</title>
    <meta
        name="description"
        content="Admin panel login for changelog management"
    />
</svelte:head>

<div class="min-h-screen bg-background flex items-center justify-center p-4">
    <div class="w-full max-w-md">
        <Card class="p-8">
            <div class="text-center mb-8">
                <h1 class="text-2xl font-bold mb-2">Admin Login</h1>
                <p class="text-muted-foreground">
                    Enter your credentials to access the admin panel
                </p>
            </div>

            {#if error}
                <div
                    class="bg-destructive/10 border border-destructive/20 text-destructive px-4 py-3 rounded-md mb-6"
                >
                    {error}
                </div>
            {/if}

            <form on:submit|preventDefault={handleLogin} class="space-y-6">
                <div>
                    <label for="username" class="label block mb-2"
                        >Username</label
                    >
                    <div class="relative">
                        <User
                            class="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground"
                        />
                        <Input
                            id="username"
                            type="text"
                            bind:value={username}
                            on:keydown={handleKeyDown}
                            placeholder="Enter username"
                            class="pl-10"
                            disabled={loading}
                            autocomplete="username"
                        />
                    </div>
                </div>

                <div>
                    <label for="password" class="label block mb-2"
                        >Password</label
                    >
                    <div class="relative">
                        <Lock
                            class="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground"
                        />
                        {#if showPassword}
                            <Input
                                id="password"
                                type="text"
                                bind:value={password}
                                on:keydown={handleKeyDown}
                                placeholder="Enter password"
                                class="pl-10 pr-10"
                                disabled={loading}
                                autocomplete="current-password"
                            />
                        {:else}
                            <Input
                                id="password"
                                type="password"
                                bind:value={password}
                                on:keydown={handleKeyDown}
                                placeholder="Enter password"
                                class="pl-10 pr-10"
                                disabled={loading}
                                autocomplete="current-password"
                            />
                        {/if}
                        <button
                            type="button"
                            on:click={() => (showPassword = !showPassword)}
                            class="absolute right-3 top-1/2 transform -translate-y-1/2 text-muted-foreground hover:text-foreground"
                        >
                            {#if showPassword}
                                <EyeOff class="h-4 w-4" />
                            {:else}
                                <Eye class="h-4 w-4" />
                            {/if}
                        </button>
                    </div>
                </div>

                <Button
                    type="submit"
                    class="w-full"
                    disabled={loading || !username || !password}
                >
                    {#if loading}
                        <div
                            class="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"
                        ></div>
                        Signing in...
                    {:else}
                        Sign In
                    {/if}
                </Button>
            </form>

            <div class="mt-6 text-center">
                <a
                    href="/"
                    class="text-sm text-muted-foreground hover:text-foreground"
                >
                    ← Back to changelog
                </a>
            </div>
        </Card>
    </div>
</div>
