<script lang="ts">
    import { Card } from "$lib/components/ui";
    import { Palette, Tag } from "lucide-svelte";
</script>

<svelte:head>
    <title>Customization - Admin</title>
</svelte:head>

<div class="max-w-4xl mx-auto">
    <div class="mb-6">
        <h1 class="text-2xl font-semibold mb-2">Customization</h1>
        <p class="text-muted-foreground">
            Customize your changelog's appearance and manage tags
        </p>
    </div>

    <div class="grid grid-cols-1 md:grid-cols-2 gap-8">
        <!-- Branding Card -->
        <a href="/admin/customization/branding" class="block group">
            <Card
                class="p-8 h-full transition-all duration-300 group-hover:shadow-lg group-hover:scale-[1.02] border group-hover:border-primary/30"
            >
                <div class="flex flex-col gap-4">
                    <div class="flex items-center gap-4">
                        <div
                            class="p-4 bg-gradient-to-br from-primary/10 to-primary/5 rounded-xl"
                        >
                            <Palette class="h-7 w-7 text-primary" />
                        </div>
                        <div>
                            <h3 class="text-xl font-semibold mb-1">Branding</h3>
                            <p class="text-muted-foreground text-sm">
                                Visual identity & appearance
                            </p>
                        </div>
                    </div>
                    <p class="text-muted-foreground text-sm leading-relaxed">
                        Customize your site's logo, colors, favicon, and overall
                        visual identity
                    </p>
                    <div class="flex flex-wrap gap-2 mt-2">
                        <span
                            class="inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium bg-primary/10 text-primary"
                        >
                            Logo & Colors
                        </span>
                        <span
                            class="inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium bg-muted text-muted-foreground"
                        >
                            Favicon
                        </span>
                        <span
                            class="inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium bg-muted text-muted-foreground"
                        >
                            Website URL
                        </span>
                    </div>
                </div>
            </Card>
        </a>

        <!-- Tags Card -->
        <a href="/admin/customization/tags" class="block group">
            <Card
                class="p-8 h-full transition-all duration-300 group-hover:shadow-lg group-hover:scale-[1.02] border group-hover:border-primary/30"
            >
                <div class="flex flex-col gap-4">
                    <div class="flex items-center gap-4">
                        <div
                            class="p-4 bg-gradient-to-br from-primary/10 to-primary/5 rounded-xl"
                        >
                            <Tag class="h-7 w-7 text-primary" />
                        </div>
                        <div>
                            <h3 class="text-xl font-semibold mb-1">Tags</h3>
                            <p class="text-muted-foreground text-sm">
                                Organize & categorize content
                            </p>
                        </div>
                    </div>
                    <p class="text-muted-foreground text-sm leading-relaxed">
                        Create and manage tags for organizing your changelog
                        entries and events
                    </p>
                    <div class="flex flex-wrap gap-2 mt-2">
                        <span
                            class="inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium bg-primary/10 text-primary"
                        >
                            Create Tags
                        </span>
                        <span
                            class="inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium bg-muted text-muted-foreground"
                        >
                            Edit Colors
                        </span>
                        <span
                            class="inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium bg-muted text-muted-foreground"
                        >
                            Manage Usage
                        </span>
                    </div>
                </div>
            </Card>
        </a>
    </div>
</div>
