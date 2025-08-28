<script lang="ts">
    import { onMount } from "svelte";
    import { api } from "$lib/api";
    import { ExternalLink } from "lucide-svelte";
    import type { FooterLink } from "$lib/types";

    let footerLinks: FooterLink[] = [];
    let loading = false;

    onMount(async () => {
        await loadFooterLinks();
    });

    async function loadFooterLinks() {
        try {
            loading = true;
            const response = await fetch("/api/footer-links/by-column");
            if (response.ok) {
                const data = await response.json();
                // Flatten the response into a single array for easier filtering
                footerLinks = Object.values(
                    data.links || {},
                ).flat() as FooterLink[];
            } else {
                console.error("Failed to fetch footer links:", response.status);
            }
        } catch (err) {
            console.error("Failed to load footer links:", err);
        } finally {
            loading = false;
        }
    }

    function getLinksForColumn(column: "left" | "middle" | "right") {
        return footerLinks.filter((link) => link.column === column);
    }
</script>

<footer
    class="border-t border-gray-200 dark:border-neutral-800 bg-white/80 dark:bg-neutral-950/80 backdrop-blur-sm"
>
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        {#if footerLinks.length > 0}
            <!-- Footer with custom links -->
            <div class="py-8">
                <div
                    class="grid grid-cols-1 md:grid-cols-3 gap-8 md:gap-12 lg:gap-16 items-start"
                >
                    <!-- Left Column -->
                    <div
                        class="space-y-3 text-center md:text-left max-w-xs mx-auto md:mx-0"
                    >
                        {#each getLinksForColumn("left") as link}
                            <a
                                href={link.url}
                                target={link.open_in_new_window
                                    ? "_blank"
                                    : "_self"}
                                rel={link.open_in_new_window
                                    ? "noopener noreferrer"
                                    : ""}
                                class="block text-sm text-gray-500 dark:text-neutral-400 hover:text-gray-700 dark:hover:text-neutral-300 transition-colors"
                            >
                                <span class="flex items-center gap-1">
                                    {link.name}
                                    {#if link.open_in_new_window}
                                        <ExternalLink class="h-3 w-3" />
                                    {/if}
                                </span>
                            </a>
                        {/each}
                    </div>

                    <!-- Middle Column (with ShipShipShip attribution) -->
                    <div
                        class="space-y-3 text-center px-4 md:px-8 max-w-xs mx-auto"
                    >
                        <!-- ShipShipShip Attribution (first to align with other columns) -->
                        {#if getLinksForColumn("middle").length === 0}
                            <div class="text-center">
                                <p
                                    class="text-sm text-gray-500 dark:text-neutral-400 flex items-center justify-center gap-2"
                                >
                                    Shipped with
                                    <a
                                        href="https://github.com/GauthierNelkinsky/ShipShipShip"
                                        target="_blank"
                                        rel="noopener noreferrer"
                                        class="font-medium text-gray-700 dark:text-neutral-300 hover:text-gray-900 dark:hover:text-neutral-100 transition-colors flex items-center gap-1"
                                    >
                                        ShipShipShip 🚢
                                    </a>
                                </p>
                            </div>
                        {:else}
                            {#each getLinksForColumn("middle") as link}
                                <a
                                    href={link.url}
                                    target={link.open_in_new_window
                                        ? "_blank"
                                        : "_self"}
                                    rel={link.open_in_new_window
                                        ? "noopener noreferrer"
                                        : ""}
                                    class="block text-sm text-gray-500 dark:text-neutral-400 hover:text-gray-700 dark:hover:text-neutral-300 transition-colors text-center"
                                >
                                    <span
                                        class="flex items-center justify-center gap-1"
                                    >
                                        {link.name}
                                        {#if link.open_in_new_window}
                                            <ExternalLink class="h-3 w-3" />
                                        {/if}
                                    </span>
                                </a>
                            {/each}

                            <!-- ShipShipShip Attribution after middle links -->
                            <div class="text-center">
                                <p
                                    class="text-sm text-gray-500 dark:text-neutral-400 flex items-center justify-center gap-2"
                                >
                                    Shipped with
                                    <a
                                        href="https://github.com/GauthierNelkinsky/ShipShipShip"
                                        target="_blank"
                                        rel="noopener noreferrer"
                                        class="font-medium text-gray-700 dark:text-neutral-300 hover:text-gray-900 dark:hover:text-neutral-100 transition-colors flex items-center gap-1"
                                    >
                                        ShipShipShip 🚢
                                    </a>
                                </p>
                            </div>
                        {/if}
                    </div>

                    <!-- Right Column -->
                    <div
                        class="space-y-3 text-center md:text-right max-w-xs mx-auto md:mx-0"
                    >
                        {#each getLinksForColumn("right") as link}
                            <a
                                href={link.url}
                                target={link.open_in_new_window
                                    ? "_blank"
                                    : "_self"}
                                rel={link.open_in_new_window
                                    ? "noopener noreferrer"
                                    : ""}
                                class="block text-sm text-gray-500 dark:text-neutral-400 hover:text-gray-700 dark:hover:text-neutral-300 transition-colors text-center md:text-right"
                            >
                                <span
                                    class="flex items-center justify-center md:justify-end gap-1"
                                >
                                    {link.name}
                                    {#if link.open_in_new_window}
                                        <ExternalLink class="h-3 w-3" />
                                    {/if}
                                </span>
                            </a>
                        {/each}
                    </div>
                </div>
            </div>
        {:else}
            <!-- Default footer without custom links -->
            <div class="flex items-center justify-center h-16 px-4">
                <p
                    class="text-sm text-gray-500 dark:text-neutral-400 flex items-center gap-2"
                >
                    Shipped with
                    <a
                        href="https://github.com/GauthierNelkinsky/ShipShipShip"
                        target="_blank"
                        rel="noopener noreferrer"
                        class="font-medium text-gray-700 dark:text-neutral-300 hover:text-gray-900 dark:hover:text-neutral-100 transition-colors flex items-center gap-1"
                    >
                        ShipShipShip 🚢
                    </a>
                </p>
            </div>
        {/if}
    </div>
</footer>
