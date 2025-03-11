<template>
    <div>
        <FireRouter />
        <LayoutContent :title="$t('ssh.session')">
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
import { GlobalStore } from '@/store';
const globalStore = GlobalStore();

const sshSearch = reactive({
    type: 'ssh',
    loginUser: '',
});

const buttons = [
    {
        label: i18n.global.t('commons.button.disConn'),
        click: function (row: any) {
            stop(row.PID);
        },
    },
];

let processSocket = ref(null) as unknown as WebSocket;
const data = ref([]);
const loading = ref(false);
const tableRef = ref();

const isWsOpen = () => {
    const readyState = processSocket && processSocket.readyState;
    return readyState === 1;
};
const closeSocket = () => {
    if (isWsOpen()) {
        processSocket && processSocket.close();
    }
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

const initProcess = () => {
    let href = window.location.href;
    let protocol = href.split('//')[0] === 'http:' ? 'ws' : 'wss';
    let ipLocal = href.split('//')[1].split('/')[0];
    let currentNode = globalStore.currentNode;
    processSocket = new WebSocket(`${protocol}://${ipLocal}/api/v2/process/ws?currentNode=${currentNode}`);
    processSocket.onopen = onOpenProcess;
    processSocket.onmessage = onMessage;
    processSocket.onerror = onerror;
    processSocket.onclose = onClose;
    search();
    sendMsg();
};

const sendMsg = () => {
    loading.value = true;
    setInterval(() => {
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
    closeSocket();
});
</script>
