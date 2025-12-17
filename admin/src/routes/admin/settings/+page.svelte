<script lang="ts">
    import { onMount } from "svelte";
    import { api, getImageUrl } from "$lib/api";
    import type { UpdateSettingsRequest } from "$lib/types";
    import { Save, Loader2, Globe, Image } from "lucide-svelte";
    import { Button, Input } from "$lib/components/ui";
    import { toast } from "svelte-sonner";
    import * as m from "$lib/paraglide/messages";

    interface SettingSection {
        id: string;
        title: string;
        description: string;
    }

    let loading = true;
    let saving = false;

    // Form data
    let title = "";
    let websiteUrl = "";
    let faviconUrl = "";

    // URL validation state
    $: websiteUrlValid = validateUrl(websiteUrl);

    // Sidebar navigation
    let activeSection = "branding";
    let sidebarTop: number | null = null;
    let sidebarElement: HTMLElement;

    const sections: SettingSection[] = [
        {
            id: "branding",
            title: m.settings_section_branding(),
            description: m.settings_branding_description(),
        },
        {
            id: "favicon",
            title: m.settings_section_favicon(),
            description: m.settings_favicon_description(),
        },
    ];

    function handleScroll() {
        if (!sidebarElement) return;

        const scrollTop =
            window.pageYOffset || document.documentElement.scrollTop;

        // Get initial position of sidebar from its parent container
        const initialTop = sidebarElement.offsetTop || 0;

        // If scrolled past initial position, stick to top with padding
        if (scrollTop > initialTop - 24) {
            sidebarTop = 24; // 1.5rem top padding from viewport top
        } else {
            sidebarTop = initialTop - scrollTop + 24;
        }

        updateActiveSectionOnScroll();
    }

    function updateActiveSectionOnScroll() {
        const scrollPosition = window.scrollY + 150;

        let closestSection: string | null = null;
        let closestDistance = Infinity;

        sections.forEach((section) => {
            const element = document.getElementById(`section-${section.id}`);
            if (!element) return;

            const rect = element.getBoundingClientRect();
            const distance = Math.abs(
                rect.top + window.scrollY - scrollPosition,
            );

            if (distance < closestDistance) {
                closestDistance = distance;
                closestSection = section.id;
            }
        });

        if (closestSection && closestSection !== activeSection) {
            activeSection = closestSection;
        }
    }

    onMount(() => {
        loadSettings();
        loading = false;

        // Set first section as active by default
        if (sections.length > 0) {
            activeSection = sections[0].id;
        }

        // Wait for DOM to be fully rendered before calculating position
        setTimeout(() => {
            const onScroll = (() => {
                let ticking = false;
                return () => {
                    if (!ticking) {
                        window.requestAnimationFrame(() => {
                            handleScroll();
                            ticking = false;
                        });
                        ticking = true;
                    }
                };
            })();

            window.addEventListener("scroll", onScroll, { passive: true });

            // Initial call to set sidebar position
            handleScroll();

            return () => {
                window.removeEventListener("scroll", onScroll);
            };
        }, 100);
    });

    async function loadSettings() {
        try {
            const data = await api.getSettings();
            title = data.title || "";
            websiteUrl = data.website_url || "";
            faviconUrl = data.favicon_url || "";
        } catch (err) {
            console.error("Failed to load settings:", err);
            toast.error(m.settings_load_failed());
        }
    }

    async function handleSave() {
        if (saving) return;

        // Validate required fields
        if (!title.trim()) {
            toast.error(m.settings_name_required());
            return;
        }

        try {
            saving = true;

            const updateData: UpdateSettingsRequest = {
                title: title.trim(),
                website_url: websiteUrl.trim() || undefined,
                favicon_url: faviconUrl.trim() || undefined,
            };

            await api.updateSettings(updateData);
            toast.success(m.settings_saved());
        } catch (err: any) {
            console.error("Failed to save settings:", err);
            const errorMessage = err.message || m.settings_save_failed();
            toast.error(errorMessage);
        } finally {
            saving = false;
        }
    }

    function validateUrl(url: string): boolean {
        if (!url || url.trim() === "") return true;
        try {
            new URL(url);
            return true;
        } catch {
            return false;
        }
    }

    async function handleImageUpload(
        event: Event & { currentTarget: HTMLInputElement },
    ) {
        const file = event.currentTarget.files?.[0];
        if (!file) return;

        try {
            const result = await api.uploadImage(file);
            faviconUrl = result.url;
            toast.success(m.settings_image_upload_success());
        } catch (err) {
            toast.error(
                err instanceof Error
                    ? err.message
                    : m.settings_image_upload_failed(),
            );
        }
    }

    function scrollToSection(sectionId: string) {
        const element = document.getElementById(`section-${sectionId}`);
        if (element) {
            element.scrollIntoView({ behavior: "smooth", block: "start" });
        }
    }
</script>

<svelte:head>
    <title>{m.settings_page_title()}</title>
</svelte:head>

<div class="w-full">
    <div class="mb-8">
        <h1 class="text-xl font-semibold mb-1">{m.settings_heading()}</h1>
        <p class="text-muted-foreground text-sm">
            {m.settings_subheading()}
        </p>
    </div>

    {#if loading}
        <div class="flex-1 flex items-center justify-center py-16">
            <div class="flex items-center gap-2 text-sm">
                <Loader2 class="h-4 w-4 animate-spin" />
                <span class="text-muted-foreground">{m.settings_loading()}</span
                >
            </div>
        </div>
    {:else}
        <div class="w-full flex gap-6">
            <!-- Sidebar Navigation -->
            <aside class="w-48 flex-shrink-0" bind:this={sidebarElement}>
                <div
                    class="fixed w-48 transition-opacity duration-200 {sidebarTop ===
                    null
                        ? 'opacity-0'
                        : 'opacity-100'}"
                    style="top: {sidebarTop !== null
                        ? `${sidebarTop}px`
                        : '0'};"
                >
                    <nav class="space-y-1">
                        {#each sections as section}
                            <button
                                on:click={() => scrollToSection(section.id)}
                                class="w-full text-left px-3 py-2 rounded-md text-sm transition-colors {activeSection ===
                                section.id
                                    ? 'bg-accent text-accent-foreground font-medium'
                                    : 'text-muted-foreground hover:text-foreground hover:bg-accent/50'}"
                            >
                                {section.title}
                            </button>
                        {/each}
                    </nav>
                </div>
            </aside>

            <!-- Main Content -->
            <div class="flex-1 min-w-0 space-y-12 ml-6">
                <!-- Website Information Section -->
                <div id="section-branding" class="scroll-mt-6">
                    <div class="mb-6">
                        <div class="flex items-center gap-3 mb-1.5">
                            <Globe class="h-5 w-5 text-primary" />
                            <h3 class="text-base font-semibold">
                                {m.settings_branding_title()}
                            </h3>
                        </div>
                        <p class="text-sm text-muted-foreground mt-1.5">
                            {m.settings_branding_description()}
                        </p>
                    </div>

                    <div class="space-y-4">
                        <div>
                            <label
                                for="title"
                                class="text-sm font-medium mb-2 block"
                            >
                                {m.settings_website_name()}
                            </label>
                            <Input
                                id="title"
                                type="text"
                                bind:value={title}
                                placeholder={m.settings_website_name_placeholder()}
                                class="max-w-md"
                            />
                            <p class="text-xs text-muted-foreground mt-1">
                                {m.settings_website_name_help()}
                            </p>
                        </div>

                        <div>
                            <label
                                for="website-url"
                                class="text-sm font-medium mb-2 block"
                            >
                                {m.settings_website_url()}
                            </label>
                            <Input
                                id="website-url"
                                type="url"
                                bind:value={websiteUrl}
                                placeholder={m.settings_website_url_placeholder()}
                                class="max-w-md"
                            />
                            {#if !websiteUrlValid}
                                <p class="text-xs text-destructive mt-1">
                                    {m.settings_website_url_invalid()}
                                </p>
                            {:else}
                                <p class="text-xs text-muted-foreground mt-1">
                                    {m.settings_website_url_help()}
                                </p>
                            {/if}
                        </div>
                    </div>
                </div>

                <!-- Favicon Section -->
                <div id="section-favicon" class="scroll-mt-6 pt-12 border-t">
                    <div class="mb-6">
                        <div class="flex items-center gap-3 mb-1.5">
                            <Image class="h-5 w-5 text-primary" />
                            <h3 class="text-base font-semibold">
                                {m.settings_favicon_title()}
                            </h3>
                        </div>
                        <p class="text-sm text-muted-foreground mt-1.5">
                            {m.settings_favicon_description()}
                        </p>
                    </div>

                    <div class="space-y-3">
                        <div class="p-3 border rounded-lg bg-muted/30">
                            {#if faviconUrl && faviconUrl !== ""}
                                {@const imageUrl = getImageUrl(faviconUrl)}
                                <div class="space-y-2">
                                    <img
                                        src={imageUrl}
                                        alt={m.settings_favicon_alt()}
                                        class="max-h-24 w-auto rounded border"
                                        on:error={(e) => {
                                            const target =
                                                e.currentTarget as HTMLImageElement;
                                            target.src =
                                                "data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='100' height='100'%3E%3Crect fill='%23ddd' width='100' height='100'/%3E%3Ctext x='50%25' y='50%25' text-anchor='middle' dy='.3em' fill='%23999'%3ENo Image%3C/text%3E%3C/svg%3E";
                                        }}
                                    />
                                    <button
                                        on:click={() => (faviconUrl = "")}
                                        class="text-xs text-destructive hover:underline"
                                        type="button"
                                    >
                                        {m.settings_remove()}
                                    </button>
                                </div>
                            {:else}
                                <div
                                    class="text-sm text-muted-foreground py-8 text-center"
                                >
                                    {m.settings_no_image()}
                                </div>
                            {/if}
                        </div>
                        <div>
                            <input
                                type="file"
                                accept="image/*,.ico"
                                on:change={handleImageUpload}
                                class="hidden"
                                id="favicon-upload"
                            />
                            <Button
                                variant="outline"
                                size="sm"
                                on:click={() =>
                                    document
                                        .getElementById("favicon-upload")
                                        ?.click()}
                                type="button"
                            >
                                {faviconUrl && faviconUrl !== ""
                                    ? m.settings_change_image()
                                    : m.settings_upload_image()}
                            </Button>
                        </div>
                    </div>
                </div>

                <!-- Save Button -->
                <div class="flex justify-end pt-6 mt-6 border-t">
                    <Button on:click={handleSave} disabled={saving}>
                        {#if saving}
                            <Loader2 class="h-4 w-4 animate-spin mr-2" />
                            {m.settings_saving()}
                        {:else}
                            <Save class="h-4 w-4 mr-2" />
                            {m.settings_save_changes()}
                        {/if}
                    </Button>
                </div>
            </div>
        </div>
    {/if}
</div>
