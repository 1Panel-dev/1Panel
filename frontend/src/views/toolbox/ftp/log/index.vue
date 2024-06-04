<template>
    <el-drawer
        v-model="drawerVisible"
        :destroy-on-close="true"
        :close-on-click-modal="false"
        :close-on-press-escape="false"
        size="50%"
    >
        <template #header>
            <DrawerHeader header="FTP" :resource="paginationConfig.user" :back="handleClose" />
        </template>
        <el-select @change="search" class="p-w-200" clearable v-model="paginationConfig.operation">
            <template #prefix>{{ $t('commons.table.operate') }}</template>
            <el-option value="PUT" :label="$t('file.upload')" />
            <el-option value="GET" :label="$t('file.download')" />
        </el-select>

        <ComplexTable class="mt-2" :pagination-config="paginationConfig" :data="data" @search="search">
            <el-table-column label="ip" prop="ip" show-overflow-tooltip />
            <el-table-column :label="$t('commons.table.status')" min-width="50" show-overflow-tooltip prop="status">
                <template #default="{ row }">
                    <el-tag v-if="row.status === '200'">{{ $t('commons.status.success') }}</el-tag>
                    <el-tag v-else type="danger">{{ $t('commons.status.failed') }}</el-tag>
                </template>
            </el-table-column>
            <el-table-column :label="$t('commons.table.operate')" min-width="40" show-overflow-tooltip>
                <template #default="{ row }">
                    {{ loadOperation(row.operation) }}
                </template>
            </el-table-column>
            <el-table-column :label="$t('file.file')" show-overflow-tooltip>
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
    </el-drawer>
</template>

<script lang="ts" setup>
import { reactive, ref } from 'vue';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { searchFtpLog } from '@/api/modules/toolbox';
import { computeSizeFromByte } from '@/utils/util';
import i18n from '@/lang';

const paginationConfig = reactive({
    cacheSizeKey: 'ftp-log-page-size',
    currentPage: 1,
    pageSize: 10,
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
        return i18n.global.t('file.upload');
    }
    if (operation.startsWith('"GET')) {
        return i18n.global.t('file.download');
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
