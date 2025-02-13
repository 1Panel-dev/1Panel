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
                    @is-exist="checkExist"
                    ref="appStatusRef"
                ></AppStatus>
            </template>
            <template #toolbar v-if="modelInfo.isExist">
                <div class="flex justify-between gap-2 flex-wrap sm:flex-row">
                    <div class="flex flex-wrap gap-3">
                        <el-button :disabled="modelInfo.status !== 'Running'" type="primary" @click="onCreate()">
                            {{ $t('ai_tools.model.create') }}
                        </el-button>
                        <el-button :disabled="modelInfo.status !== 'Running'" @click="onLoadConn" type="primary" plain>
                            {{ $t('database.databaseConnInfo') }}
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
                    </div>
                    <div>
                        <TableSearch @search="search()" v-model:searchName="searchName" />
                    </div>
                </div>
            </template>
            <template #main v-if="modelInfo.isExist">
                <ComplexTable
                    :pagination-config="paginationConfig"
                    :class="{ mask: maskShow }"
                    @sort-change="search"
                    @search="search"
                    :data="data"
                >
                    <el-table-column :label="$t('commons.table.name')" prop="name" min-width="90">
                        <template #default="{ row }">
                            <el-text type="primary" class="cursor-pointer" @click="onLoad(row.name)">
                                {{ row.name }}
                            </el-text>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('file.size')" prop="size" />
                    <el-table-column :label="$t('commons.button.log')">
                        <template #default="{ row }">
                            <el-button @click="onLoadLog(row.name)" link type="primary">
                                {{ $t('website.check') }}
                            </el-button>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('commons.table.createdAt')" prop="modified" />
                    <fu-table-operations
                        :ellipsis="mobile ? 0 : 10"
                        :min-width="mobile ? 'auto' : 400"
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

        <AddDialog ref="addRef" @search="search" @log="onLoadLog" />
        <Log ref="logRef" @close="search" />
        <Conn ref="connRef" />
        <CodemirrorDialog ref="detailRef" />
        <PortJumpDialog ref="dialogPortJumpRef" />
    </div>
</template>

<script lang="ts" setup>
import AppStatus from '@/components/app-status/index.vue';
import AddDialog from '@/views/ai-tools/model/add/index.vue';
import Conn from '@/views/ai-tools/model/conn/index.vue';
import Log from '@/components/log-dialog/index.vue';
import PortJumpDialog from '@/components/port-jump/index.vue';
import CodemirrorDialog from '@/components/codemirror-dialog/index.vue';
import { computed, onMounted, reactive, ref } from 'vue';
import i18n from '@/lang';
import { App } from '@/api/interface/app';
import { GlobalStore } from '@/store';
import { deleteOllamaModel, loadOllamaModel, searchOllamaModel } from '@/api/modules/ai-tool';
import { AITool } from '@/api/interface/ai-tool';
import { GetAppPort } from '@/api/modules/app';
import router from '@/routers';
import { MsgSuccess } from '@/utils/message';
const globalStore = GlobalStore();

const loading = ref(false);
const maskShow = ref(true);

const addRef = ref();
const logRef = ref();
const detailRef = ref();
const connRef = ref();
const openWebUIPort = ref();
const dashboardVisible = ref(false);
const dialogPortJumpRef = ref();

const appStatusRef = ref();

const data = ref();
const paginationConfig = reactive({
    cacheSizeKey: 'model-page-size',
    currentPage: 1,
    pageSize: Number(localStorage.getItem('page-size')) || 10,
    total: 0,
});
const searchName = ref();

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

const onLoadConn = async () => {
    connRef.value.acceptParams({ port: modelInfo.port, containerName: modelInfo.container });
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

const onDelete = async (row: AITool.OllamaModelInfo) => {
    ElMessageBox.confirm(i18n.global.t('commons.msg.delete'), i18n.global.t('commons.button.delete'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
        type: 'info',
    }).then(async () => {
        loading.value = true;
        await deleteOllamaModel(row.name)
            .then(() => {
                loading.value = false;
                search();
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            })
            .catch(() => {
                loading.value = false;
            });
    });
};

const onLoadLog = (name: string) => {
    logRef.value.acceptParams({ id: 0, type: 'ollama-model', name: name, tail: true });
};

const buttons = [
    {
        label: i18n.global.t('commons.button.delete'),
        click: (row: AITool.OllamaModelInfo) => {
            onDelete(row);
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
