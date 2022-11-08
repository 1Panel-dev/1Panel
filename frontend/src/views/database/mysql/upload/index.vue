<template>
    <div>
        <el-dialog v-model="upVisiable" :destroy-on-close="true" :close-on-click-modal="false" width="70%">
            <template #header>
                <div class="card-header">
                    <span>{{ $t('database.backup') }}</span>
                </div>
            </template>
            <el-upload
                ref="uploadRef"
                :on-change="fileOnChange"
                :before-upload="beforeAvatarUpload"
                class="upload-demo"
                :auto-upload="false"
            >
                <template #trigger>
                    <el-button>选择文件</el-button>
                </template>
                <el-button style="margin-left: 10px" @click="onSubmit">上传</el-button>
            </el-upload>
            <div style="margin-left: 10px">
                <span class="input-help">仅支持sql、zip、sql.gz、(tar.gz|gz|tgz)</span>
                <span class="input-help">zip、tar.gz压缩包结构：test.zip或test.tar.gz压缩包内，必需包含test.sql</span>
            </div>
            <el-divider />
            <ComplexTable :pagination-config="paginationConfig" v-model:selects="selects" :data="data">
                <template #toolbar>
                    <el-button
                        style="margin-left: 10px"
                        type="danger"
                        plain
                        :disabled="selects.length === 0"
                        @click="onBatchDelete(null)"
                    >
                        {{ $t('commons.button.delete') }}
                    </el-button>
                </template>
                <el-table-column type="selection" fix />
                <el-table-column :label="$t('commons.table.name')" show-overflow-tooltip prop="fileName" />
                <el-table-column :label="$t('file.dir')" show-overflow-tooltip prop="fileDir" />
                <el-table-column :label="$t('file.size')" prop="size">
                    <template #default="{ row }">
                        {{ computeSize(row.size) }}
                    </template>
                </el-table-column>
                <el-table-column :label="$t('commons.table.createdAt')" prop="createdAt" />
                <fu-table-operations
                    width="300px"
                    :buttons="buttons"
                    :ellipsis="10"
                    :label="$t('commons.table.operate')"
                    fix
                />
            </ComplexTable>
        </el-dialog>
    </div>
</template>

<script lang="ts" setup>
import ComplexTable from '@/components/complex-table/index.vue';
import { reactive, ref } from 'vue';
import { computeSize } from '@/utils/util';
import { useDeleteData } from '@/hooks/use-delete-data';
import { recover, searchUpList, uploadFile } from '@/api/modules/database';
import i18n from '@/lang';
import { ElMessage, UploadFile, UploadFiles, UploadInstance, UploadProps } from 'element-plus';
import { deleteBackupRecord } from '@/api/modules/backup';
import { Backup } from '@/api/interface/backup';

const selects = ref<any>([]);

const data = ref();
const paginationConfig = reactive({
    currentPage: 1,
    pageSize: 10,
    total: 0,
});

const upVisiable = ref(false);
const mysqlName = ref();
const dbName = ref();
interface DialogProps {
    mysqlName: string;
    dbName: string;
}
const acceptParams = (params: DialogProps): void => {
    mysqlName.value = params.mysqlName;
    dbName.value = params.dbName;
    upVisiable.value = true;
    search();
};

const search = async () => {
    let params = {
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
        mysqlName: mysqlName.value,
    };
    const res = await searchUpList(params);
    data.value = res.data.items || [];
    paginationConfig.total = res.data.total;
};

const onRecover = async (row: Backup.RecordInfo) => {
    let params = {
        mysqlName: mysqlName.value,
        dbName: dbName.value,
        backupName: row.fileDir + row.fileName,
    };
    await recover(params);
    ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
};

const uploaderFiles = ref<UploadFiles>([]);
const uploadRef = ref<UploadInstance>();

const beforeAvatarUpload: UploadProps['beforeUpload'] = (rawFile) => {
    if (rawFile.name.endsWith('.sql') || rawFile.name.endsWith('gz') || rawFile.name.endsWith('.zip')) {
        ElMessage.error(i18n.global.t('database.unSupportType'));
        return false;
    }
    return true;
};

const fileOnChange = (_uploadFile: UploadFile, uploadFiles: UploadFiles) => {
    uploaderFiles.value = uploadFiles;
};

const handleClose = () => {
    uploadRef.value!.clearFiles();
};

const onSubmit = () => {
    const formData = new FormData();
    if (uploaderFiles.value.length !== 1) {
        return;
    }
    if (uploaderFiles.value[0]!.raw != undefined) {
        formData.append('file', uploaderFiles.value[0]!.raw);
    }
    uploadFile(mysqlName.value, formData).then(() => {
        ElMessage.success(i18n.global.t('file.uploadSuccess'));
        handleClose();
        search();
    });
};

const onBatchDelete = async (row: Backup.RecordInfo | null) => {
    let ids: Array<number> = [];
    if (row) {
        ids.push(row.id);
    } else {
        selects.value.forEach((item: Backup.RecordInfo) => {
            ids.push(item.id);
        });
    }
    await useDeleteData(deleteBackupRecord, { ids: ids }, 'commons.msg.delete', true);
    search();
};

const buttons = [
    {
        label: i18n.global.t('commons.button.recover'),
        click: (row: Backup.RecordInfo) => {
            onRecover(row);
        },
    },
    {
        label: i18n.global.t('commons.button.delete'),
        click: (row: Backup.RecordInfo) => {
            onBatchDelete(row);
        },
    },
];

defineExpose({
    acceptParams,
});
</script>
