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
                    <div v-for="item in tags" :key="item.key" style="display: inline">
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
            <el-badge is-dot class="item" :hidden="!canUpdate">
                <el-button @click="sync" type="primary" link :plain="true">{{ $t('app.syncAppList') }}</el-button>
            </el-badge>
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
                        <el-card class="e-card">
                            <el-row :gutter="20">
                                <el-col :xs="8" :sm="6" :md="6" :lg="6" :xl="5">
                                    <div class="app-icon">
                                        <el-avatar
                                            shape="square"
                                            :size="60"
                                            :src="'data:image/png;base64,' + app.icon"
                                        />
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
                                                @click="getAppDetail(app.key)"
                                                :disabled="app.status === 'TakeDown'"
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
                                            <el-tag v-for="(tag, ind) in app.tags" :key="ind" style="margin-right: 5px">
                                                <span :style="{ color: getColor(ind) }">
                                                    {{ language == 'zh' || language == 'tw' ? tag.name : tag.key }}
                                                </span>
                                            </el-tag>
                                            <el-tag v-if="app.status === 'TakeDown'" style="margin-right: 5px">
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
                    :layout="'total, sizes, prev, pager, next, jumper'"
                />
            </div>
        </template>
    </LayoutContent>
    <Detail v-if="showDetail" :id="appId"></Detail>
</template>

<script lang="ts" setup>
import { App } from '@/api/interface/app';
import { onMounted, reactive, ref } from 'vue';
import { GetAppTags, SearchApp, SyncApp } from '@/api/modules/app';
import i18n from '@/lang';
import Detail from '../detail/index.vue';
import router from '@/routers';
import { MsgSuccess } from '@/utils/message';
import { useI18n } from 'vue-i18n';

const language = useI18n().locale.value;

const paginationConfig = reactive({
    currentPage: 1,
    pageSize: 50,
    total: 0,
});

const req = reactive({
    name: '',
    tags: [],
    page: 1,
    pageSize: 50,
});

const apps = ref<App.AppDTO[]>([]);
const tags = ref<App.Tag[]>([]);
const colorArr = ['#005eeb', '#008B45', '#BEBEBE', '#FFF68F', '#FFFF00', '#8B0000'];
const loading = ref(false);
const activeTag = ref('all');
const showDetail = ref(false);
const appId = ref(0);
const canUpdate = ref(false);

const getColor = (index: number) => {
    return colorArr[index];
};

const search = async (req: App.AppReq) => {
    loading.value = true;
    req.pageSize = paginationConfig.pageSize;
    req.page = paginationConfig.currentPage;
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

const getAppDetail = (key: string) => {
    router.push({ name: 'AppDetail', params: { appKey: key } });
};

const sync = () => {
    SyncApp().then((res) => {
        if (res.message != '') {
            MsgSuccess(res.message);
        } else {
            MsgSuccess(i18n.global.t('app.syncStart'));
        }
        canUpdate.value = false;
        search(req);
    });
};

const changeTag = (key: string) => {
    req.tags = [];
    activeTag.value = key;
    if (key !== 'all') {
        req.tags = [key];
    }
    search(req);
};

const searchByName = (name: string) => {
    req.name = name;
    search(req);
};

onMounted(() => {
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

    .app-icon {
        margin-top: 10px;
        margin-left: 10px;
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
            margin-top: 5px;
            overflow: hidden;
            display: -webkit-box;
            -webkit-line-clamp: 2;
            -webkit-box-orient: vertical;

            text-overflow: ellipsis;
            height: 45px;

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

.page-button {
    float: right;
    margin-bottom: 10px;
    margin-top: 10px;
}

@media only screen and (min-width: 768px) and (max-width: 1200px) {
    .app-col-12 {
        max-width: 50%;
        flex: 0 0 50%;
    }
}
</style>
