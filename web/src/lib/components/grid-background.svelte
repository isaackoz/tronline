<div class="grid-bg-container z-0">
	<div class="the-grid"></div>
	<div class="horizon"></div>
</div>

<style>
	.grid-bg-container {
		position: fixed;
		width: 100vw;
		height: 100vh;
		perspective: 500px;
		background: linear-gradient(to bottom, #000000 0%, #001a33 50%, #000000 100%);
	}

	.the-grid {
		position: absolute;
		width: 200%;
		height: 200%;
		left: -50%;
		top: -80vh;
		transform: rotateX(75deg);
		transform-style: preserve-3d;
		animation: moveGrid 20s linear infinite;
	}

	.the-grid::before,
	.the-grid::after {
		content: '';
		position: absolute;
		width: 100%;
		height: 100%;
		background-image:
			linear-gradient(
				0deg,
				transparent 24%,
				rgba(0, 255, 255, 0.3) 25%,
				rgba(0, 255, 255, 0.3) 26%,
				transparent 27%,
				transparent 74%,
				rgba(0, 255, 255, 0.3) 75%,
				rgba(0, 255, 255, 0.3) 76%,
				transparent 77%,
				transparent
			),
			linear-gradient(
				90deg,
				transparent 24%,
				rgba(0, 255, 255, 0.3) 25%,
				rgba(0, 255, 255, 0.3) 26%,
				transparent 27%,
				transparent 74%,
				rgba(0, 255, 255, 0.3) 75%,
				rgba(0, 255, 255, 0.3) 76%,
				transparent 77%,
				transparent
			);
		background-size: 25px 25px;
		background-position: 0 0;
		image-rendering: crisp-edges;
	}

	.the-grid::after {
		background-image:
			linear-gradient(
				0deg,
				transparent 49%,
				rgba(0, 200, 255, 0.6) 50%,
				rgba(0, 200, 255, 0.6) 51%,
				transparent 52%
			),
			linear-gradient(
				90deg,
				transparent 49%,
				rgba(0, 200, 255, 0.6) 50%,
				rgba(0, 200, 255, 0.6) 51%,
				transparent 52%
			);
		background-size: 25px 25px;
		image-rendering: crisp-edges;
	}

	@keyframes moveGrid {
		0% {
			transform: rotateX(60deg) translateY(0);
		}
		100% {
			transform: rotateX(60deg) translateY(100px);
		}
	}

	.horizon {
		position: absolute;
		bottom: 0;
		width: 100%;
		height: 50%;
		background: linear-gradient(to top, rgba(0, 255, 255, 0.1) 0%, transparent 100%);
		pointer-events: none;
	}
</style>
