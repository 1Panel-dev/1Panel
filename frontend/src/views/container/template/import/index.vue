<template>
    <DialogPro v-model="visible" :title="$t('commons.button.import')" size="large">
        <div>
            <el-alert :closable="false" show-icon type="info">
                <template #default>
                    <div>{{ $t('commons.msg.importHelper') }}</div>
                </template>
            </el-alert>
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

            <el-card class="mt-2 w-full" v-loading="loading">
                <el-table :data="displayData" @selection-change="handleSelectionChange">
                    <el-table-column type="selection" fix />
                    <el-table-column :label="$t('commons.table.status')" :min-width="80">
                        <template #default="{ row }">
                            <Status :status="row.status" />
                        </template>
                    </el-table-column>
                    <el-table-column
                        :label="$t('commons.table.name')"
                        :min-width="70"
                        show-overflow-tooltip
                        prop="name"
                    />
                    <el-table-column
                        :label="$t('commons.table.description')"
                        show-overflow-tooltip
                        :min-width="120"
                        prop="description"
                    />
                    <el-table-column :label="$t('container.content')" :min-width="70" prop="content">
                        <template #default="{ row }">
                            <el-button type="primary" link @click="onOpenDetail(row)">
                                {{ $t('commons.button.view') }}
                            </el-button>
                        </template>
                    </el-table-column>
                </el-table>
            </el-card>
        </div>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="visible = false">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button type="primary" :disabled="selects.length === 0" @click="onImport">
                    {{ $t('commons.button.import') }}
                </el-button>
            </span>
        </template>
    </DialogPro>
    <DetailDialog ref="detailRef" />
</template>

<script lang="ts" setup>
import { ref } from 'vue';
import { genFileId, UploadFile, UploadFiles, UploadProps, UploadRawFile } from 'element-plus';
import { MsgError, MsgSuccess } from '@/utils/message';
import DetailDialog from '@/views/container/template/detail/index.vue';
import i18n from '@/lang';
import { Container } from '@/api/interface/container';
import { batchComposeTemplate, searchComposeTemplate } from '@/api/modules/container';

const emit = defineEmits<{ (e: 'search'): void }>();

const visible = ref(false);
const loading = ref(false);
const selects = ref<any>([]);
const displayData = ref<any>([]);
const currentData = ref<Container.TemplateInfo[]>([]);

const uploadRef = ref();
const uploaderFiles = ref();
const detailRef = ref();

const acceptParams = async (): Promise<void> => {
    visible.value = true;
    displayData.value = [];
    selects.value = [];
    loadTemplates();
};

const loadTemplates = async () => {
    const res = await searchComposeTemplate({
        info: '',
        page: 1,
        pageSize: 10000,
    });
    currentData.value = res.data.items || [];
};

const handleSelectionChange = (val: any) => {
    selects.value = val;
};

const onOpenDetail = async (row: Container.TemplateInfo) => {
    detailRef.value.acceptParams({ content: row.content });
};

const fileOnChange = (_uploadFile: UploadFile, uploadFiles: UploadFiles) => {
    loading.value = true;
    displayData.value = [];
    uploaderFiles.value = uploadFiles;

    const reader = new FileReader();
    reader.onload = (e) => {
        try {
            const content = e.target.result as string;
            const parsed = JSON.parse(content);

            if (!Array.isArray(parsed)) {
                MsgError(i18n.global.t('commons.msg.errJsonImportFormat'));
                loading.value = false;
                return;
            }

            for (const item of parsed) {
                if (!checkDataFormat(item)) {
                    MsgError(i18n.global.t('commons.msg.errJsonImportFormat'));
                    loading.value = false;
                    return;
                }
            }

            compareData(parsed);
            loading.value = false;
        } catch (error) {
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

const checkDataFormat = (item: any): boolean => {
    return item.name && item.content;
};

const compareData = (importLists: any[]) => {
    const news: any[] = [];
    const conflicts: any[] = [];
    const duplicates: any[] = [];

    for (const importItem of importLists) {
        const key = `${importItem.name}:${importItem.content}`;

        const existing = currentData.value.find((item) => {
            const existingKey = `${item.name}:${item.content}`;
            return existingKey === key;
        });
        const conflict = currentData.value.find((item) => {
            if (existing) {
                return false;
            }
            return item.name === importItem.name && item.content !== importItem.content;
        });

        if (existing) {
            duplicates.push({ ...importItem, status: 'duplicate' });
            continue;
        }
        if (conflict) {
            conflicts.push({ ...importItem, status: 'conflict' });
            continue;
        }
        news.push({ ...importItem, status: 'new' });
    }

    displayData.value = [...news, ...conflicts, ...duplicates];
};

const onImport = async () => {
    loading.value = true;
    await batchComposeTemplate(selects.value)
        .then(() => {
            loading.value = false;
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            emit('search');
            visible.value = false;
        })
        .catch(() => {
            loading.value = false;
        });
};

defineExpose({
    acceptParams,
});
</script>
