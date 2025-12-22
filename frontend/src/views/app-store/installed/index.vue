<template>
    <LayoutContent v-loading="loading || syncLoading" :title="activeName">
        <template #search>
            <Tags @change="changeTag" hideKey="Runtime" />
        </template>
        <template #leftToolBar>
            <el-button @click="sync" type="primary" plain v-if="mode === 'installed' && data != null">
                {{ $t('commons.button.refresh') }}
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
                <MainDiv :heightDiff="mode === 'upgrade' ? 280 : 300">
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
                    <div class="update-prompt" v-if="data === null">
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
                            <AppCard
                                :installed="installed"
                                :mode="mode"
                                :defaultLink="defaultLink"
                                :currentNode="currentNode"
                                @open-detail="openDetail(installed.appKey)"
                                @open-backups="openBackups(installed)"
                                @open-log="openLog(installed)"
                                @open-terminal="openTerminal(installed)"
                                @open-operate="openOperate(installed, 'upgrade')"
                                @favorite-install="favoriteInstall(installed)"
                                @to-folder="routerToFileWithPath(installed.path)"
                                @open-uploads="openUploads(installed.appKey, installed.name)"
                                @jump-to-path="jumpToPath(router, '/settings/panel')"
                                @to-container="toContainer(installed)"
                                @ignore-app="ignoreApp(installed)"
                            >
                                <template #buttons>
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
                                </template>
                            </AppCard>
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
    <AppIgnore ref="ignoreRef" @close="search" />
    <ComposeLogs ref="composeLogRef" />
    <TerminalDialog ref="dialogTerminalRef" />
    <TaskLog ref="taskLogRef" @close="search" />
    <Detail ref="detailRef" />
    <IgnoreApp ref="ignoreAppRef" @close="search" />
</template>

<script lang="ts" setup>
import AppCard from '@/views/app-store/installed/app/card.vue';
import Backups from '@/components/backup/index.vue';
import Uploads from '@/components/upload/index.vue';
import AppResources from './check/index.vue';
import AppDelete from './delete/index.vue';
import AppParams from './detail/index.vue';
import AppUpgrade from './upgrade/index.vue';
import AppIgnore from './ignore/index.vue';
import TaskLog from '@/components/log/task/index.vue';
import Detail from '@/views/app-store/detail/index.vue';
import Tags from '@/views/app-store/components/tag.vue';
import MainDiv from '@/components/main-div/index.vue';
import ComposeLogs from '@/components/log/compose/index.vue';
import IgnoreApp from '@/views/app-store/installed/ignore/create/index.vue';
import TerminalDialog from '@/views/container/container/terminal/index.vue';

import { searchAppInstalled, installedOp, appInstalledDeleteCheck } from '@/api/modules/app';
import { onMounted, onUnmounted, reactive, ref } from 'vue';
import i18n from '@/lang';
import { ElMessageBox } from 'element-plus';
import { App } from '@/api/interface/app';
import { jumpToPath } from '@/utils/util';
import { useRouter } from 'vue-router';
import { MsgSuccess } from '@/utils/message';
import { getAgentSettingByKey } from '@/api/modules/setting';
import { routerToFileWithPath, routerToNameWithQuery } from '@/utils/router';
import { useGlobalStore } from '@/composables/useGlobalStore';
const { currentNode, isMaster, currentNodeAddr } = useGlobalStore();

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
const composeLogRef = ref();
const dialogTerminalRef = ref();
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

    localStorage.setItem('app-installed-page-size', String(searchReq.pageSize));

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
        upgradeRef.value.acceptParams(row, op);
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

const ignoreApp = (row: App.AppInstalled) => {
    ignoreAppRef.value.acceptParams(row);
};

const toContainer = async (row: App.AppInstalled) => {
    routerToNameWithQuery('ContainerItem', {
        filters: 'com.docker.compose.project=' + row.serviceName,
        uncached: true,
    });
};

const openBackups = (row: App.AppInstalled) => {
    let params = {
        type: 'app',
        name: row.appKey,
        detailName: row.name,
        status: row.status,
        appInstallID: row.id,
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

const openTerminal = (row: any) => {
    const title = i18n.global.t('app.app') + ' ' + row.name;
    dialogTerminalRef.value!.acceptParams({ containerID: row.container, title: title });
};

const getConfig = async () => {
    try {
        const res = await getAgentSettingByKey('SystemIP');
        if (res.data != '') {
            defaultLink.value = res.data;
            return;
        }
        if (!isMaster.value || currentNodeAddr.value != '127.0.0.1') {
            defaultLink.value = currentNodeAddr.value;
        }
    } catch (error) {}
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

.d-button {
    .el-button + .el-button {
        margin-left: 0;
    }
}
</style>
