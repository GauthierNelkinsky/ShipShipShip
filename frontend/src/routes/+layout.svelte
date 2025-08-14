<script lang="ts">
	import '../app.css';
	import { onMount } from 'svelte';
	import { settings, loadSettings } from '$lib/stores/settings';
	import { theme } from '$lib/stores/theme';

	onMount(async () => {
		// Initialize theme
		theme.init();

		// Load project settings
		await loadSettings();
	});
</script>

<svelte:head>
	<title>Changelog</title>
	<meta name="description" content="Product roadmap and changelog" />
</svelte:head>

<div class="min-h-screen bg-background text-foreground">
	<slot />
</div>
