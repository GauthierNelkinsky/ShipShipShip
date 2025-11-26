<script lang="ts">
    import { getLocale, locales, setLocale } from "$lib/paraglide/runtime";
    import Icon from "@iconify/svelte";
    import * as m from "$lib/paraglide/messages";

    export let collapsed = false;

    const languages: Record<string, string> = {
        en: "English",
        de: "Deutsch",
        fr: "Français",
        es: "Español",
        zh: "中文",
    };

    const flags: Record<string, string> = {
        en: "circle-flags:uk",
        de: "circle-flags:de",
        fr: "circle-flags:fr",
        es: "circle-flags:es",
        zh: "circle-flags:cn",
    };

    let isOpen = false;

    function toggleDropdown() {
        isOpen = !isOpen;
    }

    async function switchLanguage(lang: "en" | "de" | "fr" | "es" | "zh") {
        await setLocale(lang);
        isOpen = false;
    }

    function handleClickOutside(event: MouseEvent) {
        const target = event.target as HTMLElement;
        const dropdown = document.getElementById("language-dropdown");
        if (dropdown && !dropdown.contains(target)) {
            isOpen = false;
        }
    }
</script>

<svelte:window on:click={handleClickOutside} />

<div class="relative" id="language-dropdown">
    <button
        on:click|stopPropagation={toggleDropdown}
        class="flex items-center gap-3 px-3 py-2 rounded-lg text-sm font-medium text-muted-foreground hover:text-foreground hover:bg-accent transition-all duration-200 w-full {collapsed
            ? 'justify-center'
            : ''}"
        title={collapsed ? m.sidebar_language() : ""}
    >
        <Icon icon={flags[getLocale()]} class="h-4 w-4 flex-shrink-0" />
        {#if !collapsed}
            <span class="flex-1 text-left">{languages[getLocale()]}</span>
            <svg
                class="h-4 w-4 flex-shrink-0 transition-transform {isOpen
                    ? 'rotate-180'
                    : ''}"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
            >
                <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M19 9l-7 7-7-7"
                />
            </svg>
        {/if}
    </button>

    {#if isOpen}
        <div
            class="absolute bottom-full {collapsed
                ? 'left-0'
                : 'left-0'} mb-1 {collapsed
                ? 'w-auto min-w-[160px]'
                : 'w-full'} rounded-md shadow-lg bg-popover border border-border overflow-hidden z-50"
        >
            {#each locales as lang}
                <button
                    on:click|stopPropagation={() => switchLanguage(lang)}
                    class="flex items-center gap-2 w-full text-left px-3 py-2 text-sm hover:bg-accent transition-colors {lang ===
                    getLocale()
                        ? 'bg-accent font-medium'
                        : ''}"
                >
                    <Icon icon={flags[lang]} class="h-4 w-4 flex-shrink-0" />
                    {languages[lang]}
                </button>
            {/each}
        </div>
    {/if}
</div>
