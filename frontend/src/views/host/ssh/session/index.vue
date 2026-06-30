<template>
    <div>
        <FireRouter />
        <LayoutContent :title="$t('ssh.session', 2)">
            <template #rightToolBar>
                <TableSearch @search="search()" v-model:searchName="sshSearch.loginUser" />
            </template>
            <template #main>
                <ComplexTable :data="data" ref="tableRef" v-loading="loading" :heightDiff="260">
                    <el-table-column :label="$t('commons.table.user')" fix prop="username"></el-table-column>
                    <el-table-column :label="'TTY'" fix prop="terminal"></el-table-column>
                    <el-table-column :label="$t('ssh.loginIP')" fix prop="host"></el-table-column>
                    <el-table-column
                        :label="$t('ssh.loginTime')"
                        fix
                        prop="loginTime"
                        min-width="120px"
                    ></el-table-column>
                    <fu-table-operations :ellipsis="10" :buttons="buttons" :label="$t('commons.table.operate')" fix />
                </ComplexTable>
            </template>
        </LayoutContent>
    </div>
</template>

<script setup lang="ts">
import FireRouter from '@/views/host/ssh/index.vue';
import { ref, onMounted, onUnmounted, reactive } from 'vue';
import i18n from '@/lang';
import { stopProcess } from '@/api/modules/process';
import { MsgError, MsgSuccess } from '@/utils/message';
import { useGlobalStore } from '@/composables/useGlobalStore';
import { checkStreamAuth } from '@/utils/stream-auth';
const { currentNode: globalCurrentNode } = useGlobalStore();

const sshSearch = reactive({
    type: 'ssh',
    loginUser: '',
});

const buttons = [
    {
        label: i18n.global.t('commons.button.disConn'),
        permission: true,
        nodeAdmin: true,
        click: function (row: any) {
            stop(row.PID);
        },
    },
];

let processSocket: WebSocket | null = null;
let sendTimer: ReturnType<typeof setInterval> | null = null;
let initProcessToken = 0;
const data = ref([]);
const loading = ref(false);
const tableRef = ref();

const isWsOpen = () => {
    return processSocket?.readyState === WebSocket.OPEN;
};
const clearSendTimer = () => {
    if (sendTimer) {
        clearInterval(sendTimer);
        sendTimer = null;
    }
};
const closeSocket = () => {
    clearSendTimer();
    if (isWsOpen()) {
        processSocket.close();
    }
    processSocket = null;
};

const onOpenProcess = () => {};
const onMessage = (message: any) => {
    let result: any[] = JSON.parse(message.data);
    data.value = result;
    loading.value = false;
};

const onerror = () => {};
const onClose = () => {
    closeSocket();
};

const initProcess = async () => {
    const token = ++initProcessToken;
    let href = window.location.href;
    let protocol = href.split('//')[0] === 'http:' ? 'ws' : 'wss';
    let ipLocal = href.split('//')[1].split('/')[0];
    let currentNode = globalCurrentNode.value;
    const url = `${protocol}://${ipLocal}/api/v2/process/ws?operateNode=${currentNode}`;
    const authError = await checkStreamAuth(url, currentNode);
    if (token !== initProcessToken) {
        return;
    }
    if (authError) {
        loading.value = false;
        MsgError(authError);
        return;
    }
    closeSocket();
    processSocket = new WebSocket(url);
    processSocket.onopen = onOpenProcess;
    processSocket.onmessage = onMessage;
    processSocket.onerror = onerror;
    processSocket.onclose = onClose;
    search();
    sendMsg();
};

const sendMsg = () => {
    loading.value = true;
    clearSendTimer();
    sendTimer = setInterval(() => {
        search();
    }, 3000);
};

const search = () => {
    if (isWsOpen()) {
        processSocket.send(JSON.stringify(sshSearch));
    }
};

const stop = async (PID: number) => {
    ElMessageBox.confirm(i18n.global.t('ssh.stopSSHWarn'), i18n.global.t('commons.button.disConn'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
        type: 'info',
    })
        .then(async () => {
            try {
                await stopProcess({ PID: PID });
                data.value = data.value.filter((item: any) => item.PID !== PID);
                search();
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            } catch (error) {
                MsgError(error);
            }
        })
        .catch(() => {});
};

onMounted(() => {
    initProcess();
});

onUnmounted(() => {
    initProcessToken++;
    closeSocket();
});
</script>
