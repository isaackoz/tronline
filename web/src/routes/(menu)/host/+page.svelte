<script lang="ts">
	import { onMount } from 'svelte';
	import { getConnectionState } from '$lib/state/connection.svelte';

	const cs = getConnectionState();
	// let connectionError = $derived('Uh oh');
	let isLoading = $state(false);
	onMount(() => {
		// connect to the
		cs.connect('host');
		return () => {
			cs.disconnect();
		};
	});
</script>

<div class="w-full mt-4">
	<div class="bg-primary/20 rounded-lg p-4">
		<div class="">
			<a href="/">&larr; Back</a>
		</div>
		{#if cs.connectionError}
			<div class="text-red-500 mt-4">
				There was an error creating a room:
				<div class="font-mono">
					{cs.connectionError}
				</div>
			</div>
		{:else if cs.isConnected}
			{#if cs.isGuestConnected}
				<div class="flex items-center space-x-2 mt-4">
					<div aria-label="success" class="status status-success animate-pulse"></div>
					<span class="">Connected</span>
				</div>
				<button
					class="btn btn-xl btn-secondary mt-4 w-full"
					onclick={async () => {
						await cs.initiateP2P();
					}}>Start game</button
				>
			{:else}
				<div class="text-xl text-center animate-pulse mb-4 text-secondary">
					Waiting for player 2...
				</div>
				<div>Room code</div>
				<div
					class="input input-xl w-full text-center font-mono flex items-center justify-center text-secondary select-all"
				>
					{cs.roomId}
				</div>
				<div class="mt-4">
					<div class="w-full input-sm font-mono input select-all">
						{'https://tronline.app/join?code=' + cs.roomId}
					</div>
					<p class="text-sm text-secondary">Send this link to a friend to join your game.</p>
					<p class="mt-4 text-xs text-center">This room will expire in 10 minutes</p>
				</div>
			{/if}
		{:else}
			<span class="animate-pulse text-secondary">Creating a room...</span>
		{/if}
	</div>
</div>
