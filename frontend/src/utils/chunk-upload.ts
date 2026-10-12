import { chunkUploadFileData, type FileUploadRequestConfig } from '@/api/modules/files';

export const CHUNK_SIZE = 5 * 1024 * 1024;
const MAX_CHUNK_RETRIES = 3;

const shouldRetryChunkUpload = (error: unknown) => {
    const item = error as {
        code?: string | number;
        data?: { retryable?: boolean };
        response?: { status?: number };
    };
    if (item?.code === 'ERR_CANCELED') {
        return false;
    }
    if (typeof item?.data?.retryable === 'boolean') {
        return item.data.retryable;
    }
    const status = item?.response?.status ?? (typeof item?.code === 'number' ? item.code : undefined);
    return !status || status === 408 || status === 429 || status >= 500;
};

const waitForChunkRetry = (attempt: number, signal: AbortSignal) => {
    return new Promise<void>((resolve) => {
        if (signal.aborted) {
            resolve();
            return;
        }
        const timer = window.setTimeout(done, 500 * 2 ** attempt);
        function done() {
            window.clearTimeout(timer);
            signal.removeEventListener('abort', done);
            resolve();
        }
        signal.addEventListener('abort', done, { once: true });
    });
};

export const uploadChunkWithRetry = async (
    formData: FormData,
    config: FileUploadRequestConfig,
    signal: AbortSignal,
) => {
    let retryCount = 0;
    while (true) {
        signal.throwIfAborted();
        try {
            await chunkUploadFileData(formData, config);
            return;
        } catch (error) {
            if (signal.aborted || retryCount >= MAX_CHUNK_RETRIES || !shouldRetryChunkUpload(error)) {
                throw error;
            }
            await waitForChunkRetry(retryCount, signal);
            retryCount++;
        }
    }
};
