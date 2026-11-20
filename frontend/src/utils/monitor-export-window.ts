// Splits a long export range into consecutive time windows fetched one after another, so every
// request stays small enough to finish well inside the API timeout. The window size adapts to how
// dense the data is and how fast responses arrive, and shrinks automatically when a request fails.

const MINUTE = 60 * 1000;
const HOUR = 60 * MINUTE;
const DAY = 24 * HOUR;

export interface ExportWindow {
    // inclusive start of the slice this window owns (epoch ms)
    from: number;
    // exclusive end of the slice (epoch ms)
    to: number;
    // start time to send to the API: the monitor search filter is exclusive on both ends, so every
    // window after the first reaches back 1ms to keep rows that sit exactly on a boundary
    queryFrom: number;
}

export interface ExportWindowResult<T> {
    data: T;
    // how full the window was relative to what one request should carry: 1 is ideal, 2 means halve it
    load: number;
    // the response is unusable at this size (e.g. the API downsampled it); refetch a smaller window
    overflow?: boolean;
}

export interface ExportWindowOptions<T> {
    start: number;
    end: number;
    fetch: (window: ExportWindow) => Promise<ExportWindowResult<T>>;
    // called once per window, in time order, only after the whole window was fetched successfully
    commit: (data: T, window: ExportWindow) => void | Promise<void>;
    onProgress?: (ratio: number, window: ExportWindow) => void;
    isRetryable?: (error: unknown) => boolean;
    signal?: AbortSignal;
    initialSpan?: number;
    minSpan?: number;
    maxSpan?: number;
    // a window that takes longer than this is treated as overloaded, so slow links get smaller windows
    targetDuration?: number;
    // attempts per window before giving up; each failed attempt halves the window
    maxAttempts?: number;
    retryDelay?: number;
    now?: () => number;
    sleep?: (ms: number, signal?: AbortSignal) => Promise<void>;
}

export class ExportAbortedError extends Error {
    constructor() {
        super('export aborted');
        this.name = 'ExportAbortedError';
    }
}

export class ExportWindowError extends Error {
    window: ExportWindow;
    cause: unknown;
    constructor(window: ExportWindow, cause: unknown) {
        super(cause instanceof Error ? cause.message : String((cause as any)?.message ?? cause));
        this.name = 'ExportWindowError';
        this.window = window;
        this.cause = cause;
    }
}

const clamp = (value: number, min: number, max: number) => Math.min(max, Math.max(min, value));

const abortableSleep = (ms: number, signal?: AbortSignal) =>
    new Promise<void>((resolve, reject) => {
        if (signal?.aborted) {
            reject(new ExportAbortedError());
            return;
        }
        const onAbort = () => {
            clearTimeout(timer);
            reject(new ExportAbortedError());
        };
        const timer = setTimeout(() => {
            signal?.removeEventListener('abort', onAbort);
            resolve();
        }, ms);
        signal?.addEventListener('abort', onAbort, { once: true });
    });

// network failures, timeouts and 5xx/408/429 are worth retrying with a smaller window; business errors
// (a resolved response carrying a numeric code) and other 4xx answers would only repeat themselves
export const isTransientRequestError = (error: unknown): boolean => {
    const err = error as any;
    if (err && typeof err.code === 'number') {
        return false;
    }
    const status = err?.response?.status;
    if (typeof status === 'number' && status < 500 && status !== 408 && status !== 429) {
        return false;
    }
    return true;
};

export async function runExportWindows<T>(options: ExportWindowOptions<T>): Promise<void> {
    const { start, end, signal } = options;
    const minSpan = options.minSpan ?? 5 * MINUTE;
    const maxSpan = Math.max(minSpan, options.maxSpan ?? 31 * DAY);
    const targetDuration = options.targetDuration ?? 15 * 1000;
    const maxAttempts = Math.max(1, options.maxAttempts ?? 5);
    const retryDelay = options.retryDelay ?? 1000;
    const isRetryable = options.isRetryable ?? (() => true);
    const now = options.now ?? (() => Date.now());
    const sleep = options.sleep ?? abortableSleep;

    const throwIfAborted = () => {
        if (signal?.aborted) {
            throw new ExportAbortedError();
        }
    };

    let span = clamp(options.initialSpan ?? 6 * HOUR, minSpan, maxSpan);
    let from = start;
    let attempt = 0;
    options.onProgress?.(0, { from: start, to: start, queryFrom: start });
    while (from < end) {
        throwIfAborted();
        const to = Math.min(end, from + span);
        const window: ExportWindow = { from, to, queryFrom: from === start ? start : from - 1 };
        // the last window may be clipped by the range end, so adapt from the span actually requested
        const used = to - from;
        const began = now();
        let result: ExportWindowResult<T>;
        try {
            result = await options.fetch(window);
        } catch (error) {
            throwIfAborted();
            attempt++;
            if (attempt >= maxAttempts || !isRetryable(error)) {
                throw new ExportWindowError(window, error);
            }
            span = Math.max(minSpan, Math.floor(used / 2));
            await sleep(retryDelay * 2 ** (attempt - 1), signal);
            continue;
        }
        throwIfAborted();
        if (result.overflow && used > minSpan) {
            span = Math.max(minSpan, Math.floor(used / Math.max(2, result.load)));
            continue;
        }
        await options.commit(result.data, window);
        attempt = 0;
        from = to;
        options.onProgress?.(end > start ? (to - start) / (end - start) : 1, window);

        const pressure = Math.max(result.load, (now() - began) / targetDuration);
        // limit each step to 4x either way so one sparse or one slow window does not swing the size wildly
        const factor = pressure > 0 ? clamp(1 / pressure, 0.25, 4) : 4;
        span = clamp(Math.round(used * factor), minSpan, maxSpan);
    }
}

// windows overlap by 1ms, so a row sitting in that slice comes back twice; this drops rows the previous
// window already committed. Call it exactly once per committed window, empty windows included.
export function createBoundaryDeduper() {
    let previous = new Set<string>();
    return <R>(rows: R[], key: (row: R) => string): R[] => {
        const current = new Set<string>();
        const kept = rows.filter((row) => {
            const k = key(row);
            if (previous.has(k) || current.has(k)) {
                return false;
            }
            current.add(k);
            return true;
        });
        previous = current;
        return kept;
    };
}
