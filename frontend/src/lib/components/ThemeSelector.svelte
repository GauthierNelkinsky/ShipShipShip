<script lang="ts">
    import { theme, type ThemePreference } from "$lib/stores/theme";
    import { Sun, Moon, ChevronDown } from "lucide-svelte";
    import { Button } from "$lib/components/ui";
    import { onMount } from "svelte";
    import { fly } from "svelte/transition";

    let selectedTheme: ThemePreference = "light";
    let isOpen = false;

    onMount(() => {
        selectedTheme = theme.getPreference();
    });

    const themeOptions = [
        {
            value: "light" as const,
            label: "Light",
            icon: Sun,
        },
        {
            value: "dark" as const,
            label: "Dark",
            icon: Moon,
        },
    ];

    function handleThemeSelect(themeValue: ThemePreference) {
        selectedTheme = themeValue;
        theme.setPreference(themeValue);
        isOpen = false;
    }

    function toggleDropdown() {
        isOpen = !isOpen;
    }

    function handleClickOutside(event: MouseEvent) {
        const target = event.target as HTMLElement;
        if (!target.closest(".theme-selector")) {
            isOpen = false;
        }
    }

    onMount(() => {
        const handleDocumentClick = (event: MouseEvent) =>
            handleClickOutside(event);
        document.addEventListener("click", handleDocumentClick);
        return () => document.removeEventListener("click", handleDocumentClick);
    });

    $: currentOption = themeOptions.find(
        (option) => option.value === selectedTheme,
    );
</script>

<div class="relative theme-selector">
    <!-- Theme Toggle Button -->
    <Button
        variant="outline"
        size="sm"
        on:click={toggleDropdown}
        class="h-8 px-2 gap-1"
        type="button"
        title="Select theme"
        aria-expanded={isOpen}
        aria-haspopup="true"
    >
        {#if currentOption}
            <svelte:component this={currentOption.icon} class="h-3 w-3" />
            <span class="text-xs">{currentOption.label}</span>
        {/if}
        <ChevronDown
            class="h-3 w-3 transition-transform duration-200 {isOpen
                ? 'rotate-180'
                : ''}"
        />
    </Button>

    <!-- Dropdown Menu -->
    {#if isOpen}
        <div
            transition:fly={{ duration: 200, y: -10 }}
            class="absolute right-0 mt-1 w-24 rounded-md border border-border bg-popover shadow-md z-50"
            role="menu"
            aria-orientation="vertical"
        >
            <div class="p-1">
                {#each themeOptions as option}
                    <Button
                        variant="ghost"
                        on:click={() => handleThemeSelect(option.value)}
                        class="flex items-center w-full px-2 py-1.5 text-xs rounded-sm justify-start h-auto {selectedTheme ===
                        option.value
                            ? 'bg-accent text-accent-foreground'
                            : ''}"
                        role="menuitem"
                    >
                        <svelte:component
                            this={option.icon}
                            class="h-3 w-3 mr-2 flex-shrink-0"
                        />
                        <span class="font-medium">{option.label}</span>
                        {#if selectedTheme === option.value}
                            <div
                                class="w-1 h-1 bg-primary rounded-full ml-auto"
                            ></div>
                        {/if}
                    </Button>
                {/each}
            </div>
        </div>
    {/if}
</div>
