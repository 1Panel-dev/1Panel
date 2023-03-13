<template>
    <div>
        <el-drawer v-model="upVisiable" :destroy-on-close="true" :close-on-click-modal="false" size="50%">
            <template #header>
                <DrawerHeader :header="$t('commons.button.import')" :resource="title" :back="handleClose" />
            </template>
            <div v-loading="loading">
                <el-upload ref="uploadRef" drag :on-change="fileOnChange" class="upload-demo" :auto-upload="false">
                    <el-icon class="el-icon--upload"><upload-filled /></el-icon>
                    <div class="el-upload__text">
                        {{ $t('database.dropHelper') }}
                        <em>{{ $t('database.clickHelper') }}</em>
                    </div>
                    <template #tip>
                        <div v-if="type === 'mysql'" class="el-upload__tip">
                            <span class="input-help">{{ $t('database.supportUpType') }}</span>
                            <span class="input-help">
                                {{ $t('database.zipFormat') }}
                            </span>
                        </div>
                        <div v-else class="el-upload__tip">
                            <span class="input-help">{{ $t('website.supportUpType') }}</span>
                            <span class="input-help">
                                {{ $t('website.zipFormat', [type + '.json']) }}
                            </span>
                        </div>
                    </template>
                </el-upload>
                <el-button v-if="uploaderFiles.length === 1" icon="Upload" @click="onSubmit">
                    {{ $t('commons.button.upload') }}
                </el-button>

                <el-divider />
                <ComplexTable
                    :pagination-config="paginationConfig"
                    @search="search"
                    v-model:selects="selects"
                    :data="data"
                >
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
                            {{ row.createdAt }}
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
        </el-drawer>
    </div>
</template>

<script lang="ts" setup>
import ComplexTable from '@/components/complex-table/index.vue';
import { reactive, ref } from 'vue';
import { computeSize } from '@/utils/util';
import { useDeleteData } from '@/hooks/use-delete-data';
import { handleRecoverByUpload } from '@/api/modules/setting';
import i18n from '@/lang';
import { UploadFile, UploadFiles, UploadInstance } from 'element-plus';
import { File } from '@/api/interface/file';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { BatchDeleteFile, CheckFile, GetUploadList, UploadFileData } from '@/api/modules/files';
import { loadBaseDir } from '@/api/modules/setting';
import { MsgError, MsgSuccess } from '@/utils/message';

const loading = ref();
const selects = ref<any>([]);
const baseDir = ref();

const data = ref();
const title = ref();
const paginationConfig = reactive({
    currentPage: 1,
    pageSize: 10,
    total: 0,
});

const upVisiable = ref(false);
const type = ref();
const name = ref();
const detailName = ref();
interface DialogProps {
    type: string;
    name: string;
    detailName: string;
}
const acceptParams = async (params: DialogProps): Promise<void> => {
    type.value = params.type;
    name.value = params.name;
    detailName.value = params.detailName;

    const pathRes = await loadBaseDir();
    if (type.value === 'mysql') {
        title.value = name.value + ' [ ' + detailName.value + ' ]';
        baseDir.value = `${pathRes.data}/uploads/database/mysql/${name.value}/${detailName.value}/`;
    }
    if (type.value === 'website' || type.value === 'app') {
        title.value = name.value;
        baseDir.value = `${pathRes.data}/uploads/${type.value}/${name.value}/`;
    }
    upVisiable.value = true;
    search();
};

const search = async () => {
    let params = {
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
        path: baseDir.value,
    };
    const res = await GetUploadList(params);
    data.value = res.data.items || [];
    paginationConfig.total = res.data.total;
};

const onRecover = async (row: File.File) => {
    let params = {
        type: type.value,
        name: name.value,
        detailName: detailName.value,
        file: baseDir.value + row.name,
    };
    loading.value = true;
    await handleRecoverByUpload(params)
        .then(() => {
            loading.value = false;
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        })
        .catch(() => {
            loading.value = false;
        });
};

const uploaderFiles = ref<UploadFiles>([]);
const uploadRef = ref<UploadInstance>();

const beforeAvatarUpload = (rawFile) => {
    if (type.value === 'app' || type.value === 'website') {
        if (!rawFile.name.endsWith('.tar.gz')) {
            MsgError(i18n.global.t('commons.msg.unSupportType'));
            return false;
        }
        if (rawFile.size / 1024 / 1024 > 50) {
            MsgError(i18n.global.t('commons.msg.unSupportSize', [50]));
            return false;
        }
        return true;
    }
    if (!rawFile.name.endsWith('.sql') && !rawFile.name.endsWith('.tar.gz') && !rawFile.name.endsWith('.sql.gz')) {
        MsgError(i18n.global.t('commons.msg.unSupportType'));
        return false;
    }
    if (rawFile.size / 1024 / 1024 > 10) {
        MsgError(i18n.global.t('commons.msg.unSupportSize', [10]));
        return false;
    }
    return true;
};

const fileOnChange = (_uploadFile: UploadFile, uploadFiles: UploadFiles) => {
    uploaderFiles.value = uploadFiles;
};

const handleClose = () => {
    uploaderFiles.value = [];
    uploadRef.value!.clearFiles();
    upVisiable.value = false;
};

const onSubmit = async () => {
    const formData = new FormData();
    if (uploaderFiles.value.length !== 1) {
        return;
    }
    if (!uploaderFiles.value[0]!.raw.name) {
        MsgError(i18n.global.t('commons.msg.fileNameErr'));
        return;
    }
    let reg = /^[a-zA-Z0-9\u4e00-\u9fa5]{1}[a-z:A-Z0-9_.\u4e00-\u9fa5-]{0,50}$/;
    if (!reg.test(uploaderFiles.value[0]!.raw.name)) {
        MsgError(i18n.global.t('commons.msg.fileNameErr'));
        return;
    }
    const res = await CheckFile(baseDir.value + uploaderFiles.value[0]!.raw.name);
    if (!res.data) {
        MsgError(i18n.global.t('commons.msg.fileExist'));
        return;
    }
    formData.append('file', uploaderFiles.value[0]!.raw);
    let isOk = beforeAvatarUpload(uploaderFiles.value[0]!.raw);
    if (!isOk) {
        return;
    }
    formData.append('path', baseDir.value);
    loading.value = true;
    UploadFileData(formData, {})
        .then(() => {
            loading.value = false;
            uploadRef.value?.clearFiles();
            uploaderFiles.value = [];
            MsgSuccess(i18n.global.t('file.uploadSuccess'));
            search();
        })
        .catch(() => {
            loading.value = false;
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
