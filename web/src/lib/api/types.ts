export type APISuccessResponse<T> = {
	success: true;
	data: T;
};

export type APIErrorResponse = {
	success: false;
	error: string;
};

export type APIResponse<T> = APISuccessResponse<T> | APIErrorResponse;
