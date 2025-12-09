import { defineStore } from 'pinia';
import { ref, reactive } from 'vue';

export interface PsSearch {
    type: 'ps';
    pid: number | undefined;
    username: string;
    name: string;
}

export interface NetSearch {
    type: 'net';
    processID: number | undefined;
    processName: string;
    port: number | undefined;
}

export const ProcessStore = defineStore('ProcessStore', () => {
    let websocket: WebSocket | null = null;
    let pollingTimer: ReturnType<typeof setInterval> | null = null;
    let disconnectTimer: ReturnType<typeof setTimeout> | null = null;

    let connectionRefCount = 0;

    const isConnected = ref(false);
    const isConnecting = ref(false);

    const psData = ref<any[]>([]);
    const psLoading = ref(false);
    const psSearch = reactive<PsSearch>({
        type: 'ps',
        pid: undefined,
        username: '',
        name: '',
    });

    const netData = ref<any[]>([]);
    const netLoading = ref(false);
    const netSearch = reactive<NetSearch>({
        type: 'net',
        processID: undefined,
        processName: '',
        port: undefined,
    });

    let pendingRequestType: 'ps' | 'net' | null = null;

    let queuedRequestType: 'ps' | 'net' | null = null;

    const isPsFetching = ref(false);
    const isNetFetching = ref(false);

    const activePollingType = ref<'ps' | 'net' | null>(null);

    const isWsOpen = () => {
        return websocket && websocket.readyState === WebSocket.OPEN;
    };

    const onOpen = () => {
        isConnected.value = true;
        isConnecting.value = false;
    };

    const doSendMessage = (type: 'ps' | 'net') => {
        pendingRequestType = type;

        if (type === 'ps') {
            isPsFetching.value = true;
            psLoading.value = psData.value.length === 0;

            const searchParams = { ...psSearch };
            if (typeof searchParams.pid === 'string') {
                searchParams.pid = Number(searchParams.pid);
            }
            websocket!.send(JSON.stringify(searchParams));
        } else {
            isNetFetching.value = true;
            netLoading.value = netData.value.length === 0;

            const searchParams = { ...netSearch };
            if (typeof searchParams.processID === 'string') {
                searchParams.processID = Number(searchParams.processID);
            }
            if (typeof searchParams.port === 'string') {
                searchParams.port = Number(searchParams.port);
            }
            websocket!.send(JSON.stringify(searchParams));
        }
    };

    const onMessage = (event: MessageEvent) => {
        try {
            const data = JSON.parse(event.data);
            const responseType = pendingRequestType;

            if (pendingRequestType === 'ps') {
                isPsFetching.value = false;
            } else if (pendingRequestType === 'net') {
                isNetFetching.value = false;
            }
            pendingRequestType = null;

            if (responseType === activePollingType.value) {
                if (responseType === 'ps') {
                    psData.value = data || [];
                    psLoading.value = false;
                } else if (responseType === 'net') {
                    netData.value = data || [];
                    netLoading.value = false;
                }
            }

            if (queuedRequestType && isWsOpen()) {
                const typeToSend = queuedRequestType;
                queuedRequestType = null;
                doSendMessage(typeToSend);
            }
        } catch (e) {
            console.error('Failed to parse WebSocket message:', e);
        }
    };

    const onError = () => {
        console.error('WebSocket error');
    };

    const onClose = () => {
        isConnected.value = false;
        isConnecting.value = false;
        websocket = null;
    };

    const initWebSocket = (currentNode: string) => {
        if (websocket || isConnecting.value) {
            return;
        }

        isConnecting.value = true;

        const href = window.location.href;
        const protocol = href.split('//')[0] === 'http:' ? 'ws' : 'wss';
        const ipLocal = href.split('//')[1].split('/')[0];

        websocket = new WebSocket(`${protocol}://${ipLocal}/api/v2/process/ws?operateNode=${currentNode}`);
        websocket.onopen = onOpen;
        websocket.onmessage = onMessage;
        websocket.onerror = onError;
        websocket.onclose = onClose;
    };

    const closeWebSocket = () => {
        stopPolling();

        if (websocket) {
            websocket.close();
            websocket = null;
        }

        isConnected.value = false;
        isConnecting.value = false;
    };

    const connect = (currentNode: string) => {
        if (disconnectTimer) {
            clearTimeout(disconnectTimer);
            disconnectTimer = null;
        }

        connectionRefCount++;

        if (!websocket && !isConnecting.value) {
            initWebSocket(currentNode);
        }
    };

    const disconnect = () => {
        connectionRefCount = Math.max(0, connectionRefCount - 1);

        if (connectionRefCount === 0) {
            disconnectTimer = setTimeout(() => {
                if (connectionRefCount === 0) {
                    closeWebSocket();
                }
            }, 500);
        }
    };

    const sendPsMessage = () => {
        if (!isWsOpen()) {
            return;
        }

        if (pendingRequestType !== null) {
            queuedRequestType = 'ps';
            return;
        }

        if (isPsFetching.value) {
            return;
        }

        doSendMessage('ps');
    };

    const sendNetMessage = () => {
        if (!isWsOpen()) {
            return;
        }

        if (pendingRequestType !== null) {
            queuedRequestType = 'net';
            return;
        }

        if (isNetFetching.value) {
            return;
        }

        doSendMessage('net');
    };

    const startPolling = (type: 'ps' | 'net', interval = 3000, initialDelay = 0) => {
        stopPolling();
        activePollingType.value = type;

        const sendInitial = () => {
            if (type === 'ps') {
                sendPsMessage();
            } else {
                sendNetMessage();
            }
        };

        const scheduleInitialFetch = () => {
            if (initialDelay > 0) {
                setTimeout(sendInitial, initialDelay);
            } else {
                sendInitial();
            }
        };

        if (isWsOpen()) {
            scheduleInitialFetch();
        } else {
            const checkConnection = setInterval(() => {
                if (isWsOpen()) {
                    clearInterval(checkConnection);
                    scheduleInitialFetch();
                }
            }, 100);
            setTimeout(() => clearInterval(checkConnection), 5000);
        }

        pollingTimer = setInterval(() => {
            if (type === 'ps') {
                sendPsMessage();
            } else {
                sendNetMessage();
            }
        }, interval);
    };

    const stopPolling = () => {
        if (pollingTimer) {
            clearInterval(pollingTimer);
            pollingTimer = null;
        }
        activePollingType.value = null;
    };

    const updatePsSearch = (params: Partial<Omit<PsSearch, 'type'>>) => {
        Object.assign(psSearch, params);
    };

    const updateNetSearch = (params: Partial<Omit<NetSearch, 'type'>>) => {
        Object.assign(netSearch, params);
    };

    const resetPsSearch = () => {
        psSearch.pid = undefined;
        psSearch.username = '';
        psSearch.name = '';
    };

    const resetNetSearch = () => {
        netSearch.processID = undefined;
        netSearch.processName = '';
        netSearch.port = undefined;
    };

    return {
        isConnected,
        isConnecting,
        psData,
        psLoading,
        psSearch,
        netData,
        netLoading,
        netSearch,
        isPsFetching,
        isNetFetching,
        activePollingType,

        isWsOpen,
        connect,
        disconnect,
        initWebSocket,
        closeWebSocket,
        sendPsMessage,
        sendNetMessage,
        startPolling,
        stopPolling,
        updatePsSearch,
        updateNetSearch,
        resetPsSearch,
        resetNetSearch,
    };
});

export default ProcessStore;
