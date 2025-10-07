<script lang="ts">
	import { getServerStatus } from '$lib/api/status';
	import GridBackground from '$lib/components/grid-background.svelte';
	import { onMount } from 'svelte';

	let isLoading = $state(true);
	let isServerOnline = $state(false);
	onMount(async () => {
		isLoading = true;
		const serverStatus = await getServerStatus();
		if (!serverStatus.success) {
			isServerOnline = false;
		} else {
			isServerOnline = true;
		}
		isLoading = false;
	});
</script>

<GridBackground />
<main class="flex items-center w-full h-svh justify-center flex-col max-w-md mx-auto relative">
	<h1 class="text-8xl text-secondary">Tronline</h1>
	<div class="h-24 w-full mt-4">
		{#if isLoading}
			<div class="skeleton h-full w-full flex items-center justify-center text-sm text-primary">
				<p>Checking if the master control program is online...</p>
			</div>
		{:else if !isServerOnline}
			<div class="inline-grid *:[grid-area:1/1]">
				<div class="status status-error animate-ping"></div>
				<div class="status status-error"></div>
			</div>
			<div class="text-center text-sm">
				Tronline's server is currently offline. Please try again later.
			</div>
		{:else}
			<div class="flex items-center justify-center space-x-2">
				<div class="inline-grid *:[grid-area:1/1]">
					<div class="status status-success animate-ping"></div>
					<div class="status status-success"></div>
				</div>
				<span class="">Server Online</span>
			</div>
			<div class="text-center text-sm"></div>
		{/if}
	</div>
</main>

<style>
</style>
