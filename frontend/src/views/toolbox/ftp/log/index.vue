<template>
    <DrawerPro v-model="drawerVisible" header="FTP" :resource="paginationConfig.user" @close="handleClose" size="large">
        <el-select @change="search" class="p-w-200" clearable v-model="paginationConfig.operation">
            <template #prefix>{{ $t('commons.table.operate') }}</template>
            <el-option value="PUT" :label="$t('commons.button.upload')" />
            <el-option value="GET" :label="$t('commons.button.download')" />
        </el-select>
        <ComplexTable class="mt-2" :pagination-config="paginationConfig" :data="data" @search="search">
            <el-table-column label="ip" prop="ip" show-overflow-tooltip />
            <el-table-column :label="$t('commons.table.status')" min-width="50" show-overflow-tooltip prop="status">
                <template #default="{ row }">
                    <Status :status="row.status === '200' ? 'success' : 'failed'" />
                </template>
            </el-table-column>
            <el-table-column :label="$t('commons.table.operate')" min-width="40" show-overflow-tooltip>
                <template #default="{ row }">
                    {{ loadOperation(row.operation) }}
                </template>
            </el-table-column>
            <el-table-column :label="$t('menu.files')" show-overflow-tooltip>
                <template #default="{ row }">
                    {{ loadFileName(row.operation) }}
                </template>
            </el-table-column>
            <el-table-column :label="$t('file.size')" show-overflow-tooltip prop="size" min-width="60">
                <template #default="{ row }">
                    {{ computeSizeFromByte(Number(row.size)) }}
                </template>
            </el-table-column>
            <el-table-column :label="$t('commons.table.date')" prop="time" show-overflow-tooltip min-width="100" />
        </ComplexTable>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="drawerVisible = false">{{ $t('commons.button.cancel') }}</el-button>
            </span>
        </template>
    </DrawerPro>
</template>

<script lang="ts" setup>
import { reactive, ref } from 'vue';
import { searchFtpLog } from '@/api/modules/toolbox';
import { computeSizeFromByte } from '@/utils/util';
import i18n from '@/lang';

const paginationConfig = reactive({
    cacheSizeKey: 'ftp-log-page-size',
    currentPage: 1,
    pageSize: Number(localStorage.getItem('ftp-log-page-size')) || 20,
    total: 0,
    user: '',
    operation: '',
});
const data = ref();

const itemPath = ref();
interface DialogProps {
    user: string;
    path: string;
}
const loading = ref();
const drawerVisible = ref(false);

const acceptParams = (params: DialogProps): void => {
    paginationConfig.user = params.user;
    paginationConfig.operation = '';
    itemPath.value = params.path;
    search();
    drawerVisible.value = true;
};

const handleClose = () => {
    drawerVisible.value = false;
};

const search = async () => {
    let params = {
        user: paginationConfig.user,
        operation: paginationConfig.operation,
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
    };
    loading.value = true;
    await searchFtpLog(params)
        .then((res) => {
            loading.value = false;
            data.value = res.data.items || [];
            paginationConfig.total = res.data.total;
        })
        .catch(() => {
            loading.value = false;
        });
};

const loadOperation = (operation: string) => {
    if (operation.startsWith('"PUT')) {
        return i18n.global.t('commons.button.upload');
    }
    if (operation.startsWith('"GET')) {
        return i18n.global.t('commons.button.download');
    }
};
const loadFileName = (operation: string) => {
    return operation
        .replaceAll('"', '')
        .replaceAll('PUT', '')
        .replaceAll('GET', '')
        .replaceAll(itemPath.value + '/', '');
};

defineExpose({
    acceptParams,
});
</script>
