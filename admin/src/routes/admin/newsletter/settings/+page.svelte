<script lang="ts">
    import { onMount } from "svelte";
    import { api } from "$lib/api";
    import type {
        UpdateMailSettingsRequest,
        EventStatus,
        StatusDefinition,
    } from "$lib/types";
    import {
        Button,
        Card,
        Input,
        Textarea,
        Collapsible,
        Switch,
    } from "$lib/components/ui";
    import {
        Save,
        Send,
        Eye,
        EyeOff,
        FileText,
        Mail,
        ChevronDown,
        ChevronRight,
        UserCheck,
        Zap,
    } from "lucide-svelte";
    import { toast } from "svelte-sonner";
    import * as m from "$lib/paraglide/messages";

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

    // Template settings
    // Template content
    let templateSaving = false;
    // Template data
    let eventTemplate = "";
    let eventSubject = "";
    let welcomeTemplate = "";
    let welcomeSubject = "";

    // UI state
    let eventTemplateOpen = false;
    let welcomeTemplateOpen = false;

    const encryptionOptions = [
        { value: "none", label: "None" },
        { value: "tls", label: "TLS" },
        { value: "ssl", label: "SSL" },
    ];

    // Email template constants - matches backend constants
    const TEMPLATE_TYPES = {
        EVENT: "event",
        WELCOME: "welcome",
    };

    const DEFAULT_SUBJECTS = {
        [TEMPLATE_TYPES.EVENT]: "{{status}}: {{event_name}} - {{project_name}}",
        [TEMPLATE_TYPES.WELCOME]: "Welcome to {{project_name}}!",
    };

    const mobileTemplateStructure = `
        <div style="margin-bottom: 20px;">
            <div style="margin-bottom: 8px; color: #6b7280; font-size: 14px;">
                {{event_date}}
            </div>
            <div style="display: flex; flex-wrap: wrap; gap: 6px; align-items: center;">
                {{event_tags}}
            </div>
        </div>`;

    const defaultTemplates = {
        [TEMPLATE_TYPES.EVENT]: `<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px;">
    <h1 style="color: #3B82F6; text-align: center; font-size: 28px; font-weight: bold; margin: 20px 0;">{{status}}</h1>

    <div style="padding: 20px; margin-bottom: 20px;">
        <h2 style="color: #000000; margin-top: 0; font-size: 48px; font-weight: bold; margin-bottom: 15px; text-align: center;">{{event_name}}</h2>

        ${mobileTemplateStructure}
        </div>

        <div style="margin: 15px 0; font-size: 16px; line-height: 1.6;">
            {{event_content}}
        </div>
        <div style="text-align: center; margin-top: 30px;">
            <a href="{{event_url}}" style="background: {{primary_color}}; color: white; padding: 14px 28px; text-decoration: none; border-radius: 6px; display: inline-block; font-weight: bold; font-size: 16px;">See Details</a>
        </div>
    </div>

    <hr style="border: none; border-top: 1px solid #eee; margin: 30px 0;">

    <div style="text-align: center; font-size: 12px; color: #666;">
        <p style="margin: 5px 0;">
            <a href="{{project_url}}" style="color: #2563eb; text-decoration: none;">{{project_name}}</a>
            <br><a href="{{unsubscribe_url}}" style="color: #2563eb; text-decoration: none;">Unsubscribe</a>
        </p>
    </div>
</body>`,
        [TEMPLATE_TYPES.WELCOME]: `<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px;">
    <h1 style="color: #000000; text-align: center; font-size: 28px; font-weight: bold; margin: 20px 0;">🎉 Welcome to {{project_name}}!</h1>

    <div style="padding: 20px; margin-bottom: 20px;">
        <h2 style="color: #000000; margin-top: 0; font-size: 22px; font-weight: bold; margin-bottom: 15px;">Thanks for subscribing!</h2>

        <div style="margin: 15px 0; font-size: 16px; line-height: 1.6;">
            You've successfully subscribed to our newsletter. We'll keep you updated with the latest features, releases, and important announcements.
        </div>
    </div>

    <hr style="border: none; border-top: 1px solid #eee; margin: 30px 0;">

    <div style="text-align: center; font-size: 12px; color: #666;">
        <p style="margin: 5px 0;">
            <a href="{{project_url}}" style="color: #2563eb; text-decoration: none;">{{project_name}}</a>
            <br><a href="{{unsubscribe_url}}" style="color: #2563eb; text-decoration: none;">Unsubscribe</a>
        </p>
    </div>
</body>`,
    };

    onMount(async () => {
        await loadData();
    });

    async function loadData() {
        loading = true;

        try {
            await loadMailSettings();
            await loadTemplateSettings();
            await loadNewsletterAutomationSettings();
            await loadStatuses();
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
            // Filter out reserved statuses (Backlogs and Archived)
            statuses = allStatuses.filter((s) => !s.is_reserved);
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

    async function handleAutomationSave() {
        automationSaving = true;

        // Validate form
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

    function toggleStatusSelection(status: EventStatus) {
        if (automationTriggerStatuses.includes(status)) {
            automationTriggerStatuses = automationTriggerStatuses.filter(
                (s) => s !== status,
            );
        } else {
            automationTriggerStatuses = [...automationTriggerStatuses, status];
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

    async function loadTemplateSettings() {
        try {
            const response = await api.getEmailTemplates();
            const templates = response.templates;

            if (templates[TEMPLATE_TYPES.EVENT]) {
                eventTemplate = templates[TEMPLATE_TYPES.EVENT].content;
                eventSubject = templates[TEMPLATE_TYPES.EVENT].subject;
            } else {
                eventTemplate = defaultTemplates[TEMPLATE_TYPES.EVENT];
                eventSubject = DEFAULT_SUBJECTS[TEMPLATE_TYPES.EVENT];
            }

            if (templates[TEMPLATE_TYPES.WELCOME]) {
                welcomeTemplate = templates[TEMPLATE_TYPES.WELCOME].content;
                welcomeSubject = templates[TEMPLATE_TYPES.WELCOME].subject;
            } else {
                welcomeTemplate = defaultTemplates[TEMPLATE_TYPES.WELCOME];
                welcomeSubject = DEFAULT_SUBJECTS[TEMPLATE_TYPES.WELCOME];
            }
        } catch {
            console.log("No templates found, using defaults");
            eventTemplate = defaultTemplates[TEMPLATE_TYPES.EVENT];
            eventSubject = DEFAULT_SUBJECTS[TEMPLATE_TYPES.EVENT];
            welcomeTemplate = defaultTemplates[TEMPLATE_TYPES.WELCOME];
            welcomeSubject = DEFAULT_SUBJECTS[TEMPLATE_TYPES.WELCOME];
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

    async function handleTemplateSave() {
        templateSaving = true;

        try {
            const templateData = {
                [TEMPLATE_TYPES.EVENT]: {
                    subject: eventSubject,
                    content: eventTemplate,
                },
                [TEMPLATE_TYPES.WELCOME]: {
                    subject: welcomeSubject,
                    content: welcomeTemplate,
                },
            };

            await api.updateEmailTemplates(templateData);
            toast.success(m.newsletter_settings_templates_saved(), {
                description:
                    m.newsletter_settings_templates_saved_description(),
            });
        } catch (err) {
            const errorMessage =
                err instanceof Error
                    ? err.message
                    : m.newsletter_settings_templates_save_failed();
            toast.error(m.newsletter_settings_templates_save_failed(), {
                description: errorMessage,
            });
        } finally {
            templateSaving = false;
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

    function resetToDefault(templateType: string) {
        switch (templateType) {
            case "event":
                eventTemplate = defaultTemplates[TEMPLATE_TYPES.EVENT];
                eventSubject = DEFAULT_SUBJECTS[TEMPLATE_TYPES.EVENT];
                break;
            case "welcome":
                welcomeTemplate = defaultTemplates[TEMPLATE_TYPES.WELCOME];
                welcomeSubject = DEFAULT_SUBJECTS[TEMPLATE_TYPES.WELCOME];
                break;
        }
    }
</script>

<svelte:head>
    <title>{m.newsletter_settings_page_title()}</title>
</svelte:head>

{#if loading}
    <div class="flex items-center justify-center min-h-32">
        <div
            class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"
        ></div>
    </div>
{:else}
    <div class="space-y-6">
        <!-- Newsletter Display Settings -->

        <!-- Newsletter Automation Settings -->
        <Card class="p-6">
            <div class="flex items-center gap-4 mb-6">
                <Zap class="h-6 w-6 text-primary" />
                <div>
                    <h2 class="text-lg font-semibold">
                        {m.newsletter_settings_automation()}
                    </h2>
                    <p class="text-sm text-muted-foreground">
                        {m.newsletter_settings_automation_description()}
                    </p>
                </div>
            </div>

            <form
                on:submit|preventDefault={handleAutomationSave}
                class="space-y-6"
            >
                <div class="space-y-4">
                    <div>
                        <div class="block text-sm font-medium mb-3">
                            {m.newsletter_settings_trigger_statuses()}
                        </div>
                        <div class="grid gap-3">
                            {#if statuses.length === 0}
                                <div
                                    class="text-sm text-muted-foreground text-center py-4"
                                >
                                    {m.newsletter_settings_no_statuses()}
                                </div>
                            {:else}
                                {#each statuses as status}
                                    <label
                                        class="flex items-center gap-3 p-3 border border-border rounded-lg hover:bg-muted/30 cursor-pointer"
                                    >
                                        <input
                                            type="checkbox"
                                            checked={automationTriggerStatuses.includes(
                                                status.display_name,
                                            )}
                                            on:change={() =>
                                                toggleStatusSelection(
                                                    status.display_name,
                                                )}
                                            class="h-4 w-4 text-primary border-border rounded focus:ring-2 focus:ring-primary"
                                        />
                                        <div>
                                            <div class="text-sm font-medium">
                                                {status.display_name}
                                            </div>
                                            <div
                                                class="text-xs text-muted-foreground"
                                            >
                                                {m.newsletter_settings_send_when(
                                                    {
                                                        status: status.display_name,
                                                    },
                                                )}
                                            </div>
                                        </div>
                                    </label>
                                {/each}
                            {/if}
                        </div>
                        <p class="text-xs text-muted-foreground mt-2">
                            {m.newsletter_settings_trigger_help()}
                        </p>
                    </div>
                </div>

                <div class="flex justify-end">
                    <Button
                        type="submit"
                        disabled={automationSaving ||
                            (automationEnabled &&
                                automationTriggerStatuses.length === 0)}
                    >
                        {#if automationSaving}
                            <div
                                class="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"
                            ></div>
                        {:else}
                            <Save class="h-4 w-4 mr-2" />
                        {/if}
                        {m.newsletter_settings_save_automation()}
                    </Button>
                </div>
            </form>
        </Card>

        <!-- SMTP Configuration -->
        <!-- SMTP Settings -->
        <Card class="p-6">
            <div class="flex items-center gap-4 mb-6">
                <Mail class="h-6 w-6 text-primary" />
                <div>
                    <h2 class="text-lg font-semibold">
                        {m.newsletter_settings_smtp()}
                    </h2>
                    <p class="text-sm text-muted-foreground">
                        {m.newsletter_settings_smtp_description()}
                    </p>
                </div>
            </div>

            <form on:submit|preventDefault={handleMailSave} class="space-y-6">
                <!-- SMTP Configuration -->
                <div class="grid gap-4 md:grid-cols-3">
                    <div>
                        <label
                            for="smtp-host"
                            class="block text-sm font-medium mb-2"
                        >
                            {m.newsletter_settings_smtp_host()}
                        </label>
                        <Input
                            id="smtp-host"
                            bind:value={smtpHost}
                            placeholder="smtp.gmail.com"
                            disabled={mailSaving}
                        />
                    </div>
                    <div>
                        <label
                            for="smtp-port"
                            class="block text-sm font-medium mb-2"
                        >
                            {m.newsletter_settings_smtp_port()}
                        </label>
                        <Input
                            id="smtp-port"
                            bind:value={smtpPort}
                            placeholder="587"
                            disabled={mailSaving}
                        />
                    </div>
                    <div>
                        <label
                            for="smtp-encryption"
                            class="block text-sm font-medium mb-2"
                        >
                            {m.newsletter_settings_encryption()}
                        </label>
                        <select
                            id="smtp-encryption"
                            bind:value={smtpEncryption}
                            disabled={mailSaving}
                            class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                        >
                            {#each encryptionOptions as option}
                                <option value={option.value}
                                    >{option.label}</option
                                >
                            {/each}
                        </select>
                    </div>
                </div>

                <!-- Authentication -->
                <div class="grid gap-4 md:grid-cols-2">
                    <div>
                        <label
                            for="smtp-username"
                            class="block text-sm font-medium mb-2"
                        >
                            {m.newsletter_settings_username()}
                        </label>
                        <Input
                            id="smtp-username"
                            bind:value={smtpUsername}
                            placeholder="your@email.com"
                            disabled={mailSaving}
                        />
                    </div>
                    <div>
                        <label
                            for="smtp-password"
                            class="block text-sm font-medium mb-2"
                        >
                            {m.newsletter_settings_password()}
                        </label>
                        <div class="relative">
                            <Input
                                id="smtp-password"
                                type={showPassword ? "text" : "password"}
                                bind:value={smtpPassword}
                                placeholder="••••••••"
                                disabled={mailSaving}
                                class="pr-10"
                            />
                            <button
                                type="button"
                                on:click={() => (showPassword = !showPassword)}
                                class="absolute inset-y-0 right-0 flex items-center pr-3"
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

                <!-- From Information -->
                <div class="grid gap-4 md:grid-cols-2">
                    <div>
                        <label
                            for="from-email"
                            class="block text-sm font-medium mb-2"
                        >
                            {m.newsletter_settings_from_email()}
                        </label>
                        <Input
                            id="from-email"
                            type="email"
                            bind:value={fromEmail}
                            placeholder="noreply@yourdomain.com"
                            disabled={mailSaving}
                        />
                    </div>
                    <div>
                        <label
                            for="from-name"
                            class="block text-sm font-medium mb-2"
                        >
                            {m.newsletter_settings_from_name()}
                        </label>
                        <Input
                            id="from-name"
                            bind:value={fromName}
                            placeholder="Your Company"
                            disabled={mailSaving}
                        />
                    </div>
                </div>

                <!-- Test Email -->
                <div class="space-y-4">
                    <h3 class="text-sm font-medium">
                        {m.newsletter_settings_test_config()}
                    </h3>
                    <div class="flex gap-2 flex-1">
                        <Input
                            type="email"
                            bind:value={testEmail}
                            placeholder="test@example.com"
                            disabled={mailTesting}
                            class="flex-1"
                        />
                        <Button
                            type="button"
                            on:click={handleMailTest}
                            disabled={mailTesting || !testEmail.trim()}
                            variant="outline"
                        >
                            {#if mailTesting}
                                <div
                                    class="animate-spin rounded-full h-4 w-4 border-b-2 border-current mr-2"
                                ></div>
                            {:else}
                                <Send class="h-4 w-4 mr-2" />
                            {/if}
                            {m.newsletter_settings_send_test()}
                        </Button>
                    </div>
                </div>

                <!-- Save Button -->
                <div class="flex justify-end">
                    <Button type="submit" disabled={mailSaving}>
                        {#if mailSaving}
                            <div
                                class="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"
                            ></div>
                        {:else}
                            <Save class="h-4 w-4 mr-2" />
                        {/if}
                        {m.newsletter_settings_save_smtp()}
                    </Button>
                </div>
            </form>
        </Card>

        <!-- Email Templates -->
        <!-- Email Template Settings -->
        <Card class="p-6">
            <div class="flex items-center gap-4 mb-6">
                <FileText class="h-6 w-6 text-primary" />
                <div>
                    <h2 class="text-lg font-semibold">
                        {m.newsletter_settings_templates()}
                    </h2>
                    <p class="text-sm text-muted-foreground">
                        {m.newsletter_settings_templates_description()}
                    </p>
                </div>
            </div>

            <form
                on:submit|preventDefault={handleTemplateSave}
                class="space-y-6"
            >
                <!-- Event Template -->
                <Collapsible bind:open={eventTemplateOpen}>
                    <div
                        slot="trigger"
                        let:toggle
                        class="flex items-center justify-between w-full p-4 text-left bg-muted/30 hover:bg-muted/50 rounded-lg cursor-pointer transition-colors"
                        on:click={toggle}
                        on:keydown={(e) => {
                            if (e.key === "Enter" || e.key === " ") toggle();
                        }}
                        role="button"
                        tabindex="0"
                    >
                        <div class="flex items-center gap-3">
                            <FileText class="h-5 w-5 text-primary" />
                            <h3 class="text-sm font-medium">
                                {m.newsletter_settings_event_template()}
                            </h3>
                        </div>
                        <div class="flex items-center gap-2">
                            {#if eventTemplateOpen}
                                <ChevronDown
                                    class="h-4 w-4 text-muted-foreground transition-transform"
                                />
                            {:else}
                                <ChevronRight
                                    class="h-4 w-4 text-muted-foreground transition-transform"
                                />
                            {/if}
                        </div>
                    </div>

                    <div class="mt-4 space-y-4">
                        <div>
                            <label
                                for="eventSubject"
                                class="block text-sm font-medium mb-2"
                            >
                                {m.newsletter_settings_email_subject()}
                            </label>
                            <Input
                                id="eventSubject"
                                bind:value={eventSubject}
                                placeholder="&#123;&#123;status&#125;&#125;: &#123;&#123;event_name&#125;&#125; - &#123;&#123;project_name&#125;&#125;"
                            />
                        </div>

                        <div>
                            <label
                                for="eventTemplate"
                                class="block text-sm font-medium mb-2"
                            >
                                {m.newsletter_settings_email_template_html()}
                            </label>
                            <Textarea
                                id="eventTemplate"
                                bind:value={eventTemplate}
                                rows={15}
                                placeholder="Enter your event email template..."
                            />
                        </div>
                        <div class="flex items-center justify-between mt-1">
                            <p class="text-xs text-muted-foreground">
                                {m.newsletter_settings_event_variables()}
                            </p>
                            <Button
                                type="button"
                                variant="outline"
                                size="sm"
                                on:click={() => resetToDefault("event")}
                            >
                                {m.newsletter_settings_reset_default()}
                            </Button>
                        </div>
                    </div>
                </Collapsible>

                <!-- Welcome Template -->
                <Collapsible bind:open={welcomeTemplateOpen}>
                    <div
                        slot="trigger"
                        let:toggle
                        class="flex items-center justify-between w-full p-4 text-left bg-muted/30 hover:bg-muted/50 rounded-lg cursor-pointer transition-colors"
                        on:click={toggle}
                        on:keydown={(e) => {
                            if (e.key === "Enter" || e.key === " ") toggle();
                        }}
                        role="button"
                        tabindex="0"
                    >
                        <div class="flex items-center gap-3">
                            <UserCheck class="h-5 w-5 text-primary" />
                            <h3 class="text-sm font-medium">
                                {m.newsletter_settings_welcome_template()}
                            </h3>
                        </div>
                        <div class="flex items-center gap-2">
                            {#if welcomeTemplateOpen}
                                <ChevronDown
                                    class="h-4 w-4 text-muted-foreground transition-transform"
                                />
                            {:else}
                                <ChevronRight
                                    class="h-4 w-4 text-muted-foreground transition-transform"
                                />
                            {/if}
                        </div>
                    </div>

                    <div class="mt-4 space-y-4">
                        <div>
                            <label
                                for="welcome-subject"
                                class="block text-sm font-medium mb-2"
                            >
                                {m.newsletter_settings_email_subject()}
                            </label>
                            <Input
                                id="welcome-subject"
                                bind:value={welcomeSubject}
                                placeholder="Welcome to &#123;&#123;project_name&#125;&#125;!"
                            />
                        </div>

                        <div>
                            <label
                                for="welcome-template"
                                class="block text-sm font-medium mb-2"
                            >
                                {m.newsletter_settings_email_template_html()}
                            </label>
                            <Textarea
                                id="welcome-template"
                                bind:value={welcomeTemplate}
                                rows={15}
                                placeholder="Enter your welcome email template..."
                            />
                        </div>
                        <div class="flex items-center justify-between mt-1">
                            <p class="text-xs text-muted-foreground">
                                {m.newsletter_settings_welcome_variables()}
                            </p>
                            <Button
                                type="button"
                                variant="outline"
                                size="sm"
                                on:click={() => resetToDefault("welcome")}
                            >
                                {m.newsletter_settings_reset_default()}
                            </Button>
                        </div>
                    </div>
                </Collapsible>

                <!-- Save Button -->
                <div class="flex justify-end pt-4 border-t border-border">
                    <Button type="submit" disabled={templateSaving}>
                        {#if templateSaving}
                            <div
                                class="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"
                            ></div>
                        {:else}
                            <Save class="h-4 w-4 mr-2" />
                        {/if}
                        {m.newsletter_settings_save_templates()}
                    </Button>
                </div>
            </form>
        </Card>
    </div>
{/if}
