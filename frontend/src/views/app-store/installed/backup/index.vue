<template>
    <el-drawer
        :close-on-click-modal="false"
        v-model="open"
        size="50%"
        :destroy-on-close="true"
        :before-close="handleClose"
    >
        <template #header>
            <Header :header="$t('app.backup')" :resource="installData.appInstallName" :back="handleClose"></Header>
        </template>
        <ComplexTable
            :pagination-config="paginationConfig"
            :data="data"
            @search="search"
            v-loading="loading"
            v-model:selects="selects"
        >
            <template #toolbar>
                <el-button type="primary" @click="backup">{{ $t('app.backup') }}</el-button>
                <el-button type="danger" plain :disabled="selects.length === 0" @click="onBatchDelete()">
                    {{ $t('commons.button.delete') }}
                </el-button>
            </template>
            <el-table-column type="selection" fix />

            <el-table-column
                :label="$t('app.backupName')"
                min-width="120px"
                prop="name"
                show-overflow-tooltip
            ></el-table-column>
            <el-table-column
                :label="$t('app.backupPath')"
                min-width="120px"
                prop="path"
                show-overflow-tooltip
            ></el-table-column>
            <el-table-column
                prop="createdAt"
                :label="$t('app.backupdate')"
                :formatter="dateFormat"
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
        </ComplexTable>
        <el-dialog
            v-model="openRestorePage"
            :destroy-on-close="true"
            :close-on-click-modal="false"
            :title="$t('commons.msg.operate')"
            width="30%"
        >
            <el-alert :title="$t('app.restoreWarn')" type="warning" :closable="false" show-icon />
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="openRestorePage = false" :loading="loading">
                        {{ $t('commons.button.cancel') }}
                    </el-button>
                    <el-button type="primary" @click="restore" :loading="loading">
                        {{ $t('commons.button.confirm') }}
                    </el-button>
                </span>
            </template>
        </el-dialog>
    </el-drawer>
</template>

<script lang="ts" setup name="installBackup">
import { DelAppBackups, GetAppBackups, InstalledOp } from '@/api/modules/app';
import { reactive, ref } from 'vue';
import ComplexTable from '@/components/complex-table/index.vue';
import Header from '@/components/drawer-header/index.vue';
import { dateFormat } from '@/utils/util';
import { ElMessage } from 'element-plus';
import i18n from '@/lang';
import { useDeleteData } from '@/hooks/use-delete-data';

interface InstallRrops {
    appInstallId: number;
    appInstallName: string;
}
const installData = ref<InstallRrops>({
    appInstallId: 0,
    appInstallName: '',
});
const selects = ref<any>([]);
let open = ref(false);
let loading = ref(false);
let data = ref<any>();
let openRestorePage = ref(false);
const paginationConfig = reactive({
    currentPage: 1,
    pageSize: 20,
    total: 0,
});
let req = reactive({
    installId: installData.value.appInstallId,
    operate: 'restore',
    backupId: -1,
});

const em = defineEmits(['close']);
const handleClose = () => {
    open.value = false;
    em('close', open);
};

const acceptParams = (props: InstallRrops) => {
    installData.value.appInstallId = props.appInstallId;
    installData.value.appInstallName = props.appInstallName;
    search();
    open.value = true;
};

const search = async () => {
    const req = {
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
        appInstallId: installData.value.appInstallId,
    };
    await GetAppBackups(req).then((res) => {
        data.value = res.data.items;
        paginationConfig.total = res.data.total;
    });
};

const backup = async () => {
    const req = {
        installId: installData.value.appInstallId,
        operate: 'backup',
    };
    loading.value = true;
    await InstalledOp(req)
        .then(() => {
            ElMessage.success(i18n.global.t('commons.msg.backupSuccess'));
            search();
        })
        .finally(() => {
            loading.value = false;
        });
};

const openRestore = (backupId: number) => {
    openRestorePage.value = true;
    req.backupId = backupId;
    req.operate = 'restore';
    req.installId = installData.value.appInstallId;
};

const restore = async () => {
    loading.value = true;
    await InstalledOp(req)
        .then(() => {
            ElMessage.success(i18n.global.t('commons.msg.restoreSuccess'));
            openRestorePage.value = false;
            search();
        })
        .finally(() => {
            loading.value = false;
        });
};

const deleteBackup = async (ids: number[]) => {
    await useDeleteData(DelAppBackups, { ids: ids }, 'commons.msg.delete');
    search();
};

const onBatchDelete = () => {
    let ids: Array<number> = [];
    selects.value.forEach((item: any) => {
        ids.push(item.id);
    });
    deleteBackup(ids);
};

const buttons = [
    {
        label: i18n.global.t('app.delete'),
        click: (row: any) => {
            deleteBackup([row.id]);
        },
    },
    {
        label: i18n.global.t('app.restore'),
        click: (row: any) => {
            openRestore(row.id);
        },
    },
];

defineExpose({
    acceptParams,
});
</script>
