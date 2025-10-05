import { getContext, setContext } from 'svelte';

export class GameState {
	constructor() {
		//
	}
}

// unique key for the game state context
const GAME_STATE_KEY = Symbol('GAMESTATE');

// helper to set the game state
export function setGameState() {
	return setContext(GAME_STATE_KEY, new GameState());
}

// helper to get the game state
export function getGameState() {
	return getContext<ReturnType<typeof setGameState>>(GAME_STATE_KEY);
}
