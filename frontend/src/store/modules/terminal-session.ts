import { ref, reactive, shallowReactive, markRaw, nextTick } from 'vue';
import { defineStore } from 'pinia';
import { newUUID } from '@/utils/id';
import { searchTerminalSessions } from '@/api/modules/terminal';
import { useGlobalStore } from '@/composables/useGlobalStore';
import i18n from '@/lang';

export interface TerminalSessionEntry {
    key: string;
    title: string;
    wsID: number;
    endpoint: string;
    args: string;
    sessionId: string;
    status: 'online' | 'closed';
    latency: number;
    refresh: number;
}

const localEndpoint = '/api/v2/hosts/terminal/local';
const sshEndpoint = '/api/v2/hosts/terminal/ssh';

const TerminalSessionStore = defineStore('TerminalSessionStore', () => {
    const entries = ref<TerminalSessionEntry[]>([]);
    const instances = reactive<Record<string, any>>({});
    const slots = shallowReactive<Record<string, HTMLElement | undefined>>({});

    const find = (key: string) => entries.value.find((e) => e.key === key);

    const instanceOf = async (key: string) => {
        for (let i = 0; i < 5 && !instances[key]; i++) {
            await nextTick();
        }
        return instances[key];
    };

    const add = (init: {
        title: string;
        wsID: number;
        nodeName?: string;
        args?: string;
        status?: 'online' | 'closed';
    }) => {
        const key = newUUID();
        const q = `title=${encodeURIComponent(init.title)}`;
        let title = init.title;
        let args = init.args || '';
        if (init.wsID === 0) {
            const { currentNode } = useGlobalStore();
            const node =
                /operateNode=([^&]+)/.exec(args)?.[1] ||
                encodeURIComponent(init.nodeName || currentNode.value || 'local');
            if (!args.includes('operateNode=')) args = [args, `operateNode=${node}`].filter(Boolean).join('&');
            if (node !== 'local' && !init.nodeName) title = `${title} (${decodeURIComponent(node)})`;
        }
        entries.value.push({
            key,
            title,
            wsID: init.wsID,
            endpoint: init.wsID === 0 ? localEndpoint : sshEndpoint,
            args: [init.wsID === 0 ? '' : `id=${init.wsID}`, args, q].filter(Boolean).join('&'),
            sessionId: '',
            status: init.status || 'online',
            latency: 0,
            refresh: 0,
        });
        return key;
    };

    const open = async (init: { title: string; wsID: number; nodeName?: string; initCmd?: string; error?: string }) => {
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

    const reconnect = async (key: string, error = '', initCmd = '') => {
        const e = find(key);
        if (!e) return;
        e.refresh++;
        await nextTick();
        const inst = await instanceOf(key);
        inst?.acceptParams({ endpoint: e.endpoint, args: e.args, initCmd, sessionId: e.sessionId, error });
    };

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
                if (s.kind !== 'local' && s.kind !== 'ssh') continue;
                if (s.attached || entries.value.some((e) => e.sessionId === s.id)) continue;
                if (s.hostId > 0 && !fromLocalNode) continue;
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

    const removeWhere = (match: (e: TerminalSessionEntry) => boolean) => {
        for (const e of entries.value.filter(match)) {
            delete instances[e.key];
            delete slots[e.key];
        }
        entries.value = entries.value.filter((e) => !match(e));
    };
    const remove = (key: string) => removeWhere((e) => e.key === key);
    const closeAll = () => removeWhere(() => true);

    const setSessionId = (key: string, id: string) => {
        const e = find(key);
        if (!e) return;
        e.sessionId = id;
        e.status = 'online';
    };

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
