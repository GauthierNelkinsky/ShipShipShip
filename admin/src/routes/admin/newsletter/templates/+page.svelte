<script lang="ts">
    import { onMount } from "svelte";
    import { api } from "$lib/api";
    import { Button, Input, Textarea } from "$lib/components/ui";
    import {
        Save,
        Loader2,
        RotateCcw,
        FileText,
        Mail,
        Pencil,
        Eye,
    } from "lucide-svelte";
    import { toast } from "svelte-sonner";
    import * as m from "$lib/paraglide/messages";

    let loading = true;
    let templateSaving = false;
    let eventTemplate = "";
    let eventSubject = "";
    let welcomeTemplate = "";
    let welcomeSubject = "";
    let selectedTemplate: "event" | "welcome" = "event";
    let previewMode = false;

    // Sidebar state
    let sidebarTop: number | null = null;
    let sidebarElement: HTMLElement;

    interface TemplateOption {
        id: "event" | "welcome";
        title: string;
        icon: typeof FileText | typeof Mail;
    }

    const templates: TemplateOption[] = [
        {
            id: "event",
            title: m.newsletter_settings_event_template(),
            icon: FileText,
        },
        {
            id: "welcome",
            title: m.newsletter_settings_welcome_template(),
            icon: Mail,
        },
    ];

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
            <a href="{{event_url}}" style="background: #3b82f6; color: white; padding: 14px 28px; text-decoration: none; border-radius: 6px; display: inline-block; font-weight: bold; font-size: 16px;">See Details</a>
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

    $: currentSubject =
        selectedTemplate === "event" ? eventSubject : welcomeSubject;
    $: currentTemplate =
        selectedTemplate === "event" ? eventTemplate : welcomeTemplate;

    $: availableVariables =
        selectedTemplate === "event"
            ? [
                  "{{event_name}}",
                  "{{status}}",
                  "{{event_date}}",
                  "{{event_content}}",
                  "{{event_url}}",
                  "{{event_tags}}",
                  "{{project_name}}",
                  "{{project_url}}",
                  "{{unsubscribe_url}}",
              ]
            : ["{{project_name}}", "{{project_url}}", "{{unsubscribe_url}}"];

    // Sample data for preview
    const sampleData = {
        "{{event_name}}": "New Feature: Dark Mode Support",
        "{{status}}": "New Feature",
        "{{event_date}}":
            '<span style="color: #6b7280; font-size: 14px; font-weight: 500;">15 Jan. 2024</span>',
        "{{event_content}}":
            "<p>We're excited to announce dark mode support across the entire application. This feature has been highly requested and is now available to all users.</p><p>You can toggle between light and dark themes in your settings.</p>",
        "{{event_url}}": "https://example.com/updates/dark-mode",
        "{{event_tags}}":
            '<span style="display: inline-flex; align-items: center; border-radius: 12px; border: 1px solid #3b82f6; background-color: #3b82f620; color: #3b82f6; padding: 2px 8px; font-size: 11px; font-weight: 600; margin-right: 6px;">Feature</span><span style="display: inline-flex; align-items: center; border-radius: 12px; border: 1px solid #10b981; background-color: #10b98120; color: #10b981; padding: 2px 8px; font-size: 11px; font-weight: 600; margin-right: 6px;">UI</span>',
        "{{project_name}}": "ShipShipShip",
        "{{project_url}}": "https://example.com",
        "{{unsubscribe_url}}": "https://example.com/unsubscribe",
    };

    // Generate preview with replaced variables
    $: previewContent = (() => {
        let content = currentTemplate;
        for (const [variable, value] of Object.entries(sampleData)) {
            content = content.replaceAll(variable, value);
        }
        return content;
    })();

    function handleDragStart(event: DragEvent, variable: string) {
        if (event.dataTransfer) {
            event.dataTransfer.effectAllowed = "copy";
            event.dataTransfer.setData("text/plain", variable);
        }
    }

    // Native browser drag-and-drop will handle insertion automatically
    // We just need to update our reactive state after the drop
    function handleInput(event: Event, type: "subject" | "template") {
        const target = event.target as HTMLInputElement | HTMLTextAreaElement;
        if (type === "subject") {
            updateSubject(target.value);
        } else {
            updateTemplate(target.value);
        }
    }

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
    }

    onMount(async () => {
        await loadData();

        // Setup scroll handler for sidebar
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

    async function loadData() {
        try {
            await loadTemplateSettings();
        } catch (err) {
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

    function resetToDefault() {
        if (selectedTemplate === "event") {
            eventTemplate = defaultTemplates[TEMPLATE_TYPES.EVENT];
            eventSubject = DEFAULT_SUBJECTS[TEMPLATE_TYPES.EVENT];
        } else {
            welcomeTemplate = defaultTemplates[TEMPLATE_TYPES.WELCOME];
            welcomeSubject = DEFAULT_SUBJECTS[TEMPLATE_TYPES.WELCOME];
        }
        toast.success("Template reset to default");
    }

    function updateSubject(value: string) {
        if (selectedTemplate === "event") {
            eventSubject = value;
        } else {
            welcomeSubject = value;
        }
    }

    function updateTemplate(value: string) {
        if (selectedTemplate === "event") {
            eventTemplate = value;
        } else {
            welcomeTemplate = value;
        }
    }
</script>

<svelte:head>
    <title
        >{m.newsletter_settings_templates()} - {m.newsletter_page_title()}</title
    >
</svelte:head>

<div class="w-full">
    {#if loading}
        <div class="flex-1 flex items-center justify-center py-16">
            <div class="flex items-center gap-2 text-sm">
                <Loader2 class="h-4 w-4 animate-spin" />
                <span class="text-muted-foreground"
                    >{m.event_modal_loading()}</span
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
                        {#each templates as template}
                            <button
                                on:click={() =>
                                    (selectedTemplate = template.id)}
                                class="w-full text-left px-3 py-2 rounded-md text-sm transition-colors {selectedTemplate ===
                                template.id
                                    ? 'bg-accent text-accent-foreground font-medium'
                                    : 'text-muted-foreground hover:text-foreground hover:bg-accent/50'}"
                            >
                                {template.title}
                            </button>
                        {/each}
                    </nav>
                </div>
            </aside>

            <!-- Main Content -->
            <div class="flex-1 min-w-0 space-y-12 ml-6">
                <!-- Template Header -->
                <div class="scroll-mt-6">
                    <div class="mb-6">
                        <div class="flex items-center gap-3 mb-1.5">
                            {#if selectedTemplate === "event"}
                                <FileText class="h-5 w-5 text-primary" />
                                <h3 class="text-base font-semibold">
                                    {m.newsletter_settings_event_template()}
                                </h3>
                            {:else}
                                <Mail class="h-5 w-5 text-primary" />
                                <h3 class="text-base font-semibold">
                                    {m.newsletter_settings_welcome_template()}
                                </h3>
                            {/if}
                        </div>
                        <p class="text-sm text-muted-foreground mt-1.5">
                            {#if selectedTemplate === "event"}
                                {m.newsletter_settings_templates_description()}
                            {:else}
                                {m.newsletter_settings_templates_description()}
                            {/if}
                        </p>
                    </div>

                    <form
                        on:submit|preventDefault={handleTemplateSave}
                        class="space-y-12"
                    >
                        <!-- Subject -->
                        <div>
                            <label
                                for="template-subject"
                                class="text-sm font-medium mb-2 block"
                            >
                                {m.newsletter_settings_email_subject()}
                            </label>
                            <Input
                                id="template-subject"
                                type="text"
                                value={currentSubject}
                                on:input={(e) => handleInput(e, "subject")}
                                placeholder={m.event_modal_subject_placeholder()}
                            />
                        </div>

                        <!-- Template Content -->
                        <div>
                            <div class="flex items-center justify-between mb-2">
                                <label
                                    for="template-content"
                                    class="text-sm font-medium"
                                >
                                    {m.newsletter_settings_email_template_html()}
                                </label>
                                <div
                                    class="flex items-center gap-1 bg-muted rounded-md p-0.5"
                                >
                                    <button
                                        type="button"
                                        class="h-7 px-2 rounded flex items-center gap-1.5 text-xs transition-colors {!previewMode
                                            ? 'bg-background shadow-sm'
                                            : 'hover:bg-background/50'}"
                                        on:click={() => (previewMode = false)}
                                    >
                                        <Pencil class="h-3.5 w-3.5" />
                                        <span>{m.event_modal_edit()}</span>
                                    </button>
                                    <button
                                        type="button"
                                        class="h-7 px-2 rounded flex items-center gap-1.5 text-xs transition-colors {previewMode
                                            ? 'bg-background shadow-sm'
                                            : 'hover:bg-background/50'}"
                                        on:click={() => (previewMode = true)}
                                    >
                                        <Eye class="h-3.5 w-3.5" />
                                        <span>{m.event_modal_preview()}</span>
                                    </button>
                                </div>
                            </div>
                            {#if previewMode}
                                <div
                                    class="border rounded-lg p-6 bg-background min-h-[500px] overflow-auto"
                                >
                                    {@html previewContent}
                                </div>
                            {:else}
                                <Textarea
                                    id="template-content"
                                    value={currentTemplate}
                                    on:input={(e) => handleInput(e, "template")}
                                    placeholder={m.event_modal_newsletter_content_placeholder()}
                                    class="font-mono text-sm min-h-[500px]"
                                    rows={20}
                                />
                            {/if}
                            <div class="mt-3">
                                <p
                                    class="text-xs font-medium text-muted-foreground mb-2"
                                >
                                    Available variables (drag and drop):
                                </p>
                                <div class="flex flex-wrap gap-2">
                                    {#each availableVariables as variable}
                                        <button
                                            type="button"
                                            draggable="true"
                                            on:dragstart={(e) =>
                                                handleDragStart(e, variable)}
                                            class="px-2 py-1 text-xs font-mono bg-secondary text-secondary-foreground rounded border border-border hover:bg-accent cursor-grab active:cursor-grabbing transition-colors"
                                        >
                                            {variable}
                                        </button>
                                    {/each}
                                </div>
                            </div>
                        </div>

                        <!-- Actions -->
                        <div class="flex justify-end gap-2">
                            <Button
                                type="button"
                                variant="ghost"
                                on:click={resetToDefault}
                            >
                                <RotateCcw class="h-4 w-4 mr-2" />
                                {m.newsletter_settings_reset_default()}
                            </Button>
                            <Button type="submit" disabled={templateSaving}>
                                {#if templateSaving}
                                    <Loader2
                                        class="h-4 w-4 animate-spin mr-2"
                                    />
                                    Saving...
                                {:else}
                                    <Save class="h-4 w-4 mr-2" />
                                    {m.newsletter_settings_save_templates()}
                                {/if}
                            </Button>
                        </div>
                    </form>
                </div>
            </div>
        </div>
    {/if}
</div>
