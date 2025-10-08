<script lang="ts">
	import { getServerStatus } from '$lib/api/status';
	import GridBackground from '$lib/components/grid-background.svelte';
	import GridBg from '$lib/components/grid-bg.svelte';
	import { sleep } from '$lib/util/sleep';
	import { onMount } from 'svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

<div class=" w-full mt-4">
	<div class="bg-primary/10 rounded-lg p-4">
		<div class="flex items-center space-x-2">
			<div class="inline-grid *:[grid-area:1/1]">
				<div class="status status-success animate-ping"></div>
				<div class="status status-success"></div>
			</div>
			<span class="text-sm">Server Online</span>
		</div>
		<div class="mt-4 space-y-5">
			<a href="/host" class="inline-flex w-full btn btn-primary btn-xl">Host</a>
			<a href="/join" class="inline-flex w-full btn btn-secondary btn-xl">Join</a>
		</div>
		<div class="mt-4"></div>
	</div>
</div>
