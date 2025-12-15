<script lang="ts">
    import { Button } from "$lib/components/ui";
    import { Loader2, AlertCircle } from "lucide-svelte";
    import { api } from "$lib/api";
    import { toast } from "svelte-sonner";
    import * as m from "$lib/paraglide/messages";

    export let currentThemeId: string | null = null;

    let redownloading = false;

    async function redownloadTheme() {
        redownloading = true;
        try {
            const result = await api.redownloadTheme();

            if (result.success) {
                toast.success(m.theme_redownload_success(), {
                    description: `${result.themeName} v${result.version}`,
                });
                // Reload the page to show the fixed settings
                setTimeout(() => {
                    window.location.reload();
                }, 1000);
            }
        } catch (err) {
            console.error("Failed to redownload theme:", err);
            const errorMessage =
                err instanceof Error
                    ? err.message
                    : m.theme_redownload_failed();
            toast.error(m.theme_redownload_failed(), {
                description: errorMessage,
            });
        } finally {
            redownloading = false;
        }
    }
</script>

<div class="flex-1 p-6">
    <div
        class="flex items-start gap-3 bg-amber-50 dark:bg-amber-950/20 border border-amber-200 dark:border-amber-900 text-amber-900 dark:text-amber-200 p-6 rounded-lg"
    >
        <AlertCircle class="h-5 w-5 flex-shrink-0 mt-0.5" />
        <div class="flex-1">
            <p class="font-semibold text-lg mb-3">
                {m.theme_error_title()}
            </p>
            <p class="text-sm mb-3 leading-relaxed">
                <strong>{m.theme_error_what_is_theme()}</strong>
                {m.theme_error_theme_explanation()}
            </p>
            <p class="text-sm mb-4 leading-relaxed">
                {m.theme_error_description()}
            </p>
            <div
                class="bg-white dark:bg-gray-900 border border-amber-200 dark:border-amber-800 rounded-md p-4 mb-4"
            >
                <p class="text-sm font-semibold mb-3">
                    {m.theme_error_how_to_fix()}
                </p>
                <ol class="text-sm space-y-2 list-decimal list-inside ml-1">
                    <li class="pl-2">
                        {@html m.theme_error_step_1()}
                    </li>
                    <li class="pl-2">
                        {m.theme_error_step_2()}
                    </li>
                    <li class="pl-2">
                        {m.theme_error_step_3()}
                    </li>
                </ol>
            </div>
            <div class="flex gap-3">
                <Button
                    on:click={() =>
                        (window.location.href = "/admin/appearance/themes")}
                    class="mt-2"
                >
                    {m.theme_error_go_to_themes()}
                </Button>
                {#if currentThemeId}
                    <Button
                        on:click={redownloadTheme}
                        disabled={redownloading}
                        variant="outline"
                        class="mt-2"
                    >
                        {#if redownloading}
                            <Loader2 class="h-4 w-4 animate-spin mr-2" />
                            {m.theme_error_redownloading()}
                        {:else}
                            {m.theme_error_redownload()}
                        {/if}
                    </Button>
                {/if}
            </div>
            <p class="text-xs mt-4 opacity-75">
                {m.theme_error_need_help()}
                <a
                    href="https://github.com/GauthierNelkinsky/ShipShipShip/issues/new"
                    target="_blank"
                    rel="noopener noreferrer"
                    class="underline hover:text-foreground"
                >
                    {m.theme_error_create_issue()}
                </a>
                {m.theme_error_create_issue_suffix()}
            </p>
        </div>
    </div>
</div>
