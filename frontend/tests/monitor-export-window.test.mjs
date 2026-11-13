import assert from 'node:assert/strict';
import test from 'node:test';
import {
    createBoundaryDeduper,
    ExportAbortedError,
    ExportWindowError,
    isTransientRequestError,
    runExportWindows,
} from '../src/utils/monitor-export-window.ts';

const MINUTE = 60 * 1000;
const HOUR = 60 * MINUTE;
const DAY = 24 * HOUR;

const noSleep = async () => {};

// a fake monitor table: one row per minute, filtered like the backend (created_at > start AND created_at < end)
const minuteRows = (start, end) => {
    const rows = [];
    for (let time = start; time <= end; time += MINUTE) {
        rows.push({ id: time, time });
    }
    return rows;
};
const search = (rows, window) => rows.filter((row) => row.time > window.queryFrom && row.time < window.to);

test('fetches a short range in a single window with the original bounds', async () => {
    const windows = [];
    await runExportWindows({
        start: 0,
        end: HOUR,
        fetch: async (window) => {
            windows.push(window);
            return { data: null, load: 0.1 };
        },
        commit: () => {},
    });
    assert.deepEqual(windows, [{ from: 0, to: HOUR, queryFrom: 0 }]);
});

test('splits a long range into contiguous windows that export every row exactly once', async () => {
    const start = 0;
    const end = 30 * DAY;
    const rows = minuteRows(start, end);
    const dedupe = createBoundaryDeduper();
    const exported = [];
    const windows = [];
    const progress = [];
    await runExportWindows({
        start,
        end,
        fetch: async (window) => {
            windows.push(window);
            const data = search(rows, window);
            return { data, load: data.length / 1000 };
        },
        commit: (data) => {
            exported.push(...dedupe(data, (row) => String(row.id)));
        },
        onProgress: (ratio) => progress.push(ratio),
    });

    assert.ok(windows.length > 1);
    assert.equal(windows[0].queryFrom, start);
    for (let i = 1; i < windows.length; i++) {
        assert.equal(windows[i].from, windows[i - 1].to);
        assert.equal(windows[i].queryFrom, windows[i].from - 1);
    }
    assert.equal(windows.at(-1).to, end);
    // the range itself stays exclusive on both ends, like a single request over the whole range
    assert.deepEqual(
        exported.map((row) => row.time),
        rows.filter((row) => row.time > start && row.time < end).map((row) => row.time),
    );
    assert.equal(progress[0], 0);
    assert.equal(progress.at(-1), 1);
});

test('keeps a row that sits exactly on a window boundary', async () => {
    const rows = [{ id: 'edge', time: 2 * HOUR }];
    const dedupe = createBoundaryDeduper();
    const exported = [];
    await runExportWindows({
        start: 0,
        end: 4 * HOUR,
        initialSpan: 2 * HOUR,
        fetch: async (window) => ({ data: search(rows, window), load: 1 }),
        commit: (data) => exported.push(...dedupe(data, (row) => row.id)),
    });
    assert.deepEqual(exported, rows);
});

test('grows sparse windows and shrinks dense ones, at most 4x per step', async () => {
    const spans = [];
    const loads = [0.01, 0.5, 8, 1];
    await runExportWindows({
        start: 0,
        end: 365 * DAY,
        initialSpan: HOUR,
        maxSpan: 365 * DAY,
        fetch: async (window) => {
            spans.push(window.to - window.from);
            return { data: null, load: loads[spans.length - 1] ?? 1 };
        },
        commit: () => {},
    });
    assert.equal(spans[0], HOUR);
    assert.equal(spans[1], 4 * HOUR);
    assert.equal(spans[2], 8 * HOUR);
    assert.equal(spans[3], 2 * HOUR);
    assert.equal(spans[4], 2 * HOUR);
});

test('shrinks windows whose responses are slow even when they carry few rows', async () => {
    let clock = 0;
    const spans = [];
    await runExportWindows({
        start: 0,
        end: 10 * DAY,
        initialSpan: 8 * HOUR,
        targetDuration: 10_000,
        now: () => clock,
        fetch: async (window) => {
            spans.push(window.to - window.from);
            clock += spans.length === 1 ? 40_000 : 10_000;
            return { data: null, load: 0.1 };
        },
        commit: () => {},
    });
    assert.equal(spans[1], 2 * HOUR);
});

test('retries a failed window with half the span and commits it only once', async () => {
    const calls = [];
    const commits = [];
    const delays = [];
    await runExportWindows({
        start: 0,
        end: 4 * HOUR,
        initialSpan: 4 * HOUR,
        retryDelay: 100,
        sleep: async (ms) => delays.push(ms),
        fetch: async (window) => {
            calls.push(window);
            if (calls.length <= 2) {
                throw new Error('timeout of 60000ms exceeded');
            }
            return { data: window.to, load: 1 };
        },
        commit: (data) => commits.push(data),
    });
    assert.deepEqual(
        calls.map((window) => [window.from, window.to - window.from]),
        [
            [0, 4 * HOUR],
            [0, 2 * HOUR],
            [0, HOUR],
            [HOUR, HOUR],
            [2 * HOUR, HOUR],
            [3 * HOUR, HOUR],
        ],
    );
    assert.deepEqual(delays, [100, 200]);
    assert.deepEqual(commits, [HOUR, 2 * HOUR, 3 * HOUR, 4 * HOUR]);
});

test('gives up after the configured attempts and reports the failing window', async () => {
    let calls = 0;
    const failure = new Error('Network Error');
    await assert.rejects(
        runExportWindows({
            start: 0,
            end: DAY,
            maxAttempts: 3,
            sleep: noSleep,
            fetch: async () => {
                calls++;
                throw failure;
            },
            commit: () => assert.fail('nothing should be committed'),
        }),
        (error) => {
            assert.ok(error instanceof ExportWindowError);
            assert.equal(error.cause, failure);
            assert.equal(error.window.from, 0);
            return true;
        },
    );
    assert.equal(calls, 3);
});

test('does not retry errors that would only repeat themselves', async () => {
    let calls = 0;
    await assert.rejects(
        runExportWindows({
            start: 0,
            end: DAY,
            sleep: noSleep,
            isRetryable: isTransientRequestError,
            fetch: async () => {
                calls++;
                throw { code: 400, message: 'invalid time range' };
            },
            commit: () => {},
        }),
        ExportWindowError,
    );
    assert.equal(calls, 1);
});

test('refetches an overflowing window at a smaller size without committing it', async () => {
    const calls = [];
    const commits = [];
    await runExportWindows({
        start: 0,
        end: 6 * HOUR,
        initialSpan: 6 * HOUR,
        maxAttempts: 1,
        fetch: async (window) => {
            const span = window.to - window.from;
            calls.push(span);
            // pretend the API downsamples anything wider than 2 hours
            return span > 2 * HOUR ? { data: 'bucketed', load: 3, overflow: true } : { data: span, load: 1 };
        },
        commit: (data) => commits.push(data),
    });
    assert.deepEqual(calls, [6 * HOUR, 2 * HOUR, 2 * HOUR, 2 * HOUR]);
    assert.deepEqual(commits, [2 * HOUR, 2 * HOUR, 2 * HOUR]);
});

test('accepts an overflowing window once it reaches the minimum span', async () => {
    const commits = [];
    await runExportWindows({
        start: 0,
        end: 10 * MINUTE,
        initialSpan: 10 * MINUTE,
        minSpan: 10 * MINUTE,
        fetch: async () => ({ data: 'bucketed', load: 5, overflow: true }),
        commit: (data) => commits.push(data),
    });
    assert.deepEqual(commits, ['bucketed']);
});

test('stops with ExportAbortedError when cancelled between windows or while waiting to retry', async () => {
    const controller = new AbortController();
    let calls = 0;
    await assert.rejects(
        runExportWindows({
            start: 0,
            end: 30 * DAY,
            initialSpan: DAY,
            signal: controller.signal,
            fetch: async () => {
                calls++;
                return { data: null, load: 1 };
            },
            commit: () => {
                if (calls === 2) {
                    controller.abort();
                }
            },
        }),
        ExportAbortedError,
    );
    assert.equal(calls, 2);

    const retrying = new AbortController();
    const pending = runExportWindows({
        start: 0,
        end: DAY,
        retryDelay: 60_000,
        signal: retrying.signal,
        fetch: async () => {
            throw new Error('Network Error');
        },
        commit: () => {},
    });
    setTimeout(() => retrying.abort(), 10);
    await assert.rejects(pending, ExportAbortedError);
});

test('a request rejected because of cancellation is reported as an abort, not a failure', async () => {
    const controller = new AbortController();
    await assert.rejects(
        runExportWindows({
            start: 0,
            end: DAY,
            signal: controller.signal,
            fetch: async () => {
                controller.abort();
                throw new Error('canceled');
            },
            commit: () => {},
        }),
        ExportAbortedError,
    );
});

test('deduper drops rows already committed by the previous window only', () => {
    const dedupe = createBoundaryDeduper();
    const key = (row) => row;
    assert.deepEqual(dedupe(['a', 'b'], key), ['a', 'b']);
    assert.deepEqual(dedupe(['b', 'c', 'c'], key), ['c']);
    assert.deepEqual(dedupe([], key), []);
    // nothing carries over past an empty window
    assert.deepEqual(dedupe(['c'], key), ['c']);
});

test('classifies transient request errors', () => {
    assert.equal(isTransientRequestError(new Error('timeout of 60000ms exceeded')), true);
    assert.equal(isTransientRequestError({ code: 'ERR_NETWORK', message: 'Network Error' }), true);
    assert.equal(isTransientRequestError({ response: { status: 502 } }), true);
    assert.equal(isTransientRequestError({ response: { status: 429 } }), true);
    assert.equal(isTransientRequestError({ response: { status: 403 } }), false);
    assert.equal(isTransientRequestError({ code: 500, message: 'record not found' }), false);
});
