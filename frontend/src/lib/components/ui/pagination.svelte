<script lang="ts">
    import { createEventDispatcher } from "svelte";
    import { Button } from "$lib/components/ui";
    import { ChevronLeft, ChevronRight, MoreHorizontal } from "lucide-svelte";

    export let currentPage: number = 1;
    export let totalPages: number = 1;
    export let siblingCount: number = 1;
    export let showFirstLast: boolean = true;
    export let disabled: boolean = false;

    const dispatch = createEventDispatcher<{
        pageChange: number;
    }>();

    function onPageChange(page: number) {
        if (disabled || page === currentPage || page < 1 || page > totalPages) return;
        dispatch("pageChange", page);
    }

    function getPageNumbers() {
        const pages: (number | "ellipsis")[] = [];

        if (totalPages <= 7) {
            // Show all pages if total is 7 or less
            for (let i = 1; i <= totalPages; i++) {
                pages.push(i);
            }
        } else {
            // Always show first page
            if (showFirstLast) pages.push(1);

            const leftSiblingIndex = Math.max(currentPage - siblingCount, 1);
            const rightSiblingIndex = Math.min(currentPage + siblingCount, totalPages);

            const shouldShowLeftEllipsis = leftSiblingIndex > (showFirstLast ? 3 : 2);
            const shouldShowRightEllipsis = rightSiblingIndex < (showFirstLast ? totalPages - 2 : totalPages - 1);

            if (!shouldShowLeftEllipsis && shouldShowRightEllipsis) {
                const leftItemCount = 3 + 2 * siblingCount;
                const leftRange = Array.from({ length: leftItemCount }, (_, i) => i + 1);
                pages.push(...leftRange);
                pages.push("ellipsis");
                if (showFirstLast) pages.push(totalPages);
            } else if (shouldShowLeftEllipsis && !shouldShowRightEllipsis) {
                const rightItemCount = 3 + 2 * siblingCount;
                const rightRange = Array.from(
                    { length: rightItemCount },
                    (_, i) => totalPages - rightItemCount + i + 1
                );
                if (showFirstLast && pages.length === 1) pages.push("ellipsis");
                pages.push(...rightRange);
            } else if (shouldShowLeftEllipsis && shouldShowRightEllipsis) {
                if (showFirstLast && pages.length === 1) pages.push("ellipsis");
                const middleRange = Array.from(
                    { length: rightSiblingIndex - leftSiblingIndex + 1 },
                    (_, i) => leftSiblingIndex + i
                );
                pages.push(...middleRange);
                pages.push("ellipsis");
                if (showFirstLast) pages.push(totalPages);
            }
        }

        return pages;
    }

    $: pageNumbers = getPageNumbers();
</script>

<nav class="flex items-center justify-center space-x-2" aria-label="Pagination">
    <Button
        variant="outline"
        size="icon"
        on:click={() => onPageChange(currentPage - 1)}
        disabled={disabled || currentPage === 1}
        class="h-8 w-8"
    >
        <ChevronLeft class="h-4 w-4" />
        <span class="sr-only">Previous page</span>
    </Button>

    {#each pageNumbers as page}
        {#if page === "ellipsis"}
            <span class="flex h-8 w-8 items-center justify-center">
                <MoreHorizontal class="h-4 w-4" />
            </span>
        {:else}
            <Button
                variant={page === currentPage ? "default" : "outline"}
                size="icon"
                on:click={() => onPageChange(page)}
                disabled={disabled}
                class="h-8 w-8"
            >
                {page}
            </Button>
        {/if}
    {/each}

    <Button
        variant="outline"
        size="icon"
        on:click={() => onPageChange(currentPage + 1)}
        disabled={disabled || currentPage === totalPages}
        class="h-8 w-8"
    >
        <ChevronRight class="h-4 w-4" />
        <span class="sr-only">Next page</span>
    </Button>
</nav>
