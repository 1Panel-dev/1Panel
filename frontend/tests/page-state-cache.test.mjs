import assert from 'node:assert/strict';
import test from 'node:test';
import { clearPageStateCache, getPageState } from '../src/utils/page-state-cache.ts';

test('reuses state for the same page key', () => {
    clearPageStateCache();
    const first = getPageState('local:Website', () => ({ name: '' }));
    first.name = 'example.com';

    const second = getPageState('local:Website', () => ({ name: 'ignored' }));

    assert.equal(second.name, 'example.com');
    assert.equal(second, first);
});

test('isolates state by page key', () => {
    clearPageStateCache();
    const local = getPageState('local:Website', () => ({ name: 'local' }));
    const remote = getPageState('remote:Website', () => ({ name: 'remote' }));

    assert.notEqual(remote, local);
    assert.equal(remote.name, 'remote');
});

test('creates fresh state after clearing the cache', () => {
    const first = getPageState('local:Website', () => ({ page: 2 }));
    clearPageStateCache();

    const second = getPageState('local:Website', () => ({ page: 1 }));

    assert.notEqual(second, first);
    assert.equal(second.page, 1);
});
