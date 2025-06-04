<template>
    <div>
        <RouterMenu />
        <LayoutContent :title="'Servers'" v-loading="loading">
            <template #toolbar>
                <div class="flex flex-wrap gap-3">
                    <el-button type="primary" @click="openCreate">
                        {{ $t('aiTools.mcp.create') }}
                    </el-button>
                    <el-button type="primary" plain @click="openDomain">
                        {{ $t('aiTools.mcp.bindDomain') }}
                    </el-button>
                </div>
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
                            {{ row.baseUrl + row.ssePath }}
                            <CopyButton :content="row.baseUrl + row.ssePath" type="icon" />
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
                        :ellipsis="mobile ? 0 : 2"
                        :min-width="mobile ? 'auto' : 200"
                        :buttons="buttons"
                        :label="$t('commons.table.operate')"
                        fixed="right"
                        fix
                    />
                </ComplexTable>
            </template>
        </LayoutContent>
        <McpServerOperate ref="createRef" @close="searchWithTimeOut" />
        <OpDialog ref="opRef" @search="search" />
        <ComposeLogs ref="composeLogRef" />
        <BindDomain ref="bindDomainRef" @close="searchWithTimeOut" />
        <Config ref="configRef" />
    </div>
</template>

<script lang="ts" setup>
import { AI } from '@/api/interface/ai';
import { deleteMcpServer, operateMcpServer, pageMcpServer } from '@/api/modules/ai';
import RouterMenu from '@/views/ai/mcp/index.vue';
import { computed, onMounted, reactive, ref } from 'vue';
import { dateFormat } from '@/utils/util';
import McpServerOperate from './operate/index.vue';
import ComposeLogs from '@/components/log/compose/index.vue';
import { GlobalStore } from '@/store';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import BindDomain from './bind/index.vue';
import Config from './config/index.vue';
const globalStore = GlobalStore();

const loading = ref(false);
const createRef = ref();
const opRef = ref();
const composeLogRef = ref();
const bindDomainRef = ref();
const configRef = ref();
const items = ref<AI.McpServer[]>([]);
const paginationConfig = reactive({
    cacheSizeKey: 'mcp-server-page-size',
    currentPage: 1,
    pageSize: 10,
    total: 0,
});
const mobile = computed(() => {
    return globalStore.isMobile();
});

const buttons = [
    {
        label: i18n.global.t('menu.config'),
        click: (row: AI.McpServer) => {
            openConfig(row);
        },
    },
    {
        label: i18n.global.t('commons.button.edit'),
        click: (row: AI.McpServer) => {
            openDetail(row);
        },
    },
    {
        label: i18n.global.t('commons.button.start'),
        click: (row: AI.McpServer) => {
            opServer(row, 'start');
        },
        disabled: (row: AI.McpServer) => {
            return row.status === 'Running';
        },
    },
    {
        label: i18n.global.t('commons.button.stop'),
        click: (row: AI.McpServer) => {
            opServer(row, 'stop');
        },
        disabled: (row: AI.McpServer) => {
            return row.status === 'Stopped';
        },
    },
    {
        label: i18n.global.t('commons.button.restart'),
        click: (row: AI.McpServer) => {
            opServer(row, 'restart');
        },
    },
    {
        label: i18n.global.t('commons.button.delete'),
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
    });
};

const openDetail = (row: AI.McpServer) => {
    createRef.value.acceptParams(row);
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
