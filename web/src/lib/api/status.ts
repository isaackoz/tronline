import { PUBLIC_API_SERVER_URL } from '$env/static/public';
import { retryWithExponentialBackoff } from '$lib/util/exponential-backoff';
import type { APIResponse } from './types';

export async function getServerStatus(): Promise<APIResponse<'OK'>> {
	try {
		const res = await retryWithExponentialBackoff(() =>
			fetch(`${PUBLIC_API_SERVER_URL}/healthz`, {
				method: 'GET'
			})
		);

		if (!res.ok) {
			return { success: false, error: `HTTP error! status: ${res.status}` };
		}

		return { success: true, data: 'OK' };
	} catch (error) {
		console.error(error);
		return { success: false, error: 'Failed to reach server' };
	}
}
