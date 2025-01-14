<template>
    <LayoutContent v-loading="loading" v-if="!showDetail" :title="$t('app.app', 2)">
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
                    <div v-for="item in tags.slice(0, 7)" :key="item.key" class="inline">
                        <el-button
                            class="tag-button"
                            :class="activeTag === item.key ? '' : 'no-active'"
                            @click="changeTag(item.key)"
                            :type="activeTag === item.key ? 'primary' : ''"
                            :plain="activeTag !== item.key"
                        >
                            {{ item.name }}
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
                                        {{ item.name }}
                                    </el-dropdown-item>
                                </el-dropdown-menu>
                            </template>
                        </el-dropdown>
                    </div>
                </el-col>
                <el-col :xs="24" :sm="4" :md="4" :lg="4" :xl="4">
                    <TableSearch @search="searchByName()" v-model:searchName="req.name" />
                </el-col>
            </el-row>
        </template>
        <template #rightButton>
            <div class="flex justify-end flex-col sm:flex-row">
                <fu-table-pagination
                    v-model:current-page="paginationConfig.currentPage"
                    v-model:page-size="paginationConfig.pageSize"
                    v-bind="paginationConfig"
                    @change="search(req)"
                    :layout="mobile ? ' prev, pager, next' : ' prev, pager, next'"
                />
                <div>
                    <el-checkbox v-model="req.resource" true-value="all" false-value="remote" @change="search(req)">
                        {{ $t('app.showLocal') }}
                    </el-checkbox>
                </div>
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
                                                {{ app.description }}
                                            </span>
                                        </div>
                                        <div class="app-tag">
                                            <el-tag v-for="(tag, ind) in app.tags" :key="ind" class="p-mr-5">
                                                <span>
                                                    {{ tag.name }}
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
import { GetAppTags, SearchApp } from '@/api/modules/app';
import Detail from '../detail/index.vue';
import Install from '../detail/install/index.vue';
import router from '@/routers';
import { GlobalStore } from '@/store';
const globalStore = GlobalStore();

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
});

const apps = ref<App.AppDTO[]>([]);
const tags = ref<App.Tag[]>([]);
const loading = ref(false);
const activeTag = ref('all');
const showDetail = ref(false);
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
        return tag.name;
    }
};

const searchByName = () => {
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

<style scoped lang="scss">
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
                line-height: 1.5;
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
            height: 43px;
            .desc {
                overflow: hidden;
                display: -webkit-box;
                -webkit-line-clamp: 2;
                -webkit-box-orient: vertical;
                text-overflow: ellipsis;
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
