<template>
    <LayoutContent v-loading="loading || syncLoading" :title="activeName">
        <template #search>
            <Tags @change="changeTag" hideKey="Runtime" />
        </template>
        <template #leftToolBar>
            <el-button @click="sync" type="primary" plain v-if="mode === 'installed' && data != null">
                {{ $t('commons.button.sync') }}
            </el-button>
            <el-button @click="openIgnore" type="primary" plain v-if="mode === 'upgrade'">
                {{ $t('app.showIgnore') }}
            </el-button>
        </template>
        <template #rightToolBar>
            <TableSearch @search="search()" v-model:searchName="searchReq.name" />
        </template>
        <template #main>
            <div>
                <MainDiv :heightDiff="mode === 'upgrade' ? 270 : 300">
                    <el-alert
                        type="info"
                        :title="$t('app.upgradeHelper')"
                        :closable="false"
                        v-if="mode === 'upgrade'"
                    />

                    <el-alert type="info" :closable="false" v-if="mode === 'installed'">
                        <template #title>
                            <span class="flx-align-center">
                                {{ $t('app.installHelper') }}
                                <el-link
                                    class="ml-5"
                                    icon="Position"
                                    @click="jumpToPath(router, '/containers/setting')"
                                    type="primary"
                                >
                                    {{ $t('firewall.quickJump') }}
                                </el-link>
                                　
                            </span>
                        </template>
                    </el-alert>
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
                        >
                            <div class="install-card">
                                <el-card class="e-card">
                                    <el-row :gutter="10">
                                        <el-col :xs="3" :sm="3" :md="3" :lg="4" :xl="4">
                                            <div class="icon">
                                                <el-avatar
                                                    @click="openDetail(installed.appKey)"
                                                    shape="square"
                                                    :size="66"
                                                    :src="'data:image/png;base64,' + installed.icon"
                                                />
                                            </div>
                                        </el-col>
                                        <el-col :xs="24" :sm="21" :md="21" :lg="20" :xl="20">
                                            <div class="a-detail">
                                                <div class="d-name">
                                                    <div class="flex items-center justify-between">
                                                        <div class="min-w-50 flex items-center justify-start gap-1">
                                                            <el-button link type="info">
                                                                <el-tooltip :content="installed.name" placement="top">
                                                                    <span class="name">{{ installed.name }}</span>
                                                                </el-tooltip>
                                                            </el-button>
                                                            <span class="status">
                                                                <Status
                                                                    :key="installed.status"
                                                                    :status="installed.status"
                                                                ></Status>
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
                                                                <el-tooltip
                                                                    effect="dark"
                                                                    :content="$t('app.toFolder')"
                                                                    placement="top"
                                                                >
                                                                    <el-button
                                                                        type="primary"
                                                                        link
                                                                        @click="toFolder(installed.path)"
                                                                    >
                                                                        <el-icon>
                                                                            <FolderOpened />
                                                                        </el-icon>
                                                                    </el-button>
                                                                </el-tooltip>
                                                            </span>
                                                            <span class="ml-1">
                                                                <el-tooltip
                                                                    v-if="mode !== 'upgrade'"
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
                                                            <span class="ml-1" v-if="mode === 'installed'">
                                                                <el-tooltip
                                                                    effect="dark"
                                                                    :content="$t('website.cancelFavorite')"
                                                                    placement="top-start"
                                                                    v-if="installed.favorite"
                                                                >
                                                                    <el-button
                                                                        link
                                                                        size="large"
                                                                        icon="StarFilled"
                                                                        type="warning"
                                                                        @click="favoriteInstall(installed)"
                                                                    ></el-button>
                                                                </el-tooltip>
                                                                <el-tooltip
                                                                    effect="dark"
                                                                    :content="$t('website.favorite')"
                                                                    placement="top-start"
                                                                    v-else
                                                                >
                                                                    <el-button
                                                                        link
                                                                        icon="Star"
                                                                        type="info"
                                                                        @click="favoriteInstall(installed)"
                                                                    ></el-button>
                                                                </el-tooltip>
                                                            </span>
                                                        </div>
                                                        <div class="flex flex-wrap items-center justify-end gap-1">
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
                                                                @click="
                                                                    openBackups(
                                                                        installed.appKey,
                                                                        installed.name,
                                                                        installed.status,
                                                                    )
                                                                "
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
                                                                @click="ignoreApp(installed)"
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
                                                    </div>
                                                </div>
                                                <div
                                                    class="d-description flex flex-wrap items-center justify-start gap-1.5"
                                                >
                                                    <el-button class="mr-1" plain size="small">
                                                        {{ $t('app.version') }}{{ $t('commons.colon')
                                                        }}{{ installed.version }}
                                                    </el-button>
                                                    <el-button
                                                        v-if="installed.httpPort > 0"
                                                        class="mr-1"
                                                        plain
                                                        size="small"
                                                    >
                                                        {{ $t('commons.table.port') }}{{ $t('commons.colon')
                                                        }}{{ installed.httpPort }}
                                                    </el-button>
                                                    <el-button v-if="installed.httpsPort > 0" plain size="small">
                                                        {{ $t('commons.table.port') }}：{{ installed.httpsPort }}
                                                    </el-button>

                                                    <el-popover
                                                        placement="top-start"
                                                        trigger="hover"
                                                        v-if="
                                                            installed.appType == 'website' ||
                                                            installed.appKey?.startsWith('local')
                                                        "
                                                        :width="400"
                                                    >
                                                        <template #reference>
                                                            <el-button
                                                                plain
                                                                icon="Promotion"
                                                                size="small"
                                                                @click="openLink(defaultLink, installed)"
                                                            >
                                                                {{ $t('app.toLink') }}
                                                            </el-button>
                                                        </template>
                                                        <table>
                                                            <tbody>
                                                                <tr v-if="defaultLink != ''">
                                                                    <td v-if="installed.httpPort > 0">
                                                                        <el-button
                                                                            type="primary"
                                                                            link
                                                                            @click="
                                                                                toLink(
                                                                                    'http://' +
                                                                                        defaultLink +
                                                                                        ':' +
                                                                                        installed.httpPort,
                                                                                )
                                                                            "
                                                                        >
                                                                            {{
                                                                                'http://' +
                                                                                defaultLink +
                                                                                ':' +
                                                                                installed.httpPort
                                                                            }}
                                                                        </el-button>
                                                                    </td>
                                                                </tr>
                                                                <tr v-if="defaultLink != ''">
                                                                    <td v-if="installed.httpsPort > 0">
                                                                        <el-button
                                                                            type="primary"
                                                                            link
                                                                            @click="
                                                                                toLink(
                                                                                    'https://' +
                                                                                        defaultLink +
                                                                                        ':' +
                                                                                        installed.httpsPort,
                                                                                )
                                                                            "
                                                                        >
                                                                            {{
                                                                                'https://' +
                                                                                defaultLink +
                                                                                ':' +
                                                                                installed.httpsPort
                                                                            }}
                                                                        </el-button>
                                                                    </td>
                                                                </tr>
                                                                <tr v-if="installed.webUI != ''">
                                                                    <td>
                                                                        <el-button
                                                                            type="primary"
                                                                            link
                                                                            @click="toLink(installed.webUI)"
                                                                        >
                                                                            {{ installed.webUI }}
                                                                        </el-button>
                                                                    </td>
                                                                </tr>
                                                            </tbody>
                                                        </table>
                                                        <span v-if="defaultLink == '' && installed.webUI == ''">
                                                            {{ $t('app.webUIConfig') }}
                                                            <el-link
                                                                icon="Position"
                                                                @click="jumpToPath(router, '/settings/panel')"
                                                                type="primary"
                                                            >
                                                                {{ $t('firewall.quickJump') }}
                                                            </el-link>
                                                        </span>
                                                    </el-popover>
                                                </div>
                                                <div class="description">
                                                    <span>
                                                        {{ $t('app.alreadyRun') }}{{ $t('commons.colon') }}
                                                        {{ getAge(installed.createdAt) }}
                                                    </span>
                                                </div>
                                                <div class="app-divider" />
                                                <div
                                                    class="d-button flex flex-wrap items-center justify-start gap-1.5"
                                                    v-if="mode === 'installed' && installed.status != 'Installing'"
                                                >
                                                    <el-button
                                                        class="app-button"
                                                        v-for="(button, key) in buttons"
                                                        :key="key"
                                                        :type="
                                                            button.disabled && button.disabled(installed) ? 'info' : ''
                                                        "
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
                </MainDiv>
            </div>
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
    <Backups ref="backupRef" />
    <Uploads ref="uploadRef" />
    <AppResources ref="checkRef" @close="search" />
    <AppDelete ref="deleteRef" @close="search" />
    <AppParams ref="appParamRef" @close="search" />
    <AppUpgrade ref="upgradeRef" @close="search" />
    <PortJumpDialog ref="dialogPortJumpRef" />
    <AppIgnore ref="ignoreRef" @close="search" />
    <ComposeLogs ref="composeLogRef" />
    <TaskLog ref="taskLogRef" @close="search" />
    <Detail ref="detailRef" />
    <IgnoreApp ref="ignoreAppRef" @close="search" />
</template>

<script lang="ts" setup>
import { searchAppInstalled, installedOp, appInstalledDeleteCheck } from '@/api/modules/app';
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
import ComposeLogs from '@/components/log/compose/index.vue';
import { App } from '@/api/interface/app';
import Status from '@/components/status/index.vue';
import { getAge, jumpToPath, toLink } from '@/utils/util';
import { useRouter } from 'vue-router';
import { MsgSuccess } from '@/utils/message';
import { toFolder } from '@/global/business';
import TaskLog from '@/components/log/task/index.vue';
import Detail from '@/views/app-store/detail/index.vue';
import IgnoreApp from '@/views/app-store/installed/ignore/create/index.vue';
import { getAgentSettingByKey } from '@/api/modules/setting';
import Tags from '@/views/app-store/components/tag.vue';

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
    favorite: false,
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
const taskLogRef = ref();
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
const defaultLink = ref('');
const detailRef = ref();
const ignoreAppRef = ref();

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

const openDetail = (key: string) => {
    detailRef.value.acceptParams(key, 'install');
};

const changeTag = (key: string) => {
    searchReq.tags = [];
    if (key !== 'all') {
        searchReq.tags = [key];
    }
    search();
};

const search = async () => {
    searchReq.page = paginationConfig.currentPage;
    searchReq.pageSize = paginationConfig.pageSize;
    const res = await searchAppInstalled(searchReq);
    data.value = res.data.items;
    paginationConfig.total = res.data.total;
};

const sync = async () => {
    loading.value = true;
    const searchItem = {
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
        name: searchReq.name,
        tags: searchReq.tags,
        update: false,
        sync: true,
    };
    const res = await searchAppInstalled(searchItem);
    loading.value = false;
    data.value = res.data.items;
    paginationConfig.total = res.data.total;
};

const openOperate = (row: any, op: string) => {
    operateReq.installId = row.id;
    operateReq.operate = op;
    if (op == 'upgrade') {
        upgradeRef.value.acceptParams(row.id, row.name, row.dockerCompose, op, row.app);
    } else if (op == 'delete') {
        appInstalledDeleteCheck(row.id).then(async (res) => {
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

const favoriteInstall = (row: App.AppInstalled) => {
    operateReq.installId = row.id;
    operateReq.operate = 'favorite';
    operateReq.favorite = !row.favorite;
    operate();
};

const openIgnore = () => {
    ignoreRef.value.acceptParams();
};

const ignoreApp = (row: App.AppInstalled) => {
    ignoreAppRef.value.acceptParams(row);
};

const operate = async () => {
    open.value = false;
    loading.value = true;
    await installedOp(operateReq)
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
        i18n.global.t('app.operatorHelper', [i18n.global.t('commons.operate.' + operation)]),
        i18n.global.t('commons.operate.' + operation),
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
        label: i18n.global.t('commons.operate.rebuild'),
        click: (row: any) => {
            openOperate(row, 'rebuild');
        },
        disabled: (row: any) => {
            return (
                row.status === 'DownloadErr' ||
                row.status === 'Upgrading' ||
                row.status === 'Rebuilding' ||
                row.status === 'Uninstalling'
            );
        },
    },
    {
        label: i18n.global.t('commons.operate.restart'),
        click: (row: any) => {
            openOperate(row, 'restart');
        },
        disabled: (row: any) => {
            return (
                row.status === 'DownloadErr' ||
                row.status === 'Upgrading' ||
                row.status === 'Rebuilding' ||
                row.status === 'Uninstalling'
            );
        },
    },
    {
        label: i18n.global.t('commons.operate.start'),
        click: (row: any) => {
            openOperate(row, 'start');
        },
        disabled: (row: any) => {
            return (
                row.status === 'Running' ||
                row.status === 'Error' ||
                row.status === 'DownloadErr' ||
                row.status === 'Upgrading' ||
                row.status === 'Rebuilding' ||
                row.status === 'Uninstalling'
            );
        },
    },
    {
        label: i18n.global.t('commons.operate.stop'),
        click: (row: any) => {
            openOperate(row, 'stop');
        },
        disabled: (row: any) => {
            return (
                row.status !== 'Running' ||
                row.status === 'DownloadErr' ||
                row.status === 'Upgrading' ||
                row.status === 'Rebuilding' ||
                row.status === 'Uninstalling'
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
            return (
                row.status === 'DownloadErr' ||
                row.status === 'Upgrading' ||
                row.status === 'Rebuilding' ||
                row.status === 'Uninstalling'
            );
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

const openLog = (row: any) => {
    switch (row.status) {
        case 'Installing':
            taskLogRef.value.openWithResourceID('App', 'TaskInstall', row.id);
            break;
        default:
            composeLogRef.value.acceptParams({
                compose: row.path + '/docker-compose.yml',
                resource: row.name,
                container: row.container,
            });
    }
};

const getConfig = async () => {
    try {
        const res = await getAgentSettingByKey('SystemIP');
        if (res.data != '') {
            defaultLink.value = res.data;
        }
    } catch (error) {}
};

const openLink = (defaultLink: string, installed: App.AppInstalled) => {
    if (defaultLink != '' && installed.webUI != '') {
        return;
    }
    if (defaultLink == '' && installed.webUI == '') {
        return;
    }
    if (installed.webUI != '') {
        toLink(installed.webUI);
        return;
    }
    if (installed.httpsPort > 0) {
        toLink('https://' + defaultLink + ':' + installed.httpsPort);
        return;
    }
    if (installed.httpPort > 0) {
        toLink('http://' + defaultLink + ':' + installed.httpPort);
        return;
    }
};

onMounted(() => {
    getConfig();
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

.app-error {
    max-height: 500px;
    overflow-y: auto;
}
.d-name {
    .el-button + .el-button {
        margin-left: 0;
    }
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
