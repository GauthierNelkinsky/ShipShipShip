<script lang="ts">
    import { onMount } from "svelte";
    import { api } from "$lib/api";
    import type {
        UpdateMailSettingsRequest,
        EventStatus,
        StatusDefinition,
    } from "$lib/types";
    import { Button, Input } from "$lib/components/ui";
    import {
        Save,
        Send,
        Eye,
        EyeOff,
        Mail,
        ChevronDown,
        Zap,
        Loader2,
        Check,
    } from "lucide-svelte";
    import { fly } from "svelte/transition";
    import { cn } from "$lib/utils";
    import { toast } from "svelte-sonner";
    import * as m from "$lib/paraglide/messages";

    interface SettingSection {
        id: string;
        title: string;
        description: string;
    }

    let loading = true;

    // Newsletter automation settings
    let automationEnabled = false;
    let automationTriggerStatuses: EventStatus[] = [];
    let automationSaving = false;
    let statuses: StatusDefinition[] = [];

    // Mail settings
    let mailSaving = false;
    let mailTesting = false;
    let smtpHost = "";
    let smtpPort = "587";
    let smtpUsername = "";
    let smtpPassword = "";
    let smtpEncryption = "tls";
    let fromEmail = "";
    let fromName = "";
    let showPassword = false;
    let testEmail = "";

    // UI state

    // Multi-select state
    let statusSelectOpen = false;
    let statusSearchTerm = "";
    let statusButtonElement: HTMLButtonElement;
    let statusDropdownElement: HTMLDivElement;

    $: filteredStatuses = statuses.filter((status) =>
        status.display_name
            .toLowerCase()
            .includes(statusSearchTerm.toLowerCase()),
    );

    function toggleStatusSelect() {
        statusSelectOpen = !statusSelectOpen;
        if (!statusSelectOpen) {
            statusSearchTerm = "";
        }
    }

    function toggleStatus(statusName: EventStatus) {
        if (automationTriggerStatuses.includes(statusName)) {
            automationTriggerStatuses = automationTriggerStatuses.filter(
                (s) => s !== statusName,
            );
        } else {
            automationTriggerStatuses = [
                ...automationTriggerStatuses,
                statusName,
            ];
        }
    }

    function handleStatusClickOutside(event: MouseEvent) {
        const target = event.target as Element;
        if (
            statusButtonElement &&
            !statusButtonElement.contains(target) &&
            statusDropdownElement &&
            !statusDropdownElement.contains(target)
        ) {
            statusSelectOpen = false;
            statusSearchTerm = "";
        }
    }

    // Sidebar navigation
    let activeSection = "mail";
    let sidebarTop: number | null = null;
    let sidebarElement: HTMLElement;

    const sections: SettingSection[] = [
        {
            id: "mail",
            title: m.newsletter_settings_mail_settings(),
            description: m.newsletter_settings_mail_settings_description(),
        },
        {
            id: "automation",
            title: m.newsletter_settings_automation(),
            description: m.newsletter_settings_automation_description(),
        },
    ];

    const encryptionOptions = [
        { value: "none", label: "None" },
        { value: "tls", label: "TLS" },
        { value: "ssl", label: "SSL" },
    ];

    function handleScroll() {
        if (!sidebarElement) return;

        const scrollTop =
            window.pageYOffset || document.documentElement.scrollTop;
        const initialTop = sidebarElement.offsetTop || 0;

        if (scrollTop > initialTop - 24) {
            sidebarTop = 24;
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

    onMount(async () => {
        await loadData();

        if (sections.length > 0) {
            activeSection = sections[0].id;
        }

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
            handleScroll();

            return () => {
                window.removeEventListener("scroll", onScroll);
            };
        }, 100);
    });

    onMount(() => {
        // Add click outside handler for status select
        document.addEventListener("click", handleStatusClickOutside);
        return () => {
            document.removeEventListener("click", handleStatusClickOutside);
        };
    });

    async function loadData() {
        loading = true;

        try {
            await loadMailSettings();
            await loadStatuses();
            await loadNewsletterAutomationSettings();
            // Filter out any statuses that no longer exist
            cleanupAutomationStatuses();
        } catch (err) {
            console.error("Error loading data:", err);
            const errorMessage =
                err instanceof Error
                    ? err.message
                    : m.newsletter_settings_load_failed();
            toast.error(m.newsletter_settings_load_failed(), {
                description: errorMessage,
            });
        } finally {
            loading = false;
        }
    }

    async function loadStatuses() {
        try {
            const allStatuses = await api.getStatuses();
            statuses = allStatuses;
        } catch (err) {
            console.error("Failed to load statuses:", err);
            statuses = [];
        }
    }

    async function loadNewsletterAutomationSettings() {
        try {
            const settings = await api.getNewsletterAutomationSettings();
            automationEnabled = settings.enabled;
            automationTriggerStatuses = settings.trigger_statuses || [];
        } catch (err) {
            console.error(
                "Failed to load newsletter automation settings:",
                err,
            );
            automationEnabled = false;
            automationTriggerStatuses = [];
        }
    }

    function cleanupAutomationStatuses() {
        // Filter out statuses that no longer exist in the system
        const validStatusNames = new Set(statuses.map((s) => s.display_name));
        automationTriggerStatuses = automationTriggerStatuses.filter((status) =>
            validStatusNames.has(status),
        );
    }

    async function handleAutomationSave() {
        automationSaving = true;

        if (automationEnabled && automationTriggerStatuses.length === 0) {
            toast.error(m.newsletter_settings_automation_validation());
            automationSaving = false;
            return;
        }

        try {
            await api.updateNewsletterAutomationSettings({
                enabled: automationEnabled,
                trigger_statuses: automationEnabled
                    ? automationTriggerStatuses
                    : [],
            });
            toast.success(m.newsletter_settings_automation_saved(), {
                description:
                    m.newsletter_settings_automation_saved_description(),
            });
        } catch (err) {
            const errorMessage =
                err instanceof Error
                    ? err.message
                    : m.newsletter_settings_automation_save_failed();
            toast.error(m.newsletter_settings_automation_save_failed(), {
                description: errorMessage,
            });
        } finally {
            automationSaving = false;
        }
    }

    async function loadMailSettings() {
        try {
            const settings = await api.getMailSettings();
            if (settings) {
                smtpHost = settings.smtp_host || "";
                smtpPort = String(settings.smtp_port || 587);
                smtpUsername = settings.smtp_username || "";
                smtpPassword = settings.smtp_password || "";
                smtpEncryption = settings.smtp_encryption || "tls";
                fromEmail = settings.from_email || "";
                fromName = settings.from_name || "";
            }
        } catch {
            console.log("No mail settings found");
        }
    }

    async function handleMailSave() {
        if (!validateMailForm()) return;

        mailSaving = true;

        try {
            const settings: UpdateMailSettingsRequest = {
                smtp_host: smtpHost.trim(),
                smtp_port: parseInt(smtpPort),
                smtp_username: smtpUsername.trim(),
                smtp_password: smtpPassword,
                smtp_encryption: smtpEncryption,
                from_email: fromEmail.trim(),
                from_name: fromName.trim(),
            };

            await api.updateMailSettings(settings);
            toast.success(m.newsletter_settings_mail_saved(), {
                description: m.newsletter_settings_mail_saved_description(),
            });
        } catch (err) {
            const errorMessage =
                err instanceof Error
                    ? err.message
                    : m.newsletter_settings_mail_save_failed();
            toast.error(m.newsletter_settings_mail_save_failed(), {
                description: errorMessage,
            });
        } finally {
            mailSaving = false;
        }
    }

    async function handleMailTest() {
        if (!testEmail.trim()) {
            toast.error(m.newsletter_settings_test_email_required());
            return;
        }

        mailTesting = true;

        try {
            await api.testMailSettings(testEmail);
            toast.success(m.newsletter_settings_test_sent(), {
                description: m.newsletter_settings_test_sent_description({
                    email: testEmail,
                }),
            });
        } catch (err) {
            const errorMessage =
                err instanceof Error
                    ? err.message
                    : m.newsletter_settings_test_failed();
            toast.error(m.newsletter_settings_test_failed(), {
                description: errorMessage,
            });
        } finally {
            mailTesting = false;
        }
    }

    function validateMailForm() {
        if (!smtpHost) {
            toast.error(m.newsletter_settings_smtp_host_required());
            return false;
        }
        if (!smtpPort || isNaN(parseInt(smtpPort))) {
            toast.error(m.newsletter_settings_smtp_port_required());
            return false;
        }
        if (!smtpUsername) {
            toast.error(m.newsletter_settings_smtp_username_required());
            return false;
        }
        if (!smtpPassword) {
            toast.error(m.newsletter_settings_smtp_password_required());
            return false;
        }
        if (!fromEmail) {
            toast.error(m.newsletter_settings_from_email_required());
            return false;
        }
        return true;
    }

    function scrollToSection(sectionId: string) {
        const element = document.getElementById(`section-${sectionId}`);
        if (element) {
            element.scrollIntoView({ behavior: "smooth", block: "start" });
        }
    }
</script>

<svelte:head>
    <title>{m.newsletter_settings_page_title()}</title>
</svelte:head>

<div class="w-full">
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
                <!-- Mail Settings Section -->
                <div id="section-mail" class="scroll-mt-6">
                    <div class="mb-6">
                        <div class="flex items-center gap-3 mb-1.5">
                            <Mail class="h-5 w-5 text-primary" />
                            <h3 class="text-base font-semibold">
                                {m.newsletter_settings_smtp()}
                            </h3>
                        </div>
                        <p class="text-sm text-muted-foreground mt-1.5">
                            {m.newsletter_settings_smtp_description()}
                        </p>
                    </div>

                    <form
                        on:submit|preventDefault={handleMailSave}
                        class="space-y-6"
                    >
                        <div class="grid gap-4 md:grid-cols-3">
                            <div>
                                <label
                                    for="smtp-host"
                                    class="text-sm font-medium mb-2 block"
                                >
                                    {m.newsletter_settings_smtp_host()}
                                </label>
                                <Input
                                    id="smtp-host"
                                    type="text"
                                    bind:value={smtpHost}
                                    placeholder="smtp.example.com"
                                />
                            </div>
                            <div>
                                <label
                                    for="smtp-port"
                                    class="text-sm font-medium mb-2 block"
                                >
                                    {m.newsletter_settings_smtp_port()}
                                </label>
                                <Input
                                    id="smtp-port"
                                    type="text"
                                    bind:value={smtpPort}
                                    placeholder="587"
                                />
                            </div>
                            <div>
                                <label
                                    for="smtp-encryption"
                                    class="text-sm font-medium mb-2 block"
                                >
                                    {m.newsletter_settings_encryption()}
                                </label>
                                <select
                                    id="smtp-encryption"
                                    bind:value={smtpEncryption}
                                    class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                                >
                                    {#each encryptionOptions as option}
                                        <option value={option.value}>
                                            {option.label}
                                        </option>
                                    {/each}
                                </select>
                            </div>
                        </div>

                        <div class="grid gap-4 md:grid-cols-2">
                            <div>
                                <label
                                    for="smtp-username"
                                    class="text-sm font-medium mb-2 block"
                                >
                                    {m.newsletter_settings_username()}
                                </label>
                                <Input
                                    id="smtp-username"
                                    type="text"
                                    bind:value={smtpUsername}
                                    placeholder="username"
                                />
                            </div>
                            <div>
                                <label
                                    for="smtp-password"
                                    class="text-sm font-medium mb-2 block"
                                >
                                    {m.newsletter_settings_password()}
                                </label>
                                <div class="relative">
                                    <Input
                                        id="smtp-password"
                                        type={showPassword
                                            ? "text"
                                            : "password"}
                                        bind:value={smtpPassword}
                                        placeholder="••••••••"
                                        class="pr-10"
                                    />
                                    <button
                                        type="button"
                                        on:click={() =>
                                            (showPassword = !showPassword)}
                                        class="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                                    >
                                        {#if showPassword}
                                            <EyeOff class="h-4 w-4" />
                                        {:else}
                                            <Eye class="h-4 w-4" />
                                        {/if}
                                    </button>
                                </div>
                            </div>
                        </div>

                        <div class="grid gap-4 md:grid-cols-2">
                            <div>
                                <label
                                    for="from-email"
                                    class="text-sm font-medium mb-2 block"
                                >
                                    {m.newsletter_settings_from_email()}
                                </label>
                                <Input
                                    id="from-email"
                                    type="email"
                                    bind:value={fromEmail}
                                    placeholder="noreply@example.com"
                                />
                            </div>
                            <div>
                                <label
                                    for="from-name"
                                    class="text-sm font-medium mb-2 block"
                                >
                                    {m.newsletter_settings_from_name()}
                                </label>
                                <Input
                                    id="from-name"
                                    type="text"
                                    bind:value={fromName}
                                    placeholder="My Project"
                                />
                            </div>
                        </div>

                        <div class="space-y-4 pt-6 border-t">
                            <h3 class="text-sm font-medium">
                                {m.newsletter_settings_test_config()}
                            </h3>
                            <div class="flex gap-2 flex-1">
                                <Input
                                    type="email"
                                    bind:value={testEmail}
                                    placeholder="test@example.com"
                                    class="flex-1"
                                />
                                <Button
                                    type="button"
                                    variant="outline"
                                    size="default"
                                    on:click={handleMailTest}
                                    disabled={mailTesting}
                                >
                                    {#if mailTesting}
                                        <Loader2
                                            class="h-4 w-4 animate-spin mr-2"
                                        />
                                    {:else}
                                        <Send class="h-4 w-4 mr-2" />
                                    {/if}
                                    {m.newsletter_settings_send_test()}
                                </Button>
                            </div>
                        </div>

                        <div class="flex justify-end">
                            <Button type="submit" disabled={mailSaving}>
                                {#if mailSaving}
                                    <Loader2
                                        class="h-4 w-4 animate-spin mr-2"
                                    />
                                    Saving...
                                {:else}
                                    <Save class="h-4 w-4 mr-2" />
                                    {m.newsletter_settings_save_smtp()}
                                {/if}
                            </Button>
                        </div>
                    </form>
                </div>

                <!-- Automation Section -->
                <div id="section-automation" class="scroll-mt-6 pt-12 border-t">
                    <div class="mb-6">
                        <div class="flex items-center gap-3 mb-1.5">
                            <Zap class="h-5 w-5 text-primary" />
                            <h3 class="text-base font-semibold">
                                {m.newsletter_settings_automation()}
                            </h3>
                        </div>
                        <p class="text-sm text-muted-foreground mt-1.5">
                            {m.newsletter_settings_automation_description()}
                        </p>
                    </div>

                    <form
                        on:submit|preventDefault={handleAutomationSave}
                        class="space-y-4"
                    >
                        <div>
                            <div class="text-sm font-medium mb-2">
                                {m.newsletter_settings_trigger_statuses()}
                            </div>

                            <div class="relative">
                                <button
                                    bind:this={statusButtonElement}
                                    type="button"
                                    on:click={toggleStatusSelect}
                                    class="h-9 px-3 text-sm border rounded-md bg-background hover:bg-muted flex items-center gap-2 transition-colors w-full justify-between"
                                    aria-haspopup="true"
                                    aria-expanded={statusSelectOpen}
                                >
                                    <span class="truncate text-left flex-1">
                                        {#if automationTriggerStatuses.length === 0}
                                            <span class="text-muted-foreground"
                                                >{m.newsletter_settings_select_statuses()}</span
                                            >
                                        {:else if automationTriggerStatuses.length === 1}
                                            {automationTriggerStatuses[0]}
                                        {:else}
                                            {m.newsletter_settings_statuses_selected(
                                                {
                                                    count: automationTriggerStatuses.length,
                                                },
                                            )}
                                        {/if}
                                    </span>
                                    <ChevronDown
                                        class={cn(
                                            "h-4 w-4 shrink-0 opacity-50 transition-transform duration-200",
                                            statusSelectOpen && "rotate-180",
                                        )}
                                    />
                                </button>

                                {#if statusSelectOpen}
                                    <div
                                        bind:this={statusDropdownElement}
                                        transition:fly={{
                                            duration: 200,
                                            y: 10,
                                        }}
                                        class="absolute left-0 bottom-full mb-1 w-full rounded-md border bg-background shadow-md p-2 text-sm space-y-1 z-50"
                                        role="menu"
                                    >
                                        {#if filteredStatuses.length === 0}
                                            <div
                                                class="py-6 text-center text-sm text-muted-foreground"
                                            >
                                                No statuses found
                                            </div>
                                        {:else}
                                            {#each filteredStatuses as status}
                                                <button
                                                    type="button"
                                                    class="w-full text-left px-2 py-1.5 rounded hover:bg-muted transition-colors flex items-center justify-between gap-2"
                                                    on:click={() =>
                                                        toggleStatus(
                                                            status.display_name,
                                                        )}
                                                    role="menuitem"
                                                >
                                                    <span class="truncate"
                                                        >{status.display_name}</span
                                                    >
                                                    <span
                                                        class={cn(
                                                            "flex h-4 w-4 items-center justify-center shrink-0",
                                                            automationTriggerStatuses.includes(
                                                                status.display_name,
                                                            )
                                                                ? "opacity-100"
                                                                : "opacity-0",
                                                        )}
                                                    >
                                                        <Check
                                                            class="h-4 w-4"
                                                        />
                                                    </span>
                                                </button>
                                            {/each}
                                        {/if}
                                    </div>
                                {/if}
                            </div>

                            <p class="text-xs text-muted-foreground mt-2">
                                {m.newsletter_settings_trigger_help()}
                            </p>
                        </div>

                        <div class="flex justify-end">
                            <Button
                                type="submit"
                                disabled={automationSaving}
                                size="sm"
                            >
                                {#if automationSaving}
                                    <Loader2
                                        class="h-4 w-4 animate-spin mr-2"
                                    />
                                    Saving...
                                {:else}
                                    <Save class="h-4 w-4 mr-2" />
                                    {m.newsletter_settings_save_automation()}
                                {/if}
                            </Button>
                        </div>
                    </form>
                </div>
            </div>
        </div>
    {/if}
</div>
