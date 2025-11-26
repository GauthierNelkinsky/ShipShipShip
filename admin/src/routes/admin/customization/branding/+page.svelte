<script lang="ts">
    import { onMount } from "svelte";
    import { goto } from "$app/navigation";
    import { api } from "$lib/api";
    import { authStore } from "$lib/stores/auth";
    import type {
        ProjectSettings,
        UpdateSettingsRequest,
        FooterLink,
    } from "$lib/types";
    import {
        Save,
        Upload,
        Palette,
        Type,
        Image,
        Link,
        Plus,
        Edit,
        Trash2,
        ExternalLink,
    } from "lucide-svelte";
    import { Button, Card, Input } from "$lib/components/ui";
    import ImageUploadModal from "$lib/components/ImageUploadModal.svelte";
    import { toast } from "svelte-sonner";
    import * as m from "$lib/paraglide/messages";
    import { localizeUrl } from "$lib/paraglide/runtime";

    let loading = true;
    let saving = false;
    let settings: ProjectSettings | null = null;

    // Form data
    let title = "";
    let logoUrl = "";
    let darkLogoUrl = "";
    let faviconUrl = "";
    let websiteUrl = "";
    let primaryColor = "#3b82f6";

    // Footer links state
    let footerLinks: FooterLink[] = [];
    let footerLinksLoading = false;
    let footerLinksSaving = false;
    let editingLink: FooterLink | null = null;
    let showAddLinkFormForColumn: "left" | "middle" | "right" | null = null;
    let newLinkName = "";
    let newLinkUrl = "";
    let newLinkOpenInNewWindow = false;

    // Image upload state
    let imageUploadModalOpen = false;
    let currentUploadTarget: "logo" | "darkLogo" | "favicon" | null = null;

    // Color presets
    const colorPresets = [
        { name: "Blue", value: "#3b82f6" },
        { name: "Green", value: "#10b981" },
        { name: "Purple", value: "#8b5cf6" },
        { name: "Red", value: "#ef4444" },
        { name: "Orange", value: "#f97316" },
        { name: "Pink", value: "#ec4899" },
        { name: "Indigo", value: "#6366f1" },
        { name: "Teal", value: "#14b8a6" },
    ];

    onMount(async () => {
        // Wait for authentication to be initialized before loading settings
        const unsubscribe = authStore.subscribe(async (auth) => {
            if (auth.initialized && auth.isAuthenticated) {
                await loadSettings();
                await loadFooterLinks();
                unsubscribe();
            } else if (auth.initialized && !auth.isAuthenticated) {
                // User is not authenticated, redirect to login
                goto(localizeUrl("/login").toString());
                unsubscribe();
            }
        });
    });

    async function loadFooterLinks() {
        try {
            footerLinksLoading = true;
            const response = await api.getFooterLinks();
            footerLinks = response.links || [];
        } catch (err) {
            console.error("Failed to load footer links:", err);
            toast.error("Failed to load footer links");
        } finally {
            footerLinksLoading = false;
        }
    }

    function getLinksForColumn(column: string) {
        return footerLinks.filter((link) => link.column === column);
    }

    async function addFooterLink(column: "left" | "middle" | "right") {
        if (!newLinkName.trim() || !newLinkUrl.trim()) return;

        try {
            footerLinksSaving = true;
            await api.createFooterLink({
                name: newLinkName.trim(),
                url: newLinkUrl.trim(),
                column: column,
                open_in_new_window: newLinkOpenInNewWindow,
            });

            // Reload the footer links to ensure consistency with backend
            await loadFooterLinks();

            // Reset form
            newLinkName = "";
            newLinkUrl = "";
            newLinkOpenInNewWindow = false;
            showAddLinkFormForColumn = null;

            toast.success(m.branding_footer_link_added(), {
                description: m.branding_footer_link_added_description({
                    linkName: newLinkName.trim(),
                }),
            });
        } catch (err) {
            console.error("Failed to add footer link:", err);
            const errorMessage =
                err instanceof Error ? err.message : m.branding_unknown_error();
            toast.error(m.branding_footer_link_add_failed(), {
                description: errorMessage,
            });
        } finally {
            footerLinksSaving = false;
        }
    }

    function startEditingLink(link: FooterLink) {
        editingLink = { ...link };
    }

    function cancelEditingLink() {
        editingLink = null;
    }

    async function updateFooterLink(updatedLink: FooterLink) {
        if (!updatedLink.name.trim() || !updatedLink.url.trim()) return;

        try {
            footerLinksSaving = true;
            await api.updateFooterLink(updatedLink.id, {
                name: updatedLink.name.trim(),
                url: updatedLink.url.trim(),
                column: updatedLink.column,
                open_in_new_window: updatedLink.open_in_new_window,
            });

            // Reload the footer links to ensure consistency with backend
            await loadFooterLinks();

            editingLink = null;
            toast.success(m.branding_footer_link_updated(), {
                description: m.branding_footer_link_updated_description({
                    linkName: updatedLink.name,
                }),
            });
        } catch (err) {
            console.error("Failed to update footer link:", err);
            const errorMessage =
                err instanceof Error ? err.message : m.branding_unknown_error();
            toast.error(m.branding_footer_link_update_failed(), {
                description: errorMessage,
            });
        } finally {
            footerLinksSaving = false;
        }
    }

    async function deleteFooterLink(linkId: number) {
        if (!confirm(m.branding_delete_footer_link_confirm())) return;

        try {
            footerLinksSaving = true;
            await api.deleteFooterLink(linkId);

            // Reload the footer links to ensure consistency with backend
            await loadFooterLinks();

            toast.success(m.branding_footer_link_deleted(), {
                description: m.branding_footer_link_deleted_description(),
            });
        } catch (err) {
            console.error("Failed to delete footer link:", err);
            const errorMessage =
                err instanceof Error ? err.message : m.branding_unknown_error();
            toast.error(m.branding_footer_link_delete_failed(), {
                description: errorMessage,
            });
        } finally {
            footerLinksSaving = false;
        }
    }

    function handleShowAddForm(column: string) {
        showAddLinkFormForColumn = column as "left" | "middle" | "right";
    }

    async function loadSettings() {
        try {
            loading = true;
            settings = await api.getSettings();

            // Populate form with current settings
            title = settings.title;
            logoUrl = settings.logo_url;
            darkLogoUrl = settings.dark_logo_url;
            faviconUrl = settings.favicon_url;
            websiteUrl = settings.website_url;
            primaryColor = settings.primary_color;
        } catch (err) {
            const errorMessage =
                err instanceof Error ? err.message : "Failed to load settings";
            console.error("Failed to load settings:", err);
            toast.error(m.branding_settings_load_failed());
        } finally {
            loading = false;
        }
    }

    async function handleSave() {
        if (!title.trim()) {
            toast.error(m.branding_title_required());
            return;
        }

        // Validate and normalize color format
        const colorRegex = /^#[0-9A-Fa-f]{6}$/;
        if (!colorRegex.test(primaryColor)) {
            toast.error(m.branding_invalid_color());
            return;
        }

        // Normalize to lowercase
        primaryColor = primaryColor.toLowerCase();

        saving = true;

        try {
            const updateData: UpdateSettingsRequest = {
                title: title.trim(),
                logo_url: logoUrl.trim(),
                dark_logo_url: darkLogoUrl.trim(),
                favicon_url: faviconUrl.trim(),
                website_url: websiteUrl.trim(),
                primary_color: primaryColor,
            };

            settings = await api.updateSettings(updateData);
            toast.success(m.branding_settings_saved(), {
                description: m.branding_settings_saved_description(),
            });

            // Update CSS custom properties for immediate preview
            updateCSSVariables();
        } catch (err) {
            const errorMessage =
                err instanceof Error
                    ? err.message
                    : m.branding_settings_save_failed();
            toast.error(m.branding_settings_save_failed(), {
                description: errorMessage,
            });
        } finally {
            saving = false;
        }
    }

    function updateCSSVariables() {
        // Convert hex to HSL for CSS custom properties
        const hsl = hexToHsl(primaryColor);
        document.documentElement.style.setProperty(
            "--primary",
            `${hsl.h} ${hsl.s}% ${hsl.l}%`,
        );
    }

    function hexToHsl(hex: string) {
        const r = parseInt(hex.slice(1, 3), 16) / 255;
        const g = parseInt(hex.slice(3, 5), 16) / 255;
        const b = parseInt(hex.slice(5, 7), 16) / 255;

        const max = Math.max(r, g, b);
        const min = Math.min(r, g, b);
        let h,
            s,
            l = (max + min) / 2;

        if (max === min) {
            h = s = 0;
        } else {
            const d = max - min;
            s = l > 0.5 ? d / (2 - max - min) : d / (max + min);
            switch (max) {
                case r:
                    h = (g - b) / d + (g < b ? 6 : 0);
                    break;
                case g:
                    h = (b - r) / d + 2;
                    break;
                case b:
                    h = (r - g) / d + 4;
                    break;
                default:
                    h = 0;
            }
            h /= 6;
        }

        return {
            h: Math.round(h * 360),
            s: Math.round(s * 100),
            l: Math.round(l * 100),
        };
    }

    function selectColorPreset(color: string) {
        primaryColor = color;
        updateCSSVariables();
    }

    function validateUrl(url: string): boolean {
        if (!url) return true; // Empty is valid
        // Accept relative paths (starting with /) or full URLs
        if (url.startsWith("/")) return true;
        try {
            new URL(url);
            return true;
        } catch {
            return false;
        }
    }

    $: logoUrlValid = validateUrl(logoUrl);
    $: darkLogoUrlValid = validateUrl(darkLogoUrl);
    $: faviconUrlValid = validateUrl(faviconUrl);

    function openImageUpload(target: "logo" | "darkLogo" | "favicon") {
        currentUploadTarget = target;
        imageUploadModalOpen = true;
    }

    function handleImageSelected(event: CustomEvent) {
        if (currentUploadTarget === "logo") {
            logoUrl = event.detail.url;
        } else if (currentUploadTarget === "darkLogo") {
            darkLogoUrl = event.detail.url;
        } else if (currentUploadTarget === "favicon") {
            faviconUrl = event.detail.url;
        }
        currentUploadTarget = null;
        imageUploadModalOpen = false;
    }

    $: websiteUrlValid = validateUrl(websiteUrl);
</script>

<svelte:head>
    <title>{m.branding_page_title()}</title>
</svelte:head>

<div class="max-w-4xl mx-auto">
    <div class="mb-8">
        <h1 class="text-xl font-semibold mb-1">{m.branding_heading()}</h1>
        <p class="text-muted-foreground text-sm">
            {m.branding_subheading()}
        </p>
    </div>

    {#if loading}
        <div class="flex items-center justify-center min-h-64">
            <div
                class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"
            ></div>
        </div>
    {:else}
        <form on:submit|preventDefault={handleSave} class="space-y-6">
            <!-- Project Title -->
            <Card class="p-6">
                <div class="flex items-center gap-4 mb-6">
                    <Type class="h-6 w-6 text-primary" />
                    <div>
                        <h2 class="text-lg font-semibold">
                            {m.branding_project_title()}
                        </h2>
                        <p class="text-sm text-muted-foreground">
                            {m.branding_project_title_description()}
                        </p>
                    </div>
                </div>

                <div class="space-y-4">
                    <div>
                        <label
                            for="title"
                            class="text-sm font-medium block mb-2"
                            >{m.branding_title_label()}</label
                        >
                        <Input
                            id="title"
                            type="text"
                            bind:value={title}
                            placeholder={m.branding_title_placeholder()}
                            disabled={saving}
                            required
                        />
                        <p class="text-xs text-muted-foreground mt-1">
                            {m.branding_title_help()}
                        </p>
                    </div>
                </div>
            </Card>

            <!-- Project Logo -->
            <Card class="p-6">
                <div class="flex items-center gap-4 mb-6">
                    <Image class="h-6 w-6 text-primary" />
                    <div>
                        <h2 class="text-lg font-semibold">
                            {m.branding_project_logo()}
                        </h2>
                        <p class="text-sm text-muted-foreground">
                            {m.branding_project_logo_description()}
                        </p>
                    </div>
                </div>

                <div class="space-y-4">
                    <!-- Light Theme Logo -->
                    <div>
                        <div class="flex items-center justify-between mb-3">
                            <span class="text-sm font-medium"
                                >{m.branding_light_logo()}</span
                            >
                            <Button
                                type="button"
                                variant="outline"
                                size="sm"
                                on:click={(e) => {
                                    e.preventDefault();
                                    e.stopPropagation();
                                    openImageUpload("logo");
                                }}
                                disabled={saving}
                                class="gap-1.5 text-xs h-8 px-2.5"
                            >
                                <Upload class="h-4 w-4" />
                                {logoUrl
                                    ? m.branding_change_logo()
                                    : m.branding_upload_logo()}
                            </Button>
                        </div>

                        {#if logoUrl && logoUrlValid}
                            <div
                                class="border border-border rounded-lg p-3 bg-muted/10 flex items-center justify-between"
                            >
                                <img
                                    src={logoUrl}
                                    alt={m.branding_logo_alt()}
                                    class="h-16 w-auto object-contain"
                                    on:error={() => (logoUrl = "")}
                                />
                                <Button
                                    type="button"
                                    variant="ghost"
                                    size="sm"
                                    on:click={() => (logoUrl = "")}
                                    disabled={saving}
                                    class="text-destructive hover:text-destructive"
                                >
                                    {m.branding_remove()}
                                </Button>
                            </div>
                        {:else}
                            <div
                                class="border border-dashed border-border rounded-lg p-6 text-center bg-muted/10"
                            >
                                <Image
                                    class="h-8 w-8 mx-auto text-muted-foreground mb-2"
                                />
                                <p class="text-sm text-muted-foreground">
                                    {m.branding_no_logo()}
                                </p>
                                <p class="text-xs text-muted-foreground">
                                    {m.branding_upload_logo_help()}
                                </p>
                            </div>
                        {/if}
                    </div>

                    <!-- Dark Theme Logo -->
                    <div>
                        <div class="flex items-center justify-between mb-3">
                            <span class="text-sm font-medium"
                                >{m.branding_dark_logo()}</span
                            >
                            <Button
                                type="button"
                                variant="outline"
                                size="sm"
                                on:click={(e) => {
                                    e.preventDefault();
                                    e.stopPropagation();
                                    openImageUpload("darkLogo");
                                }}
                                disabled={saving}
                                class="gap-1.5 text-xs h-8 px-2.5"
                            >
                                <Upload class="h-4 w-4" />
                                {darkLogoUrl
                                    ? m.branding_change_logo()
                                    : m.branding_upload_logo()}
                            </Button>
                        </div>

                        {#if darkLogoUrl && darkLogoUrlValid}
                            <div
                                class="border border-border rounded-lg p-3 bg-gray-900 flex items-center justify-between"
                            >
                                <img
                                    src={darkLogoUrl}
                                    alt={m.branding_dark_logo_alt()}
                                    class="h-16 w-auto object-contain"
                                    on:error={() => (darkLogoUrl = "")}
                                />
                                <Button
                                    type="button"
                                    variant="ghost"
                                    size="sm"
                                    on:click={() => (darkLogoUrl = "")}
                                    disabled={saving}
                                    class="text-destructive hover:text-destructive"
                                >
                                    {m.branding_remove()}
                                </Button>
                            </div>
                        {:else}
                            <div
                                class="border border-dashed border-border rounded-lg p-6 text-center bg-gray-900"
                            >
                                <Image
                                    class="h-8 w-8 mx-auto text-gray-400 mb-2"
                                />
                                <p class="text-sm text-gray-300">
                                    {m.branding_no_dark_logo()}
                                </p>
                                <p class="text-xs text-gray-400">
                                    {m.branding_dark_logo_help()}
                                </p>
                            </div>
                        {/if}
                    </div>
                </div>
            </Card>

            <!-- Favicon -->
            <Card class="p-6">
                <div class="flex items-center gap-4 mb-6">
                    <svg
                        class="h-6 w-6 text-primary"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                        xmlns="http://www.w3.org/2000/svg"
                    >
                        <path
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            stroke-width="2"
                            d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z"
                        />
                    </svg>
                    <div>
                        <h2 class="text-lg font-semibold">
                            {m.branding_favicon()}
                        </h2>
                        <p class="text-sm text-muted-foreground">
                            {m.branding_favicon_description()}
                        </p>
                    </div>
                </div>

                <div class="space-y-4">
                    <div>
                        <div class="flex items-center justify-between mb-3">
                            <span class="text-sm font-medium"
                                >{m.branding_favicon()}</span
                            >
                            <Button
                                type="button"
                                variant="outline"
                                size="sm"
                                on:click={(e) => {
                                    e.preventDefault();
                                    e.stopPropagation();
                                    openImageUpload("favicon");
                                }}
                                disabled={saving}
                                class="gap-1.5 text-xs h-8 px-2.5"
                            >
                                <Upload class="h-4 w-4" />
                                {faviconUrl
                                    ? m.branding_change_favicon()
                                    : m.branding_upload_favicon()}
                            </Button>
                        </div>

                        {#if faviconUrl && faviconUrlValid}
                            <div
                                class="border border-border rounded-lg p-3 bg-muted/10 flex items-center justify-between"
                            >
                                <div class="flex items-center gap-3">
                                    <img
                                        src={faviconUrl}
                                        alt={m.branding_favicon_alt()}
                                        class="h-8 w-8 object-contain"
                                        on:error={() => (faviconUrl = "")}
                                    />
                                    <div>
                                        <p class="text-sm font-medium">
                                            {m.branding_favicon_uploaded()}
                                        </p>
                                        <p
                                            class="text-xs text-muted-foreground"
                                        >
                                            {m.branding_favicon_size()}
                                        </p>
                                    </div>
                                </div>
                                <Button
                                    type="button"
                                    variant="ghost"
                                    size="sm"
                                    on:click={() => (faviconUrl = "")}
                                    disabled={saving}
                                    class="text-destructive hover:text-destructive"
                                >
                                    {m.branding_remove()}
                                </Button>
                            </div>
                        {:else}
                            <div
                                class="border border-dashed border-border rounded-lg p-6 text-center bg-muted/10"
                            >
                                <svg
                                    class="h-8 w-8 mx-auto text-muted-foreground mb-2"
                                    fill="none"
                                    stroke="currentColor"
                                    viewBox="0 0 24 24"
                                >
                                    <path
                                        stroke-linecap="round"
                                        stroke-linejoin="round"
                                        stroke-width="2"
                                        d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z"
                                    />
                                </svg>
                                <p class="text-sm text-muted-foreground">
                                    {m.branding_no_favicon()}
                                </p>
                                <p class="text-xs text-muted-foreground">
                                    {m.branding_favicon_help()}
                                </p>
                            </div>
                        {/if}
                    </div>
                </div>
            </Card>

            <!-- Website URL -->
            <Card class="p-6">
                <div class="flex items-center gap-4 mb-6">
                    <svg
                        class="h-6 w-6 text-primary"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                        xmlns="http://www.w3.org/2000/svg"
                    >
                        <path
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            stroke-width="2"
                            d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1"
                        />
                    </svg>
                    <div>
                        <h2 class="text-lg font-semibold">
                            {m.branding_website_url()}
                        </h2>
                        <p class="text-sm text-muted-foreground">
                            {m.branding_website_url_description()}
                        </p>
                    </div>
                </div>

                <div class="space-y-4">
                    <div>
                        <label
                            for="websiteUrl"
                            class="text-sm font-medium block mb-2"
                            >{m.branding_website_url()}</label
                        >
                        <Input
                            id="websiteUrl"
                            type="url"
                            bind:value={websiteUrl}
                            placeholder="https://yourwebsite.com"
                            class={!websiteUrlValid ? "border-destructive" : ""}
                            disabled={saving}
                        />
                        {#if !websiteUrlValid}
                            <p class="text-xs text-destructive mt-1">
                                {m.branding_invalid_url()}
                            </p>
                        {:else}
                            <p class="text-xs text-muted-foreground mt-1">
                                {m.branding_website_url_help()}
                            </p>
                        {/if}
                    </div>
                </div>
            </Card>

            <!-- Primary Color -->
            <Card class="p-6">
                <div class="flex items-center gap-4 mb-6">
                    <Palette class="h-6 w-6 text-primary" />
                    <div>
                        <h2 class="text-lg font-semibold">
                            {m.branding_primary_color()}
                        </h2>
                        <p class="text-sm text-muted-foreground">
                            {m.branding_primary_color_description()}
                        </p>
                    </div>
                </div>

                <div class="space-y-4">
                    <!-- Color Input -->
                    <div>
                        <label
                            for="primaryColor"
                            class="text-sm font-medium block mb-2"
                            >{m.branding_color_label()}</label
                        >
                        <div class="flex gap-3">
                            <input
                                id="primaryColor"
                                type="color"
                                bind:value={primaryColor}
                                class="w-12 h-10 rounded border border-input cursor-pointer"
                                disabled={saving}
                                on:input={() => {
                                    updateCSSVariables();
                                    // Ensure proper format
                                    if (
                                        primaryColor &&
                                        !primaryColor.startsWith("#")
                                    ) {
                                        primaryColor = "#" + primaryColor;
                                    }
                                }}
                            />
                            <Input
                                type="text"
                                bind:value={primaryColor}
                                placeholder="#3b82f6"
                                class="flex-1"
                                disabled={saving}
                                maxlength={7}
                                on:input={() => {
                                    // Auto-add # if missing
                                    if (
                                        primaryColor &&
                                        !primaryColor.startsWith("#") &&
                                        primaryColor.length > 0
                                    ) {
                                        primaryColor =
                                            "#" + primaryColor.replace("#", "");
                                    }
                                    // Ensure uppercase hex
                                    if (primaryColor.length === 7) {
                                        primaryColor =
                                            primaryColor.toLowerCase();
                                    }
                                    updateCSSVariables();
                                }}
                            />
                        </div>
                        <p class="text-xs text-muted-foreground mt-1">
                            {m.branding_color_help()}
                        </p>
                    </div>

                    <!-- Color Presets -->
                    <div>
                        <p class="text-sm font-medium mb-3">
                            {m.branding_quick_presets()}
                        </p>
                        <div class="flex flex-wrap gap-3">
                            {#each colorPresets as preset}
                                <button
                                    type="button"
                                    on:click={() =>
                                        selectColorPreset(preset.value)}
                                    class="group relative flex items-center gap-2 px-3 py-2 rounded-lg border border-border bg-card transition-all hover:bg-accent hover:border-primary/50 focus:outline-none focus:ring-2 focus:ring-primary focus:ring-offset-1
										{primaryColor === preset.value ? 'border-primary bg-primary/5' : ''}"
                                    title={preset.name}
                                    disabled={saving}
                                >
                                    <div
                                        class="w-4 h-4 rounded-full border border-border/50 shadow-sm
											{primaryColor === preset.value ? 'ring-2 ring-primary ring-offset-1' : ''}"
                                        style="background-color: {preset.value}"
                                    ></div>
                                    <span
                                        class="text-sm font-medium text-muted-foreground group-hover:text-foreground
										{primaryColor === preset.value ? 'text-foreground' : ''}"
                                    >
                                        {preset.name}
                                    </span>
                                </button>
                            {/each}
                        </div>
                    </div>

                    <!-- Color Preview -->
                    <div
                        class="border border-border rounded-lg p-3 bg-muted/10"
                    >
                        <p class="text-sm text-muted-foreground mb-3">
                            {m.branding_component_preview()}
                        </p>
                        <div class="space-y-3">
                            <Button
                                type="button"
                                style="background-color: {primaryColor}; border-color: {primaryColor}"
                            >
                                {m.branding_primary_button()}
                            </Button>
                            <div class="flex items-center gap-2">
                                <span
                                    class="inline-flex items-center rounded-full px-3 py-1 text-sm font-medium"
                                    style="background-color: {primaryColor}20; color: {primaryColor}"
                                >
                                    {m.branding_tag_example()}
                                </span>
                                <span
                                    class="text-sm"
                                    style="color: {primaryColor}"
                                >
                                    {m.branding_accent_text()}
                                </span>
                            </div>
                        </div>
                    </div>
                </div>
            </Card>

            <!-- Footer Links Management -->
            <Card class="p-6">
                <div class="flex items-center gap-4 mb-6">
                    <Link class="h-6 w-6 text-primary" />
                    <div>
                        <h2 class="text-lg font-semibold">
                            {m.branding_footer_links()}
                        </h2>
                        <p class="text-sm text-muted-foreground">
                            {m.branding_footer_links_description()}
                        </p>
                    </div>
                </div>

                <div class="space-y-6">
                    <!-- Footer Links by Column -->
                    {#if footerLinksLoading}
                        <div class="flex items-center justify-center py-8">
                            <div
                                class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"
                            ></div>
                        </div>
                    {:else}
                        <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
                            {#each ["left", "middle", "right"] as column}
                                <div class="space-y-3">
                                    <div
                                        class="flex items-center justify-between"
                                    >
                                        <h3
                                            class="font-medium text-sm uppercase tracking-wide text-muted-foreground"
                                        >
                                            {column === "left"
                                                ? m.branding_column_left()
                                                : column === "middle"
                                                  ? m.branding_column_middle()
                                                  : m.branding_column_right()}
                                        </h3>
                                        <Button
                                            variant="ghost"
                                            size="sm"
                                            on:click={() =>
                                                handleShowAddForm(column)}
                                            class="h-8 w-8 p-0"
                                            disabled={footerLinksSaving}
                                        >
                                            <Plus class="h-3 w-3" />
                                        </Button>
                                    </div>
                                    <div
                                        class="space-y-2 min-h-[100px] border border-dashed border-border rounded-lg p-3"
                                    >
                                        <!-- Add Form for this column -->
                                        {#if showAddLinkFormForColumn === column}
                                            <Card
                                                class="p-3 border-2 border-dashed border-primary/50 bg-primary/5"
                                            >
                                                <div class="space-y-3">
                                                    <div
                                                        class="flex items-center justify-between"
                                                    >
                                                        <h4
                                                            class="text-sm font-medium"
                                                        >
                                                            {m.branding_add_link_to(
                                                                {
                                                                    column:
                                                                        column
                                                                            .charAt(
                                                                                0,
                                                                            )
                                                                            .toUpperCase() +
                                                                        column.slice(
                                                                            1,
                                                                        ),
                                                                },
                                                            )}
                                                        </h4>
                                                        <Button
                                                            variant="ghost"
                                                            size="sm"
                                                            on:click={() => {
                                                                showAddLinkFormForColumn =
                                                                    null;
                                                                newLinkName =
                                                                    "";
                                                                newLinkUrl = "";
                                                                newLinkOpenInNewWindow = false;
                                                            }}
                                                            class="h-6 w-6 p-0"
                                                        >
                                                            ×
                                                        </Button>
                                                    </div>
                                                    <Input
                                                        bind:value={newLinkName}
                                                        placeholder={m.branding_link_name_placeholder()}
                                                        class="text-sm"
                                                    />
                                                    <Input
                                                        bind:value={newLinkUrl}
                                                        placeholder="https://example.com"
                                                        type="url"
                                                        class="text-sm"
                                                    />
                                                    <div
                                                        class="flex items-center space-x-2"
                                                    >
                                                        <input
                                                            type="checkbox"
                                                            id="newLinkOpenInNewWindow-{column}"
                                                            bind:checked={
                                                                newLinkOpenInNewWindow
                                                            }
                                                            class="rounded border-border"
                                                        />
                                                        <label
                                                            for="newLinkOpenInNewWindow-{column}"
                                                            class="text-xs text-muted-foreground cursor-pointer"
                                                        >
                                                            {m.branding_open_in_new_window()}
                                                        </label>
                                                    </div>
                                                    <Button
                                                        size="sm"
                                                        on:click={() =>
                                                            addFooterLink(
                                                                column,
                                                            )}
                                                        disabled={!newLinkName.trim() ||
                                                            !newLinkUrl.trim() ||
                                                            footerLinksSaving}
                                                        class="w-full"
                                                    >
                                                        {#if footerLinksSaving}
                                                            <div
                                                                class="h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent mr-2"
                                                            ></div>
                                                        {/if}
                                                        {m.branding_add_link()}
                                                    </Button>
                                                </div>
                                            </Card>
                                        {/if}
                                        {#each getLinksForColumn(column) as link (link.id)}
                                            <div
                                                class="group border border-border rounded-lg p-3 bg-card hover:bg-accent/50 transition-colors"
                                            >
                                                {#if editingLink && editingLink.id === link.id}
                                                    <!-- Edit Mode -->
                                                    <div class="space-y-3">
                                                        <Input
                                                            bind:value={
                                                                editingLink.name
                                                            }
                                                            placeholder={m.branding_link_name_placeholder()}
                                                            class="text-sm"
                                                        />
                                                        <Input
                                                            bind:value={
                                                                editingLink.url
                                                            }
                                                            placeholder="URL"
                                                            type="url"
                                                            class="text-sm"
                                                        />
                                                        <select
                                                            bind:value={
                                                                editingLink.column
                                                            }
                                                            class="w-full p-2 text-sm border border-input rounded-md bg-background"
                                                        >
                                                            <option value="left"
                                                                >{m.branding_column_left()}</option
                                                            >
                                                            <option
                                                                value="middle"
                                                                >{m.branding_column_middle()}</option
                                                            >
                                                            <option
                                                                value="right"
                                                                >{m.branding_column_right()}</option
                                                            >
                                                        </select>
                                                        <div
                                                            class="flex items-center space-x-2"
                                                        >
                                                            <input
                                                                type="checkbox"
                                                                id="editLinkOpenInNewWindow-{editingLink.id}"
                                                                bind:checked={
                                                                    editingLink.open_in_new_window
                                                                }
                                                                class="rounded border-border"
                                                            />
                                                            <label
                                                                for="editLinkOpenInNewWindow-{editingLink.id}"
                                                                class="text-xs text-muted-foreground cursor-pointer"
                                                            >
                                                                {m.branding_open_in_new_window()}
                                                            </label>
                                                        </div>
                                                        <div class="flex gap-2">
                                                            <Button
                                                                size="sm"
                                                                on:click={() =>
                                                                    editingLink &&
                                                                    updateFooterLink(
                                                                        editingLink,
                                                                    )}
                                                                disabled={!editingLink?.name.trim() ||
                                                                    !editingLink?.url.trim() ||
                                                                    footerLinksSaving}
                                                            >
                                                                {#if footerLinksSaving}
                                                                    <div
                                                                        class="h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent mr-2"
                                                                    ></div>
                                                                {/if}
                                                                {m.branding_save()}
                                                            </Button>
                                                            <Button
                                                                size="sm"
                                                                variant="ghost"
                                                                on:click={cancelEditingLink}
                                                            >
                                                                {m.branding_cancel()}
                                                            </Button>
                                                        </div>
                                                    </div>
                                                {:else}
                                                    <!-- Display Mode -->
                                                    <div
                                                        class="flex items-start justify-between"
                                                    >
                                                        <div
                                                            class="flex-1 min-w-0"
                                                        >
                                                            <div
                                                                class="flex items-center gap-1"
                                                            >
                                                                <p
                                                                    class="font-medium text-sm truncate"
                                                                >
                                                                    {link.name}
                                                                </p>
                                                                {#if link.open_in_new_window}
                                                                    <ExternalLink
                                                                        class="h-3 w-3 text-muted-foreground flex-shrink-0"
                                                                    />
                                                                {/if}
                                                            </div>
                                                            <p
                                                                class="text-xs text-muted-foreground truncate"
                                                            >
                                                                {link.url}
                                                            </p>
                                                        </div>
                                                        <div
                                                            class="flex gap-1 opacity-0 group-hover:opacity-100 transition-opacity"
                                                        >
                                                            <Button
                                                                size="sm"
                                                                variant="ghost"
                                                                class="h-8 w-8 p-0"
                                                                on:click={() =>
                                                                    startEditingLink(
                                                                        link,
                                                                    )}
                                                            >
                                                                <Edit
                                                                    class="h-3 w-3"
                                                                />
                                                            </Button>
                                                            <Button
                                                                size="sm"
                                                                variant="ghost"
                                                                class="h-8 w-8 p-0 text-destructive hover:text-destructive"
                                                                disabled={footerLinksSaving}
                                                                on:click={() =>
                                                                    deleteFooterLink(
                                                                        link.id,
                                                                    )}
                                                            >
                                                                <Trash2
                                                                    class="h-3 w-3"
                                                                />
                                                            </Button>
                                                        </div>
                                                    </div>
                                                {/if}
                                            </div>
                                        {:else}
                                            <div
                                                class="text-center py-8 text-muted-foreground text-sm"
                                            >
                                                {m.branding_no_links_column()}
                                            </div>
                                        {/each}
                                    </div>
                                </div>
                            {/each}
                        </div>
                    {/if}
                </div>
            </Card>

            <div class="flex justify-end gap-2 pt-4 border-t border-border/50">
                <Button
                    type="submit"
                    size="sm"
                    class="flex items-center gap-1.5"
                    disabled={saving ||
                        !title.trim() ||
                        !logoUrlValid ||
                        !darkLogoUrlValid ||
                        !websiteUrlValid}
                >
                    {#if saving}
                        <div
                            class="animate-spin rounded-full h-4 w-4 border-b-2 border-white"
                        ></div>
                        {m.branding_saving()}
                    {:else}
                        <Save class="h-4 w-4" />
                        {m.branding_save_settings()}
                    {/if}
                </Button>
            </div>
        </form>

        <!-- Single Image Upload Modal -->
        <ImageUploadModal
            bind:isOpen={imageUploadModalOpen}
            on:imageSelected={handleImageSelected}
        />
    {/if}
</div>
