<template>
    <div>
        <docker-status v-model:isActive="isActive" v-model:isExist="isExist" />
        <LayoutContent v-loading="loading" v-if="isExist" :class="{ mask: !isActive }" :title="$t('app.app', 2)">
            <template #search>
                <Tags @change="changeTag" />
            </template>
            <template #leftToolBar>
                <el-button @click="sync" type="primary" plain :disabled="syncing">
                    <span>{{ syncCustomAppstore || isOffLine ? $t('app.syncCustomApp') : $t('app.syncAppList') }}</span>
                </el-button>
                <el-button @click="syncLocal" type="primary" plain :disabled="syncing" class="ml-2">
                    {{ $t('app.syncLocalApp') }}
                </el-button>
            </template>
            <template #rightToolBar>
                <el-checkbox class="!mr-2.5" v-model="req.showCurrentArch" @change="search(req)">
                    {{ $t('app.showCurrentArch') }}
                </el-checkbox>
                <el-checkbox
                    class="!mr-2.5"
                    v-model="req.resource"
                    true-value="all"
                    false-value="remote"
                    @change="search(req)"
                >
                    {{ $t('app.showLocal') }}
                </el-checkbox>
                <TableSearch @search="searchByName()" v-model:searchName="req.name" />
            </template>
            <template #main>
                <div>
                    <MainDiv :heightDiff="300">
                        <el-alert type="info" :title="$t('app.appHelper')" :closable="false" />
                        <el-row :gutter="5" v-if="apps.length > 0">
                            <el-col
                                class="app-col-12"
                                v-for="(app, index) in apps"
                                :key="index"
                                :xs="24"
                                :sm="12"
                                :md="8"
                                :lg="8"
                                :xl="6"
                            >
                                <AppCard :app="app" @open-install="openInstall" @open-detail="openDetail" />
                            </el-col>
                        </el-row>
                        <NoApp v-if="noApp" />
                    </MainDiv>
                    <div class="page-button">
                        <fu-table-pagination
                            v-model:current-page="paginationConfig.currentPage"
                            v-model:page-size="paginationConfig.pageSize"
                            v-bind="paginationConfig"
                            @change="search(req)"
                            :page-sizes="[30, 60, 90]"
                            :layout="mobile ? 'total, prev, pager, next' : 'total, sizes, prev, pager, next, jumper'"
                        />
                    </div>
                </div>
            </template>
        </LayoutContent>
    </div>
    <Install ref="installRef" />
    <Detail ref="detailRef" />
    <TaskLog ref="taskLogRef" @close="refresh" />
</template>

<script lang="ts" setup>
import { App } from '@/api/interface/app';
import { onMounted, reactive, ref, computed } from 'vue';
import { searchApp, syncApp, syncCutomAppStore, syncLocalApp, getCurrentNodeCustomAppConfig } from '@/api/modules/app';
import Install from '../detail/install/index.vue';
import router from '@/routers';
import { MsgSuccess } from '@/utils/message';
import { newUUID } from '@/utils/util';
import Detail from '../detail/index.vue';
import TaskLog from '@/components/log/task/index.vue';
import bus from '@/global/bus';
import Tags from '@/views/app-store/components/tag.vue';
import DockerStatus from '@/views/container/docker-status/index.vue';
import NoApp from '@/views/app-store/apps/no-app/index.vue';
import AppCard from '@/views/app-store/apps/app/index.vue';
import MainDiv from '@/components/main-div/index.vue';
import { jumpToInstall } from '@/utils/app';
import { useGlobalStore } from '@/composables/useGlobalStore';
const { globalStore, isProductPro, isOffLine } = useGlobalStore();

const mobile = computed(() => {
    return globalStore.isMobile();
});

const paginationConfig = reactive({
    cacheSizeKey: 'app-page-size',
    currentPage: 1,
    pageSize: Number(localStorage.getItem('app-page-size')) || 60,
    total: 0,
});

const req = reactive({
    name: '',
    tags: [],
    page: 1,
    pageSize: 60,
    resource: 'all',
    showCurrentArch: false,
});

const apps = ref<App.AppDTO[]>([]);
const loading = ref(false);
const canUpdate = ref(false);
const syncing = ref(false);
const installRef = ref();
const installKey = ref('');
const mainHeight = ref(0);
const detailRef = ref();
const taskLogRef = ref();
const syncCustomAppstore = ref(false);
const isActive = ref(false);
const isExist = ref(false);
const noApp = ref(false);

const refresh = () => {
    search(req);
};

const search = async (req: App.AppReq) => {
    loading.value = true;
    req.pageSize = paginationConfig.pageSize;
    req.page = paginationConfig.currentPage;
    localStorage.setItem('app-page-size', req.pageSize + '');

    const customReq = {
        page: req.page,
        pageSize: req.pageSize,
        tags: req.tags,
        name: req.name,
        resource: req.resource,
        showCurrentArch: req.showCurrentArch,
    };
    if (syncCustomAppstore.value && req.resource === 'remote') {
        customReq.resource = 'custom';
    }
    await searchApp(customReq)
        .then((res) => {
            apps.value = res.data.items;
            paginationConfig.total = res.data.total;
            if (noApp.value && apps.value.length > 0) {
                noApp.value = false;
            }
        })
        .finally(() => {
            loading.value = false;
        });
};

const openInstall = (app: App.App) => {
    if (!jumpToInstall(app.type, app.key)) {
        const params = {
            app: app,
        };
        installRef.value.acceptParams(params);
    }
};

const openDetail = (key: string) => {
    detailRef.value.acceptParams(key, 'install');
};

const openTaskLog = (taskID: string) => {
    taskLogRef.value.openWithTaskID(taskID);
};

const sync = async () => {
    syncing.value = true;
    const taskID = newUUID();
    const syncReq = {
        taskID: taskID,
    };
    try {
        let res;
        if (isOffLine.value || (isProductPro.value && syncCustomAppstore.value)) {
            res = await syncCutomAppStore(syncReq);
        } else {
            res = await syncApp(syncReq);
        }
        if (res.message != '' && res.message != 'success') {
            MsgSuccess(res.message);
        } else {
            openTaskLog(taskID);
        }
        canUpdate.value = false;
        search(req);
    } finally {
        syncing.value = false;
    }
};

const syncLocal = () => {
    const taskID = newUUID();
    const syncReq = {
        taskID: taskID,
    };
    syncing.value = true;
    syncLocalApp(syncReq)
        .then(() => {
            openTaskLog(taskID);
            canUpdate.value = false;
            search(req);
        })
        .finally(() => {
            syncing.value = false;
        });
};

const changeTag = (key: string) => {
    req.tags = [];
    if (key !== 'all') {
        req.tags = [key];
    }
    search(req);
};

const searchByName = () => {
    search(req);
};

onMounted(async () => {
    bus.on('refreshApp', () => {
        search(req);
    });
    if (router.currentRoute.value.query.install) {
        installKey.value = String(router.currentRoute.value.query.install);
        const params = {
            app: {
                key: installKey.value,
            },
        };
        installRef.value.acceptParams(params);
    }
    search(req);
    if (isProductPro.value) {
        const res = await getCurrentNodeCustomAppConfig();
        if (res && res.data) {
            syncCustomAppstore.value = res.data.status === 'Enable';
        }
    }
    if (isOffLine.value) {
        syncCustomAppstore.value = true;
    }
    mainHeight.value = window.innerHeight - 380;
    window.onresize = () => {
        return (() => {
            mainHeight.value = window.innerHeight - 380;
        })();
    };
});
</script>

<style lang="scss" scoped>
.header {
    padding-bottom: 10px;
}

@media only screen and (min-width: 768px) and (max-width: 1200px) {
    .app-col-12 {
        max-width: 50%;
        flex: 0 0 50%;
    }
}

.page-button {
    float: right;
    margin-bottom: 10px;
    margin-top: 10px;
}
</style>
