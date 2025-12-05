<script lang="ts">
    import { onMount } from "svelte";
    import { api } from "$lib/api";
    import type { UpdateSettingsRequest } from "$lib/types";
    import { Save, Loader2, Upload } from "lucide-svelte";
    import { Button, Input } from "$lib/components/ui";
    import ImageUploadModal from "$lib/components/ImageUploadModal.svelte";
    import { toast } from "svelte-sonner";

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
    $: faviconUrlValid = validateUrl(faviconUrl);

    // Image upload state
    let imageUploadModalOpen = false;

    // Sidebar navigation
    let activeSection = "branding";
    let sidebarTop: number | null = null;
    let sidebarElement: HTMLElement;

    const sections: SettingSection[] = [
        {
            id: "branding",
            title: "Website Information",
            description: "Configure your website name and URL",
        },
        {
            id: "favicon",
            title: "Favicon",
            description: "Upload your favicon",
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

    onMount(() => {
        loadSettings();
        loading = false;

        window.addEventListener("scroll", onScroll, { passive: true });

        return () => {
            window.removeEventListener("scroll", onScroll);
        };
    });

    async function loadSettings() {
        try {
            const data = await api.getSettings();
            title = data.title || "";
            websiteUrl = data.website_url || "";
            faviconUrl = data.favicon_url || "";
        } catch (err) {
            console.error("Failed to load settings:", err);
            toast.error("Failed to load settings");
        }
    }

    async function handleSave() {
        if (saving) return;

        // Validate required fields
        if (!title.trim()) {
            toast.error("Please enter a website name");
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
            toast.success("Settings saved successfully");
        } catch (err: any) {
            console.error("Failed to save settings:", err);
            const errorMessage =
                err.message ||
                "Failed to save settings. Please check your inputs and try again.";
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

    function openImageUpload() {
        imageUploadModalOpen = true;
    }

    function handleImageSelected(event: CustomEvent<{ url: string }>) {
        const { url } = event.detail;
        faviconUrl = url;
        imageUploadModalOpen = false;
    }

    function scrollToSection(sectionId: string) {
        const element = document.getElementById(`section-${sectionId}`);
        if (element) {
            element.scrollIntoView({ behavior: "smooth", block: "start" });
        }
    }
</script>

<svelte:head>
    <title>Settings - Admin</title>
</svelte:head>

<div class="w-full">
    <div class="mb-8">
        <h1 class="text-xl font-semibold mb-1">Settings</h1>
        <p class="text-muted-foreground text-sm">
            Configure your project settings
        </p>
    </div>

    {#if loading}
        <div class="flex-1 flex items-center justify-center py-16">
            <div class="flex items-center gap-2 text-sm">
                <Loader2 class="h-4 w-4 animate-spin" />
                <span class="text-muted-foreground">Loading settings...</span>
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
                        <h3 class="text-base font-semibold">
                            Website Information
                        </h3>
                        <p class="text-sm text-muted-foreground mt-1.5">
                            Configure your website name and URL. These will be
                            used in emails, browser tabs, and throughout the
                            application.
                        </p>
                    </div>

                    <div class="space-y-4">
                        <div>
                            <label
                                for="title"
                                class="text-sm font-medium mb-2 block"
                            >
                                Website Name
                            </label>
                            <Input
                                id="title"
                                type="text"
                                bind:value={title}
                                placeholder="My Awesome Project"
                                class="max-w-md"
                            />
                            <p class="text-xs text-muted-foreground mt-1">
                                This name will be used in email notifications
                                and as the main title of your project.
                            </p>
                        </div>

                        <div>
                            <label
                                for="website-url"
                                class="text-sm font-medium mb-2 block"
                            >
                                Website URL
                            </label>
                            <Input
                                id="website-url"
                                type="url"
                                bind:value={websiteUrl}
                                placeholder="https://example.com"
                                class="max-w-md"
                            />
                            {#if !websiteUrlValid}
                                <p class="text-xs text-destructive mt-1">
                                    Please enter a valid URL
                                </p>
                            {:else}
                                <p class="text-xs text-muted-foreground mt-1">
                                    The main website URL that will be included
                                    in email communications and links.
                                </p>
                            {/if}
                        </div>
                    </div>
                </div>

                <ImageUploadModal
                    bind:isOpen={imageUploadModalOpen}
                    on:imageSelected={handleImageSelected}
                />

                <!-- Favicon Section -->
                <div id="section-favicon" class="scroll-mt-6">
                    <div class="mb-6">
                        <h3 class="text-base font-semibold">Favicon</h3>
                        <p class="text-sm text-muted-foreground mt-1.5">
                            Upload a favicon for your website. This small icon
                            appears in browser tabs and bookmarks.
                        </p>
                    </div>

                    <div class="space-y-4">
                        <div class="flex items-center justify-between mb-3">
                            <span class="text-sm font-medium">Favicon</span>
                            <Button
                                type="button"
                                variant="outline"
                                size="sm"
                                on:click={openImageUpload}
                            >
                                <Upload class="h-4 w-4 mr-2" />
                                Upload
                            </Button>
                        </div>

                        {#if faviconUrl && faviconUrlValid}
                            <div
                                class="border border-border rounded-lg p-4 bg-muted/30"
                            >
                                <div class="flex items-center gap-3">
                                    <img
                                        src={faviconUrl}
                                        alt="Favicon"
                                        class="h-8 w-8"
                                        on:error={() => (faviconUrl = "")}
                                    />
                                    <div>
                                        <p class="text-sm font-medium">
                                            Favicon uploaded
                                        </p>
                                        <p
                                            class="text-xs text-muted-foreground truncate max-w-xs"
                                        >
                                            {faviconUrl}
                                        </p>
                                    </div>
                                </div>
                                <Button
                                    type="button"
                                    variant="ghost"
                                    size="sm"
                                    class="mt-2"
                                    on:click={() => (faviconUrl = "")}
                                >
                                    Remove
                                </Button>
                            </div>
                        {:else}
                            <div
                                class="border-2 border-dashed border-border rounded-lg p-8 text-center bg-muted/30"
                            >
                                <svg
                                    class="h-8 w-8 text-muted-foreground mx-auto mb-2"
                                    fill="none"
                                    stroke="currentColor"
                                    viewBox="0 0 24 24"
                                >
                                    <path
                                        stroke-linecap="round"
                                        stroke-linejoin="round"
                                        stroke-width="2"
                                        d="M7 21a4 4 0 01-4-4V5a2 2 0 012-2h4a2 2 0 012 2v12a4 4 0 01-4 4zm0 0h12a2 2 0 002-2v-4a2 2 0 00-2-2h-2.343M11 7.343l1.657-1.657a2 2 0 012.828 0l2.829 2.829a2 2 0 010 2.828l-8.486 8.485M7 17h.01"
                                    />
                                </svg>
                                <p class="text-sm text-muted-foreground">
                                    No favicon uploaded
                                </p>
                                <p class="text-xs text-muted-foreground">
                                    Recommended size: 32x32px or 64x64px
                                </p>
                            </div>
                        {/if}
                    </div>
                </div>

                <!-- Save Button -->
                <div class="flex justify-end pt-6 mt-6 border-t">
                    <Button on:click={handleSave} disabled={saving}>
                        {#if saving}
                            <Loader2 class="h-4 w-4 animate-spin mr-2" />
                            Saving...
                        {:else}
                            <Save class="h-4 w-4 mr-2" />
                            Save Changes
                        {/if}
                    </Button>
                </div>
            </div>
        </div>
    {/if}
</div>
