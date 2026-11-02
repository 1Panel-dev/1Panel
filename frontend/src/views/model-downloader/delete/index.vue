<template>
    <DialogPro
        v-model="open"
        :title="$t('commons.button.delete') + ' - ' + name"
        size="small"
        :show-close="!loading"
        :close-on-press-escape="!loading"
    >
        <el-form v-loading="loading" @submit.prevent="submit">
            <el-form-item>
                <el-checkbox
                    v-model="withFile"
                    :label="$t('modelDownloadExtra.deleteFiles', [contentLabel])"
                    :disabled="loading || fileOnly"
                />
                <span class="input-help break-all">
                    {{
                        $t('modelDownloadExtra.deleteFilesPath', [
                            contentLabel,
                            item && ('id' in item ? item.targetDir : item.path),
                        ])
                    }}
                </span>
            </el-form-item>
            <el-form-item>
                <div class="text-xs break-all">
                    <span>{{ $t('commons.msg.operateConfirm') }}</span>
                    <span class="text-red-500 font-medium">{{ name }}</span>
                </div>
                <el-input v-model="deleteInfo" :placeholder="name" :disabled="loading" />
            </el-form-item>
        </el-form>
        <template #footer>
            <el-button :disabled="loading" @click="open = false">{{ $t('commons.button.cancel') }}</el-button>
            <el-button type="primary" :loading="loading" :disabled="deleteInfo !== name || loading" @click="submit">
                {{ $t('commons.button.confirm') }}
            </el-button>
        </template>
    </DialogPro>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue';
import type { ModelDownloader } from '@/api/interface/model-downloader';
import { deleteLocalDownloadedModel, removeModelDownloadTask } from '@/api/modules/model-downloader';
import { MsgSuccess } from '@/utils/message';
import i18n from '@/lang';

const emit = defineEmits<{ search: [] }>();
const open = ref(false);
const loading = ref(false);
const withFile = ref(false);
const deleteInfo = ref('');
const item = ref<ModelDownloader.TaskItem | ModelDownloader.LocalModelItem>();
const name = computed(() => item.value?.repoID || (item.value && 'name' in item.value ? item.value.name : ''));
const contentLabel = computed(() =>
    i18n.global.t(item.value?.downloadType === 'tokenizer' ? 'aiTools.benchmark.tokenizer' : 'aiTools.model.model'),
);
const fileOnly = computed(() => item.value?.downloadType === 'all' && !('id' in item.value));

const acceptParams = (row: ModelDownloader.TaskItem | ModelDownloader.LocalModelItem) => {
    item.value = row;
    withFile.value = fileOnly.value;
    deleteInfo.value = '';
    open.value = true;
};

const submit = async () => {
    if (!item.value || deleteInfo.value !== name.value || loading.value) return;
    loading.value = true;
    try {
        if ('id' in item.value) {
            await removeModelDownloadTask({ id: item.value.id, withFile: withFile.value });
        } else {
            await deleteLocalDownloadedModel({
                name: item.value.name,
                downloadType: item.value.downloadType,
                withFile: withFile.value,
            });
        }
        open.value = false;
        MsgSuccess(i18n.global.t('commons.msg.deleteSuccess'));
        emit('search');
    } finally {
        loading.value = false;
    }
};

defineExpose({ acceptParams });
</script>
