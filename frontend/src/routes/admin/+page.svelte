<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';

	onMount(() => {
		// Redirect to events page as it's now the main admin page
		goto('/admin/events', { replaceState: true });
	});
</script>

<!-- This page immediately redirects to /admin/events -->
<div class="min-h-screen flex items-center justify-center">
	<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
</div>
