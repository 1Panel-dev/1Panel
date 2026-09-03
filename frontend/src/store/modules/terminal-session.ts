import { ref, reactive, shallowReactive, markRaw, nextTick } from 'vue';
import { defineStore } from 'pinia';
import { newUUID } from '@/utils/id';
import { searchTerminalSessions } from '@/api/modules/terminal';
import { useGlobalStore } from '@/composables/useGlobalStore';
import i18n from '@/lang';

// A web terminal session is bound to the browser tab, not to the route.
// Entries live here for as long as the SPA does; the Terminal components that
// own the websocket and xterm are rendered by components/terminal/host.vue and
// teleported into whichever slot (terminal page, floating dock) claims them.
// The agent keeps a session alive for a while after the websocket drops, so
// restore() can rebuild the entries after a page refresh or a closed browser tab.
export interface TerminalSessionEntry {
    key: string;
    title: string;
    wsID: number; // 0 = local shell
    endpoint: string;
    args: string;
    sessionId: string; // agent side id, known once the hello arrived
    status: 'online' | 'closed';
    latency: number;
    refresh: number; // bump to remount the Terminal component
}

const localEndpoint = '/api/v2/hosts/terminal/local';
const sshEndpoint = '/api/v2/hosts/terminal/ssh';

const TerminalSessionStore = defineStore('TerminalSessionStore', () => {
    const entries = ref<TerminalSessionEntry[]>([]);
    // Terminal component instances and slot elements, keyed by entry key.
    const instances = reactive<Record<string, any>>({});
    const slots = shallowReactive<Record<string, HTMLElement | undefined>>({});

    const find = (key: string) => entries.value.find((e) => e.key === key);

    // The Terminal components render in the layout level host; wait for it to render ours.
    const instanceOf = async (key: string) => {
        for (let i = 0; i < 5 && !instances[key]; i++) {
            await nextTick();
        }
        return instances[key];
    };

    const add = (init: { title: string; wsID: number; args?: string; status?: 'online' | 'closed' }) => {
        const key = newUUID();
        const q = `title=${encodeURIComponent(init.title)}`;
        entries.value.push({
            key,
            title: init.title,
            wsID: init.wsID,
            endpoint: init.wsID === 0 ? localEndpoint : sshEndpoint,
            args: [init.wsID === 0 ? '' : `id=${init.wsID}`, init.args || '', q].filter(Boolean).join('&'),
            sessionId: '',
            status: init.status || 'online',
            latency: 0,
            refresh: 0,
        });
        return key;
    };

    // open adds an entry and connects it. error is shown instead of connecting when set.
    const open = async (init: { title: string; wsID: number; initCmd?: string; error?: string }) => {
        const key = add({ ...init, status: init.error ? 'closed' : 'online' });
        const e = find(key)!;
        const inst = await instanceOf(key);
        inst?.acceptParams({
            endpoint: e.endpoint,
            args: e.args,
            initCmd: init.initCmd || '',
            error: init.error || '',
        });
        return key;
    };

    // reconnect remounts the Terminal; an entry that still has an agent session reattaches to it.
    const reconnect = async (key: string, error = '', initCmd = '') => {
        const e = find(key);
        if (!e) return;
        e.refresh++;
        await nextTick();
        const inst = await instanceOf(key);
        inst?.acceptParams({ endpoint: e.endpoint, args: e.args, initCmd, sessionId: e.sessionId, error });
    };

    // restore rebuilds entries from the agent's session list and reattaches them.
    const restore = async () => {
        const { currentNode } = useGlobalStore();
        const node = currentNode.value || 'local';
        const results = await Promise.allSettled([
            searchTerminalSessions(false),
            ...(node === 'local' ? [] : [searchTerminalSessions(true)]),
        ]);
        results.forEach((r, i) => {
            if (r.status !== 'fulfilled') return;
            const fromLocalNode = i === 1 || node === 'local';
            for (const s of r.value.data || []) {
                // attached elsewhere = another browser tab is using it; do not steal it
                if (s.attached || entries.value.some((e) => e.sessionId === s.id)) continue;
                if (s.hostId > 0 && !fromLocalNode) continue; // ssh sessions are served by the local node
                const key = add({
                    title: s.title || i18n.global.t('terminal.localhost'),
                    wsID: s.hostId,
                    args: s.hostId === 0 && i === 1 ? 'operateNode=local' : '',
                });
                find(key)!.sessionId = s.id;
            }
        });
        for (const e of entries.value) {
            if (e.sessionId) await reconnect(e.key);
        }
    };

    // remove drops entries; the host unmounts their Terminals, which closes the websockets with 1000.
    const removeWhere = (match: (e: TerminalSessionEntry) => boolean) => {
        for (const e of entries.value.filter(match)) {
            delete instances[e.key];
            delete slots[e.key];
        }
        entries.value = entries.value.filter((e) => !match(e));
    };
    const remove = (key: string) => removeWhere((e) => e.key === key);
    // closeAll runs on logout (or an expired login).
    const closeAll = () => removeWhere(() => true);

    const setSessionId = (key: string, id: string) => {
        const e = find(key);
        if (!e) return;
        e.sessionId = id;
        e.status = 'online';
    };

    // onExpired: the agent no longer has the session; the next reconnect opens a fresh one.
    const onExpired = (key: string) => {
        const e = find(key);
        if (!e) return;
        e.sessionId = '';
        e.status = 'closed';
    };

    const setInstance = (key: string, inst: any) => {
        if (inst) {
            instances[key] = markRaw(inst);
        } else {
            delete instances[key];
        }
    };

    const setSlot = (key: string, el: HTMLElement | null) => {
        slots[key] = el || undefined;
    };

    // sync pulls status/latency from the live components.
    const sync = () => {
        for (const e of entries.value) {
            const inst = instances[e.key];
            if (!inst) continue;
            e.status = inst.isWsOpen() ? 'online' : 'closed';
            e.latency = inst.getLatency();
        }
    };

    return {
        entries,
        instances,
        slots,
        find,
        open,
        reconnect,
        restore,
        remove,
        closeAll,
        setSessionId,
        onExpired,
        setInstance,
        setSlot,
        sync,
    };
});

export default TerminalSessionStore;
