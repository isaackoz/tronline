<script lang="ts">
	import GridBg from '$lib/components/grid-bg.svelte';
	import type { LayoutProps } from './$types';

	let { children, data }: LayoutProps = $props();
</script>

{#snippet serverError()}
	<div class="flex items-center space-x-2">
		<div class="inline-grid *:[grid-area:1/1]">
			<div class="status status-error animate-ping"></div>
			<div class="status status-error"></div>
		</div>
		<div class="text-sm">Offline</div>
	</div>
	<p class="mt-4 text-error">The server is currently offline. Please try again later.</p>
{/snippet}

<GridBg />
<main class="flex items-center w-full h-svh justify-center flex-col max-w-md mx-auto relative">
	<h1 class="text-8xl neon-text neon-blue">Tronline</h1>
	{#await data.isServerOnline}
		<div class="flex items-center space-x-2">
			<span class="status"></span>
			<span class="text-sm animate-pulse"> Checking server status... </span>
		</div>
		<div class="mt-4 space-y-5">
			<button
				class="btn btn-primary btn-xl w-full"
				title="Host a game"
				disabled
				aria-disabled="true">Host</button
			>
			<button
				class="btn btn-primary btn-xl w-full"
				title="Host a game"
				disabled
				aria-disabled="true">Join</button
			>
		</div>
	{:then isServerOnline}
		{#if isServerOnline}
			{@render children?.()}
		{:else}
			{serverError()}
		{/if}
	{:catch}
		{serverError()}
	{/await}

	<div class="mt-4">
		<p class="text-xs font-mono">
			Tronline is not affiliated with, endorsed by, or connected to The Walt Disney Company or any
			other rights holders of the Tron franchise. All Tron-related trademarks and copyrights are the
			property of The Walt Disney Company. This is a fan-made project created for educational and
			recreational purposes only and is not intended for commercial use.
		</p>
	</div>
</main>
