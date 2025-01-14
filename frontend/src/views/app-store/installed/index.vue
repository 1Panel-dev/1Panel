<template>
    <LayoutContent v-loading="loading || syncLoading" :title="activeName">
        <template #toolbar>
            <el-row :gutter="5">
                <el-col :xs="24" :sm="20" :md="20" :lg="20" :xl="20">
                    <div>
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
                                            {{ language == 'zh' || language == 'tw' ? item.name : item.key }}
                                        </el-dropdown-item>
                                    </el-dropdown-menu>
                                </template>
                            </el-dropdown>
                        </div>
                    </div>
                </el-col>

                <el-col :xs="24" :sm="4" :md="4" :lg="4" :xl="4">
                    <TableSearch @search="search()" v-model:searchName="searchReq.name" />
                </el-col>
            </el-row>
        </template>
        <template #rightButton>
            <el-button @click="sync" type="primary" plain v-if="mode === 'installed' && data != null">
                {{ $t('app.sync') }}
            </el-button>
            <el-button @click="openIgnore" type="primary" plain v-if="mode === 'upgrade'">
                {{ $t('app.showIgnore') }}
            </el-button>
        </template>

        <template #main>
            <el-alert type="info" :closable="false" v-if="mode === 'installed'">
                <template #title>
                    <span class="flx-align-center">
                        {{ $t('app.installHelper') }}
                        <el-link class="ml-5" icon="Position" @click="quickJump()" type="primary">
                            {{ $t('firewall.quickJump') }}
                        </el-link>
                        　
                    </span>
                </template>
            </el-alert>
            <el-alert type="info" :title="$t('app.upgradeHelper')" :closable="false" v-if="mode === 'upgrade'" />
            <div class="update-prompt" v-if="data == null">
                <span>{{ mode === 'upgrade' ? $t('app.updatePrompt') : $t('app.installPrompt') }}</span>
                <div>
                    <img src="@/assets/images/no_update_app.svg" />
                </div>
            </div>
            <el-row :gutter="5">
                <el-col
                    v-for="(installed, index) in data"
                    :key="index"
                    :xs="24"
                    :sm="24"
                    :md="24"
                    :lg="12"
                    :xl="12"
                    :class="mode === 'upgrade' ? 'upgrade-card-col-12' : 'install-card-col-12'"
                >
                    <div class="install-card">
                        <el-card class="e-card">
                            <el-row :gutter="10">
                                <el-col :xs="3" :sm="3" :md="3" :lg="4" :xl="4">
                                    <div class="icon" @click.stop="openDetail(installed.appKey)">
                                        <el-avatar
                                            shape="square"
                                            :size="66"
                                            :src="'data:image/png;base64,' + installed.icon"
                                        />
                                    </div>
                                </el-col>
                                <el-col :xs="24" :sm="21" :md="21" :lg="20" :xl="20">
                                    <div class="a-detail">
                                        <div class="d-name">
                                            <el-button link type="info">
                                                <el-tooltip effect="dark" :content="installed.name" placement="top">
                                                    <span class="name">{{ installed.name }}</span>
                                                </el-tooltip>
                                            </el-button>
                                            <span class="status">
                                                <Status :key="installed.status" :status="installed.status"></Status>
                                            </span>
                                            <span class="msg">
                                                <el-popover
                                                    v-if="isAppErr(installed)"
                                                    placement="bottom"
                                                    :width="400"
                                                    trigger="hover"
                                                    :content="installed.message"
                                                    :popper-options="options"
                                                >
                                                    <template #reference>
                                                        <el-button link type="danger">
                                                            <el-icon><Warning /></el-icon>
                                                        </el-button>
                                                    </template>
                                                    <div class="app-error">
                                                        {{ installed.message }}
                                                    </div>
                                                </el-popover>
                                            </span>
                                            <span class="ml-1">
                                                <el-tooltip effect="dark" :content="$t('app.toFolder')" placement="top">
                                                    <el-button type="primary" link @click="toFolder(installed.path)">
                                                        <el-icon>
                                                            <FolderOpened />
                                                        </el-icon>
                                                    </el-button>
                                                </el-tooltip>
                                            </span>
                                            <span class="ml-1">
                                                <el-tooltip
                                                    effect="dark"
                                                    :content="$t('commons.button.log')"
                                                    placement="top"
                                                >
                                                    <el-button
                                                        type="primary"
                                                        link
                                                        @click="openLog(installed)"
                                                        :disabled="installed.status === 'DownloadErr'"
                                                    >
                                                        <el-icon><Tickets /></el-icon>
                                                    </el-button>
                                                </el-tooltip>
                                            </span>

                                            <el-button
                                                class="h-button"
                                                plain
                                                round
                                                size="small"
                                                @click="openUploads(installed.appKey, installed.name)"
                                                v-if="mode === 'installed'"
                                            >
                                                {{ $t('database.loadBackup') }}
                                            </el-button>
                                            <el-button
                                                class="h-button"
                                                plain
                                                round
                                                size="small"
                                                @click="openBackups(installed.appKey, installed.name, installed.status)"
                                                v-if="mode === 'installed'"
                                            >
                                                {{ $t('commons.button.backup') }}
                                            </el-button>
                                            <el-button
                                                class="h-button"
                                                plain
                                                round
                                                size="small"
                                                :disabled="installed.status === 'Upgrading'"
                                                @click="openOperate(installed, 'ignore')"
                                                v-if="mode === 'upgrade'"
                                            >
                                                {{ $t('commons.button.ignore') }}
                                            </el-button>
                                            <el-button
                                                class="h-button"
                                                plain
                                                round
                                                size="small"
                                                :disabled="
                                                    (installed.status !== 'Running' &&
                                                        installed.status !== 'UpgradeErr') ||
                                                    installed.appStatus === 'TakeDown'
                                                "
                                                @click="openOperate(installed, 'upgrade')"
                                                v-if="mode === 'upgrade'"
                                            >
                                                {{ $t('commons.button.upgrade') }}
                                            </el-button>
                                        </div>
                                        <div class="d-description flex flex-wrap items-center justify-start gap-1.5">
                                            <el-button class="tagMargin" plain size="small">
                                                {{ $t('app.version') }}{{ $t('commons.colon') }}{{ installed.version }}
                                            </el-button>

                                            <el-button
                                                v-if="installed.httpPort > 0"
                                                @click="goDashboard(installed.httpPort, 'http')"
                                                class="tagMargin"
                                                icon="Position"
                                                plain
                                                size="small"
                                            >
                                                {{ $t('app.busPort') }}{{ $t('commons.colon') }}{{ installed.httpPort }}
                                            </el-button>

                                            <el-button
                                                v-if="installed.httpsPort > 0"
                                                @click="goDashboard(installed.httpsPort, 'https')"
                                                class="tagMargin"
                                                icon="Position"
                                                plain
                                                size="small"
                                            >
                                                {{ $t('app.busPort') }}{{ $t('commons.colon')
                                                }}{{ installed.httpsPort }}
                                            </el-button>
                                        </div>
                                        <div class="description">
                                            <span>
                                                {{ $t('app.alreadyRun') }}{{ $t('commons.colon') }}
                                                {{ getAge(installed.createdAt) }}
                                            </span>
                                        </div>
                                        <div class="app-divider" v-if="mode === 'installed'" />
                                        <div
                                            class="d-button flex flex-wrap items-center justify-start gap-1.5"
                                            v-if="mode === 'installed' && installed.status != 'Installing'"
                                        >
                                            <el-button
                                                class="app-button"
                                                v-for="(button, key) in buttons"
                                                :key="key"
                                                :type="button.disabled && button.disabled(installed) ? 'info' : ''"
                                                plain
                                                round
                                                size="small"
                                                @click="button.click(installed)"
                                                :disabled="button.disabled && button.disabled(installed)"
                                            >
                                                {{ button.label }}
                                            </el-button>
                                        </div>
                                    </div>
                                </el-col>
                            </el-row>
                        </el-card>
                    </div>
                </el-col>
            </el-row>
            <div class="page-button" v-if="mode === 'installed'">
                <fu-table-pagination
                    v-model:current-page="paginationConfig.currentPage"
                    v-model:page-size="paginationConfig.pageSize"
                    v-bind="paginationConfig"
                    @change="search"
                    :layout="'total, sizes, prev, pager, next, jumper'"
                />
            </div>
        </template>
    </LayoutContent>
    <Backups ref="backupRef" @close="search" />
    <Uploads ref="uploadRef" />
    <AppResources ref="checkRef" @close="search" />
    <AppDelete ref="deleteRef" @close="search" />
    <AppParams ref="appParamRef" />
    <AppUpgrade ref="upgradeRef" @close="search" />
    <PortJumpDialog ref="dialogPortJumpRef" />
    <AppIgnore ref="ignoreRef" @close="search" />
    <ComposeLogs ref="composeLogRef" />
    <AppDetail ref="appDetail" />
</template>

<script lang="ts" setup>
import {
    SearchAppInstalled,
    InstalledOp,
    SyncInstalledApp,
    AppInstalledDeleteCheck,
    GetAppTags,
} from '@/api/modules/app';
import { onMounted, onUnmounted, reactive, ref } from 'vue';
import i18n from '@/lang';
import { ElMessageBox } from 'element-plus';
import Backups from '@/components/backup/index.vue';
import Uploads from '@/components/upload/index.vue';
import PortJumpDialog from '@/components/port-jump/index.vue';
import AppResources from './check/index.vue';
import AppDelete from './delete/index.vue';
import AppParams from './detail/index.vue';
import AppUpgrade from './upgrade/index.vue';
import AppIgnore from './ignore/index.vue';
import AppDetail from '../detail/index.vue';
import ComposeLogs from '@/components/compose-log/index.vue';
import { App } from '@/api/interface/app';
import Status from '@/components/status/index.vue';
import { getAge, getLanguage } from '@/utils/util';
import { useRouter } from 'vue-router';
import { MsgSuccess } from '@/utils/message';
import { toFolder } from '@/global/business';

const data = ref<any>();
const loading = ref(false);
const syncLoading = ref(false);
let timer: NodeJS.Timer | null = null;
const paginationConfig = reactive({
    cacheSizeKey: 'app-installed-page-size',
    currentPage: 1,
    pageSize: Number(localStorage.getItem('app-installed-page-size')) || 20,
    total: 0,
});
const open = ref(false);
const operateReq = reactive({
    installId: 0,
    operate: '',
    detailId: 0,
});
const backupRef = ref();
const uploadRef = ref();
const checkRef = ref();
const deleteRef = ref();
const appParamRef = ref();
const upgradeRef = ref();
const ignoreRef = ref();
const dialogPortJumpRef = ref();
const composeLogRef = ref();
const tags = ref<App.Tag[]>([]);
const activeTag = ref('all');
const searchReq = reactive({
    page: 1,
    pageSize: 20,
    name: '',
    tags: [],
    update: false,
    sync: false,
});
const router = useRouter();
const activeName = ref(i18n.global.t('app.installed'));
const mode = ref('installed');
const moreTag = ref('');
const language = getLanguage();
const appDetail = ref();
const options = {
    modifiers: [
        {
            name: 'flip',
            options: {
                padding: 5,
                fallbackPlacements: ['bottom-start', 'top-start', 'right', 'left'],
            },
        },
    ],
};

const sync = () => {
    ElMessageBox.confirm(i18n.global.t('app.syncAllAppHelper'), i18n.global.t('app.sync'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
        type: 'info',
    })
        .then(async () => {
            syncLoading.value = true;
            try {
                await SyncInstalledApp();
                MsgSuccess(i18n.global.t('app.syncSuccess'));
                await search();
            } finally {
                syncLoading.value = false;
            }
        })
        .catch(() => {});
};

const changeTag = (key: string) => {
    searchReq.tags = [];
    activeTag.value = key;
    if (key !== 'all') {
        searchReq.tags = [key];
    }
    const index = tags.value.findIndex((tag) => tag.key === key);
    if (index > 6) {
        moreTag.value = key;
    } else {
        moreTag.value = '';
    }
    search();
};

const getTagValue = (key: string) => {
    const tag = tags.value.find((tag) => tag.key === key);
    if (tag) {
        return language == 'zh' || language == 'tw' ? tag.name : tag.key;
    }
};

const search = async () => {
    searchReq.page = paginationConfig.currentPage;
    searchReq.pageSize = paginationConfig.pageSize;
    const res = await SearchAppInstalled(searchReq);
    data.value = res.data.items;
    paginationConfig.total = res.data.total;
    GetAppTags().then((res) => {
        tags.value = res.data;
    });
};

const goDashboard = async (port: any, protocol: string) => {
    dialogPortJumpRef.value.acceptParams({ port: port, protocol: protocol });
};

const openDetail = (appKey: string) => {
    appDetail.value.acceptParams(appKey, 'detail');
};

const openOperate = (row: any, op: string) => {
    operateReq.installId = row.id;
    operateReq.operate = op;
    if (op == 'upgrade' || op == 'ignore') {
        upgradeRef.value.acceptParams(row.id, row.name, row.dockerCompose, op, row.app);
    } else if (op == 'delete') {
        AppInstalledDeleteCheck(row.id).then(async (res) => {
            const items = res.data;
            if (res.data && res.data.length > 0) {
                checkRef.value.acceptParams({ items: items, key: row.appKey, installID: row.id });
            } else {
                deleteRef.value.acceptParams(row);
            }
        });
    } else {
        onOperate(op);
    }
};

const openIgnore = () => {
    ignoreRef.value.acceptParams();
};

const operate = async () => {
    open.value = false;
    loading.value = true;
    await InstalledOp(operateReq)
        .then(() => {
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            searchReq.sync = true;
            search();
            setTimeout(() => {
                search();
            }, 3000);
            setTimeout(() => {
                search();
            }, 15000);
        })
        .catch(() => {
            search();
        })
        .finally(() => {
            loading.value = false;
        });
};

const onOperate = async (operation: string) => {
    ElMessageBox.confirm(
        i18n.global.t('app.operatorHelper', [i18n.global.t('app.' + operation)]),
        i18n.global.t('app.' + operation),
        {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
            type: 'info',
        },
    ).then(() => {
        operate();
    });
};

const buttons = [
    {
        label: i18n.global.t('app.sync'),
        click: (row: any) => {
            openOperate(row, 'sync');
        },
        disabled: (row: any) => {
            return row.status === 'DownloadErr' || row.status === 'Upgrading' || row.status === 'Rebuilding';
        },
    },
    {
        label: i18n.global.t('app.rebuild'),
        click: (row: any) => {
            openOperate(row, 'rebuild');
        },
        disabled: (row: any) => {
            return row.status === 'DownloadErr' || row.status === 'Upgrading' || row.status === 'Rebuilding';
        },
    },
    {
        label: i18n.global.t('app.restart'),
        click: (row: any) => {
            openOperate(row, 'restart');
        },
        disabled: (row: any) => {
            return row.status === 'DownloadErr' || row.status === 'Upgrading' || row.status === 'Rebuilding';
        },
    },
    {
        label: i18n.global.t('app.start'),
        click: (row: any) => {
            openOperate(row, 'start');
        },
        disabled: (row: any) => {
            return (
                row.status === 'Running' ||
                row.status === 'Error' ||
                row.status === 'DownloadErr' ||
                row.status === 'Upgrading' ||
                row.status === 'Rebuilding'
            );
        },
    },
    {
        label: i18n.global.t('app.stop'),
        click: (row: any) => {
            openOperate(row, 'stop');
        },
        disabled: (row: any) => {
            return (
                row.status !== 'Running' ||
                row.status === 'DownloadErr' ||
                row.status === 'Upgrading' ||
                row.status === 'Rebuilding'
            );
        },
    },
    {
        label: i18n.global.t('commons.button.uninstall'),
        click: (row: any) => {
            openOperate(row, 'delete');
        },
    },
    {
        label: i18n.global.t('app.params'),
        click: (row: any) => {
            openParam(row);
        },
        disabled: (row: any) => {
            return row.status === 'DownloadErr' || row.status === 'Upgrading' || row.status === 'Rebuilding';
        },
    },
];

const openBackups = (key: string, name: string, status: string) => {
    let params = {
        type: 'app',
        name: key,
        detailName: name,
        status: status,
    };
    backupRef.value.acceptParams(params);
};

const openUploads = (key: string, name: string) => {
    let params = {
        type: 'app',
        name: key,
        detailName: name,
    };
    uploadRef.value.acceptParams(params);
};

const openParam = (row: any) => {
    appParamRef.value.acceptParams({ id: row.id });
};

const isAppErr = (row: any) => {
    return row.status.includes('Err') || row.status.includes('Error') || row.status.includes('UnHealthy');
};

const quickJump = () => {
    router.push({ name: 'ContainerSetting' });
};

const openLog = (row: any) => {
    composeLogRef.value.acceptParams({ compose: row.path + '/docker-compose.yml', resource: row.name });
};

onMounted(() => {
    const path = router.currentRoute.value.path;
    if (path == '/apps/upgrade') {
        activeName.value = i18n.global.t('app.canUpgrade');
        mode.value = 'upgrade';
        searchReq.update = true;
    }
    loading.value = true;
    search();
    loading.value = false;
    setTimeout(() => {
        searchReq.sync = true;
        search();
    }, 1000);
    timer = setInterval(() => {
        search();
    }, 1000 * 30);
});

onUnmounted(() => {
    clearInterval(Number(timer));
    timer = null;
});
</script>

<style scoped lang="scss">
@use '../index';
@media only screen and (max-width: 1400px) {
    .install-card-col-12 {
        max-width: 100%;
        flex: 0 0 100%;
        .a-detail {
            .d-name {
                .name {
                    max-width: 300px;
                }
            }
        }
    }
}

@media only screen and (max-width: 1499px) {
    .upgrade-card-col-12 {
        max-width: 100%;
        flex: 0 0 100%;
        .a-detail {
            .d-name {
                .name {
                    max-width: 300px;
                }
            }
        }
    }
}

.app-error {
    max-height: 500px;
    overflow-y: auto;
}
.d-button {
    .el-button + .el-button {
        margin-left: 0;
    }
}
.d-description {
    .el-button + .el-button {
        margin-left: 0;
    }
}
</style>
