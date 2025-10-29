<template>
    <DialogPro v-model="visible" :title="$t('commons.button.import')" size="large">
        <div>
            <el-alert :closable="false" show-icon type="info" :title="$t('cronjob.importHelper')" />
            <el-upload
                action="#"
                :auto-upload="false"
                ref="uploadRef"
                class="float-left mt-2"
                :show-file-list="false"
                :limit="1"
                accept=".json"
                :on-change="fileOnChange"
                :on-exceed="handleExceed"
                v-model:file-list="uploaderFiles"
            >
                <el-button class="float-left" type="primary">{{ $t('commons.button.upload') }}</el-button>
            </el-upload>

            <el-button :disabled="selects.length === 0" @click="onImport" class="ml-2 mt-2">
                {{ $t('commons.button.import') }}
            </el-button>

            <el-card class="mt-2 w-full" v-loading="loading">
                <el-table :data="data" @selection-change="handleSelectionChange">
                    <el-table-column type="selection" fix />
                    <el-table-column
                        :label="$t('cronjob.taskName')"
                        :min-width="120"
                        prop="name"
                        show-overflow-tooltip
                    />
                    <el-table-column
                        :label="$t('commons.table.type')"
                        :min-width="120"
                        prop="type"
                        show-overflow-tooltip
                    >
                        <template #default="{ row }">
                            <el-tag>{{ $t('cronjob.' + row.type) }}</el-tag>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('cronjob.cronSpec')" show-overflow-tooltip :min-width="120">
                        <template #default="{ row }">
                            <div v-for="(item, index) of row.spec.split(',')" :key="index">
                                <div v-if="row.expand || (!row.expand && index < 3)">
                                    <span>
                                        {{ row.specCustom ? item : transSpecToStr(item) }}
                                    </span>
                                </div>
                            </div>
                            <div v-if="!row.expand && row.spec.split(',').length > 3">
                                <el-button type="primary" link @click="row.expand = true">
                                    {{ $t('commons.button.expand') }}...
                                </el-button>
                            </div>
                            <div v-if="row.expand && row.spec.split(',').length > 3">
                                <el-button type="primary" link @click="row.expand = false">
                                    {{ $t('commons.button.collapse') }}
                                </el-button>
                            </div>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('cronjob.retainCopies')" :min-width="120" prop="retainCopies" />
                </el-table>
            </el-card>
        </div>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="visible = false">
                    {{ $t('commons.button.cancel') }}
                </el-button>
            </span>
        </template>
    </DialogPro>
</template>

<script lang="ts" setup>
import { ref } from 'vue';
import { transSpecToStr } from './../helper';
import { genFileId, UploadFile, UploadFiles, UploadProps, UploadRawFile } from 'element-plus';
import { MsgError, MsgSuccess } from '@/utils/message';
import i18n from '@/lang';
import { importCronjob } from '@/api/modules/cronjob';
import { Cronjob } from '@/api/interface/cronjob';

const emit = defineEmits<{ (e: 'search'): void }>();

const visible = ref(false);
const loading = ref();
const selects = ref<any>([]);
const data = ref();

const uploadRef = ref();
const uploaderFiles = ref();

const acceptParams = (): void => {
    visible.value = true;
    data.value = [];
};

const handleSelectionChange = (val: any) => {
    selects.value = val;
};

const fileOnChange = (_uploadFile: UploadFile, uploadFiles: UploadFiles) => {
    loading.value = true;
    data.value = [];
    uploaderFiles.value = uploadFiles;
    const reader = new FileReader();
    reader.onload = (e) => {
        try {
            const content = e.target.result as string;
            const parsed = JSON.parse(content) as Cronjob.CronjobTrans;
            if (!Array.isArray(parsed)) {
                MsgError(i18n.global.t('commons.msg.errImportFormat'));
                loading.value = false;
                return;
            }
            for (const item of parsed) {
                if (!checkDataFormat(item)) {
                    MsgError(i18n.global.t('commons.msg.errImportFormat'));
                    loading.value = false;
                    return;
                }
            }
            data.value = parsed;
            loading.value = false;
        } catch (error) {
            MsgError(i18n.global.t('commons.msg.errImport') + error.message);
            loading.value = false;
        }
    };
    reader.readAsText(_uploadFile.raw);
};

const handleExceed: UploadProps['onExceed'] = (files) => {
    uploadRef.value!.clearFiles();
    const file = files[0] as UploadRawFile;
    file.uid = genFileId();
    uploadRef.value!.handleStart(file);
};

const checkDataFormat = (item: any) => {
    if (!item.name) {
        return false;
    }
    const cronjobTypes = [
        'shell',
        'app',
        'website',
        'database',
        'directory',
        'log',
        'curl',
        'cutWebsiteLog',
        'clean',
        'snapshot',
        'ntp',
    ];
    if (!item.type || cronjobTypes.indexOf(item.type) === -1) {
        return false;
    }
    return true;
};

const onImport = async () => {
    await importCronjob(selects.value).then(() => {
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        visible.value = false;
        emit('search');
    });
};

defineExpose({
    acceptParams,
});
</script>
