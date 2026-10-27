<template>
    <div v-loading="loading">
        <el-button v-permission :disabled="modelInfo.status !== 'Running'" @click="bindDomain">
            {{ $t('aiTools.proxy.proxy') }}
        </el-button>
        <el-divider />
        <div class="model-section-heading">
            <h3 class="section-title">{{ $t('aiTools.model.modelManagement') }}</h3>
        </div>
        <LayoutContent title="Ollama">
            <template #leftToolBar>
                <el-button v-permission :disabled="modelInfo.status !== 'Running'" type="primary" @click="onCreate">
                    {{ $t('commons.button.add') }}
                </el-button>
                <el-button v-permission :disabled="selects.length === 0" @click="onDelete(null)">
                    {{ $t('commons.button.delete') }}
                </el-button>
            </template>
            <template #rightToolBar>
                <TableSearch @search="search" v-model:searchName="searchName" />
                <TableRefresh @search="search" />
                <TableSetting title="model-refresh" @search="!loading && search()" />
            </template>
            <template #main>
                <ComplexTable
                    :selection-context="() => [searchName]"
                    row-key="id"
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
                            <el-tooltip v-if="row.status === 'Unknown'" :content="row.message" placement="top">
                                <el-tag type="warning">{{ $t('commons.status.unknown') }}</el-tag>
                            </el-tooltip>
                            <el-tag v-if="row.status === 'Canceled'" type="danger">
                                {{ $t('commons.status.systemrestart') }}
                            </el-tag>
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
        <DialogPro v-model="detailVisible" :title="$t('commons.button.view')" size="large">
            <CodemirrorPro v-model="detailInfo" :height-diff="300" readonly mode="json" />
        </DialogPro>
        <BindDomain ref="bindDomainRef" />
        <TaskLog ref="taskLogRef" width="70%" @close="refresh" />
    </div>
</template>

<script lang="ts" setup>
import { useSearchState } from '@/utils/search-reset';
import AddDialog from '@/views/ai/model/ollama/add/index.vue';
import TaskLog from '@/components/log/task/index.vue';
import Terminal from '@/views/ai/model/ollama/terminal/index.vue';
import { toRef, onMounted, reactive, ref } from 'vue';
import i18n from '@/lang';
import { deleteOllamaModel, loadOllamaModel, recreateOllamaModel, searchOllamaModel } from '@/api/modules/ai';
import { AI } from '@/api/interface/ai';
import { checkAppInstalled } from '@/api/modules/app';
import { dateFormat } from '@/utils/date';
import { newUUID } from '@/utils/id';
import { MsgInfo, MsgSuccess } from '@/utils/message';
import BindDomain from '@/views/ai/model/ollama/domain/index.vue';
import { useGlobalStore } from '@/composables/useGlobalStore';

const props = defineProps({ instanceName: String });
const { isMobile } = useGlobalStore();
const loading = ref(false);
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
    { onRevisit: () => search() },
);
const appInstallID = ref(0);

const opRef = ref();
const operateIDs = ref();
const forceDelete = ref();

const modelInfo = reactive({ status: '' });

function selectable(row) {
    return row.status !== 'Waiting';
}

const search = async () => {
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

const refresh = async () => {
    const { data } = await checkAppInstalled('ollama', props.instanceName || '');
    appInstallID.value = data.appInstallId;
    modelInfo.status = data.status;
    await search();
};
onMounted(refresh);
</script>

<style lang="scss" scoped>
.section-title {
    margin: 0;
    font-size: 14px;
    font-weight: 600;
    color: var(--el-text-color-primary);
}
.model-section-heading {
    margin-bottom: 16px;
}
</style>
