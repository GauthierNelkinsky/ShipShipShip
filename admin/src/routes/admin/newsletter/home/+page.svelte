<script lang="ts">
    import { onMount } from "svelte";
    import { api } from "$lib/api";
    import { Button, Card, Input, Pagination } from "$lib/components/ui";
    import {
        Users,
        Download,
        Search,
        Mail,
        Trash2,
        Calendar,
    } from "lucide-svelte";
    import { toast } from "svelte-sonner";

    export let disabled = false;
    export let newsletterEnabled = false;

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
    let _editingSubscriber: any = null;
    let deleteConfirmModal = false;
    let emailToDelete = "";
    let deleteLoading = false;

    onMount(async () => {
        await loadData();
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
                    : "Failed to load newsletter data";
            toast.error("Failed to load newsletter data", {
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

    function exportSubscribers() {
        const csvContent = [
            "Email,Status,Subscribed Date",
            ...subscribers.map(
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
            toast.success("Subscriber removed", {
                description: `${emailToDelete} has been unsubscribed`,
            });
            await loadSubscriberData();
            closeDeleteConfirm();
        } catch (err) {
            const errorMessage =
                err instanceof Error
                    ? err.message
                    : "Failed to remove subscriber";
            toast.error("Failed to remove subscriber", {
                description: errorMessage,
            });
            deleteLoading = false;
        }
    }

    function formatDate(dateString: string) {
        return new Date(dateString).toLocaleDateString("en-US", {
            year: "numeric",
            month: "short",
            day: "numeric",
            hour: "2-digit",
            minute: "2-digit",
        });
    }

    function _calculateOpenRate(opens: number, recipients: number) {
        return recipients > 0 ? Math.round((opens / recipients) * 100) : 0;
    }

    function _calculateClickRate(clicks: number, opens: number) {
        return opens > 0 ? Math.round((clicks / opens) * 100) : 0;
    }

    // Pagination functions
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

    // Reactive statements
    $: (searchQuery, filterSubscribers());
</script>

{#if loading}
    <div class="flex items-center justify-center min-h-32">
        <div
            class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"
        ></div>
    </div>
{:else}
    <div class="grid gap-6">
        <!-- Subscribers Section -->
        <Card class="p-6">
            <div class="flex items-center justify-between mb-6">
                <div class="flex items-center gap-4">
                    <Users class="h-6 w-6 text-primary" />
                    <div>
                        <h2 class="text-lg font-semibold">Subscribers</h2>
                        <p class="text-sm text-muted-foreground">
                            {subscriberCount} active subscribers
                        </p>
                    </div>
                </div>

                <div class="flex items-center gap-2">
                    {#if subscribers.length > 0}
                        <Button
                            on:click={exportSubscribers}
                            variant="outline"
                            size="sm"
                            {disabled}
                        >
                            <Download class="h-4 w-4 mr-2" />
                            Export CSV
                        </Button>
                    {/if}
                </div>
            </div>

            <!-- Search -->
            <div class="mb-4">
                <div class="relative">
                    <Search
                        class="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground"
                    />
                    <Input
                        bind:value={searchQuery}
                        placeholder="Search subscribers..."
                        class="pl-10"
                        {disabled}
                    />
                </div>
            </div>

            {#if subscribersLoading}
                <div class="flex items-center justify-center py-8">
                    <div
                        class="animate-spin rounded-full h-6 w-6 border-b-2 border-primary"
                    ></div>
                </div>
            {:else if filteredSubscribers.length === 0}
                <div class="text-center py-8">
                    <Users
                        class="h-12 w-12 text-muted-foreground mx-auto mb-4"
                    />
                    <h3 class="font-medium text-lg mb-2">
                        {searchQuery
                            ? "No subscribers found"
                            : "No subscribers yet"}
                    </h3>
                    <p class="text-muted-foreground">
                        {searchQuery
                            ? "Try adjusting your search terms"
                            : "Subscribers will appear here once users sign up for your newsletter."}
                    </p>
                </div>
            {:else}
                <Card class="overflow-hidden">
                    <div class="overflow-x-auto">
                        <table class="w-full">
                            <thead class="border-b border-border">
                                <tr class="bg-muted" style="opacity: 0.5;">
                                    <th
                                        class="text-left py-2 px-3 font-medium text-sm text-muted-foreground"
                                        >Email</th
                                    >
                                    <th
                                        class="text-left py-2 px-3 font-medium text-sm text-muted-foreground"
                                        >Subscribed</th
                                    >
                                    <th
                                        class="text-right py-2 px-3 font-medium text-sm text-muted-foreground"
                                        >Actions</th
                                    >
                                </tr>
                            </thead>
                            <tbody>
                                {#each filteredSubscribers as subscriber}
                                    <tr
                                        class="border-b border-border hover:bg-muted transition-colors group"
                                        style="--hover-opacity: 0.2;"
                                        on:mouseenter={(e) =>
                                            (e.currentTarget.style.backgroundColor =
                                                "hsl(var(--muted) / 0.2)")}
                                        on:mouseleave={(e) =>
                                            (e.currentTarget.style.backgroundColor =
                                                "")}
                                    >
                                        <td class="py-2 px-3">
                                            <div
                                                class="font-medium text-sm text-foreground"
                                            >
                                                {subscriber.email}
                                            </div>
                                        </td>
                                        <td class="py-2 px-3">
                                            <div
                                                class="text-sm text-muted-foreground"
                                            >
                                                {formatDate(
                                                    subscriber.subscribed_at,
                                                )}
                                            </div>
                                        </td>
                                        <td class="py-2 px-3">
                                            <div
                                                class="flex items-center justify-end gap-1 opacity-0 group-hover:opacity-100 transition-opacity"
                                            >
                                                <Button
                                                    variant="ghost"
                                                    size="icon"
                                                    on:click={() =>
                                                        openDeleteConfirm(
                                                            subscriber.email,
                                                        )}
                                                    {disabled}
                                                    class="h-8 w-8 hover:bg-destructive hover:text-destructive-foreground"
                                                    title="Remove subscriber"
                                                >
                                                    <Trash2 class="h-3 w-3" />
                                                </Button>
                                            </div>
                                        </td>
                                    </tr>
                                {/each}
                            </tbody>
                        </table>
                    </div>
                </Card>

                <!-- Subscribers Pagination -->
                {#if totalSubscriberPages > 1}
                    <div
                        class="flex items-center justify-between mt-4 pt-4 border-t"
                    >
                        <div class="text-sm text-muted-foreground">
                            Showing {(currentSubscriberPage - 1) *
                                subscriberLimit +
                                1} to {Math.min(
                                currentSubscriberPage * subscriberLimit,
                                totalSubscribers,
                            )} of {totalSubscribers} subscribers
                        </div>
                        <Pagination
                            currentPage={currentSubscriberPage}
                            totalPages={totalSubscriberPages}
                            disabled={subscribersLoading}
                            on:pageChange={(e) => goToSubscriberPage(e.detail)}
                        />
                    </div>
                {/if}
            {/if}
        </Card>

        <!-- Newsletter History Section -->
        <Card class="p-6">
            <div class="flex items-center justify-between mb-6">
                <div class="flex items-center gap-4">
                    <Mail class="h-6 w-6 text-primary" />
                    <div>
                        <h2 class="text-lg font-semibold">
                            Newsletter History
                        </h2>
                        <p class="text-sm text-muted-foreground">
                            Past and scheduled newsletters
                        </p>
                    </div>
                </div>
            </div>

            {#if newslettersLoading}
                <div class="flex items-center justify-center py-8">
                    <div
                        class="animate-spin rounded-full h-6 w-6 border-b-2 border-primary"
                    ></div>
                </div>
            {:else if newsletters.length === 0}
                <div class="text-center py-8">
                    <Mail
                        class="h-12 w-12 text-muted-foreground mx-auto mb-4"
                    />
                    <h3 class="font-medium text-lg mb-2">No newsletters yet</h3>
                    <p class="text-muted-foreground">
                        Create your first newsletter to get started.
                    </p>
                </div>
            {:else}
                <Card class="overflow-hidden">
                    <div class="overflow-x-auto">
                        <table class="w-full">
                            <thead class="border-b border-border">
                                <tr class="bg-muted" style="opacity: 0.5;">
                                    <th
                                        class="text-left py-2 px-3 font-medium text-sm text-muted-foreground"
                                        >Subject</th
                                    >

                                    <th
                                        class="text-left py-2 px-3 font-medium text-sm text-muted-foreground"
                                        >Date</th
                                    >
                                    <th
                                        class="text-left py-2 px-3 font-medium text-sm text-muted-foreground"
                                        >Recipients</th
                                    >
                                </tr>
                            </thead>
                            <tbody>
                                {#each newsletters as newsletter}
                                    <tr
                                        class="border-b border-border hover:bg-muted transition-colors group"
                                        style="--hover-opacity: 0.2;"
                                        on:mouseenter={(e) =>
                                            (e.currentTarget.style.backgroundColor =
                                                "hsl(var(--muted) / 0.2)")}
                                        on:mouseleave={(e) =>
                                            (e.currentTarget.style.backgroundColor =
                                                "")}
                                    >
                                        <td class="py-2 px-3">
                                            <div
                                                class="font-medium text-sm text-foreground truncate max-w-xs"
                                            >
                                                {newsletter.subject}
                                            </div>
                                        </td>

                                        <td class="py-2 px-3">
                                            {#if newsletter.sent_at}
                                                <div
                                                    class="flex items-center gap-1 text-sm text-muted-foreground"
                                                >
                                                    <Calendar class="h-3 w-3" />
                                                    {formatDate(
                                                        newsletter.sent_at,
                                                    )}
                                                </div>
                                            {:else}
                                                <div
                                                    class="text-sm text-muted-foreground"
                                                >
                                                    -
                                                </div>
                                            {/if}
                                        </td>
                                        <td class="py-2 px-3">
                                            {#if newsletter.status === "sent"}
                                                <div
                                                    class="text-sm text-muted-foreground"
                                                >
                                                    {newsletter.recipient_count}
                                                </div>
                                            {:else}
                                                <div
                                                    class="text-sm text-muted-foreground"
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
                </Card>

                <!-- Newsletter History Pagination -->
                {#if totalHistoryPages > 1}
                    <div
                        class="flex items-center justify-between mt-4 pt-4 border-t"
                    >
                        <div class="text-sm text-muted-foreground">
                            Showing {(currentHistoryPage - 1) * historyLimit +
                                1} to {Math.min(
                                currentHistoryPage * historyLimit,
                                totalNewsletters,
                            )} of {totalNewsletters} newsletters
                        </div>
                        <Pagination
                            currentPage={currentHistoryPage}
                            totalPages={totalHistoryPages}
                            disabled={newslettersLoading}
                            on:pageChange={(e) => goToHistoryPage(e.detail)}
                        />
                    </div>
                {/if}
            {/if}
        </Card>
    </div>
{/if}

<!-- Delete Confirmation Modal -->
{#if deleteConfirmModal}
    <div
        class="fixed inset-0 z-50 flex items-center justify-center bg-black/40"
    >
        <div class="bg-background rounded-lg p-5 w-full max-w-sm space-y-4">
            <h2 class="text-sm font-semibold">Remove subscriber?</h2>
            <p class="text-xs text-muted-foreground">
                Are you sure you want to remove <strong>{emailToDelete}</strong>
                from the newsletter subscribers? This action cannot be undone.
            </p>
            <div class="flex justify-end gap-2 text-xs">
                <Button
                    variant="outline"
                    size="sm"
                    on:click={closeDeleteConfirm}
                    disabled={deleteLoading}
                >
                    Cancel
                </Button>
                <Button
                    size="sm"
                    on:click={confirmDeleteSubscriber}
                    disabled={deleteLoading}
                >
                    {deleteLoading ? "Removing..." : "Remove Subscriber"}
                </Button>
            </div>
        </div>
    </div>
{/if}
