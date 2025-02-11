<template>
    <div>
        <el-dialog
            v-model="dialogVisible"
            :title="$t('file.existFileTitle')"
            width="35%"
            :close-on-click-modal="false"
            :destroy-on-close="true"
        >
            <el-alert :show-icon="true" type="warning" :closable="false">
                <div class="whitespace-break-spaces">
                    <span>{{ $t('file.existFileHelper') }}</span>
                </div>
            </el-alert>
            <div>
                <el-table :data="existFiles" max-height="350">
                    <el-table-column type="index" :label="$t('commons.table.serialNumber')" width="55" />
                    <el-table-column prop="path" :label="$t('commons.table.name')" :min-width="200" />
                    <el-table-column :label="$t('file.existFileSize')" width="230">
                        <template #default="{ row }">
                            {{ getFileSize(row.uploadSize) }} -> {{ getFileSize(row.size) }}
                        </template>
                    </el-table-column>
                </el-table>
            </div>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="handleSkip">{{ $t('commons.button.skip') }}</el-button>
                    <el-button type="primary" @click="handleOverwrite()">
                        {{ $t('commons.button.cover') }}
                    </el-button>
                </span>
            </template>
        </el-dialog>
    </div>
</template>
<script lang="ts" setup>
import { ref } from 'vue';
import { computeSize } from '@/utils/util';

const dialogVisible = ref();
const existFiles = ref<DialogProps[]>([]);

interface DialogProps {
    name: string;
    path: string;
    size: number;
    uploadSize: number;
    modTime: string;
}
let onConfirmCallback = null;
const getFileSize = (size: number) => {
    return computeSize(size);
};

const handleSkip = () => {
    dialogVisible.value = false;
    if (onConfirmCallback) {
        onConfirmCallback(
            'skip',
            existFiles.value.map((file) => file.path),
        );
    }
};
const handleOverwrite = () => {
    dialogVisible.value = false;
    if (onConfirmCallback) {
        onConfirmCallback('overwrite');
    }
};
const acceptParams = async ({ paths, onConfirm }): Promise<void> => {
    existFiles.value = paths;
    onConfirmCallback = onConfirm;
    dialogVisible.value = true;
};

defineExpose({ acceptParams });
</script>
