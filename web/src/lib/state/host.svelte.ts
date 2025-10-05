import { getContext, setContext } from 'svelte';

export class HostState {
	constructor() {
		//
	}
}

// unique key for the host state context
const HOST_STATE_KEY = Symbol('HOST_STATE');

// helper to set the host state
export function setHostState() {
	return setContext(HOST_STATE_KEY, new HostState());
}

// helper to get the host state
export function getHostState() {
	return getContext<ReturnType<typeof setHostState>>(HOST_STATE_KEY);
}
