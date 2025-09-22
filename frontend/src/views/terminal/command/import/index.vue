<template>
    <DialogPro v-model="visible" :title="$t('commons.button.import')" size="large">
        <div>
            <el-upload
                action="#"
                :auto-upload="false"
                ref="uploadRef"
                class="float-left mt-2"
                :show-file-list="false"
                :limit="1"
                accept=".csv"
                :on-change="fileOnChange"
                :on-exceed="handleExceed"
                v-model:file-list="uploaderFiles"
            >
                <el-button class="float-left" type="primary">{{ $t('commons.button.upload') }}</el-button>
            </el-upload>

            <el-button :disabled="selects.length === 0" @click="onImport" class="ml-2 mt-2">
                {{ $t('commons.button.import') }}
            </el-button>

            <el-select
                filterable
                :placeholder="$t('terminal.groupChange')"
                v-model="currentGroup"
                @change="changeGroup"
                class="p-w-200 ml-2 mt-2"
            >
                <div v-for="item in groupList" :key="item.id">
                    <el-option v-if="item.name === 'Default'" :label="$t('commons.table.default')" :value="item.id" />
                    <el-option v-else :label="item.name" :value="item.id" />
                </div>
            </el-select>

            <el-card class="mt-2 w-full" v-loading="loading">
                <el-table :data="data" @selection-change="handleSelectionChange">
                    <el-table-column type="selection" fix />
                    <el-table-column
                        :label="$t('commons.table.name')"
                        :min-width="80"
                        prop="name"
                        show-overflow-tooltip
                    />
                    <el-table-column
                        :label="$t('commons.table.group')"
                        show-overflow-tooltip
                        min-width="80"
                        prop="groupBelong"
                        fix
                    >
                        <template #default="{ row }">
                            <span v-if="row.groupBelong === 'Default'">{{ $t('commons.table.default') }}</span>
                            <span v-else>{{ row.groupBelong }}</span>
                        </template>
                    </el-table-column>
                    <el-table-column
                        :label="$t('terminal.command')"
                        :min-width="120"
                        prop="command"
                        show-overflow-tooltip
                    />
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
import { genFileId, UploadFile, UploadFiles, UploadProps, UploadRawFile } from 'element-plus';
import { importCommands, uploadCommands } from '@/api/modules/command';
import { getGroupList } from '@/api/modules/group';
import { MsgError, MsgSuccess } from '@/utils/message';
import i18n from '@/lang';

const emit = defineEmits<{ (e: 'search'): void }>();

const visible = ref(false);
const loading = ref();
const selects = ref<any>([]);
const data = ref([]);

const uploadRef = ref();
const uploaderFiles = ref();

const currentGroup = ref();
const groupList = ref();

const acceptParams = (): void => {
    visible.value = true;
    loadGroups();
    data.value = [];
};

const loadGroups = async () => {
    const res = await getGroupList('command');
    groupList.value = res.data || [];
};
const changeGroup = () => {
    let itemGroup;
    for (const g of groupList.value) {
        if (g.id === currentGroup.value) {
            itemGroup = g;
            break;
        }
    }
    for (const item of data.value) {
        item.groupID = currentGroup.value;
        item.groupBelong = itemGroup.name;
    }
};

const handleSelectionChange = (val: any) => {
    selects.value = val;
};

const fileOnChange = async (_uploadFile: UploadFile, uploadFiles: UploadFiles) => {
    uploaderFiles.value = uploadFiles;
    if (uploaderFiles.value.length !== 1) {
        return;
    }
    const file = uploaderFiles.value[0];
    const formData = new FormData();
    formData.append('file', file.raw);
    loading.value = true;
    await uploadCommands(formData)
        .then((res) => {
            loading.value = false;
            uploadRef.value!.clearFiles();
            uploaderFiles.value = [];
            data.value = res.data || [];
            if (data.value.length === 0) {
                MsgError(i18n.global.t('terminal.noSuchCommand'));
            }
        })
        .catch(() => {
            loading.value = false;
            uploadRef.value!.clearFiles();
            uploaderFiles.value = [];
        });
};

const handleExceed: UploadProps['onExceed'] = (files) => {
    uploadRef.value!.clearFiles();
    const file = files[0] as UploadRawFile;
    file.uid = genFileId();
    uploadRef.value!.handleStart(file);
};

const onImport = async () => {
    loading.value = true;
    importCommands(selects.value)
        .then(() => {
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            loading.value = false;
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
