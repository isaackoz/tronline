<script module>
</script>

<!-- A canvas implementation of the grid -->
<script lang="ts">
	import { drawGrid } from '$lib/canvas/draw-grid';
	import { onMount } from 'svelte';

	let canvas: HTMLCanvasElement;
	let width = $state(0);
	let height = $state(0);

	$effect(() => {
		const ctx = canvas.getContext('2d');
		if (!ctx) return;
		canvas.width = width;
		canvas.height = height;
		requestAnimationFrame(() => {
			drawGrid(ctx, width, height);
		});
	});

	onMount(() => {
		const observer = new ResizeObserver((e) => {
			width = canvas.clientWidth;
			height = canvas.clientHeight;
		});
		observer.observe(canvas);
	});
</script>

<canvas bind:this={canvas} id="grid-canvas"></canvas>

<style>
	#grid-canvas {
		pointer-events: none;
		position: fixed;
		display: block;
		top: 0;
		left: 0;
		width: 100vw;
		height: 100vh;
		z-index: 0;
		image-rendering: crisp-edges;
	}
</style>
