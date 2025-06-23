<template>
    <div>
        <docker-status v-model:isActive="isActive" v-model:isExist="isExist" />
        <LayoutContent v-loading="loading" v-if="isExist" :class="{ mask: !isActive }" :title="$t('app.app', 2)">
            <template #search>
                <Tags @change="changeTag" />
            </template>
            <template #leftToolBar>
                <el-button @click="sync" type="primary" plain :disabled="syncing">
                    <span>{{ syncCustomAppstore ? $t('app.syncCustomApp') : $t('app.syncAppList') }}</span>
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
                        <el-row :gutter="5">
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
                                <div class="app">
                                    <el-card>
                                        <div class="app-wrapper" @click="openDetail(app.key)">
                                            <div class="app-image">
                                                <el-avatar
                                                    shape="square"
                                                    :size="60"
                                                    :src="'data:image/png;base64,' + app.icon"
                                                />
                                            </div>
                                            <div class="app-content">
                                                <div class="content-top">
                                                    <el-space wrap :size="1">
                                                        <span class="app-title">{{ app.name }}</span>
                                                        <el-tag
                                                            type="success"
                                                            v-if="app.installed"
                                                            round
                                                            size="small"
                                                            class="!ml-2"
                                                        >
                                                            {{ $t('app.allReadyInstalled') }}
                                                        </el-tag>
                                                    </el-space>
                                                </div>
                                                <div class="content-middle">
                                                    <span class="app-description">
                                                        {{ app.description }}
                                                    </span>
                                                </div>
                                                <div class="content-bottom">
                                                    <div class="app-tags">
                                                        <el-tag v-for="(tag, ind) in app.tags" :key="ind" type="info">
                                                            <span>
                                                                {{ tag.name }}
                                                            </span>
                                                        </el-tag>
                                                        <el-tag v-if="app.status === 'TakeDown'" class="p-mr-5">
                                                            <span style="color: red">{{ $t('app.takeDown') }}</span>
                                                        </el-tag>
                                                    </div>
                                                    <el-button
                                                        type="primary"
                                                        size="small"
                                                        plain
                                                        round
                                                        :disabled="
                                                            (app.installed && app.limit == 1) ||
                                                            app.status === 'TakeDown'
                                                        "
                                                        @click.stop="openInstall(app)"
                                                    >
                                                        {{ $t('commons.button.install') }}
                                                    </el-button>
                                                </div>
                                            </div>
                                        </div>
                                    </el-card>
                                </div>
                            </el-col>
                        </el-row>
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
import { GlobalStore } from '@/store';
import { newUUID, jumpToPath } from '@/utils/util';
import Detail from '../detail/index.vue';
import TaskLog from '@/components/log/task/index.vue';
import { storeToRefs } from 'pinia';
import bus from '@/global/bus';
import Tags from '@/views/app-store/components/tag.vue';
import DockerStatus from '@/views/container/docker-status/index.vue';

const globalStore = GlobalStore();
const { isProductPro } = storeToRefs(globalStore);

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
        })
        .finally(() => {
            loading.value = false;
        });
};

const openInstall = (app: App.App) => {
    switch (app.type) {
        case 'php':
        case 'node':
        case 'java':
        case 'go':
        case 'python':
        case 'dotnet':
            jumpToPath(router, '/websites/runtimes/' + app.type);
            break;
        default:
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
        if (isProductPro.value && syncCustomAppstore.value) {
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

.app {
    margin: 10px;
    .el-card {
        padding: 0 !important;
        border: var(--panel-border) !important;
        &:hover {
            border: 1px solid var(--el-color-primary) !important;
        }
    }
    .el-card__body {
        padding: 8px 8px 2px 8px !important;
    }
    .app-wrapper {
        display: flex;
        height: 100%;
        cursor: pointer;
    }
    .app-image {
        flex: 0 0 100px;
        display: flex;
        justify-content: center;
        margin-top: 14px;
        transition: transform 0.1s;
    }

    &:hover .app-image {
        transform: scale(1.2);
    }

    .el-avatar {
        width: 65px !important;
        height: 65px !important;
        max-width: 65px;
        max-height: 65px;
        object-fit: cover;
    }
    .app-content {
        flex: 1;
        display: flex;
        flex-direction: column;
        padding: 10px;
    }
    .content-top,
    .content-bottom {
        display: flex;
        justify-content: space-between;
        align-items: center;
    }
    .content-middle {
        flex: 1;
        margin: 10px 0;
        overflow: hidden; /* 防止内容溢出 */
    }
    .app-name {
        margin: 0;
        line-height: 1.5;
        font-weight: 500;
        font-size: 16px;
        color: var(--el-text-color-regular);
    }
    .app-description {
        margin: 0;
        overflow: hidden;
        display: -webkit-box;
        -webkit-line-clamp: 2;
        -webkit-box-orient: vertical;
        text-overflow: ellipsis;
        font-size: 14px;
        color: var(--el-text-color-regular);

        line-height: 1.2;
        height: calc(1.2em * 2);
        min-height: calc(1.2em * 2);
    }
    .app-tags {
        display: flex;
        gap: 5px;
    }
}

.tag-button {
    margin-right: 10px;
    &.no-active {
        background: none;
        border: none;
    }
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
