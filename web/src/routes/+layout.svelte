<script lang="ts">
	import favicon from '$lib/assets/favicon.svg';
	import '@fontsource-variable/orbitron';
	import '../app.css';
	import { setConnectionState } from '$lib/state/connection.svelte';
	import { PUBLIC_API_SERVER_URL } from '$env/static/public';

	let { children } = $props();

	// initialize state here for websocket connection management
	setConnectionState({
		baseUrl: PUBLIC_API_SERVER_URL + '/ws',
		maxReconnectAttempts: 3
	});
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
</svelte:head>

{@render children?.()}

<style>
	:global(body) {
		font-family: 'Orbitron Variable', sans-serif;
		margin: 0;
	}

	:global(h1, h2, h3, h4, h5, h6) {
		font-family: 'Orbitron Variable', sans-serif;
		font-weight: 600;
	}
</style>
