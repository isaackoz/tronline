<script lang="ts">
	import { getServerStatus } from '$lib/api/status';
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

<main class="flex items-center w-full h-svh justify-center flex-col max-w-md mx-auto">
	<h1 class="text-5xl text-secondary">Tronline</h1>
	<div class="h-24 w-full mt-4">
		{#if isLoading}
			<div class="skeleton h-full w-full flex items-center justify-center text-sm text-primary">
				<p>Checking if the master control program is online...</p>
			</div>
		{:else if !isServerOnline}
			<div class="text-center text-sm">
				Tronline's server is currently offline. Please try again later.
			</div>
		{:else}
			<div class="text-center text-sm">
				The master control program is online! You can now proceed to the <a
					href="/play"
					class="text-primary underline">game</a
				>.
			</div>
		{/if}
	</div>
</main>

<style>
</style>
