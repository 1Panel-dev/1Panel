<template>
    <LayoutContent v-loading="loading" :title="$t('app.installed')">
        <template #toolbar>
            <el-row :gutter="5">
                <el-col :span="20">
                    <!-- <div>
                        <el-button @click="changeTag('all')" type="primary" :plain="activeTag !== 'all'">
                            {{ $t('app.all') }}
                        </el-button>
                        <div v-for="item in tags" :key="item.key" style="display: inline">
                            <el-button
                                class="tag-button"
                                @click="changeTag(item.key)"
                                type="primary"
                                :plain="activeTag !== item.key"
                            >
                                {{ item.name }}
                            </el-button>
                        </div>
                    </div> -->
                </el-col>
                <el-col :span="4">
                    <div style="float: right">
                        <el-input
                            class="table-button"
                            v-model="searchName"
                            clearable
                            @clear="search()"
                            suffix-icon="Search"
                            @keyup.enter="search()"
                            @blur="search()"
                            :placeholder="$t('commons.button.search')"
                        ></el-input>
                    </div>
                </el-col>
            </el-row>
        </template>
        <template #rightButton>
            <el-button @click="sync" type="primary" link>{{ $t('app.sync') }}</el-button>
        </template>
        <template #main>
            <div class="divider"></div>
            <el-row :gutter="5">
                <el-col v-for="(installed, index) in data" :key="index" :span="12">
                    <div class="app-card">
                        <el-row :gutter="24">
                            <el-col :span="4">
                                <div class="icon">
                                    <el-avatar
                                        shape="square"
                                        :size="66"
                                        :src="'data:image/png;base64,' + installed.app.icon"
                                    />
                                </div>
                            </el-col>
                            <el-col :span="20">
                                <div class="a-detail">
                                    <div class="d-name">
                                        <span class="name">{{ installed.name }}</span>
                                        <span class="status">
                                            <el-popover
                                                v-if="installed.status === 'Error'"
                                                placement="bottom"
                                                :width="400"
                                                trigger="hover"
                                                :content="installed.message"
                                            >
                                                <template #reference>
                                                    <Status :key="installed.status" :status="installed.status"></Status>
                                                </template>
                                            </el-popover>
                                            <span v-else>
                                                <el-icon v-if="installed.status === 'Installing'" class="is-loading">
                                                    <Loading />
                                                </el-icon>
                                                <Status :key="installed.status" :status="installed.status"></Status>
                                            </span>
                                        </span>

                                        <el-button
                                            class="h-button"
                                            type="primary"
                                            link
                                            @click="openBackups(installed.id, installed.name)"
                                        >
                                            备份
                                        </el-button>
                                    </div>
                                    <div class="d-description">
                                        <el-tag>版本：{{ installed.version }}</el-tag>
                                        <el-tag>HTTP端口：{{ installed.httpPort }}</el-tag>
                                        <!-- <span class="description">
                                            {{ app.shortDesc }}
                                        </span> -->
                                        <div class="description">
                                            <span>已运行：12天</span>
                                        </div>
                                    </div>
                                    <div class="divider"></div>
                                    <div class="d-button">
                                        <el-button
                                            v-for="(button, key) in buttons"
                                            :key="key"
                                            type="primary"
                                            plain
                                            round
                                            size="small"
                                            @click="button.click(installed)"
                                            :disabled="button.disabled && button.disabled(installed)"
                                        >
                                            {{ button.label }}
                                        </el-button>
                                        <!-- <el-tag v-for="(tag, ind) in app.tags" :key="ind" :colr="getColor(ind)">
                                            {{ tag.name }}
                                        </el-tag> -->
                                    </div>
                                </div>
                            </el-col>
                        </el-row>
                    </div>
                </el-col>
            </el-row>

            <!-- <ComplexTable :pagination-config="paginationConfig" :data="data" @search="search" v-loading="loading">
                <el-table-column :label="$t('app.name')" prop="name" min-width="150px" show-overflow-tooltip>
                    <template #default="{ row }">
                        <el-link :underline="false" @click="openParam(row.id)" type="primary">
                            {{ row.name }}
                        </el-link>
                        <el-tag round effect="dark" v-if="row.canUpdate">{{ $t('app.canUpdate') }}</el-tag>
                    </template>
                </el-table-column>
                <el-table-column :label="$t('app.app')" prop="app.name" show-overflow-tooltip></el-table-column>
                <el-table-column :label="$t('app.version')" prop="version" show-overflow-tooltip></el-table-column>
                <el-table-column :label="$t('website.port')" prop="httpPort"></el-table-column>
                <el-table-column :label="$t('app.backup')">
                    <template #default="{ row }">
                        <el-link :underline="false" @click="openBackups(row.id, row.name)" type="primary">
                            {{ $t('app.backup') }} ({{ row.backups.length }})
                        </el-link>
                    </template>
                </el-table-column>

                <el-table-column :label="$t('app.status')">
                    <template #default="{ row }">
                        <el-popover
                            v-if="row.status === 'Error'"
                            placement="bottom"
                            :width="400"
                            trigger="hover"
                            :content="row.message"
                        >
                            <template #reference><Status :key="row.status" :status="row.status"></Status></template>
                        </el-popover>
                        <div v-else>
                            <el-icon v-if="row.status === 'Installing'" class="is-loading">
                                <Loading />
                            </el-icon>
                            <Status :key="row.status" :status="row.status"></Status>
                        </div>
                    </template>
                </el-table-column>
                <el-table-column
                    prop="createdAt"
                    :label="$t('commons.table.date')"
                    :formatter="dateFromat"
                    show-overflow-tooltip
                />
                <fu-table-operations
                    width="300px"
                    :ellipsis="10"
                    :buttons="buttons"
                    :label="$t('commons.table.operate')"
                    fixed="right"
                    fix
                />
            </ComplexTable> -->
        </template>
    </LayoutContent>
    <el-dialog
        v-model="open"
        :title="$t('commons.msg.operate')"
        :destroy-on-close="true"
        :close-on-click-modal="false"
        :before-close="handleClose"
        width="30%"
    >
        <div style="text-align: center">
            <p>{{ $t('app.versioneSelect') }}</p>
            <el-select v-model="operateReq.detailId">
                <el-option
                    v-for="(version, index) in versions"
                    :key="index"
                    :value="version.detailId"
                    :label="version.version"
                ></el-option>
            </el-select>
        </div>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose">{{ $t('commons.button.cancel') }}</el-button>
                <el-button
                    type="primary"
                    @click="operate"
                    :disabled="operateReq.operate == 'update' && versions == null"
                >
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </el-dialog>
    <Backups ref="backupRef" @close="search"></Backups>
    <AppResources ref="checkRef"></AppResources>
    <AppDelete ref="deleteRef" @close="search"></AppDelete>
    <AppParams ref="appParamRef"></AppParams>
</template>

<script lang="ts" setup>
import {
    SearchAppInstalled,
    InstalledOp,
    SyncInstalledApp,
    GetAppUpdateVersions,
    AppInstalledDeleteCheck,
} from '@/api/modules/app';
import LayoutContent from '@/layout/layout-content.vue';
import { onMounted, onUnmounted, reactive, ref } from 'vue';
// import ComplexTable from '@/components/complex-table/index.vue';
// import { dateFromat } from '@/utils/util';
import i18n from '@/lang';
import { ElMessage, ElMessageBox } from 'element-plus';
import Backups from './backups.vue';
import AppResources from './check/index.vue';
import AppDelete from './delete/index.vue';
import AppParams from './detail/index.vue';
import { App } from '@/api/interface/app';
import Status from '@/components/status/index.vue';

let data = ref<any>();
let loading = ref(false);
let timer: NodeJS.Timer | null = null;
const paginationConfig = reactive({
    currentPage: 1,
    pageSize: 20,
    total: 0,
});
let open = ref(false);
let operateReq = reactive({
    installId: 0,
    operate: '',
    detailId: 0,
});
let versions = ref<App.VersionDetail[]>();
const backupRef = ref();
const checkRef = ref();
const deleteRef = ref();
const appParamRef = ref();
let searchName = ref('');

const sync = () => {
    loading.value = true;
    SyncInstalledApp()
        .then(() => {
            ElMessage.success(i18n.global.t('app.syncSuccess'));
            search();
        })
        .finally(() => {
            loading.value = false;
        });
};

const search = () => {
    const req = {
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
        name: searchName.value,
    };

    SearchAppInstalled(req).then((res) => {
        data.value = res.data.items;
        paginationConfig.total = res.data.total;
    });
};

const openOperate = (row: any, op: string) => {
    operateReq.installId = row.id;
    operateReq.operate = op;
    if (op == 'update') {
        GetAppUpdateVersions(row.id).then((res) => {
            versions.value = res.data;
            if (res.data != null && res.data.length > 0) {
                operateReq.detailId = res.data[0].detailId;
            }
            open.value = true;
        });
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

const operate = async () => {
    open.value = false;
    loading.value = true;
    await InstalledOp(operateReq)
        .then(() => {
            ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
            search();
        })
        .finally(() => {
            loading.value = false;
        });
};

const handleClose = () => {
    open.value = false;
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
    },
    {
        label: i18n.global.t('app.update'),
        click: (row: any) => {
            openOperate(row, 'update');
        },
        disabled: (row: any) => {
            return !row.canUpdate;
        },
    },
    {
        label: i18n.global.t('app.restart'),
        click: (row: any) => {
            openOperate(row, 'restart');
        },
    },
    {
        label: i18n.global.t('app.up'),
        click: (row: any) => {
            openOperate(row, 'up');
        },
        disabled: (row: any) => {
            return row.status === 'Running';
        },
    },
    {
        label: i18n.global.t('app.down'),
        click: (row: any) => {
            openOperate(row, 'down');
        },
    },
    {
        label: i18n.global.t('app.delete'),
        click: (row: any) => {
            openOperate(row, 'delete');
        },
    },
    {
        label: i18n.global.t('app.detail'),
        click: (row: any) => {
            openParam(row.id);
        },
    },
];

const openBackups = (installId: number, installName: string) => {
    let params = {
        appInstallId: installId,
        appInstallName: installName,
    };
    backupRef.value.acceptParams(params);
};

const openParam = (installId: number) => {
    appParamRef.value.acceptParams({ id: installId });
};

onMounted(() => {
    search();
    timer = setInterval(() => {
        search();
    }, 1000 * 8);
});

onUnmounted(() => {
    clearInterval(Number(timer));
});
</script>

<style scoped lang="scss">
@import '../index.scss';
</style>
