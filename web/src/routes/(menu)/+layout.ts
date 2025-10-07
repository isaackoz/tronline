import { getServerStatus } from '$lib/api/status';
import type { LayoutLoad } from './$types';

export const load: LayoutLoad = async () => {
	return {
		// stream the server status response to all child routes
		isServerOnline: getServerStatus()
	};
};
