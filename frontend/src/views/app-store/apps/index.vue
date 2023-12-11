<template>
    <LayoutContent v-loading="loading" v-if="!showDetail" :title="$t('app.app')">
        <template #toolbar>
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
                    <div v-for="item in tags.slice(0, 6)" :key="item.key" class="inline">
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
                                        v-for="item in tags.slice(6)"
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
                <el-col :xs="24" :sm="4" :md="4" :lg="4" :xl="4">
                    <div class="search-button">
                        <el-input
                            v-model="req.name"
                            clearable
                            @clear="searchByName('')"
                            suffix-icon="Search"
                            @change="searchByName(req.name)"
                            :placeholder="$t('commons.button.search')"
                        ></el-input>
                    </div>
                </el-col>
            </el-row>
        </template>
        <template #rightButton>
            <div class="flex justify-end">
                <div class="mr-10">
                    <el-checkbox v-model="req.resource" true-label="all" false-label="remote" @change="search(req)">
                        {{ $t('app.showLocal') }}
                    </el-checkbox>
                </div>
                <fu-table-pagination
                    v-model:current-page="paginationConfig.currentPage"
                    v-model:page-size="paginationConfig.pageSize"
                    v-bind="paginationConfig"
                    @change="search(req)"
                    :layout="mobile ? ' prev, pager, next' : ' prev, pager, next'"
                />
                <el-badge is-dot :hidden="!canUpdate" class="ml-5">
                    <el-button @click="sync" type="primary" plain :disabled="syncing">
                        {{ $t('app.syncAppList') }}
                    </el-button>
                </el-badge>
            </div>
        </template>
        <template #main>
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
                    :xl="8"
                >
                    <div class="app-card">
                        <el-card class="e-card" @click.stop="openDetail(app.key)">
                            <el-row :gutter="20">
                                <el-col :xs="8" :sm="6" :md="6" :lg="6" :xl="5">
                                    <div class="app-icon-container">
                                        <div class="app-icon">
                                            <el-avatar
                                                shape="square"
                                                :size="60"
                                                :src="'data:image/png;base64,' + app.icon"
                                            />
                                        </div>
                                    </div>
                                </el-col>
                                <el-col :xs="16" :sm="18" :md="18" :lg="18" :xl="19">
                                    <div class="app-content">
                                        <div class="app-header">
                                            <span class="app-title">{{ app.name }}</span>
                                            <el-text type="success" style="margin-left: 10px" v-if="app.installed">
                                                {{ $t('app.allReadyInstalled') }}
                                            </el-text>
                                            <el-button
                                                class="app-button"
                                                type="primary"
                                                plain
                                                round
                                                size="small"
                                                :disabled="
                                                    (app.installed && app.limit == 1) || app.status === 'TakeDown'
                                                "
                                                @click.stop="openInstall(app)"
                                            >
                                                {{ $t('app.install') }}
                                            </el-button>
                                        </div>
                                        <div class="app-desc">
                                            <span class="desc">
                                                {{
                                                    language == 'zh' || language == 'tw'
                                                        ? app.shortDescZh
                                                        : app.shortDescEn
                                                }}
                                            </span>
                                        </div>
                                        <div class="app-tag">
                                            <el-tag v-for="(tag, ind) in app.tags" :key="ind" class="p-mr-5">
                                                <span>
                                                    {{ language == 'zh' || language == 'tw' ? tag.name : tag.key }}
                                                </span>
                                            </el-tag>
                                            <el-tag v-if="app.status === 'TakeDown'" class="p-mr-5">
                                                <span style="color: red">{{ $t('app.takeDown') }}</span>
                                            </el-tag>
                                        </div>
                                    </div>
                                </el-col>
                            </el-row>
                        </el-card>
                    </div>
                </el-col>
            </el-row>
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
        </template>
    </LayoutContent>
    <Detail ref="detailRef"></Detail>
    <Install ref="installRef" />
</template>

<script lang="ts" setup>
import { App } from '@/api/interface/app';
import { onMounted, reactive, ref, computed } from 'vue';
import { GetAppTags, SearchApp, SyncApp } from '@/api/modules/app';
import i18n from '@/lang';
import Detail from '../detail/index.vue';
import Install from '../detail/install/index.vue';
import router from '@/routers';
import { MsgSuccess } from '@/utils/message';
import { useI18n } from 'vue-i18n';
import { GlobalStore } from '@/store';

const globalStore = GlobalStore();

const mobile = computed(() => {
    return globalStore.isMobile();
});

const language = useI18n().locale.value;

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
});

const apps = ref<App.AppDTO[]>([]);
const tags = ref<App.Tag[]>([]);
const loading = ref(false);
const activeTag = ref('all');
const showDetail = ref(false);
const canUpdate = ref(false);
const syncing = ref(false);
const detailRef = ref();
const installRef = ref();
const installKey = ref('');
const moreTag = ref('');

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
            router.push({ path: '/websites/runtimes/php' });
            break;
        case 'node':
            router.push({ path: '/websites/runtimes/node' });
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

const sync = () => {
    syncing.value = true;
    SyncApp()
        .then((res) => {
            if (res.message != '') {
                MsgSuccess(res.message);
            } else {
                MsgSuccess(i18n.global.t('app.syncStart'));
            }
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
    if (index > 5) {
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

const searchByName = (name: string) => {
    req.name = name;
    search(req);
};

onMounted(() => {
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
});
</script>

<style lang="scss">
.header {
    padding-bottom: 10px;
}

.app-card {
    margin-top: 10px;
    cursor: pointer;
    padding: 5px;

    .app-icon-container {
        margin-top: 10px;
        margin-left: 15px;
    }

    &:hover .app-icon {
        transform: scale(1.2);
    }

    .app-icon {
        transition: transform 0.1s;
        transform-origin: center center;
    }

    .app-content {
        margin-top: 10px;
        height: 100%;
        width: 100%;

        .app-header {
            height: 20%;
            .app-title {
                font-weight: 500;
                font-size: 16px;
                color: var(--el-text-color-regular);
            }
            .app-button {
                float: right;
                margin-right: 20px;
            }
        }

        .app-desc {
            margin-top: 8px;
            overflow: hidden;
            display: -webkit-box;
            -webkit-line-clamp: 2;
            -webkit-box-orient: vertical;

            text-overflow: ellipsis;
            height: 43px;

            .desc {
                font-size: 14px;
                color: var(--el-text-color-regular);
            }
        }

        .app-tag {
            margin-top: 5px;
        }
    }

    .e-card {
        border: var(--panel-border) !important;
        &:hover {
            cursor: pointer;
            border: 1px solid var(--el-color-primary) !important;
        }
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
