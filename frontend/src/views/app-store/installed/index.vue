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
                    </div>
                </el-col>
                <el-col :xs="24" :sm="4" :md="4" :lg="4" :xl="4">
                    <div class="search-button">
                        <el-input
                            class="table-button"
                            v-model="searchReq.name"
                            clearable
                            @clear="search()"
                            suffix-icon="Search"
                            @keyup.enter="search()"
                            @change="search()"
                            :placeholder="$t('commons.button.search')"
                        ></el-input>
                    </div>
                </el-col>
            </el-row>
        </template>
        <template #rightButton>
            <el-button @click="sync" type="primary" link v-if="mode === 'installed' && data != null">
                {{ $t('app.sync') }}
            </el-button>
            <el-button @click="openIngore" type="primary" link v-if="mode === 'upgrade'">
                {{ $t('app.showIgnore') }}
            </el-button>
        </template>

        <template #main>
            <el-alert type="info" :closable="false" v-if="mode === 'installed'">
                <template #default>
                    <span>
                        <span>{{ $t('app.installHelper') }}</span>
                        <el-link
                            style="font-size: 12px; margin-left: 5px"
                            icon="Position"
                            @click="quickJump()"
                            type="primary"
                        >
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
                    class="install-card-col-12"
                >
                    <div class="install-card">
                        <el-card class="e-card">
                            <el-row :gutter="20">
                                <el-col :xs="3" :sm="3" :md="3" :lg="4" :xl="4">
                                    <div class="icon">
                                        <el-avatar
                                            shape="square"
                                            :size="66"
                                            :src="'data:image/png;base64,' + installed.app.icon"
                                        />
                                    </div>
                                </el-col>
                                <el-col :xs="24" :sm="21" :md="21" :lg="20" :xl="20">
                                    <div class="a-detail">
                                        <div class="d-name">
                                            <el-button link type="info">
                                                <span class="name">{{ installed.name }}</span>
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
                                                >
                                                    <template #reference>
                                                        <el-button link type="primary">
                                                            {{ $t('app.detail') }}
                                                        </el-button>
                                                    </template>
                                                </el-popover>
                                            </span>

                                            <el-tooltip effect="dark" :content="$t('app.toFolder')" placement="top">
                                                <el-button type="primary" link @click="toFolder(installed.path)">
                                                    <el-icon>
                                                        <FolderOpened />
                                                    </el-icon>
                                                </el-button>
                                            </el-tooltip>

                                            <el-button
                                                class="h-button"
                                                type="primary"
                                                plain
                                                round
                                                size="small"
                                                :disabled="
                                                    installed.status !== 'Running' ||
                                                    installed.app.status === 'TakeDown'
                                                "
                                                @click="openUploads(installed.app.key, installed.name)"
                                                v-if="mode === 'installed'"
                                            >
                                                {{ $t('database.loadBackup') }}
                                            </el-button>
                                            <el-button
                                                class="h-button"
                                                type="primary"
                                                plain
                                                round
                                                size="small"
                                                :disabled="
                                                    installed.status !== 'Running' ||
                                                    installed.app.status === 'TakeDown'
                                                "
                                                @click="openBackups(installed.app.key, installed.name)"
                                                v-if="mode === 'installed'"
                                            >
                                                {{ $t('commons.button.backup') }}
                                            </el-button>
                                            <el-button
                                                class="h-button"
                                                type="primary"
                                                plain
                                                round
                                                size="small"
                                                @click="openOperate(installed, 'ignore')"
                                                v-if="mode === 'upgrade'"
                                            >
                                                {{ $t('commons.button.ignore') }}
                                            </el-button>
                                            <el-button
                                                class="h-button"
                                                type="primary"
                                                plain
                                                round
                                                size="small"
                                                :disabled="
                                                    (installed.status !== 'Running' &&
                                                        installed.status !== 'UpgradeErr') ||
                                                    installed.app.status === 'TakeDown'
                                                "
                                                @click="openOperate(installed, 'upgrade')"
                                                v-if="mode === 'upgrade'"
                                            >
                                                {{ $t('commons.button.upgrade') }}
                                            </el-button>
                                        </div>
                                        <div class="d-description">
                                            <el-tag class="middle-center">
                                                {{ $t('app.version') }}：{{ installed.version }}
                                            </el-tag>
                                            <el-tag
                                                class="middle-center"
                                                v-if="installed.httpPort > 0"
                                                @click="goDashboard(installed.httpPort)"
                                            >
                                                <el-icon class="middle-center"><Position /></el-icon>
                                                {{ $t('app.busPort') }}：{{ installed.httpPort }}
                                            </el-tag>
                                            <div class="description">
                                                <span>
                                                    {{ $t('app.alreadyRun') }}： {{ getAge(installed.createdAt) }}
                                                </span>
                                            </div>
                                        </div>
                                        <div class="app-divider" />
                                        <div
                                            class="d-button"
                                            v-if="mode === 'installed' && installed.status != 'Installing'"
                                        >
                                            <el-button
                                                v-for="(button, key) in buttons"
                                                :key="key"
                                                :type="
                                                    button.disabled && button.disabled(installed) ? 'info' : 'primary'
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
        </template>
    </LayoutContent>
    <Backups ref="backupRef" @close="search" />
    <Uploads ref="uploadRef" />
    <AppResources ref="checkRef" />
    <AppDelete ref="deleteRef" @close="search" />
    <AppParams ref="appParamRef" />
    <AppUpgrade ref="upgradeRef" @close="search" />
    <PortJumpDialog ref="dialogPortJumpRef" />
    <AppIgnore ref="ignoreRef" @close="search" />
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
import { App } from '@/api/interface/app';
import Status from '@/components/status/index.vue';
import { getAge } from '@/utils/util';
import { useRouter } from 'vue-router';
import { MsgSuccess } from '@/utils/message';
import { toFolder } from '@/global/business';
import { useI18n } from 'vue-i18n';

const data = ref<any>();
const loading = ref(false);
const syncLoading = ref(false);
let timer: NodeJS.Timer | null = null;
const paginationConfig = reactive({
    currentPage: 1,
    pageSize: 20,
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
const tags = ref<App.Tag[]>([]);
const activeTag = ref('all');
const searchReq = reactive({
    page: 1,
    pageSize: 15,
    name: '',
    tags: [],
    update: false,
});
const router = useRouter();
const activeName = ref(i18n.global.t('app.installed'));
const mode = ref('installed');

const language = useI18n().locale.value;

const sync = () => {
    syncLoading.value = true;
    SyncInstalledApp()
        .then(() => {
            MsgSuccess(i18n.global.t('app.syncSuccess'));
            search();
        })
        .finally(() => {
            syncLoading.value = false;
        });
};

const changeTag = (key: string) => {
    searchReq.tags = [];
    activeTag.value = key;
    if (key !== 'all') {
        searchReq.tags = [key];
    }
    search();
};

const search = () => {
    loading.value = true;
    searchReq.page = paginationConfig.currentPage;
    searchReq.pageSize = paginationConfig.pageSize;
    SearchAppInstalled(searchReq)
        .then((res) => {
            data.value = res.data.items;
            paginationConfig.total = res.data.total;
        })
        .finally(() => {
            loading.value = false;
        });
    GetAppTags().then((res) => {
        tags.value = res.data;
    });
};

const goDashboard = async (port: any) => {
    dialogPortJumpRef.value.acceptParams({ port: port });
};

const openOperate = (row: any, op: string) => {
    operateReq.installId = row.id;
    operateReq.operate = op;
    if (op == 'upgrade' || op == 'ignore') {
        upgradeRef.value.acceptParams(row.id, row.name, op);
    } else if (op == 'delete') {
        AppInstalledDeleteCheck(row.id).then(async (res) => {
            const items = res.data;
            if (res.data && res.data.length > 0) {
                checkRef.value.acceptParams({ items: items });
            } else {
                deleteRef.value.acceptParams(row);
            }
        });
    } else {
        onOperate(op);
    }
};

const openIngore = () => {
    ignoreRef.value.acceptParams();
};

const operate = async () => {
    open.value = false;
    loading.value = true;
    await InstalledOp(operateReq)
        .then(() => {
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            search();
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
            return row.status === 'DownloadErr' || row.status === 'Upgrading';
        },
    },
    {
        label: i18n.global.t('app.rebuild'),
        click: (row: any) => {
            openOperate(row, 'rebuild');
        },
        disabled: (row: any) => {
            return row.status === 'DownloadErr' || row.status === 'Upgrading';
        },
    },
    {
        label: i18n.global.t('app.restart'),
        click: (row: any) => {
            openOperate(row, 'restart');
        },
        disabled: (row: any) => {
            return row.status === 'DownloadErr' || row.status === 'Upgrading';
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
                row.status === 'Upgrading'
            );
        },
    },
    {
        label: i18n.global.t('app.stop'),
        click: (row: any) => {
            openOperate(row, 'stop');
        },
        disabled: (row: any) => {
            return row.status !== 'Running' || row.status === 'DownloadErr' || row.status === 'Upgrading';
        },
    },
    {
        label: i18n.global.t('commons.button.delete'),
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
            return row.status === 'DownloadErr' || row.status === 'Upgrading';
        },
    },
];

const openBackups = (key: string, name: string) => {
    let params = {
        type: 'app',
        name: key,
        detailName: name,
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
    appParamRef.value.acceptParams({ app: row.app, id: row.id });
};

const isAppErr = (row: any) => {
    return row.status.includes('Err') || row.status.includes('Error') || row.status.includes('UnHealthy');
};

const quickJump = () => {
    router.push({ name: 'ContainerSetting' });
};

onMounted(() => {
    const path = router.currentRoute.value.path;
    if (path == '/apps/upgrade') {
        activeName.value = i18n.global.t('app.canUpgrade');
        mode.value = 'upgrade';
        searchReq.update = true;
    }
    search();
    timer = setInterval(() => {
        search();
    }, 10000 * 6);
});

onUnmounted(() => {
    clearInterval(Number(timer));
    timer = null;
});
</script>

<style scoped lang="scss">
@import '../index.scss';
@media only screen and (max-width: 1300px) {
    .install-card-col-12 {
        max-width: 100%;
        flex: 0 0 100%;
    }
}
</style>
