/**
 * Retries an API call with exponential backoff.
 * @param apiCall A function that returns a promise for the API call
 * @param maxRetries The maximum number of retries
 * @param baseDelay The base delay in milliseconds
 * @returns The result of the API call
 */
export async function retryWithExponentialBackoff<T>(
	apiCall: () => Promise<T>,
	maxRetries = 5,
	baseDelay = 100
) {
	for (let attempt = 0; attempt < maxRetries; attempt++) {
		try {
			return await apiCall();
		} catch (error) {
			if (attempt === maxRetries - 1) {
				throw error;
			}
			const delay = baseDelay * Math.pow(2, attempt) + Math.random() * baseDelay;
			await new Promise((res) => setTimeout(res, delay));
		}
	}
	// This line should never be reached, but satisfies TypeScript
	throw new Error('Max retries reached');
}
