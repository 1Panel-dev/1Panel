<template>
    <div v-loading="loading">
        <RouterButton
            :buttons="[
                {
                    label: i18n.global.t('ai_tools.model.model'),
                    path: '/ai-tools/model',
                },
            ]"
        />
        <LayoutContent title="Ollama">
            <template #app>
                <AppStatus
                    app-key="ollama"
                    v-model:loading="loading"
                    :hide-setting="true"
                    v-model:mask-show="maskShow"
                    v-model:appInstallID="appInstallID"
                    @is-exist="checkExist"
                    ref="appStatusRef"
                ></AppStatus>
            </template>
            <template #prompt>
                <el-alert type="info" :closable="false">
                    <template #title>
                        <span>{{ $t('runtime.systemRestartHelper') }}</span>
                    </template>
                </el-alert>
            </template>
            <template #toolbar v-if="modelInfo.isExist">
                <div class="flex justify-between gap-2 flex-wrap sm:flex-row">
                    <div class="flex flex-wrap gap-3">
                        <el-button :disabled="modelInfo.status !== 'Running'" type="primary" @click="onCreate()">
                            {{ $t('ai_tools.model.create') }}
                        </el-button>
                        <el-button plain type="primary" :disabled="modelInfo.status !== 'Running'" @click="bindDomain">
                            {{ $t('aitool.proxy') }}
                        </el-button>
                        <el-button :disabled="modelInfo.status !== 'Running'" @click="onLoadConn" type="primary" plain>
                            {{ $t('database.databaseConnInfo') }}
                        </el-button>
                        <el-button :disabled="modelInfo.status !== 'Running'" type="primary" plain @click="onSync()">
                            {{ $t('database.loadFromRemote') }}
                        </el-button>
                        <el-button
                            :disabled="modelInfo.status !== 'Running'"
                            icon="Position"
                            @click="goDashboard()"
                            type="primary"
                            plain
                        >
                            OpenWebUI
                        </el-button>

                        <el-button plain :disabled="selects.length === 0" type="primary" @click="onDelete(null)">
                            {{ $t('commons.button.delete') }}
                        </el-button>
                    </div>
                    <div>
                        <TableSearch @search="search()" v-model:searchName="searchName" />
                    </div>
                </div>
            </template>
            <template #main v-if="modelInfo.isExist">
                <ComplexTable
                    :pagination-config="paginationConfig"
                    v-model:selects="selects"
                    :class="{ mask: maskShow }"
                    @sort-change="search"
                    @search="search"
                    :data="data"
                >
                    <el-table-column type="selection" :selectable="selectable" fix />
                    <el-table-column :label="$t('ai_tools.model.model')" prop="name" min-width="90">
                        <template #default="{ row }">
                            <el-text v-if="row.size" type="primary" class="cursor-pointer" @click="onLoad(row.name)">
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
                                {{ $t('database.isDelete') }}
                            </el-tag>
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
                    <el-table-column :label="$t('commons.button.log')">
                        <template #default="{ row }">
                            <el-button @click="onLoadLog(row)" link type="primary">
                                {{ $t('website.check') }}
                            </el-button>
                        </template>
                    </el-table-column>
                    <el-table-column
                        min-width="80"
                        :label="$t('commons.table.date')"
                        prop="createdAt"
                        :formatter="dateFormat"
                    />
                    <fu-table-operations
                        :ellipsis="mobile ? 0 : 10"
                        :min-width="mobile ? 'auto' : 200"
                        :buttons="buttons"
                        :label="$t('commons.table.operate')"
                        fixed="right"
                        fix
                    />
                </ComplexTable>
            </template>
        </LayoutContent>

        <el-card v-if="modelInfo.isExist && modelInfo.status != 'Running' && !loading && maskShow" class="mask-prompt">
            <span>
                {{ $t('commons.service.serviceNotStarted', ['Ollama']) }}
            </span>
        </el-card>

        <el-dialog
            v-model="dashboardVisible"
            :title="$t('app.checkTitle')"
            width="30%"
            :close-on-click-modal="false"
            :destroy-on-close="true"
        >
            <div class="flex justify-center items-center gap-2 flex-wrap">
                {{ $t('app.checkInstalledWarn', ['OpenWebUI']) }}
                <el-link icon="Position" @click="goInstall('ollama-webui')" type="primary">
                    {{ $t('database.goInstall') }}
                </el-link>
            </div>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="dashboardVisible = false">{{ $t('commons.button.cancel') }}</el-button>
                </span>
            </template>
        </el-dialog>

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
        <Log ref="logRef" @close="search" />
        <Del ref="delRef" @search="search" />
        <Terminal ref="terminalRef" />
        <Conn ref="connRef" />
        <CodemirrorDialog ref="detailRef" />
        <PortJumpDialog ref="dialogPortJumpRef" />
        <BindDomain ref="bindDomainRef" />
    </div>
</template>

<script lang="ts" setup>
import AppStatus from '@/components/app-status/index.vue';
import AddDialog from '@/views/ai/model/add/index.vue';
import Conn from '@/views/ai/model/conn/index.vue';
import Terminal from '@/views/ai/model/terminal/index.vue';
import Del from '@/views/ai/model/del/index.vue';
import Log from '@/components/log-dialog/index.vue';
import PortJumpDialog from '@/components/port-jump/index.vue';
import CodemirrorDialog from '@/components/codemirror-dialog/index.vue';
import { computed, onMounted, reactive, ref } from 'vue';
import i18n from '@/lang';
import { App } from '@/api/interface/app';
import { GlobalStore } from '@/store';
import {
    deleteOllamaModel,
    loadOllamaModel,
    recreateOllamaModel,
    searchOllamaModel,
    syncOllamaModel,
} from '@/api/modules/ai';
import { AI } from '@/api/interface/ai';
import { GetAppPort } from '@/api/modules/app';
import { dateFormat } from '@/utils/util';
import router from '@/routers';
import { MsgInfo, MsgSuccess } from '@/utils/message';
import BindDomain from '@/views/ai/model/domain/index.vue';
const globalStore = GlobalStore();

const loading = ref(false);
const selects = ref<any>([]);
const maskShow = ref(true);
const addRef = ref();
const logRef = ref();
const detailRef = ref();
const delRef = ref();
const connRef = ref();
const terminalRef = ref();
const openWebUIPort = ref();
const dashboardVisible = ref(false);
const dialogPortJumpRef = ref();
const appStatusRef = ref();
const bindDomainRef = ref();
const data = ref();
const paginationConfig = reactive({
    cacheSizeKey: 'model-page-size',
    currentPage: 1,
    pageSize: Number(localStorage.getItem('page-size')) || 10,
    total: 0,
});
const searchName = ref();
const appInstallID = ref(0);

const opRef = ref();
const operateIDs = ref();
const forceDelete = ref();

const modelInfo = reactive({
    status: '',
    container: '',
    isExist: null,
    version: '',
    port: 11434,
});

const mobile = computed(() => {
    return globalStore.isMobile();
});

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
            loading.value = false;
        });
};

const onCreate = async () => {
    addRef.value.acceptParams();
};

const onSync = async () => {
    loading.value = true;
    await syncOllamaModel()
        .then((res) => {
            loading.value = false;
            if (res.data) {
                delRef.value.acceptParams({ list: res.data });
            } else {
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                search();
            }
        })
        .catch(() => {
            loading.value = false;
        });
};

const onLoadConn = async () => {
    connRef.value.acceptParams({
        port: modelInfo.port,
        containerName: modelInfo.container,
        appinstallID: appInstallID.value,
    });
};

const onLoad = async (name: string) => {
    const res = await loadOllamaModel(name);
    let detailInfo = res.data;
    let param = {
        header: i18n.global.t('commons.button.view'),
        detailInfo: detailInfo,
    };
    detailRef.value!.acceptParams(param);
};

const goDashboard = async () => {
    if (openWebUIPort.value === 0) {
        dashboardVisible.value = true;
        return;
    }
    dialogPortJumpRef.value.acceptParams({ port: openWebUIPort.value });
};

const bindDomain = () => {
    bindDomainRef.value.acceptParams(appInstallID.value);
};

const goInstall = (name: string) => {
    router.push({ name: 'AppAll', query: { install: name } });
};

const loadWebUIPort = async () => {
    const res = await GetAppPort('ollama-webui', '');
    openWebUIPort.value = res.data;
};

const checkExist = (data: App.CheckInstalled) => {
    modelInfo.isExist = data.isExist;
    modelInfo.status = data.status;
    modelInfo.version = data.version;
    modelInfo.container = data.containerName;
    modelInfo.port = data.httpPort;

    if (modelInfo.isExist && modelInfo.status === 'Running') {
        search();
    }
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
    await recreateOllamaModel(name)
        .then(() => {
            loading.value = false;
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            search();
        })
        .catch(() => {
            loading.value = false;
        });
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
            i18n.global.t('cronjob.cronTask'),
            i18n.global.t('commons.button.delete'),
        ]),
        api: null,
        params: null,
    });
};

const onLoadLog = (row: any) => {
    if (row.from === 'remote') {
        MsgInfo(i18n.global.t('ai_tools.model.from_remote'));
        return;
    }
    if (!row.logFileExist) {
        MsgInfo(i18n.global.t('ai_tools.model.no_logs'));
        return;
    }
    logRef.value.acceptParams({ id: 0, type: 'ollama-model', name: row.name, tail: true });
};

const buttons = [
    {
        label: i18n.global.t('commons.button.run'),
        click: (row: AI.OllamaModelInfo) => {
            terminalRef.value.acceptParams({ name: row.name });
        },
        disabled: (row: any) => {
            return row.status !== 'Success';
        },
    },
    {
        label: i18n.global.t('commons.button.retry'),
        click: (row: AI.OllamaModelInfo) => {
            onReCreate(row.name);
        },
        disabled: (row: any) => {
            return row.status === 'Success' || row.status === 'Waiting';
        },
    },
    {
        label: i18n.global.t('commons.button.delete'),
        click: (row: AI.OllamaModelInfo) => {
            onDelete(row);
        },
        disabled: (row: any) => {
            return row.status === 'Waiting';
        },
    },
];

onMounted(() => {
    loadWebUIPort();
});
</script>

<style lang="scss" scoped>
.iconInTable {
    margin-left: 5px;
    margin-top: 3px;
}
.jumpAdd {
    margin-top: 10px;
    margin-left: 15px;
    margin-bottom: 5px;
    font-size: 12px;
}
.tagClass {
    float: right;
    font-size: 12px;
    margin-top: 5px;
}
</style>
