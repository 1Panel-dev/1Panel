<template>
    <LayoutContent v-loading="loading" v-if="!showDetail" :title="$t('app.app')">
        <template #search>
            <el-row :gutter="5">
                <el-col :xs="24" :sm="20" :md="20" :lg="20" :xl="20">
                    <el-button
                        class="tag-button"
                        :class="activeTag === 'all' ? '' : 'no-active'"
                        @click="changeTag('all')"
                        :type="activeTag === 'all' ? 'primary' : ''"
                        :plain="activeTag !== 'all'"
                    >
                        {{ $t('app.all') }}
                    </el-button>
                    <div v-for="item in tags.slice(0, 7)" :key="item.key" class="inline">
                        <el-button
                            class="tag-button"
                            :class="activeTag === item.key ? '' : 'no-active'"
                            @click="changeTag(item.key)"
                            :type="activeTag === item.key ? 'primary' : ''"
                            :plain="activeTag !== item.key"
                        >
                            {{ language == 'zh' || language == 'tw' ? item.name : item.key }}
                        </el-button>
                    </div>
                    <div class="inline">
                        <el-dropdown>
                            <el-button
                                class="tag-button"
                                :type="moreTag !== '' ? 'primary' : ''"
                                :class="moreTag !== '' ? '' : 'no-active'"
                            >
                                {{ moreTag == '' ? $t('tabs.more') : getTagValue(moreTag) }}
                                <el-icon class="el-icon--right">
                                    <arrow-down />
                                </el-icon>
                            </el-button>
                            <template #dropdown>
                                <el-dropdown-menu>
                                    <el-dropdown-item
                                        v-for="item in tags.slice(7)"
                                        @click="changeTag(item.key)"
                                        :key="item.key"
                                    >
                                        {{ language == 'zh' || language == 'tw' ? item.name : item.key }}
                                    </el-dropdown-item>
                                </el-dropdown-menu>
                            </template>
                        </el-dropdown>
                    </div>
                </el-col>
                <el-col :xs="24" :sm="4" :md="4" :lg="4" :xl="4"></el-col>
            </el-row>
        </template>
        <template #leftToolBar>
            <el-button @click="sync" type="primary" plain :disabled="syncing">
                {{ $t('app.syncAppList') }}
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
                <MainDiv :heightDiff="350">
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
                                    <div class="app-wrapper">
                                        <div class="app-image">
                                            <el-avatar
                                                @click="openDetail(app.key)"
                                                shape="square"
                                                :size="60"
                                                :src="'data:image/png;base64,' + app.icon"
                                            />
                                        </div>
                                        <div class="app-content">
                                            <div class="content-top">
                                                <div>
                                                    <span class="app-name">{{ app.name }}</span>
                                                    <el-text type="success" class="!ml-2" v-if="app.installed">
                                                        {{ $t('app.allReadyInstalled') }}
                                                    </el-text>
                                                </div>
                                            </div>
                                            <div class="content-middle">
                                                <span class="app-description">
                                                    {{
                                                        language == 'zh' || language == 'tw'
                                                            ? app.shortDescZh
                                                            : app.shortDescEn
                                                    }}
                                                </span>
                                            </div>
                                            <div class="content-bottom">
                                                <div class="app-tags">
                                                    <el-tag v-for="(tag, ind) in app.tags" :key="ind" type="info">
                                                        <span>
                                                            {{
                                                                language == 'zh' || language == 'tw'
                                                                    ? tag.name
                                                                    : tag.key
                                                            }}
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
                                                        (app.installed && app.limit == 1) || app.status === 'TakeDown'
                                                    "
                                                    @click.stop="openInstall(app)"
                                                >
                                                    {{ $t('app.install') }}
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
    <Install ref="installRef" />
    <Detail ref="detailRef" />
    <TaskLog ref="taskLogRef" />
</template>

<script lang="ts" setup>
import { App } from '@/api/interface/app';
import { onMounted, reactive, ref, computed } from 'vue';
import { GetAppTags, SearchApp, SyncApp, SyncCutomAppStore, SyncLocalApp } from '@/api/modules/app';
import Install from '../detail/install/index.vue';
import router from '@/routers';
import { MsgSuccess } from '@/utils/message';
import { GlobalStore } from '@/store';
import { getLanguage, newUUID } from '@/utils/util';
import Detail from '../detail/index.vue';
import TaskLog from '@/components/task-log/index.vue';
import { storeToRefs } from 'pinia';
import { GetCustomAppStoreConfig } from '@/xpack/api/modules/app';

const globalStore = GlobalStore();
const { isProductPro } = storeToRefs(globalStore);

const mobile = computed(() => {
    return globalStore.isMobile();
});

const language = getLanguage();

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
const tags = ref<App.Tag[]>([]);
const loading = ref(false);
const activeTag = ref('all');
const showDetail = ref(false);
const canUpdate = ref(false);
const syncing = ref(false);
const installRef = ref();
const installKey = ref('');
const moreTag = ref('');
const mainHeight = ref(0);
const detailRef = ref();
const taskLogRef = ref();
const syncCustomAppstore = ref(false);

const search = async (req: App.AppReq) => {
    loading.value = true;
    req.pageSize = paginationConfig.pageSize;
    req.page = paginationConfig.currentPage;
    localStorage.setItem('app-page-size', req.pageSize + '');
    await SearchApp(req)
        .then((res) => {
            apps.value = res.data.items;
            paginationConfig.total = res.data.total;
        })
        .finally(() => {
            loading.value = false;
        });
    GetAppTags().then((res) => {
        tags.value = res.data;
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
            router.push({ path: '/websites/runtimes/' + app.type });
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
            res = await SyncCutomAppStore(syncReq);
        } else {
            res = await SyncApp(syncReq);
        }
        if (res.message != '') {
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
    SyncLocalApp(syncReq)
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
    activeTag.value = key;
    if (key !== 'all') {
        req.tags = [key];
    }
    const index = tags.value.findIndex((tag) => tag.key === key);
    if (index > 6) {
        moreTag.value = key;
    } else {
        moreTag.value = '';
    }
    search(req);
};

const getTagValue = (key: string) => {
    const tag = tags.value.find((tag) => tag.key === key);
    if (tag) {
        return language == 'zh' || language == 'tw' ? tag.name : tag.key;
    }
};

const searchByName = () => {
    search(req);
};

onMounted(async () => {
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
        const res = await GetCustomAppStoreConfig();
        if (res && res.data) {
            syncCustomAppstore.value = res.data.status === 'enable';
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

<style lang="scss">
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
    }
    .app-image {
        cursor: pointer;
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
