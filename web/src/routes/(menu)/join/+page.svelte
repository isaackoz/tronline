<script lang="ts">
	import { getConnectionState } from '$lib/state/connection.svelte';
	import { onMount } from 'svelte';

	let roomCodeInput = $state('');
	let isInputValid = $state(false);
	let isLoading = $state(false);
	const cs = getConnectionState();

	function joinGame() {
		if (isLoading) return;
		if (!isInputValid) return;
		if (roomCodeInput.length !== 6) return;
		isLoading = true;
		cs.connect('client', roomCodeInput);
		isLoading = false;
	}
	onMount(() => {
		return () => {
			// disconnect when the component is unmounted
			cs.disconnect();
		};
	});

	$effect(() => {
		isInputValid = false;
		if (roomCodeInput.startsWith('https')) {
			roomCodeInput = roomCodeInput.slice(-6);
		}
		// sanitize input to be alphanumeric and uppercase only, max length 6
		roomCodeInput = roomCodeInput
			.toUpperCase()
			.replace(/[^A-Z0-9]/g, '')
			.slice(0, 6);
		if (roomCodeInput.length === 6) {
			isInputValid = true;
		}
	});
</script>

<div class="w-full mt-4">
	<div class="bg-primary/20 rounded-lg p-4">
		<div class="">
			<a href="/">&larr; Back</a>
		</div>
		{#if cs.isConnected}
			<div class="mt-4">
				<div class="flex items-center space-x-2">
					<div aria-label="success" class="status status-success animate-pulse"></div>
					<span class="">Connected</span>
				</div>
				<div class="mt-4">
					<p class="text-center text-secondary animate-pulse">
						Waiting for the host to start the game.
					</p>
				</div>
			</div>
		{:else}
			<h2 class="text-center text-2xl mb-4 text-secondary">Enter Room Code</h2>
			<input
				type="text"
				placeholder="Room code..."
				class="input input-xl w-full text-secondary text-center font-mono"
				disabled={isLoading || cs.isConnecting}
				bind:value={roomCodeInput}
			/>
			{#if cs.connectionError || cs.roomError}
				<div class="text-red-500">
					<div class="font-mono capitalize font-bold text-sm">
						{cs.roomError ?? cs.connectionError}
					</div>
				</div>
			{/if}
			<button
				class="btn btn-secondary w-full mt-4 text-xl"
				disabled={!isInputValid || isLoading || cs.isConnecting}
				onclick={joinGame}
			>
				{#if cs.isConnecting || isLoading}
					Joining...
				{:else}
					Join
				{/if}
			</button>
		{/if}
	</div>
</div>
