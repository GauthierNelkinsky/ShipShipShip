<script lang="ts">
    import { Link2, Unlink } from "lucide-svelte";
    import * as m from "$lib/paraglide/messages";

    type StatusCategoryMapping = {
        categoryId: string | null;
        categoryLabel: string;
        categoryDescription: string;
    };

    export let statusLabel: string;
    export let eventCount: number;
    export let categoryMapping: StatusCategoryMapping | undefined = undefined;
    export let showCount: boolean = true;
</script>

<div class="flex items-center justify-between">
    <div class="flex items-center gap-1.5 flex-1">
        <span class="text-sm font-medium">
            {statusLabel}
        </span>
    </div>
    <div class="flex items-center gap-2">
        {#if categoryMapping}
            <div
                class="relative group/mapping"
                title={categoryMapping.categoryId
                    ? m.events_page_linked_to_theme_title({
                          categoryLabel: categoryMapping.categoryLabel,
                      })
                    : m.events_page_not_linked_to_theme()}
            >
                {#if categoryMapping.categoryId}
                    <Link2
                        class="h-3.5 w-3.5 text-blue-600 dark:text-blue-500"
                    />
                {:else}
                    <Unlink class="h-3.5 w-3.5 text-muted-foreground/50" />
                {/if}
                <div
                    class="absolute right-0 top-full mt-2 w-64 p-3 rounded-lg border bg-popover text-popover-foreground shadow-lg opacity-0 invisible group-hover/mapping:opacity-100 group-hover/mapping:visible transition-all z-50 pointer-events-none"
                >
                    {#if categoryMapping.categoryId}
                        <div class="flex items-center gap-2 mb-2">
                            <Link2
                                class="h-3.5 w-3.5 text-blue-600 dark:text-blue-500 flex-shrink-0"
                            />
                            <span
                                class="text-xs font-semibold text-blue-600 dark:text-blue-500"
                            >
                                {m.events_page_linked_to_theme()}
                            </span>
                        </div>
                        <div class="text-xs mb-1">
                            <strong>{categoryMapping.categoryLabel}</strong>
                        </div>
                        {#if categoryMapping.categoryDescription}
                            <div class="text-xs text-muted-foreground">
                                {categoryMapping.categoryDescription}
                            </div>
                        {/if}
                    {:else}
                        <div class="flex items-center gap-2 mb-2">
                            <Unlink
                                class="h-3.5 w-3.5 text-muted-foreground flex-shrink-0"
                            />
                            <span class="text-xs font-semibold">
                                {m.events_page_not_linked_to_theme()}
                            </span>
                        </div>
                        <div class="text-xs text-muted-foreground">
                            {m.events_page_not_linked_description()}
                        </div>
                    {/if}
                </div>
            </div>
        {/if}
        {#if showCount}
            <span
                class="text-xs text-muted-foreground bg-muted rounded px-1.5 py-0.5"
            >
                {eventCount}
            </span>
        {/if}
    </div>
</div>
