<template>
    <div>
        <RouterMenu />
        <LayoutContent :title="$t('menu.mcp')" v-loading="loading">
            <template #leftToolBar>
                <div class="flex flex-wrap gap-3">
                    <el-button v-permission type="primary" @click="openCreate">
                        {{ $t('commons.button.create') }}
                    </el-button>
                    <el-button v-permission type="primary" plain @click="openDomain">
                        {{ $t('aiTools.mcp.bindDomain') }}
                    </el-button>
                </div>
            </template>
            <template #rightToolBar>
                <TableRefresh @search="search()" />
            </template>
            <template #main>
                <ComplexTable :pagination-config="paginationConfig" :data="items" @search="search()">
                    <el-table-column
                        :label="$t('commons.table.name')"
                        fix
                        prop="name"
                        width="200px"
                        show-overflow-tooltip
                    >
                        <template #default="{ row }">
                            <el-text type="primary" class="cursor-pointer" @click="openDetail(row)">
                                {{ row.name }}
                            </el-text>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('aiTools.mcp.externalUrl')" prop="baseUrl" min-width="200px">
                        <template #default="{ row }">
                            {{ getUrl(row) }}
                            <CopyButton :content="getUrl(row)" />
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('commons.table.status')" prop="status" width="120px">
                        <template #default="{ row }">
                            <el-popover
                                v-if="row.status === 'error'"
                                placement="bottom"
                                :width="400"
                                trigger="hover"
                                :content="row.message"
                                popper-class="max-h-[300px] overflow-auto"
                            >
                                <template #reference>
                                    <Status :key="row.status" :status="row.status"></Status>
                                </template>
                            </el-popover>
                            <div v-else>
                                <Status :key="row.status" :status="row.status"></Status>
                            </div>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('commons.button.log')" prop="path" width="120px">
                        <template #default="{ row }">
                            <el-button
                                @click="openLog(row)"
                                link
                                type="primary"
                                :disabled="
                                    row.status !== 'Running' && row.status !== 'Rrror' && row.status !== 'Restarting'
                                "
                            >
                                {{ $t('website.check') }}
                            </el-button>
                        </template>
                    </el-table-column>
                    <el-table-column
                        prop="createdAt"
                        :label="$t('commons.table.date')"
                        :formatter="dateFormat"
                        show-overflow-tooltip
                        width="180"
                        fix
                    />
                    <fu-table-operations
                        :ellipsis="isMobile ? 0 : 2"
                        :min-width="isMobile ? 'auto' : 200"
                        :buttons="buttons"
                        :label="$t('commons.table.operate')"
                        fixed="right"
                        fix
                    />
                </ComplexTable>
            </template>
        </LayoutContent>
        <McpServerOperate ref="createRef" @close="searchWithTimeOut" @task="openTaskLog" />
        <OpDialog ref="opRef" @search="search" />
        <ComposeLogs ref="composeLogRef" />
        <BindDomain ref="bindDomainRef" @close="searchWithTimeOut" />
        <Config ref="configRef" />
        <TaskLog ref="taskLogRef" width="70%" @close="search" />
    </div>
</template>

<script lang="ts" setup>
import { AI } from '@/api/interface/ai';
import {
    deleteMcpServer,
    loadMcpServerDetail,
    operateMcpServer,
    pageMcpServer,
    syncMcpServerStatus,
    testMcpServerConnection,
} from '@/api/modules/ai';
import RouterMenu from '@/views/ai/mcp/index.vue';
import { onMounted, reactive, ref } from 'vue';
import { dateFormat } from '@/utils/date';
import McpServerOperate from './operate/index.vue';
import ComposeLogs from '@/components/log/compose/index.vue';
import TaskLog from '@/components/log/task/index.vue';
import i18n from '@/lang';
import { MsgError, MsgSuccess } from '@/utils/message';
import BindDomain from './bind/index.vue';
import Config from './config/index.vue';
import { useGlobalStore } from '@/composables/useGlobalStore';

const { isMobile } = useGlobalStore();
const loading = ref(false);
const createRef = ref();
const opRef = ref();
const composeLogRef = ref();
const taskLogRef = ref();
const bindDomainRef = ref();
const configRef = ref();
const items = ref<AI.McpServer[]>([]);
const paginationConfig = reactive({
    cacheSizeKey: 'mcp-server-page-size',
    currentPage: 1,
    pageSize: Number(localStorage.getItem('mcp-server-page-size')) || 20,
    total: 0,
});

const getUrl = (row: AI.McpServer) => {
    if (row.outputTransport == 'sse') {
        return row.baseUrl + row.ssePath;
    } else {
        return row.baseUrl + row.streamableHttpPath;
    }
};

const buttons = [
    {
        label: i18n.global.t('menu.config'),
        permission: true,
        click: (row: AI.McpServer) => {
            openConfig(row);
        },
    },
    {
        label: i18n.global.t('commons.button.edit'),
        permission: true,
        click: (row: AI.McpServer) => {
            openDetail(row);
        },
    },
    {
        label: i18n.global.t('commons.button.start'),
        permission: true,
        click: (row: AI.McpServer) => {
            opServer(row, 'start');
        },
        disabled: (row: AI.McpServer) => {
            return row.status === 'Running';
        },
    },
    {
        label: i18n.global.t('commons.button.stop'),
        permission: true,
        click: (row: AI.McpServer) => {
            opServer(row, 'stop');
        },
        disabled: (row: AI.McpServer) => {
            return row.status === 'Stopped';
        },
    },
    {
        label: i18n.global.t('commons.button.restart'),
        permission: true,
        click: (row: AI.McpServer) => {
            opServer(row, 'restart');
        },
    },
    {
        label: i18n.global.t('aiTools.mcp.testConnection'),
        permission: true,
        click: (row: AI.McpServer) => {
            testConnection(row);
        },
    },
    {
        label: i18n.global.t('commons.button.delete'),
        permission: true,
        click: (row: AI.McpServer) => {
            deleteServer(row);
        },
    },
];

const searchWithTimeOut = () => {
    search();
    setTimeout(() => {
        search();
    }, 1000);
};

const search = () => {
    loading.value = true;
    pageMcpServer({
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
        name: '',
    }).then((res) => {
        items.value = res.data.items;
        paginationConfig.total = res.data.total;
        loading.value = false;
        syncStatus();
    });
};

const syncStatus = async () => {
    const ids = items.value.map((item) => item.id).filter(Boolean);
    if (ids.length === 0) {
        return;
    }
    try {
        const res = await syncMcpServerStatus({ ids });
        const statusMap = new Map(res.data.map((item) => [item.id, item]));
        for (const item of items.value) {
            const status = statusMap.get(item.id);
            if (status) {
                item.status = status.status;
                item.message = status.message;
            }
        }
    } catch (error) {}
};

const openDetail = async (row: AI.McpServer) => {
    loading.value = true;
    try {
        const res = await loadMcpServerDetail({ id: row.id });
        createRef.value.acceptParams(res.data);
    } finally {
        loading.value = false;
    }
};

const openCreate = () => {
    let maxPort = 7999;
    if (items.value && items.value.length > 0) {
        maxPort = Math.max(...items.value.map((item) => item.port));
    }
    createRef.value.acceptParams({ port: maxPort + 1 });
};

const openLog = (row: AI.McpServer) => {
    composeLogRef.value.acceptParams({
        compose: row.dir + '/docker-compose.yml',
        resource: row.name,
        container: row.containerName,
    });
};

const openTaskLog = (taskID: string) => {
    taskLogRef.value.openWithTaskID(taskID);
};

const deleteServer = async (row: AI.McpServer) => {
    try {
        opRef.value.acceptParams({
            title: i18n.global.t('commons.button.delete'),
            names: [row.name],
            msg: i18n.global.t('commons.msg.operatorHelper', [
                i18n.global.t('aiTools.mcp.server'),
                i18n.global.t('commons.button.delete'),
            ]),
            api: deleteMcpServer,
            params: { id: row.id },
        });
    } catch (error) {}
};

const opServer = async (row: AI.McpServer, operate: string) => {
    ElMessageBox.confirm(
        i18n.global.t('aiTools.mcp.operatorHelper', [
            i18n.global.t('aiTools.mcp.server'),
            i18n.global.t('commons.button.' + operate),
        ]),
        i18n.global.t('commons.button.' + operate),
        {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
            type: 'info',
        },
    ).then(async () => {
        try {
            await operateMcpServer({ id: row.id, operate: operate });
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            search();
        } catch (error) {}
    });
};

const testConnection = async (row: AI.McpServer) => {
    loading.value = true;
    try {
        const res = await testMcpServerConnection({ id: row.id });
        if (res.data.success) {
            MsgSuccess(res.data.message || i18n.global.t('aiTools.mcp.connectionSuccess'));
            return;
        }
        MsgError(res.data.message || i18n.global.t('aiTools.mcp.connectionFailed'));
    } finally {
        loading.value = false;
    }
};

const openDomain = () => {
    bindDomainRef.value.acceptParams();
};

const openConfig = (row: AI.McpServer) => {
    configRef.value.acceptParams(row);
};

onMounted(() => {
    search();
});
</script>
