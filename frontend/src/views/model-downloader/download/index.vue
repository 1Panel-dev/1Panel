<template>
    <DialogPro
        v-model="open"
        :title="
            $t(
                form.downloadType === 'tokenizer'
                    ? 'modelDownloadExtra.downloadTokenizer'
                    : 'modelDownloadExtra.downloadModel',
            )
        "
        size="small"
        :show-close="!loading"
        :close-on-press-escape="!loading"
    >
        <el-form label-position="top" @submit.prevent="submit">
            <el-form-item :label="$t('aiTools.model.modelName')">
                <div class="download-repo">{{ downloadRepoID }}</div>
            </el-form-item>
            <el-form-item :label="$t('aiTools.model.modelSource')">
                <SourceSelect v-model="form.source" :sources="sources" :disabled="loading" />
            </el-form-item>
        </el-form>
        <template #footer>
            <el-button :disabled="loading" @click="open = false">{{ $t('commons.button.cancel') }}</el-button>
            <el-button type="primary" :loading="loading" :disabled="loading" @click="submit">
                {{ $t('commons.button.confirm') }}
            </el-button>
        </template>
    </DialogPro>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue';
import type { ModelDownloader } from '@/api/interface/model-downloader';
import { startHFModelDownload, startModelScopeModelDownload } from '@/api/modules/model-downloader';
import { useGlobalStore } from '@/composables/useGlobalStore';
import { MsgSuccess } from '@/utils/message';
import i18n from '@/lang';
import SourceSelect from '../source-select.vue';

const emit = defineEmits<{ search: [] }>();
const { currentNode } = useGlobalStore();
const open = ref(false);
const loading = ref(false);
const sources = ref<ModelDownloader.Source[]>([]);
const sourceModelIds = ref<ModelDownloader.FeaturedModel['sourceModelIds']>();
const form = reactive({
    repoID: '',
    downloadType: 'all' as ModelDownloader.DownloadType,
    source: 'modelscope' as ModelDownloader.Source,
});
const downloadRepoID = computed(() => sourceModelIds.value?.[form.source]?.trim() || form.repoID);

const acceptParams = (
    repoID: string,
    downloadType: ModelDownloader.DownloadType,
    availableSources: ModelDownloader.Source[],
    modelIds?: ModelDownloader.FeaturedModel['sourceModelIds'],
) => {
    if (!repoID.trim() || !availableSources.length || loading.value) return;
    sources.value = availableSources;
    sourceModelIds.value = modelIds;
    Object.assign(form, {
        repoID: repoID.trim(),
        downloadType,
        source: availableSources[0],
    });
    open.value = true;
};

const submit = async () => {
    if (!open.value || loading.value) return;
    const node = currentNode.value;
    loading.value = true;
    try {
        const req = {
            repoID: downloadRepoID.value,
            downloadType: form.downloadType,
        };
        if (form.source === 'modelscope') await startModelScopeModelDownload(req);
        else await startHFModelDownload(req);
        if (node !== currentNode.value) return;
        open.value = false;
        MsgSuccess(i18n.global.t('aiTools.model.downloadQueued'));
        emit('search');
    } finally {
        loading.value = false;
    }
};

watch(currentNode, () => (open.value = false), { flush: 'sync' });

defineExpose({ acceptParams });
</script>

<style scoped lang="scss">
.download-repo {
    width: 100%;
    padding: 12px;
    border-radius: var(--el-border-radius-base);
    background: var(--el-fill-color-light);
    color: var(--el-text-color-primary);
    font-weight: 500;
    line-height: 1.6;
    overflow-wrap: anywhere;
}
</style>
