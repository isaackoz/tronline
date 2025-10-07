<script lang="ts">
	import GridBg from '$lib/components/grid-bg.svelte';

	let { children } = $props();
</script>

<GridBg />
<main class="flex items-center w-full h-svh justify-center flex-col max-w-md mx-auto relative">
	<h1 class="text-8xl text-secondary">Tronline</h1>
	{@render children?.()}
</main>
