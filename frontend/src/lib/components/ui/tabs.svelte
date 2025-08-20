<!-- TabsList Component -->
<script lang="ts" context="module">
    export function TabsList(node: HTMLElement, { className = "" } = {}) {
        if (!node) return {};

        node.className = cn(
            "inline-flex h-10 items-center justify-center rounded-md bg-muted p-1 text-muted-foreground",
            className,
        );
        return {};
    }

    export function TabsTrigger(
        node: HTMLElement,
        { value: triggerValue, activeValue, className = "" } = {} as {
            value: string;
            activeValue: string;
            className?: string;
        },
    ) {
        if (!node) return {};

        const isActive = triggerValue === activeValue;
        node.className = cn(
            "inline-flex items-center justify-center whitespace-nowrap rounded-sm px-3 py-1.5 text-sm font-medium ring-offset-background transition-all focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50",
            isActive
                ? "bg-background text-foreground shadow-sm"
                : "hover:bg-background/60 hover:text-foreground",
            className,
        );
        return {
            update({
                value: newValue,
                activeValue: newActiveValue,
                className: newClassName = "",
            }: {
                value: string;
                activeValue: string;
                className?: string;
            }) {
                if (!node) return;

                const newIsActive = newValue === newActiveValue;
                node.className = cn(
                    "inline-flex items-center justify-center whitespace-nowrap rounded-sm px-3 py-1.5 text-sm font-medium ring-offset-background transition-all focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50",
                    newIsActive
                        ? "bg-background text-foreground shadow-sm"
                        : "hover:bg-background/60 hover:text-foreground",
                    newClassName,
                );
            },
        };
    }

    export function TabsContent(
        node: HTMLElement,
        { value: contentValue, activeValue, className = "" } = {} as {
            value: string;
            activeValue: string;
            className?: string;
        },
    ) {
        if (!node || !node.style) return {};

        const isActive = contentValue === activeValue;
        node.className = cn(
            "mt-2 ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2",
            className,
        );
        node.style.display = isActive ? "block" : "none";

        return {
            update({
                value: newValue,
                activeValue: newActiveValue,
                className: newClassName = "",
            }: {
                value: string;
                activeValue: string;
                className?: string;
            }) {
                if (!node || !node.style) return;

                const newIsActive = newValue === newActiveValue;
                node.className = cn(
                    "mt-2 ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2",
                    newClassName,
                );
                node.style.display = newIsActive ? "block" : "none";
            },
        };
    }
</script>

<script lang="ts">
    import { createEventDispatcher } from "svelte";
    import { cn } from "$lib/utils";

    const dispatch = createEventDispatcher();

    export let value: string = "";
    export let className: string = "";

    function handleTabClick(tabValue: string) {
        value = tabValue;
        dispatch("change", tabValue);
    }
</script>

<div class={cn("w-full", className)}>
    <slot {value} {handleTabClick} />
</div>
