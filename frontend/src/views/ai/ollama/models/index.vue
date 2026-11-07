<template>
    <div v-loading="loading || initializing">
        <LayoutContent :title="$t('aiTools.model.ollamaManagement')">
            <template #app>
                <AppStatus
                    ref="appStatusRef"
                    app-key="ollama"
                    :hide-setting="true"
                    v-model:loading="loading"
                    @is-exist="checkExist"
                />
            </template>
            <template #leftToolBar>
                <el-button v-permission :disabled="modelInfo.status !== 'Running'" type="primary" @click="onCreate">
                    {{ $t('commons.button.add') }}
                </el-button>
                <el-button v-permission :disabled="modelInfo.status !== 'Running'" @click="bindDomain">
                    {{ $t('aiTools.proxy.proxy') }}
                </el-button>
                <el-button :disabled="modelInfo.status !== 'Running'" @click="onLoadConn">
                    {{ $t('database.databaseConnInfo') }}
                </el-button>
                <el-button :disabled="modelInfo.status !== 'Running'" icon="Position" @click="openWebUI">
                    OpenWebUI
                </el-button>
                <el-button v-permission :disabled="selects.length === 0" @click="onDelete(null)">
                    {{ $t('commons.button.delete') }}
                </el-button>
            </template>
            <template #rightToolBar>
                <TableSearch @search="search" v-model:searchName="searchName" />
                <TableRefresh @search="refresh" />
                <TableSetting title="model-refresh" @search="!loading && !initializing && search()" />
            </template>
            <template #main>
                <el-alert v-if="loadFailed" :title="$t('aiTools.model.loadFailed')" type="error" :closable="false" />
                <el-empty
                    v-else-if="!initializing && !modelInstalled"
                    :description="$t('app.checkInstalledWarn', ['Ollama'])"
                >
                    <el-button
                        v-permission
                        type="primary"
                        @click="routerToNameWithQuery('AppAll', { install: 'ollama' })"
                    >
                        {{ $t('commons.button.install') }}
                    </el-button>
                </el-empty>
                <ComplexTable
                    v-else-if="modelInstalled"
                    :selection-context="() => [searchName]"
                    row-key="id"
                    :cache-pagination="false"
                    :pagination-config="paginationConfig"
                    v-model:selects="selects"
                    @sort-change="search"
                    @search="search"
                    :data="data"
                >
                    <el-table-column type="selection" :selectable="selectable" fix />
                    <el-table-column :label="$t('aiTools.model.model')" prop="name" min-width="180">
                        <template #default="{ row }">
                            <el-text
                                v-if="row.status === 'Success' && modelInfo.status === 'Running'"
                                type="primary"
                                class="cursor-pointer"
                                @click="onLoad(row.name)"
                            >
                                {{ row.name }}
                            </el-text>
                            <span v-else>{{ row.name }}</span>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('file.size')" prop="size">
                        <template #default="{ row }">
                            <span>{{ row.size || '-' }}</span>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('commons.table.status')" prop="status">
                        <template #default="{ row }">
                            <el-tag v-if="row.status === 'Success'" type="success">
                                {{ $t('commons.status.success') }}
                            </el-tag>
                            <el-tag v-if="row.status === 'Deleted'" type="info">
                                {{ $t('commons.status.notfound') }}
                            </el-tag>
                            <el-tooltip
                                v-if="row.status === 'Unknown'"
                                :content="row.message"
                                v-bind="TABLE_TOOLTIP_OPTIONS"
                            >
                                <el-tag type="warning">{{ $t('commons.status.unknown') }}</el-tag>
                            </el-tooltip>
                            <el-tooltip
                                v-if="row.status === 'Canceled'"
                                :content="row.message"
                                :disabled="!row.message"
                                v-bind="TABLE_TOOLTIP_OPTIONS"
                            >
                                <el-tag type="danger">{{ $t('commons.status.systemrestart') }}</el-tag>
                            </el-tooltip>
                            <el-tag v-if="row.status === 'Failed'" type="danger">
                                {{ $t('commons.status.failed') }}
                            </el-tag>
                            <el-tag v-if="row.status === 'Waiting'">
                                <el-icon v-if="row.status === 'Waiting'" class="is-loading">
                                    <Loading />
                                </el-icon>
                                {{ $t('commons.status.waiting') }}
                            </el-tag>
                        </template>
                    </el-table-column>
                    <el-table-column
                        min-width="170"
                        :label="$t('commons.table.date')"
                        prop="createdAt"
                        :formatter="dateFormat"
                    />
                    <fu-table-operations
                        :ellipsis="isMobile ? 0 : 1"
                        :width="isMobile ? 'auto' : 110"
                        :buttons="buttons"
                        :label="$t('commons.table.operate')"
                        fixed="right"
                        fix
                    />
                </ComplexTable>
            </template>
        </LayoutContent>

        <OpDialog ref="opRef" @search="search" @submit="onSubmitDelete()">
            <template #content>
                <el-form class="mt-4 mb-1" ref="deleteForm" label-position="left">
                    <el-form-item>
                        <el-checkbox v-model="forceDelete" :label="$t('website.forceDelete')" />
                        <span class="input-help">
                            {{ $t('website.forceDeleteHelper') }}
                        </span>
                    </el-form-item>
                </el-form>
            </template>
        </OpDialog>
        <AddDialog ref="addRef" @search="search" @log="onLoadLog" />
        <Terminal ref="terminalRef" />
        <DrawerPro v-model="detailVisible" :header="$t('commons.button.view')" size="large">
            <template #content>
                <CodemirrorPro v-model="detailInfo" :height-diff="180" readonly mode="json" />
            </template>
        </DrawerPro>
        <BindDomain ref="bindDomainRef" />
        <Connection ref="connectionRef" />
        <PortJumpDialog ref="portRef" />
        <TaskLog ref="taskLogRef" width="50%" @close="refresh" />
    </div>
</template>

<script lang="ts" setup>
import { TABLE_TOOLTIP_OPTIONS } from '@/components/table/shared';
import { useSearchState } from '@/utils/search-reset';
import AddDialog from '@/views/ai/ollama/add/index.vue';
import TaskLog from '@/components/log/task/index.vue';
import Terminal from '@/views/ai/ollama/terminal/index.vue';
import { toRef, reactive, ref } from 'vue';
import { ElMessageBox } from 'element-plus';
import AppStatus from '@/components/app-status/index.vue';
import Connection from '@/views/ai/components/connection/index.vue';
import PortJumpDialog from '@/components/port-jump/index.vue';
import { App } from '@/api/interface/app';
import i18n from '@/lang';
import { deleteOllamaModel, loadOllamaModel, recreateOllamaModel, searchOllamaModel } from '@/api/modules/ai';
import { AI } from '@/api/interface/ai';
import { getAppPort } from '@/api/modules/app';
import { dateFormat } from '@/utils/date';
import { newUUID } from '@/utils/id';
import { MsgInfo, MsgSuccess } from '@/utils/message';
import BindDomain from '@/views/ai/ollama/domain/index.vue';
import { useGlobalStore } from '@/composables/useGlobalStore';
import { routerToNameWithQuery } from '@/utils/router';

const { isMobile, currentNode } = useGlobalStore();
const appStatusRef = ref();
const connectionRef = ref();
const portRef = ref();
const loading = ref(false);
const initializing = ref(true);
const loadFailed = ref(false);
const modelInstalled = ref(false);
const selects = ref<any>([]);
const addRef = ref();
const detailVisible = ref(false);
const detailInfo = ref('');
const terminalRef = ref();
const bindDomainRef = ref();
const taskLogRef = ref();
const data = ref();
const paginationConfig = reactive({
    currentPage: 1,
    pageSize: 20,
    total: 0,
});
const searchName = ref();
useSearchState(
    'ai/model/ollama/index',
    { search: searchName, page: toRef(paginationConfig, 'currentPage'), pageSize: toRef(paginationConfig, 'pageSize') },
    { onRevisit: () => refresh() },
);
const appInstallID = ref(0);

const opRef = ref();
const operateIDs = ref();
const forceDelete = ref();

const modelInfo = reactive({ status: '', name: '', version: '', containerName: '', httpPort: 0 });

function selectable(row) {
    return row.status !== 'Waiting';
}

const search = async () => {
    if (!modelInstalled.value) return;
    let params = {
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
        info: searchName.value,
    };
    loading.value = true;
    await searchOllamaModel(params)
        .then((res) => {
            loading.value = false;
            data.value = res.data.items || [];
            paginationConfig.total = res.data.total;
        })
        .catch(() => {
            data.value = [];
            paginationConfig.total = 0;
            loading.value = false;
        });
};

const onCreate = async () => {
    addRef.value.acceptParams();
};

const onLoad = async (name: string) => {
    const res = await loadOllamaModel(name);
    detailInfo.value = res.data;
    detailVisible.value = true;
};

const bindDomain = () => {
    bindDomainRef.value.acceptParams(appInstallID.value);
};

const onSubmitDelete = async () => {
    loading.value = true;
    await deleteOllamaModel(operateIDs.value, forceDelete.value)
        .then(() => {
            loading.value = false;
            MsgSuccess(i18n.global.t('commons.msg.deleteSuccess'));
            search();
        })
        .catch(() => {
            loading.value = false;
        });
};

const onReCreate = async (name: string) => {
    loading.value = true;
    let taskID = newUUID();
    await recreateOllamaModel(name, taskID)
        .then(() => {
            loading.value = false;
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            openTaskLog(taskID);
            search();
        })
        .catch(() => {
            loading.value = false;
        });
};
const openTaskLog = (taskID: string) => {
    taskLogRef.value.openWithTaskID(taskID);
};

const onDelete = async (row: AI.OllamaModelInfo) => {
    let names = [];
    let ids = [];
    if (row) {
        ids = [row.id];
        names = [row.name];
    } else {
        for (const item of selects.value) {
            names.push(item.name);
            ids.push(item.id);
        }
    }
    operateIDs.value = ids;
    opRef.value.acceptParams({
        title: i18n.global.t('commons.button.delete'),
        names: names,
        msg: i18n.global.t('commons.msg.operatorHelper', [
            i18n.global.t('aiTools.model.model'),
            i18n.global.t('commons.button.delete'),
        ]),
        api: null,
        params: null,
    });
};

const onLoadLog = (row: any) => {
    if (row.taskID) {
        openTaskLog(row.taskID);
        return;
    }
    if (row.from === 'remote') {
        MsgInfo(i18n.global.t('aiTools.model.from_remote'));
        return;
    }
    if (!row.logFileExist) {
        MsgInfo(i18n.global.t('aiTools.model.no_logs'));
        return;
    }

    taskLogRef.value.openWithResourceID('AI', 'TaskPull', row.id);
};

const buttons = [
    {
        label: i18n.global.t('commons.button.run'),
        click: (row: AI.OllamaModelInfo) => {
            terminalRef.value.acceptParams({ name: row.name });
        },
        disabled: (row: any) => {
            return row.status !== 'Success' || modelInfo.status !== 'Running';
        },
    },
    {
        label: i18n.global.t('commons.button.retry'),
        permission: true,
        click: (row: AI.OllamaModelInfo) => {
            onReCreate(row.name);
        },
        disabled: (row: any) => {
            return !['Deleted', 'Failed', 'Canceled'].includes(row.status) || modelInfo.status !== 'Running';
        },
    },
    {
        label: i18n.global.t('commons.button.log'),
        click: onLoadLog,
    },
    {
        label: i18n.global.t('commons.button.delete'),
        permission: true,
        click: (row: AI.OllamaModelInfo) => {
            onDelete(row);
        },
        disabled: (row: any) => {
            return row.status === 'Waiting';
        },
    },
];

const onLoadConn = () => {
    connectionRef.value.acceptParams({
        name: modelInfo.name,
        status: modelInfo.status,
        version: modelInfo.version,
        engine: 'ollama',
        port: modelInfo.httpPort,
        containerName: modelInfo.containerName,
        appinstallID: appInstallID.value,
    });
};

const openWebUI = async () => {
    const node = currentNode.value;
    const { data: port } = await getAppPort('ollama-webui', '');
    if (node !== currentNode.value) return;
    if (port) {
        portRef.value.acceptParams({ port });
        return;
    }
    const confirmed = await ElMessageBox.confirm(i18n.global.t('app.checkInstalledWarn', ['OpenWebUI']), 'OpenWebUI', {
        confirmButtonText: i18n.global.t('database.goInstall'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
    }).catch(() => false);
    if (!confirmed || node !== currentNode.value) return;
    await routerToNameWithQuery('AppAll', { install: 'ollama-webui' });
};

const checkExist = async (installed: App.CheckInstalled | false) => {
    loadFailed.value = !installed;
    selects.value = [];
    if (!installed) {
        modelInfo.status = '';
        initializing.value = false;
        return;
    }
    appInstallID.value = installed.appInstallId;
    modelInstalled.value = installed.isExist;
    modelInfo.status = installed.status;
    modelInfo.name = installed.name;
    modelInfo.version = installed.version;
    modelInfo.containerName = installed.containerName;
    modelInfo.httpPort = installed.httpPort;
    if (installed.isExist) {
        await search();
    } else {
        data.value = [];
        paginationConfig.total = 0;
    }
    initializing.value = false;
};

const refresh = async () => {
    initializing.value = true;
    await appStatusRef.value?.onCheck('ollama', '');
};
</script>
