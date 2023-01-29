<template>
    <div :v-loading="loading">
        <el-dialog v-model="upVisiable" :destroy-on-close="true" :close-on-click-modal="false" width="70%">
            <template #header>
                <div class="card-header">
                    <span>{{ $t('commons.button.import') }}</span>
                </div>
            </template>
            <div v-loading="loading">
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
                    <span class="input-help">{{ $t('website.supportUpType') }}</span>
                    <span class="input-help">
                        {{ $t('website.zipFormat') }}
                    </span>
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
                    <el-table-column :label="$t('commons.table.name')" show-overflow-tooltip prop="name" />
                    <el-table-column :label="$t('file.size')" prop="size">
                        <template #default="{ row }">
                            {{ computeSize(row.size) }}
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('commons.table.createdAt')" min-width="80" fix>
                        <template #default="{ row }">
                            {{ dateFromat(0, 0, row.modTime) }}
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
            </div>
        </el-dialog>
    </div>
</template>

<script lang="ts" setup>
import ComplexTable from '@/components/complex-table/index.vue';
import { reactive, ref } from 'vue';
import { computeSize, dateFromat } from '@/utils/util';
import { useDeleteData } from '@/hooks/use-delete-data';
import i18n from '@/lang';
import { ElMessage, UploadFile, UploadFiles, UploadInstance, UploadProps } from 'element-plus';
import { File } from '@/api/interface/file';
import { BatchDeleteFile, GetFilesList, UploadFileData } from '@/api/modules/files';
import { RecoverWebsiteByUpload } from '@/api/modules/website';
import { loadBaseDir } from '@/api/modules/setting';

const selects = ref<any>([]);
const baseDir = ref();
const loading = ref(false);

const data = ref();
const paginationConfig = reactive({
    currentPage: 1,
    pageSize: 10,
    total: 0,
});

const upVisiable = ref(false);
const websiteName = ref();
const websiteType = ref();

interface DialogProps {
    websiteName: string;
    websiteType: string;
}
const acceptParams = async (params: DialogProps): Promise<void> => {
    websiteName.value = params.websiteName;
    websiteType.value = params.websiteType;
    upVisiable.value = true;
    const pathRes = await loadBaseDir();
    baseDir.value = `${pathRes.data}/uploads/website/${websiteName.value}`;
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
        websiteName: websiteName.value,
        type: websiteType.value,
        fileDir: baseDir.value,
        fileName: row.name,
    };
    loading.value = true;
    await RecoverWebsiteByUpload(params)
        .then(() => {
            loading.value = false;
            ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
        })
        .finally(() => {
            loading.value = false;
        });
};

const uploaderFiles = ref<UploadFiles>([]);
const uploadRef = ref<UploadInstance>();

const beforeAvatarUpload: UploadProps['beforeUpload'] = (rawFile) => {
    if (rawFile.name.endsWith('.tar.gz')) {
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
    formData.append('path', baseDir.value + '/');
    loading.value = true;
    UploadFileData(formData, {})
        .then(() => {
            loading.value = false;
            ElMessage.success(i18n.global.t('file.uploadSuccess'));
            handleClose();
            search();
        })
        .finally(() => {
            loading.value = false;
        });
};

const onBatchDelete = async (row: File.File | null) => {
    let files: Array<string> = [];
    if (row) {
        files.push(baseDir.value + '/' + row.name);
    } else {
        selects.value.forEach((item: File.File) => {
            files.push(baseDir.value + '/' + item.name);
        });
    }
    await useDeleteData(BatchDeleteFile, { isDir: false, paths: files }, 'commons.msg.delete');
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
