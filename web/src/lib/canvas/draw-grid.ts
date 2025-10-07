export function drawGrid(ctx: CanvasRenderingContext2D, width: number, height: number) {
	const boxSize = 50; // size of each box in the grid

	ctx.strokeStyle = '#333';
	ctx.lineWidth = 1;

	// Draw vertical lines
	for (let i = 0; i <= width / boxSize; i++) {
		ctx.beginPath();
		ctx.moveTo(i * boxSize, 0);
		ctx.lineTo(i * boxSize, height);
		ctx.stroke();
	}

	// Draw horizontal lines
	for (let i = 0; i <= height / boxSize; i++) {
		ctx.beginPath();
		ctx.moveTo(0, i * boxSize);
		ctx.lineTo(width, i * boxSize);
		ctx.stroke();
	}
}
