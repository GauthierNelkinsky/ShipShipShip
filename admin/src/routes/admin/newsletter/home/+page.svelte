<script lang="ts">
    import { onMount } from "svelte";
    import { api } from "$lib/api";
    import { Button, Input, Pagination } from "$lib/components/ui";
    import {
        Users,
        Download,
        Search,
        Mail,
        Trash2,
        Calendar,
        Loader2,
    } from "lucide-svelte";
    import { toast } from "svelte-sonner";
    import * as m from "$lib/paraglide/messages";

    interface HomeSection {
        id: string;
        title: string;
        description: string;
    }

    export let disabled = false;

    let loading = true;

    // Subscribers data
    let subscribers: any[] = [];
    let subscriberCount = 0;
    let subscribersLoading = false;
    let searchQuery = "";
    let filteredSubscribers: any[] = [];

    // Pagination for subscribers
    let currentSubscriberPage = 1;
    let subscriberLimit = 10;
    let totalSubscribers = 0;
    let totalSubscriberPages = 0;

    // Newsletter history data
    let newsletters: any[] = [];
    let newslettersLoading = false;

    // Pagination for newsletter history
    let currentHistoryPage = 1;
    let historyLimit = 10;
    let totalNewsletters = 0;
    let totalHistoryPages = 0;

    // Modals and editing
    let deleteConfirmModal = false;
    let emailToDelete = "";
    let deleteLoading = false;

    // Sidebar navigation
    let activeSection = "subscribers";
    let sidebarTop: number | null = null;
    let sidebarElement: HTMLElement;

    const sections: HomeSection[] = [
        {
            id: "subscribers",
            title: m.newsletter_home_subscribers(),
            description: m.newsletter_home_subscribers_description(),
        },
        {
            id: "history",
            title: m.newsletter_home_history(),
            description: m.newsletter_home_history_description(),
        },
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

    async function loadData() {
        loading = true;

        try {
            await Promise.all([loadSubscriberData(), loadNewsletterHistory()]);
        } catch (err) {
            console.error("Error loading data:", err);
            const errorMessage =
                err instanceof Error
                    ? err.message
                    : m.newsletter_home_load_failed();
            toast.error(m.newsletter_home_load_failed(), {
                description: errorMessage,
            });
        } finally {
            loading = false;
        }
    }

    async function loadSubscriberData() {
        subscribersLoading = true;
        try {
            const [stats, subscriberData] = await Promise.all([
                api.getNewsletterStats(),
                api.getNewsletterSubscribersPaginated(
                    currentSubscriberPage,
                    subscriberLimit,
                ),
            ]);

            subscriberCount = stats.active_subscribers;
            subscribers = subscriberData.subscribers || [];
            totalSubscribers = subscriberData.total;
            totalSubscriberPages = subscriberData.total_pages;
            filterSubscribers();
        } catch (err) {
            console.error("Failed to load subscriber data:", err);
            subscriberCount = 0;
            subscribers = [];
            totalSubscribers = 0;
            totalSubscriberPages = 0;
        } finally {
            subscribersLoading = false;
        }
    }

    async function loadNewsletterHistory() {
        newslettersLoading = true;
        try {
            const historyData = await api.getNewsletterHistory(
                currentHistoryPage,
                historyLimit,
            );
            newsletters = historyData.newsletters || [];
            totalNewsletters = historyData.total;
            totalHistoryPages = historyData.total_pages;
        } catch (err) {
            console.error("Failed to load newsletter history:", err);
            newsletters = [];
            totalNewsletters = 0;
            totalHistoryPages = 0;
        } finally {
            newslettersLoading = false;
        }
    }

    function filterSubscribers() {
        if (!searchQuery.trim()) {
            filteredSubscribers = subscribers;
        } else {
            filteredSubscribers = subscribers.filter((sub) =>
                sub.email.toLowerCase().includes(searchQuery.toLowerCase()),
            );
        }
    }

    async function exportSubscribers() {
        try {
            const allSubscribersData = await api.getNewsletterSubscribers();
            const allSubscribers = allSubscribersData.subscribers || [];

            const csvContent = [
                "Email,Status,Subscribed Date",
                ...allSubscribers.map(
                    (sub) =>
                        `${sub.email},Active,${new Date(sub.subscribed_at).toLocaleDateString()}`,
                ),
            ].join("\n");

            const blob = new Blob([csvContent], { type: "text/csv" });
            const url = window.URL.createObjectURL(blob);
            const a = document.createElement("a");
            a.style.display = "none";
            a.href = url;
            a.download = "newsletter-subscribers.csv";
            document.body.appendChild(a);
            a.click();
            window.URL.revokeObjectURL(url);
            document.body.removeChild(a);
        } catch (err) {
            console.error("Failed to export subscribers:", err);
            const errorMessage =
                err instanceof Error
                    ? err.message
                    : "Could not export subscribers. Please try again.";
            toast.error(m.newsletter_home_load_failed(), {
                description: errorMessage,
            });
        }
    }

    function openDeleteConfirm(email: string) {
        emailToDelete = email;
        deleteConfirmModal = true;
    }

    function closeDeleteConfirm() {
        deleteConfirmModal = false;
        emailToDelete = "";
        deleteLoading = false;
    }

    async function confirmDeleteSubscriber() {
        deleteLoading = true;

        try {
            await api.deleteNewsletterSubscriber(emailToDelete);
            toast.success(m.newsletter_subscriber_removed(), {
                description: m.newsletter_subscriber_removed_description({
                    email: emailToDelete,
                }),
            });
            await loadSubscriberData();
            closeDeleteConfirm();
        } catch (err) {
            const errorMessage =
                err instanceof Error
                    ? err.message
                    : m.newsletter_subscriber_remove_failed();
            toast.error(m.newsletter_subscriber_remove_failed(), {
                description: errorMessage,
            });
            deleteLoading = false;
        }
    }

    function formatDate(dateString: string) {
        const date = new Date(dateString);
        const month = String(date.getMonth() + 1).padStart(2, "0");
        const day = String(date.getDate()).padStart(2, "0");
        const year = String(date.getFullYear()).slice(-2);
        const hours = String(date.getHours()).padStart(2, "0");
        const minutes = String(date.getMinutes()).padStart(2, "0");
        return `${month}/${day}/${year} ${hours}:${minutes}`;
    }

    async function goToSubscriberPage(page: number) {
        if (page < 1 || page > totalSubscriberPages || subscribersLoading)
            return;
        currentSubscriberPage = page;
        await loadSubscriberData();
    }

    async function goToHistoryPage(page: number) {
        if (page < 1 || page > totalHistoryPages || newslettersLoading) return;
        currentHistoryPage = page;
        await loadNewsletterHistory();
    }

    function scrollToSection(sectionId: string) {
        const element = document.getElementById(`section-${sectionId}`);
        if (element) {
            element.scrollIntoView({ behavior: "smooth", block: "start" });
        }
    }

    $: (searchQuery, filterSubscribers());
</script>

<div class="w-full">
    {#if loading}
        <div class="flex-1 flex items-center justify-center py-16">
            <div class="flex items-center gap-2 text-sm">
                <Loader2 class="h-4 w-4 animate-spin" />
                <span class="text-muted-foreground">Loading...</span>
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
                <!-- Subscribers Section -->
                <div id="section-subscribers" class="scroll-mt-6">
                    <div class="mb-6">
                        <div class="flex items-center justify-between">
                            <div>
                                <div class="flex items-center gap-3 mb-1.5">
                                    <Users class="h-5 w-5 text-primary" />
                                    <h3 class="text-base font-semibold">
                                        {m.newsletter_subscribers()}
                                    </h3>
                                </div>
                                <p class="text-sm text-muted-foreground">
                                    {m.newsletter_active_subscribers({
                                        count: subscriberCount,
                                    })}
                                </p>
                            </div>
                            {#if subscribers.length > 0}
                                <Button
                                    variant="outline"
                                    size="sm"
                                    on:click={exportSubscribers}
                                    {disabled}
                                >
                                    <Download class="h-4 w-4 mr-2" />
                                    {m.newsletter_export_csv()}
                                </Button>
                            {/if}
                        </div>
                    </div>

                    <div class="space-y-4">
                        <div class="relative">
                            <Search
                                class="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground"
                            />
                            <Input
                                type="search"
                                placeholder={m.newsletter_search_subscribers()}
                                bind:value={searchQuery}
                                class="pl-10"
                            />
                        </div>

                        {#if subscribersLoading}
                            <div class="flex items-center justify-center py-8">
                                <div class="flex items-center gap-2 text-sm">
                                    <Loader2 class="h-4 w-4 animate-spin" />
                                    <span class="text-muted-foreground"
                                        >Loading...</span
                                    >
                                </div>
                            </div>
                        {:else if filteredSubscribers.length === 0}
                            <div
                                class="text-center py-8 border rounded-lg bg-muted/30"
                            >
                                <Users
                                    class="h-8 w-8 text-muted-foreground mx-auto mb-3"
                                />
                                <h3 class="font-medium text-lg mb-2">
                                    {#if searchQuery}
                                        {m.newsletter_no_subscribers_found()}
                                    {:else}
                                        {m.newsletter_no_subscribers_yet()}
                                    {/if}
                                </h3>
                                <p class="text-muted-foreground text-sm">
                                    {#if searchQuery}
                                        {m.newsletter_adjust_search()}
                                    {:else}
                                        {m.newsletter_subscribers_appear()}
                                    {/if}
                                </p>
                            </div>
                        {:else}
                            <div class="border rounded-lg overflow-hidden">
                                <div class="overflow-x-auto">
                                    <table class="w-full">
                                        <thead class="border-b border-border">
                                            <tr
                                                class="bg-muted"
                                                style="opacity: 0.5;"
                                            >
                                                <th
                                                    class="text-left py-1.5 px-2 text-xs font-medium text-muted-foreground"
                                                >
                                                    {m.newsletter_table_email()}
                                                </th>
                                                <th
                                                    class="text-left py-1.5 px-2 text-xs font-medium text-muted-foreground"
                                                >
                                                    {m.newsletter_table_subscribed()}
                                                </th>
                                                <th
                                                    class="text-right py-1.5 px-2 text-xs font-medium text-muted-foreground w-16"
                                                >
                                                    {m.newsletter_table_actions()}
                                                </th>
                                            </tr>
                                        </thead>
                                        <tbody>
                                            {#each filteredSubscribers as subscriber}
                                                <tr
                                                    class="border-b border-border last:border-0 hover:bg-muted/30"
                                                >
                                                    <td class="py-1.5 px-2">
                                                        <div
                                                            class="text-xs font-medium truncate max-w-xs"
                                                        >
                                                            {subscriber.email}
                                                        </div>
                                                    </td>
                                                    <td class="py-1.5 px-2">
                                                        <div
                                                            class="text-xs text-muted-foreground whitespace-nowrap"
                                                        >
                                                            {formatDate(
                                                                subscriber.subscribed_at,
                                                            )}
                                                        </div>
                                                    </td>
                                                    <td class="py-1.5 px-2">
                                                        <div
                                                            class="flex justify-end"
                                                        >
                                                            <Button
                                                                variant="ghost"
                                                                size="sm"
                                                                on:click={() =>
                                                                    openDeleteConfirm(
                                                                        subscriber.email,
                                                                    )}
                                                                {disabled}
                                                                class="h-6 w-6 p-0"
                                                            >
                                                                <Trash2
                                                                    class="h-3 w-3"
                                                                />
                                                            </Button>
                                                        </div>
                                                    </td>
                                                </tr>
                                            {/each}
                                        </tbody>
                                    </table>
                                </div>
                            </div>

                            {#if totalSubscriberPages > 1}
                                <div
                                    class="flex items-center justify-between pt-4"
                                >
                                    <div class="text-sm text-muted-foreground">
                                        {m.newsletter_showing_subscribers({
                                            from:
                                                (currentSubscriberPage - 1) *
                                                    subscriberLimit +
                                                1,
                                            to: Math.min(
                                                currentSubscriberPage *
                                                    subscriberLimit,
                                                totalSubscribers,
                                            ),
                                            total: totalSubscribers,
                                        })}
                                    </div>
                                    <Pagination
                                        currentPage={currentSubscriberPage}
                                        totalPages={totalSubscriberPages}
                                        disabled={subscribersLoading}
                                        on:pageChange={(e) =>
                                            goToSubscriberPage(e.detail)}
                                    />
                                </div>
                            {/if}
                        {/if}
                    </div>
                </div>

                <!-- Newsletter History Section -->
                <div id="section-history" class="scroll-mt-6 pt-12 border-t">
                    <div class="mb-6">
                        <div class="flex items-center gap-3 mb-1.5">
                            <Mail class="h-5 w-5 text-primary" />
                            <h3 class="text-base font-semibold">
                                {m.newsletter_history()}
                            </h3>
                        </div>
                        <p class="text-sm text-muted-foreground">
                            {m.newsletter_history_description()}
                        </p>
                    </div>

                    {#if newslettersLoading}
                        <div class="flex items-center justify-center py-8">
                            <div class="flex items-center gap-2 text-sm">
                                <Loader2 class="h-4 w-4 animate-spin" />
                                <span class="text-muted-foreground"
                                    >Loading...</span
                                >
                            </div>
                        </div>
                    {:else if newsletters.length === 0}
                        <div
                            class="text-center py-8 border rounded-lg bg-muted/30"
                        >
                            <Mail
                                class="h-8 w-8 text-muted-foreground mx-auto mb-3"
                            />
                            <h3 class="font-medium text-lg mb-2">
                                {m.newsletter_no_newsletters()}
                            </h3>
                            <p class="text-muted-foreground text-sm">
                                {m.newsletter_create_first()}
                            </p>
                        </div>
                    {:else}
                        <div class="border rounded-lg overflow-hidden">
                            <div class="overflow-x-auto">
                                <table class="w-full">
                                    <thead class="border-b border-border">
                                        <tr
                                            class="bg-muted"
                                            style="opacity: 0.5;"
                                        >
                                            <th
                                                class="text-left py-1.5 px-2 text-xs font-medium text-muted-foreground"
                                            >
                                                {m.newsletter_table_subject()}
                                            </th>
                                            <th
                                                class="text-left py-1.5 px-2 text-xs font-medium text-muted-foreground w-48"
                                            >
                                                {m.newsletter_table_date()}
                                            </th>
                                            <th
                                                class="text-left py-1.5 px-2 text-xs font-medium text-muted-foreground w-20"
                                            >
                                                {m.newsletter_table_recipients()}
                                            </th>
                                        </tr>
                                    </thead>
                                    <tbody>
                                        {#each newsletters as newsletter}
                                            <tr
                                                class="border-b border-border last:border-0 hover:bg-muted/30"
                                            >
                                                <td class="py-1.5 px-2">
                                                    <div
                                                        class="text-xs font-medium truncate max-w-md"
                                                    >
                                                        {newsletter.subject}
                                                    </div>
                                                </td>
                                                <td class="py-1.5 px-2">
                                                    {#if newsletter.sent_at}
                                                        <div
                                                            class="flex items-center gap-1.5 text-xs text-muted-foreground whitespace-nowrap"
                                                        >
                                                            <Calendar
                                                                class="h-3 w-3 flex-shrink-0"
                                                            />
                                                            {formatDate(
                                                                newsletter.sent_at,
                                                            )}
                                                        </div>
                                                    {:else}
                                                        <div
                                                            class="text-xs text-muted-foreground"
                                                        >
                                                            -
                                                        </div>
                                                    {/if}
                                                </td>
                                                <td class="py-1.5 px-2">
                                                    {#if newsletter.status === "sent"}
                                                        <div
                                                            class="text-xs text-muted-foreground"
                                                        >
                                                            {newsletter.recipient_count}
                                                        </div>
                                                    {:else}
                                                        <div
                                                            class="text-xs text-muted-foreground"
                                                        >
                                                            -
                                                        </div>
                                                    {/if}
                                                </td>
                                            </tr>
                                        {/each}
                                    </tbody>
                                </table>
                            </div>
                        </div>

                        {#if totalHistoryPages > 1}
                            <div class="flex items-center justify-between pt-4">
                                <div class="text-sm text-muted-foreground">
                                    {m.newsletter_showing_newsletters({
                                        from:
                                            (currentHistoryPage - 1) *
                                                historyLimit +
                                            1,
                                        to: Math.min(
                                            currentHistoryPage * historyLimit,
                                            totalNewsletters,
                                        ),
                                        total: totalNewsletters,
                                    })}
                                </div>
                                <Pagination
                                    currentPage={currentHistoryPage}
                                    totalPages={totalHistoryPages}
                                    disabled={newslettersLoading}
                                    on:pageChange={(e) =>
                                        goToHistoryPage(e.detail)}
                                />
                            </div>
                        {/if}
                    {/if}
                </div>
            </div>
        </div>
    {/if}
</div>

{#if deleteConfirmModal}
    <div
        class="fixed inset-0 bg-black/50 flex items-center justify-center z-50"
        on:click={closeDeleteConfirm}
        on:keydown={(e) => {
            if (e.key === "Escape") closeDeleteConfirm();
        }}
        role="button"
        tabindex="-1"
    >
        <div
            class="bg-background rounded-lg p-5 w-full max-w-sm space-y-4"
            on:click|stopPropagation
            on:keydown|stopPropagation
            role="dialog"
            tabindex="-1"
        >
            <h2 class="text-sm font-semibold">
                {m.newsletter_modal_remove_title()}
            </h2>
            <p class="text-xs text-muted-foreground">
                {m.newsletter_modal_remove_message({
                    email: emailToDelete,
                })}
            </p>
            <div class="flex justify-end gap-2 text-xs">
                <Button
                    variant="outline"
                    size="sm"
                    on:click={closeDeleteConfirm}
                    disabled={deleteLoading}
                >
                    {m.newsletter_modal_cancel()}
                </Button>
                <Button
                    variant="destructive"
                    size="sm"
                    on:click={confirmDeleteSubscriber}
                    disabled={deleteLoading}
                >
                    {deleteLoading
                        ? m.newsletter_modal_removing()
                        : m.newsletter_modal_remove_button()}
                </Button>
            </div>
        </div>
    </div>
{/if}
