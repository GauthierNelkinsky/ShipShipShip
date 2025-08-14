<script lang="ts">
    import { createEventDispatcher, onMount, tick } from "svelte";
    import { fly } from "svelte/transition";
    import { Check, ChevronDown, Search } from "lucide-svelte";
    import Button from "./button.svelte";
    import Input from "./input.svelte";
    import { cn } from "$lib/utils";

    interface ComboboxOption {
        value: string;
        label: string;
        disabled?: boolean;
    }

    export let options: ComboboxOption[] = [];
    export let value: string = "";
    export let placeholder: string = "Select option...";
    export let searchPlaceholder: string = "Search...";
    export let emptyMessage: string = "No option found.";
    export let disabled: boolean = false;
    export let className: string = "";

    const dispatch = createEventDispatcher<{ select: string }>();

    let isOpen = false;
    let searchTerm = "";
    let buttonElement: HTMLButtonElement;
    let inputElement: HTMLInputElement;
    let dropdownElement: HTMLDivElement;

    $: selectedOption = options.find((option) => option.value === value);
    $: filteredOptions = options.filter((option) =>
        option.label.toLowerCase().includes(searchTerm.toLowerCase()),
    );

    function toggleOpen() {
        if (disabled) return;
        isOpen = !isOpen;
        if (isOpen) {
            tick().then(() => {
                inputElement?.focus();
            });
        } else {
            searchTerm = "";
        }
    }

    function selectOption(optionValue: string) {
        if (options.find((opt) => opt.value === optionValue)?.disabled) return;

        value = optionValue;
        isOpen = false;
        searchTerm = "";
        dispatch("select", optionValue);
        buttonElement?.focus();
    }

    function handleKeydown(event: KeyboardEvent) {
        if (disabled) return;

        if (event.key === "Escape") {
            isOpen = false;
            searchTerm = "";
            buttonElement?.focus();
        } else if (event.key === "Enter") {
            event.preventDefault();
            if (filteredOptions.length === 1 && !filteredOptions[0].disabled) {
                selectOption(filteredOptions[0].value);
            }
        } else if (event.key === "ArrowDown") {
            event.preventDefault();
            if (!isOpen) {
                toggleOpen();
            }
        }
    }

    function handleClickOutside(event: MouseEvent) {
        const target = event.target as Element;
        if (
            buttonElement &&
            !buttonElement.contains(target) &&
            dropdownElement &&
            !dropdownElement.contains(target)
        ) {
            isOpen = false;
            searchTerm = "";
        }
    }

    onMount(() => {
        document.addEventListener("click", handleClickOutside);
        return () => {
            document.removeEventListener("click", handleClickOutside);
        };
    });
</script>

<div class={cn("relative", className)}>
    <Button
        bind:this={buttonElement}
        variant="outline"
        role="combobox"
        aria-expanded={isOpen}
        aria-haspopup="listbox"
        {disabled}
        on:click={toggleOpen}
        on:keydown={handleKeydown}
        class={cn(
            "w-full justify-between",
            !selectedOption && "text-muted-foreground",
        )}
    >
        <span class="truncate">
            {selectedOption ? selectedOption.label : placeholder}
        </span>
        <ChevronDown
            class={cn(
                "ml-2 h-4 w-4 shrink-0 opacity-50 transition-transform duration-200",
                isOpen && "rotate-180",
            )}
        />
    </Button>

    {#if isOpen}
        <div
            bind:this={dropdownElement}
            transition:fly={{ duration: 200, y: -10 }}
            class="absolute z-50 mt-1 w-full rounded-md border border-border bg-popover shadow-md outline-none"
            role="listbox"
        >
            <div class="flex items-center border-b border-border px-3">
                <Search class="mr-2 h-4 w-4 shrink-0 opacity-50" />
                <Input
                    bind:this={inputElement}
                    bind:value={searchTerm}
                    placeholder={searchPlaceholder}
                    class="flex h-10 w-full rounded-md bg-transparent py-3 text-sm outline-none placeholder:text-muted-foreground disabled:cursor-not-allowed disabled:opacity-50 border-0 focus-visible:ring-0 focus-visible:ring-offset-0"
                    on:keydown={handleKeydown}
                />
            </div>

            <div class="max-h-60 overflow-auto p-1">
                {#if filteredOptions.length === 0}
                    <div class="py-6 text-center text-sm text-muted-foreground">
                        {emptyMessage}
                    </div>
                {:else}
                    {#each filteredOptions as option (option.value)}
                        <button
                            type="button"
                            role="option"
                            aria-selected={value === option.value}
                            disabled={option.disabled}
                            class={cn(
                                "relative flex w-full cursor-pointer select-none items-center rounded-sm py-1.5 pl-8 pr-2 text-sm outline-none",
                                "hover:bg-accent hover:text-accent-foreground",
                                "focus:bg-accent focus:text-accent-foreground",
                                "disabled:pointer-events-none disabled:opacity-50",
                                value === option.value &&
                                    "bg-accent text-accent-foreground",
                            )}
                            on:click={() => selectOption(option.value)}
                        >
                            <span
                                class={cn(
                                    "absolute left-2 flex h-3.5 w-3.5 items-center justify-center",
                                    value === option.value
                                        ? "opacity-100"
                                        : "opacity-0",
                                )}
                            >
                                <Check class="h-4 w-4" />
                            </span>
                            <span class="truncate">{option.label}</span>
                        </button>
                    {/each}
                {/if}
            </div>
        </div>
    {/if}
</div>
