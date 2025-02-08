<template>
    <div>
        <LayoutContent :title="$t('setting.backupAccount')">
            <template #leftToolBar>
                <el-button type="primary" @click="onOpenDialog('create')">
                    {{ $t('commons.button.add') }}
                </el-button>
            </template>
            <template #rightToolBar>
                <TableSearch @search="search()" v-model:searchName="paginationConfig.name" />
                <TableRefresh @search="search()" />
                <TableSetting title="backup-account-refresh" @search="search()" />
            </template>
            <template #main>
                <el-alert type="info" :closable="false" class="common-div">
                    <template #title>
                        <span>
                            {{ $t('setting.backupAlert') }}
                            <el-link
                                class="ml-1 text-xs"
                                type="primary"
                                target="_blank"
                                :href="globalStore.docsUrl + '/user_manual/settings/#3'"
                            >
                                {{ $t('commons.button.helpDoc') }}
                            </el-link>
                        </span>
                    </template>
                </el-alert>
                <ComplexTable :pagination-config="paginationConfig" @sort-change="search" @search="search" :data="data">
                    <el-table-column
                        :label="$t('commons.table.name')"
                        :min-width="80"
                        prop="name"
                        show-overflow-tooltip
                    />
                    <el-table-column
                        v-if="globalStore.isProductPro"
                        :label="$t('setting.scope')"
                        :min-width="80"
                        prop="isPublic"
                    >
                        <template #default="{ row }">
                            <el-button plain size="small">
                                {{ row.isPublic ? $t('setting.public') : $t('setting.private') }}
                            </el-button>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('commons.table.type')" :min-width="80" prop="type">
                        <template #default="{ row }">
                            <el-tag>{{ $t('setting.' + row.type) }}</el-tag>
                            <el-tooltip v-if="row.type === 'OneDrive'">
                                <template #content>
                                    {{ $t('setting.clickToRefresh') }}
                                    <br />
                                    <span v-if="row.varsJson['refresh_status'] === 'Success'">
                                        {{ $t('setting.refreshStatus') + ':' + $t('commons.status.success') }}
                                    </span>
                                    <div v-else>
                                        <span>
                                            {{ $t('setting.refreshStatus') + ':' + $t('commons.status.failed') }}
                                        </span>
                                        <br />
                                        <span>
                                            {{ $t('commons.table.message') + ':' + row.varsJson['refresh_msg'] }}
                                        </span>
                                    </div>
                                    <br />
                                    {{ $t('setting.refreshTime') + ':' + row.varsJson['refresh_time'] }}
                                </template>
                                <el-tag @click="refreshItemToken(row)" class="ml-1">
                                    {{ 'Token ' + $t('commons.button.refresh') }}
                                </el-tag>
                            </el-tooltip>
                        </template>
                    </el-table-column>
                    <el-table-column prop="bucket" label="Bucket" show-overflow-tooltip>
                        <template #default="{ row }">
                            {{ row.bucket || '-' }}
                        </template>
                    </el-table-column>
                    <el-table-column prop="endpoint" label="Endpoint" show-overflow-tooltip>
                        <template #default="{ row }">
                            {{ loadEndpoint(row) }}
                        </template>
                    </el-table-column>
                    <el-table-column prop="backupPath" :label="$t('setting.backupDir')" show-overflow-tooltip />
                    <el-table-column
                        prop="createdAt"
                        :label="$t('commons.table.date')"
                        :formatter="dateFormat"
                        show-overflow-tooltip
                    />
                    <fu-table-operations
                        width="300px"
                        :buttons="buttons"
                        :ellipsis="10"
                        :label="$t('commons.table.operate')"
                        fix
                    />
                </ComplexTable>
            </template>
        </LayoutContent>

        <Operate ref="dialogRef" @search="search" />
        <OpDialog ref="opRef" @search="search" />
    </div>
</template>
<script setup lang="ts">
import { dateFormat } from '@/utils/util';
import { onMounted, ref } from 'vue';
import { searchBackup, deleteBackup, refreshToken } from '@/api/modules/backup';
import Operate from '@/views/setting/backup-account/operate/index.vue';
import { Backup } from '@/api/interface/backup';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { GlobalStore } from '@/store';
const globalStore = GlobalStore();

const loading = ref();
const data = ref();
const paginationConfig = reactive({
    cacheSizeKey: 'backup-page-size',
    currentPage: 1,
    pageSize: 10,
    total: 0,
    type: '',
    name: '',
});
const opRef = ref();
const dialogRef = ref();

const search = async () => {
    let params = {
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
        type: paginationConfig.type,
        name: paginationConfig.name,
    };
    loading.value = true;
    await searchBackup(params)
        .then((res) => {
            loading.value = false;
            data.value = res.data.items || [];
            for (const bac of data.value) {
                if (bac.id !== 0) {
                    bac.varsJson = JSON.parse(bac.vars);
                }
            }
            paginationConfig.total = res.data.total;
        })
        .catch(() => {
            loading.value = false;
        });
};

const loadEndpoint = (row: any) => {
    if (row.type === 'COS' || row.type === 'MINIO' || row.type === 'OSS' || row.type === 'S3') {
        return row.varsJson['endpoint'];
    }
    if (row.type === 'KODO') {
        return row.varsJson['domain'];
    }
    return '';
};

const onDelete = async (row: Backup.BackupInfo) => {
    opRef.value.acceptParams({
        title: i18n.global.t('commons.button.delete'),
        names: ['[ ' + row.type + ' ] ' + row.name],
        msg: i18n.global.t('commons.msg.operatorHelper', [
            i18n.global.t('setting.backupAccount'),
            i18n.global.t('commons.button.delete'),
        ]),
        api: deleteBackup,
        params: { id: row.id, name: row.name, isPublic: row.isPublic },
    });
};

const onOpenDialog = async (
    title: string,
    rowData: Partial<Backup.BackupInfo> = {
        id: 0,
        isPublic: false,
        varsJson: {},
    },
) => {
    let params = {
        title,
        rowData: { ...rowData },
    };
    dialogRef.value!.acceptParams(params);
};

const refreshItemToken = async (row: any) => {
    await refreshToken({ id: row.id, name: row.name, isPublic: row.isPublic });
    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
    search();
};

const buttons = [
    {
        label: i18n.global.t('commons.button.edit'),
        click: (row: Backup.BackupInfo) => {
            onOpenDialog('edit', row);
        },
    },
    {
        label: i18n.global.t('commons.button.delete'),
        click: (row: Backup.BackupInfo) => {
            onDelete(row);
        },
    },
];

onMounted(() => {
    search();
});
</script>
