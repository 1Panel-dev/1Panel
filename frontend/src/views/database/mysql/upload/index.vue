<template>
    <div>
        <el-drawer v-model="upVisiable" :destroy-on-close="true" :close-on-click-modal="false" size="50%">
            <template #header>
                <DrawerHeader :header="$t('commons.button.import')" :back="handleClose" />
            </template>
            <el-upload
                ref="uploadRef"
                :on-change="fileOnChange"
                :before-upload="beforeAvatarUpload"
                class="upload-demo"
                :auto-upload="false"
            >
                <template #trigger>
                    <el-button type="primary" plain>{{ $t('database.selectFile') }}</el-button>
                </template>
                <el-button style="margin-left: 10px" icon="Upload" @click="onSubmit">
                    {{ $t('commons.button.upload') }}
                </el-button>
            </el-upload>
            <div style="margin-left: 10px">
                <span class="input-help">{{ $t('database.supportUpType') }}</span>
                <span class="input-help">
                    {{ $t('database.zipFormat') }}
                </span>
            </div>
            <el-divider />
            <ComplexTable :pagination-config="paginationConfig" v-model:selects="selects" :data="data">
                <template #toolbar>
                    <el-button
                        style="margin-left: 10px"
                        plain
                        :disabled="selects.length === 0"
                        @click="onBatchDelete(null)"
                    >
                        {{ $t('commons.button.delete') }}
                    </el-button>
                </template>
                <el-table-column type="selection" fix />
                <el-table-column :label="$t('commons.table.name')" show-overflow-tooltip prop="name" />
                <el-table-column :label="$t('file.size')" prop="size">
                    <template #default="{ row }">
                        {{ computeSize(row.size) }}
                    </template>
                </el-table-column>
                <el-table-column :label="$t('commons.table.createdAt')" min-width="80" fix>
                    <template #default="{ row }">
                        {{ dateFormatSimple(row.modTime) }}
                    </template>
                </el-table-column>
                <fu-table-operations
                    width="300px"
                    :buttons="buttons"
                    :ellipsis="10"
                    :label="$t('commons.table.operate')"
                    fix
                />
            </ComplexTable>
        </el-drawer>
    </div>
</template>

<script lang="ts" setup>
import ComplexTable from '@/components/complex-table/index.vue';
import { reactive, ref } from 'vue';
import { computeSize, dateFormatSimple } from '@/utils/util';
import { useDeleteData } from '@/hooks/use-delete-data';
import { recoverByUpload } from '@/api/modules/database';
import i18n from '@/lang';
import { ElMessage, UploadFile, UploadFiles, UploadInstance, UploadProps } from 'element-plus';
import { File } from '@/api/interface/file';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { BatchDeleteFile, GetFilesList, UploadFileData } from '@/api/modules/files';
import { loadBaseDir } from '@/api/modules/setting';

const selects = ref<any>([]);
const baseDir = ref();

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
const acceptParams = async (params: DialogProps): Promise<void> => {
    mysqlName.value = params.mysqlName;
    dbName.value = params.dbName;

    const pathRes = await loadBaseDir();
    baseDir.value = `${pathRes.data}/uploads/database/mysql/${mysqlName.value}/${dbName.value}/`;
    upVisiable.value = true;
    search();
};

const search = async () => {
    let params = {
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
        path: baseDir.value,
        expand: true,
    };
    const res = await GetFilesList(params);
    data.value = res.data.items || [];
    paginationConfig.total = res.data.itemTotal;
};

const onRecover = async (row: File.File) => {
    let params = {
        mysqlName: mysqlName.value,
        dbName: dbName.value,
        fileDir: baseDir.value,
        fileName: row.name,
    };
    await recoverByUpload(params);
    ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
};

const uploaderFiles = ref<UploadFiles>([]);
const uploadRef = ref<UploadInstance>();

const beforeAvatarUpload: UploadProps['beforeUpload'] = (rawFile) => {
    if (
        rawFile.name.endsWith('.sql') ||
        rawFile.name.endsWith('.gz') ||
        rawFile.name.endsWith('.zip') ||
        rawFile.name.endsWith('.tgz')
    ) {
        ElMessage.error(i18n.global.t('database.unSupportType'));
        return false;
    } else if (rawFile.size / 1024 / 1024 > 10) {
        ElMessage.error(i18n.global.t('database.unSupportSize'));
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
    formData.append('path', baseDir.value);
    UploadFileData(formData, {}).then(() => {
        ElMessage.success(i18n.global.t('file.uploadSuccess'));
        handleClose();
        search();
    });
};

const onBatchDelete = async (row: File.File | null) => {
    let files: Array<string> = [];
    if (row) {
        files.push(baseDir.value + row.name);
    } else {
        selects.value.forEach((item: File.File) => {
            files.push(baseDir.value + item.name);
        });
    }
    await useDeleteData(BatchDeleteFile, { paths: files, isDir: false }, 'commons.msg.delete');
    search();
};

const buttons = [
    {
        label: i18n.global.t('commons.button.recover'),
        click: (row: File.File) => {
            onRecover(row);
        },
    },
    {
        label: i18n.global.t('commons.button.delete'),
        click: (row: File.File) => {
            onBatchDelete(row);
        },
    },
];

defineExpose({
    acceptParams,
});
</script>
